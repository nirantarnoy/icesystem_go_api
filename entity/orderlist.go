package entity

import "time"

type OrderList struct {
	Id                  uint64    `json:"id" form:"id"`
	OrderNo             string    `json:"order_no" form:"order_no"`
	OrderDate           time.Time `json:"order_date" form:"order_date"`
	OrderStatus         uint64    `json:"order_status" form:"order_status"`
	CustomerId          uint64    `json:"customer_id" form:"customer_id"`
	CustomerCode        string    `json:"customer_code" form:"customer_code"`
	CustomerName        string    `json:"customer_name" form:"customer_name"`
	Note                string    `json:"note" form:"note"`
	SalePaymentMethodId uint64    `json:"sale_payment_method_id" form:"sale_payment_method_id"`
	LineTotal           float64   `json:"line_total" form:"line_total"`
	Qty                 float64   `json:"qty" form:"qty"`
	Price               float64   `json:"price" form:"price"`
	OrderLineId         uint64    `json:"order_line_id" form:"order_line_id"`
	ProductId           uint64    `json:"product_id" form:"product_id"`
	ProductCode         string    `json:"product_code" form:"product_code"`
	ProductName         string    `json:"product_name" form:"product_name"`
	OrderLineDate       time.Time `json:"order_line_date" form:"created_at"`
	OrderLineStatus     uint64    `json:"order_line_status" form:"order_line_status"`
}
