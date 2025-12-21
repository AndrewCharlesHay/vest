package models

import (
	"time"
)

// Position represents a holding in an account for a specific date
type Position struct {
	Date         time.Time `json:"date"`
	AccountID    string    `json:"account_id"`
	Ticker       string    `json:"ticker"`
	Quantity     float64   `json:"quantity"`
	MarketValue  float64   `json:"market_value"`
	SourceSystem string    `json:"source_system,omitempty"`
}

// TradeRecord represents a row from Format 1 (CSV)
type TradeRecord struct {
	TradeDate      string
	AccountID      string
	Ticker         string
	Quantity       float64
	Price          float64
	TradeType      string
	SettlementDate string
}

// ReportRecord represents a row from Format 2 (Pipe-delimited)
type ReportRecord struct {
	ReportDate    string
	AccountID     string
	SecurityTicker string
	Shares        float64
	MarketValue   float64
	SourceSystem  string
}

// BlotterResponse represents the simplified data for the blotter endpoint
type BlotterResponse struct {
	Date        string  `json:"date"`
	AccountID   string  `json:"account_id"`
	Ticker      string  `json:"ticker"`
	Quantity    float64 `json:"quantity"`
	MarketValue float64 `json:"market_value"`
}

// PositionResponse represents the % of funds by ticker
type PositionResponse struct {
	AccountID   string             `json:"account_id"`
	Allocations map[string]float64 `json:"allocations"` // Ticker -> Percentage
}

// AlarmResponse represents the alarm compliance check
type AlarmResponse struct {
	Date          string `json:"date"`
	AccountID     string `json:"account_id"`
	HasViolation  bool   `json:"has_violation"`
	ViolationInfo string `json:"violation_info,omitempty"`
}
