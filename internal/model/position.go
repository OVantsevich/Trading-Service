// Package model position model
package model

import "time"

// Position model
type Position struct {
	ID            string    `json:"id"`
	User          string    `json:"user"`
	Name          string    `json:"name"`
	Amount        float64   `json:"amount"`
	SellingPrice  float64   `json:"selling_price"`
	PurchasePrice float64   `json:"purchase_price"`
	StopLoss      float64   `json:"stop_loss"`
	TakeProfit    float64   `json:"take_profit"`
	ShortPosition bool      `json:"short_position"`
	Closed        int64     `json:"closed"`
	Created       time.Time `json:"created"`
	Updated       time.Time `json:"updated"`
}
