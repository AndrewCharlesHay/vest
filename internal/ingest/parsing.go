package ingest

import (
	"encoding/csv"
	"io"
	"strconv"

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
	// Pipe delimited
	reader := csv.NewReader(r)
	reader.Comma = '|'

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, err
	}

	var records []models.ReportRecord
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		
		if len(row) < 6 {
			continue
		}

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
