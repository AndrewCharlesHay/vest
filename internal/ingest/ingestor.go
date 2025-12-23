package ingest

import (
	"context"
	"database/sql"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/AndrewCharlesHay/vest/internal/models"
	"github.com/pkg/sftp"

)

type Worker struct {
	DB         *sql.DB
	SFTPClient *sftp.Client
	UploadDir  string
}

func NewWorker(db *sql.DB, sftpClient *sftp.Client, dir string) *Worker {
	return &Worker{
		DB:         db,
		SFTPClient: sftpClient,
		UploadDir:  dir,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.ProcessFiles(); err != nil {
				log.Printf("Error processing files: %v", err)
			}
		}
	}
}

func (w *Worker) ProcessFiles() error {
	files, err := w.SFTPClient.ReadDir(w.UploadDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filename := file.Name()
		// Basic check to see if we processed it? 
		// For this exercise, we'll try to process and maybe move/delete or just overwrite in DB.
		// A real system would track processed files. We will just upsert.
		// Skip dot files
		if strings.HasPrefix(filename, ".") {
			continue
		}

		log.Printf("Processing file: %s", filename)
		f, err := w.SFTPClient.Open(filepath.Join(w.UploadDir, filename))
		if err != nil {
			log.Printf("Failed to open file %s: %v", filename, err)
			continue
		}
		
		// Determine format based on strict heuristic or try both?
		// Format 1 is CSV (comma), Format 2 is Pipe.
		// Heuristic: Read first line? Or just filename/structure.
		// The requirements don't specify filenames. I'll peek content or just try parsing.
		// Simpler: Format 1 has 7 columns, Format 2 has 6. 
		// Actually Format 2 has headers with PIPE. Format 1 has commas.
		
		// Reset cursor after peeking if needed, but easier to just read all content.
		// Note: csv.Reader consumes io.Reader.
		
		// Simplification for exercise: Try Format 2 (Pipe) first as it is distinct. If fails, try Format 1.
		// Resetting seek is needed.
		
		// NOTE: pkg/sftp File supports Seek.
		
	// Try Pipe
		_, err := f.Seek(0, 0)
		if err != nil {
			log.Printf("Failed to seek file %s: %v", filename, err)
			f.Close()
			continue
		}

		records2, err2 := ParseFormat2(f)
		if err2 == nil && len(records2) > 0 {
			// It is format 2
			err = w.IngestFormat2(records2)
		} else {
			// Try Format 1
			if _, seekErr := f.Seek(0, 0); seekErr != nil {
				log.Printf("Failed to seek file %s: %v", filename, seekErr)
				f.Close()
				continue
			}

			records1, err1 := ParseFormat1(f)
			if err1 == nil && len(records1) > 0 {
				err = w.IngestFormat1(records1)
			} else {
				log.Printf("Could not parse file %s as either format", filename)
				f.Close()
				continue
			}
		}

		f.Close()
		
		if err != nil {
			log.Printf("Failed to ingest %s: %v", filename, err)
		} else {
			log.Printf("Successfully ingested %s", filename)
			// Move or delete to avoid reprocessing endlessly in this loop
			// For exercise, we delete
			if err := w.SFTPClient.Remove(filepath.Join(w.UploadDir, filename)); err != nil {
				log.Printf("Failed to remove file %s: %v", filename, err)
			}
		}
	}
	return nil
}

func (w *Worker) IngestFormat1(records []models.TradeRecord) error {
	tx, err := w.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("Failed to rollback tx: %v", err)
		}
	}()

	stmt, err := tx.Prepare(`
		INSERT INTO positions (date, account_id, ticker, quantity, market_value, shares, source_system)
		VALUES ($1, $2, $3, $4, $5, $4, 'Trade')
		ON CONFLICT (date, account_id, ticker) 
		DO UPDATE SET 
			quantity = positions.quantity + EXCLUDED.quantity,
			shares = positions.shares + EXCLUDED.shares,
			market_value = positions.market_value + EXCLUDED.market_value
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range records {
		// Calculate signed quantity based on BUY/SELL
		qty := r.Quantity
		if r.TradeType == "SELL" {
			qty = -qty
		}
		
		// Market Value needed? Format 1 has Price. MV = Qty * Price
		mv := qty * r.Price

		_, err := stmt.Exec(r.TradeDate, r.AccountID, r.Ticker, qty, mv)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (w *Worker) IngestFormat2(records []models.ReportRecord) error {
	tx, err := w.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("Failed to rollback tx: %v", err)
		}
	}()

	stmt, err := tx.Prepare(`
		INSERT INTO positions (date, account_id, ticker, quantity, market_value, shares, source_system)
		VALUES ($1, $2, $3, $4, $5, $4, $6)
		ON CONFLICT (date, account_id, ticker) 
		DO UPDATE SET 
			quantity = EXCLUDED.quantity, 
			market_value = EXCLUDED.market_value,
			shares = EXCLUDED.shares,
			source_system = EXCLUDED.source_system
	`)
	// Note: Format 2 seems to be snapshots ("Report"), so we might want to OVERWRITE or UPDATE absolute values, not add.
	// Logic above: Update with new values (snapshot).
	
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range records {
		// Date format 20250115 needs parsing to 2025-01-15 for consistency w/ DB date type
		parsedDate, _ := time.Parse("20060102", r.ReportDate)
		dateStr := parsedDate.Format("2006-01-02")

		_, err := stmt.Exec(dateStr, r.AccountID, r.SecurityTicker, r.Shares, r.MarketValue, r.SourceSystem)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
