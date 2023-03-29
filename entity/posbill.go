package entity

import "time"

type PosBillLine struct {
	Id        uint64  `json:"id"`
	ProductId uint64  `json:"product_id"`
	Qty       float64 `json:"qty"`
	Price     float64 `json:"price"`
}

type PosBill struct {
	Id         uint64        `json:"id"`
	OrderNo    string        `json:"order_no"`
	OrderDate  time.Time     `json:"order_date"`
	CustomerId uint64        `json:"customer_id"`
	OrderLine  []PosBillLine `json:"order_line"`
	TimeUse    string        `json:"time_use"`
}
