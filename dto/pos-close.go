package dto

type PosSaleLineDto struct {
	ProductId uint64  `json:"product_id"`
	Qty       float64 `json:"qty"`
	Price     float64 `json:"price"`
}

type PosCloseDataDto struct {
	CustomerId      uint64           `json:"customer_id" form:"customer_id"`
	PosList         []PosSaleLineDto `json:"data_list" form:"data_list"`
	SalePayType     uint64           `json:"sale_pay_type" form:"sale_pay_type"`
	SaleTotalAmount float64          `json:"sale_total_amount" form:"sale_total_amount"`
	SalePayAmount   float64          `json:"sale_pay_amount" form:"sale_pay_amount"`
	UserId          uint64           `json:"user_id" form:"user_id"`
	WarehouseId     uint64           `json:"warehouse_id" form:"warehouse_id"`
	CompanyID       uint64           `json:"company_id" form:"company_id"`
	BranchId        uint64           `json:"branch_id" form:"branch_id"`
	PaymentMethodId uint64           `json:"payment_method_id" form:"payment_method_id"`
}
