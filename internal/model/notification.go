// Package model notification model
package model

// Notification notify from postgres SL or TP
type Notification struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
}
