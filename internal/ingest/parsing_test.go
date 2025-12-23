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

func TestParseFormat2_UserSample(t *testing.T) {
	// User provided sample with UPPERCASE headers and potential line merging
	pipeData := `REPORT_DATE|ACCOUNT_ID|SECURITY_TICKER|SHARES|MARKET_VALUE|SOURCE_SYSTEM
20250115|ACC001|AAPL|100|18550.00|CUSTODIAN_A 20250115|ACC001|MSFT|50|21012.50|CUSTODIAN_A
20250115|ACC001|GOOGL|100|14280.00|CUSTODIAN_A
20250115|ACC002|GOOGL|75|10710.00|CUSTODIAN_B
20250115|ACC002|AAPL|200|37100.00|CUSTODIAN_B
20250115|ACC002|NVDA|120|60636.00|CUSTODIAN_B
20250115|ACC003|TSLA|-150|-35767.50|CUSTODIAN_A
20250115|ACC003|NVDA|80|40424.00|CUSTODIAN_A 20250115|ACC004|AAPL|500|92750.00|CUSTODIAN_C
20250115|ACC004|MSFT|300|126075.00|CUSTODIAN_C`

	r := strings.NewReader(pipeData)
	records, err := ParseFormat2(r)

	if err != nil {
		t.Fatalf("Parse error with user sample: %v", err)
	}

	// We expect 10 records
	if len(records) != 10 {
		t.Errorf("Got %d records, expected 10", len(records))
		for i, r := range records {
			t.Logf("Record %d: %+v", i, r)
		}
	}
	
	// Verify spot check
	// Record 0: AAPL
	if records[0].SecurityTicker != "AAPL" {
		t.Errorf("Rec 0 mismatch: %v", records[0])
	}
	// Record 1: MSFT (was merged on line 2)
	if records[1].SecurityTicker != "MSFT" {
		t.Errorf("Rec 1 mismatch: %v", records[1])
	}
	if records[1].SourceSystem != "CUSTODIAN_A" {
		t.Errorf("Rec 1 Source mismatch: %v", records[1])
	}
}
