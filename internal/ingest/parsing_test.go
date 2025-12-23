package ingest

import (
	"strings"
	"testing"
)

func TestParseFormat1(t *testing.T) {
	// Trade Data (CSV)
	csvData := `TradeDate,AccountID,Ticker,Quantity,Price,TradeType,SettlementDate
2025-01-15,1001,AMZN,10,185.50,BUY,2025-01-17`

	r := strings.NewReader(csvData)
	records, err := ParseFormat1(r)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	tr := records[0]
	if tr.Ticker != "AMZN" {
		t.Errorf("Expected ticker AMZN, got %s", tr.Ticker)
	}
	if tr.Quantity != 10 {
		t.Errorf("Expected qty 10, got %f", tr.Quantity)
	}
	if tr.Price != 185.50 {
		t.Errorf("Expected price 185.50, got %f", tr.Price)
	}
}

func TestParseFormat2(t *testing.T) {
	// Report Data (Pipe)
	pipeData := `ReportDate|AccountID|SecurityTicker|Shares|MarketValue|SourceSystem
20250115|1001|GOOG|50|140.00|ReportingSystem`

	r := strings.NewReader(pipeData)
	records, err := ParseFormat2(r)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("Expected 1 record, got %d", len(records))
	}

	rr := records[0]
	if rr.SecurityTicker != "GOOG" {
		t.Errorf("Expected ticker GOOG, got %s", rr.SecurityTicker)
	}
	if rr.Shares != 50 {
		t.Errorf("Expected shares 50, got %f", rr.Shares)
	}
	if rr.MarketValue != 140.00 {
		t.Errorf("Expected MV 140.00, got %f", rr.MarketValue)
	}
}
