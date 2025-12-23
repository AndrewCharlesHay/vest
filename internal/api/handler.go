package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AndrewCharlesHay/vest/internal/models"
)

type Handler struct {
	DB *sql.DB
}

func NewHandler(db *sql.DB) *Handler {
	return &Handler{DB: db}
}

func (h *Handler) Blotter(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "date parameter required", http.StatusBadRequest)
		return
	}

	rows, err := h.DB.Query(`
		SELECT date, account_id, ticker, quantity, market_value 
		FROM positions 
		WHERE date = $1
	`, date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var results []models.BlotterResponse
	for rows.Next() {
		var b models.BlotterResponse
		var d time.Time // Scan date as time
		if err := rows.Scan(&d, &b.AccountID, &b.Ticker, &b.Quantity, &b.MarketValue); err != nil {
			continue
		}
		b.Date = d.Format("2006-01-02")
		results = append(results, b)
	}

	if err := json.NewEncoder(w).Encode(results); err != nil {
		return
	}
}

func (h *Handler) Positions(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "date parameter required", http.StatusBadRequest)
		return
	}

	// Calculate total MV per account first
	// Or do it in SQL: Window functions usually better but let's do two queries or one complex one.
	// Complex query:
	query := `
		WITH AccountTotals AS (
			SELECT account_id, SUM(market_value) as total_mv
			FROM positions
			WHERE date = $1
			GROUP BY account_id
		)
		SELECT p.account_id, p.ticker, p.market_value, a.total_mv
		FROM positions p
		JOIN AccountTotals a ON p.account_id = a.account_id
		WHERE p.date = $1
	`
	
	rows, err := h.DB.Query(query, date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Map: Account -> Ticker -> %
	respMap := make(map[string]map[string]float64)

	for rows.Next() {
		var accID, ticker string
		var mv, totalMv float64
		if err := rows.Scan(&accID, &ticker, &mv, &totalMv); err != nil {
			continue
		}
		
		if _, ok := respMap[accID]; !ok {
			respMap[accID] = make(map[string]float64)
		}
		
		if totalMv != 0 {
			respMap[accID][ticker] = (mv / totalMv) * 100
		} else {
			respMap[accID][ticker] = 0
		}
	}

	var response []models.PositionResponse
	for acc, allocs := range respMap {
		response = append(response, models.PositionResponse{
			AccountID:   acc,
			Allocations: allocs,
		})
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		return
	}
}

func (h *Handler) Alarms(w http.ResponseWriter, r *http.Request) {
	date := r.URL.Query().Get("date")
	if date == "" {
		http.Error(w, "date parameter required", http.StatusBadRequest)
		return
	}
	
	// Re-use logic or SQL to find > 20%
	query := `
		WITH AccountTotals AS (
			SELECT account_id, SUM(market_value) as total_mv
			FROM positions
			WHERE date = $1
			GROUP BY account_id
		)
		SELECT p.account_id, p.ticker, p.market_value, a.total_mv
		FROM positions p
		JOIN AccountTotals a ON p.account_id = a.account_id
		WHERE p.date = $1
	`
	rows, err := h.DB.Query(query, date)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	
	violationMap := make(map[string][]string) // Account -> Violations

	for rows.Next() {
		var accID, ticker string
		var mv, totalMv float64
		if err := rows.Scan(&accID, &ticker, &mv, &totalMv); err != nil {
			continue
		}
		
		pct := 0.0
		if totalMv != 0 {
			pct = (mv / totalMv) * 100
		}
		
		if pct > 20.0 {
			msg := fmt.Sprintf("%s is %.2f%% of portfolio", ticker, pct)
			violationMap[accID] = append(violationMap[accID], msg)
		}
	}
	
	var response []models.AlarmResponse
	// We need to return info for ALL accounts? Req says: "Returns true for any account that has over 20%"
	// Let's return only those with violations? Or all checked?
	// Usually "Alarms" endpoint returns active alarms.
	for acc, violations := range violationMap {
		response = append(response, models.AlarmResponse{
			Date:         date,
			AccountID:    acc,
			HasViolation: true,
			ViolationInfo: fmt.Sprintf("Violations: %v", violations),
		})
	}
	// If no violations, return empty list or specific message? Empty list is standard json.
	if response == nil {
		response = []models.AlarmResponse{}
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Can't effectively change status if partly written, but should verify
		// Just log?
		// log.Printf("Failed to encode response: %v", err)
		return
	}
}
