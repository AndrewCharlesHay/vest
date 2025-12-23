package ingest

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"

	"github.com/AndrewCharlesHay/vest/internal/models"
)

func ParseFormat1(r io.Reader) ([]models.TradeRecord, error) {
	reader := csv.NewReader(r)
	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	var records []models.TradeRecord
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if len(row) < 7 {
			continue 
		}

		qty, _ := strconv.ParseFloat(row[3], 64)
		price, _ := strconv.ParseFloat(row[4], 64)

		records = append(records, models.TradeRecord{
			TradeDate:      row[0],
			AccountID:      row[1],
			Ticker:         row[2],
			Quantity:       qty,
			Price:          price,
			TradeType:      row[5],
			SettlementDate: row[6],
		})
	}
	return records, nil
}

func ParseFormat2(r io.Reader) ([]models.ReportRecord, error) {
	// Read full content to handle messy newlines/concatenated records
	content, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	
	// Split by Pipe
	// The structure is F1|F2|F3|F4|F5|F6 (End of Rec 1) (Start of Rec 2) F1|F2...
	// So we expect 5 pipes per record.
	// If we split by "|", we get tokens.
	// Index 0: F1
	// Index 1: F2
	// ...
	// Index 4: F5
	// Index 5: "F6 [whitespace] F1"  <-- Boundary
	// Index 6: F2
	// ...
	
	rawTokens := strings.Split(string(content), "|")
	var allFields []string

	for i, token := range rawTokens {
		token = strings.TrimSpace(token)
		
		// Boundary check: Every 5th index (0-based) starting from 5: 5, 10, 15...
		// But NOT the very last token (which is just the last F6).
		if i > 0 && i%5 == 0 && i != len(rawTokens)-1 {
			// This token likely contains "SourceSystem [Whitespace] NextDate"
			// We split by whitespace. 
			// Heuristic: The NextDate (F1) is 20250115 (numeric-ish). SourceSystem might have spaces?
			// Assumption: F1 (Date) does not have spaces.
			// We split by Fields and take the last part as F1, the rest as F6.
			
			parts := strings.Fields(token)
			if len(parts) >= 2 {
				fNext := parts[len(parts)-1]
				fPrev := strings.Join(parts[:len(parts)-1], " ")
				allFields = append(allFields, fPrev)
				allFields = append(allFields, fNext)
			} else {
				// Fallback if no whitespace found? just treat as one field (likely error downstream)
				allFields = append(allFields, token)
			}
		} else {
			allFields = append(allFields, token)
		}
	}

	// Now we have a flat list of fields. Group by 6.
	const fieldsPerRecord = 6
	if len(allFields) < fieldsPerRecord {
		return nil, nil // Empty or just header?
	}

	var records []models.ReportRecord
	
	// Skip Header (First 6 fields)
	// We assume input ALWAYS has header "REPORT_DATE|..."
	// Basic validation: Check if first field looks like "ReportDate" or "REPORT_DATE"
	// For robust skipping, just skip first 6.
	
	startIndex := fieldsPerRecord
	
	for i := startIndex; i < len(allFields); i += fieldsPerRecord {
		// Ensure we have enough fields remaining
		if i+fieldsPerRecord > len(allFields) {
			break // Incomplete record at end
		}
		
		row := allFields[i : i+fieldsPerRecord]
		
		// Parse
		shares, _ := strconv.ParseFloat(row[3], 64)
		mv, _ := strconv.ParseFloat(row[4], 64)

		records = append(records, models.ReportRecord{
			ReportDate:     row[0],
			AccountID:      row[1],
			SecurityTicker: row[2],
			Shares:         shares,
			MarketValue:    mv,
			SourceSystem:   row[5],
		})
	}

	return records, nil
}
