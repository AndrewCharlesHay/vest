package main

import (
	"context"
	"database/sql"
	"log"

	"net/http"
	"os"
	"time"

	"github.com/AndrewCharlesHay/vest/internal/api"
	"github.com/AndrewCharlesHay/vest/internal/ingest"
	"github.com/AndrewCharlesHay/vest/internal/middleware"
	_ "github.com/jackc/pgx/v5/stdlib" // PG driver
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

func main() {
	// 1. DB Connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		host := os.Getenv("DB_HOST")
		user := os.Getenv("DB_USER")
		pass := os.Getenv("DB_PASSWORD")
		name := os.Getenv("DB_NAME")
		if host != "" && user != "" && pass != "" && name != "" {
			// postgres://user:password@host:port/dbname
			dbURL = "postgres://" + user + ":" + pass + "@" + host + ":5432/" + name
		} else {
			log.Fatal("DATABASE_URL or (DB_HOST, DB_USER, DB_PASSWORD, DB_NAME) is required")
		}
	}
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	// Wait for DB to be ready
	for i := 0; i < 10; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		log.Println("Waiting for DB...")
		time.Sleep(2 * time.Second)
	}

	// Auto-Migrate (Create Table)
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS positions (
			date DATE NOT NULL,
			account_id TEXT NOT NULL,
			ticker TEXT NOT NULL,
			quantity DECIMAL(15, 2),
			market_value DECIMAL(15, 2),
			shares DECIMAL(15, 2),
			source_system TEXT,
			ingested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (date, account_id, ticker)
		);
	`)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
	log.Println("Database schema initialized.")

	// 2. SFTP Connection for Ingestion
	// Only start if config present (optional for running just API test?)
	sftpHost := os.Getenv("SFTP_HOST")
	if sftpHost != "" {
		go func() {
			log.Println("Starting SFTP Ingestor...")
			for {
				err := runIngestor(db, sftpHost)
				if err != nil {
					log.Printf("Ingestor failed: %v. Retrying in 5s...", err)
					time.Sleep(5 * time.Second)
				}
			}
		}()
	}

	// 3. API Server
	h := api.NewHandler(db)
	
	// Middleware
	// We can wrap specific routes or all.
	// Let's create a mux and wrap the whole thing or individual.
	mux := http.NewServeMux()
	mux.HandleFunc("/blotter", h.Blotter)
	mux.HandleFunc("/positions", h.Positions)
	mux.HandleFunc("/alarms", h.Alarms)
	// Health check usually public
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Health check write failed: %v", err)
		}
	})

	// Wrap with API Key Auth (except health?)
	// Custom wrapper to exclude /health
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			mux.ServeHTTP(w, r)
			return
		}
		middleware.APIKeyAuth(mux).ServeHTTP(w, r)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on port %s", port)
	if err := http.ListenAndServe(":"+port, finalHandler); err != nil {
		log.Fatal(err)
	}
}

func runIngestor(db *sql.DB, host string) error {
	user := os.Getenv("SFTP_USER")
	pass := os.Getenv("SFTP_PASS")
	dir := os.Getenv("SFTP_DIR")
	
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.Password(pass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For exercise
	}

	conn, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return err
	}
	defer conn.Close()

	client, err := sftp.NewClient(conn)
	if err != nil {
		return err
	}
	defer client.Close()

	worker := ingest.NewWorker(db, client, dir)
	worker.Start(context.Background()) // This blocks loop in worker, but we want it to run just once or loop inside? 
	// Worker.Start loops.
	return nil
}
