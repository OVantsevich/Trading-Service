// Package model position model
package model

import "time"

// Position model
type Position struct {
	ID         string    `json:"id"`
	User       string    `json:"user"`
	Name       string    `json:"name"`
	Amount     float64   `json:"amount"`
	StopLoss   float64   `json:"stop_loss"`
	TakeProfit float64   `json:"take_profit"`
	Closed     time.Time `json:"closed"`
	Created    time.Time `json:"created"`
	Updated    time.Time `json:"updated"`
}
