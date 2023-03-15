package entity

type OrderDetail struct {
	OrderId             int64   `json:"order_id"`
	CustomerId          int64   `json:"customer_id"`
	ProductId           int64   `json:"product_id"`
	Qty                 float64 `json:"qty"`
	Price               float64 `json:"price"`
	LineTotal           float64 `json:"line_total"`
	PriceGroupId        int64   `json:"price_group_id"`
	Status              int64   `json:"status"`
	SalePaymentMethodId int64   `json:"sale_payment_method_id"`
	IssueRefId          int64   `json:"issue_ref_id"`
	IsFree              int64   `json:"is_free"`
}
