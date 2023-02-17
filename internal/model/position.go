// Package model position model
package model

import "time"

// Position model
type Position struct {
	ID             string    `json:"id"`
	User           string    `json:"user"`
	Amount         float64   `json:"amount"`
	LowerThreshold float64   `json:"lowerThreshold"`
	UpperThreshold float64   `json:"upperThreshold"`
	Created        time.Time `json:"created"`
	Updated        time.Time `json:"updated"`
}
