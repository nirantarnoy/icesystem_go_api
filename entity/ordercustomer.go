package entity

type OrderCustomer struct {
	CarId          uint64 `json:"car_id" form:"car_id"`
	SearchCustomer uint64 `json:"searchcustomer" form:"searchcustomer"`
}
