// Package model notification model
package model

// Notification notify from postgres SL or TP
type Notification struct {
	*Position
	Type string `json:"type"`
}
