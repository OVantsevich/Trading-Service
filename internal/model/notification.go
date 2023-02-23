// Package model notification model
package model

import "time"

// Notification notify from postgres SL or TP
type Notification struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	User   string     `json:"user"`
	Type   string     `json:"type"`
	Price  float64    `json:"price"`
	Closed *time.Time `json:"closed"`
}
