package repository

import (
	"strconv"
	"time"

	"gorm.io/gorm"
	"tarlek.com/icesystem/entity"
)

type OrderRepository interface {
	CreateOrder(order entity.OrderCreate) entity.OrderCreate
	GetLastNo(company_id uint64, branch_id uint64, route_id uint64, route_code string) string
}

type orderRepository struct {
	connect *gorm.DB
}

func (db *orderRepository) CreateOrder(order entity.OrderCreate) entity.OrderCreate {
	//order.RunNo = db.GetLastNo(order.CompanyId, order.BranchId, order.RouteId, order.RouteCode)
	var data []entity.OrderLineStruct = order.DataList
	var order_master entity.OrderMaster
	var order_total_amt float64 = 0

	order_master.OrderNo = order.OrderNo
	order_master.OrderDate = time.Now()
	order_master.CustomerId = 0
	order_master.OrderChannelId = int64(order.RouteId)
	order_master.SaleChannelId = 1
	order_master.CarRefId = int64(order.CarId)
	order_master.IssueId = int64(order.IssueId)
	order_master.Status = 1
	order_master.CreatedBy = int64(order.UserId)
	order_master.CompanyId = int64(order.CompanyId)
	order_master.BranchId = int64(order.BranchId)
	order_master.SaleFromMobile = 1
	order_master.Emp_1 = int64(order.EmpId)
	order_master.Emp_2 = int64(order.EmpId2)
	order_master.OrderDate2 = time.Now()
	order_master.OrderShift = 0
	order_master.DiscountAmt = order.Discount
	order_master.PaymentMethodId = int64(order.PaymentTypeId)

	res := db.connect.Table("orders").Create(&order_master) // save and return id
	if res.RowsAffected > 0 {
		//print(res.RowsAffected)
		print(order_master.Id)

		for i := 0; i <= len(data)-1; i++ {
			//print(data[0].ProductId)
			if data[i].Qty <= 0 {
				continue
			}
			line_total := (data[i].Qty * data[i].Price)
			order_total_amt += line_total

			var line_price float64 = 0
			var line_total_price float64 = 0
			var is_free int = 0

			if order.PaymentTypeId != 3 {
				line_price = data[i].Price
				line_total_price = line_total
			} else {
				is_free = 1
			}

			var orderdetail entity.OrderDetail
			orderdetail.OrderId = order_master.Id
			orderdetail.CustomerId = int64(order.CustomerId)
			orderdetail.ProductId = int64(data[i].ProductId)
			orderdetail.Qty = data[i].Qty
			orderdetail.Price = line_price
			orderdetail.LineTotal = line_total_price
			orderdetail.PriceGroupId = int64(data[i].PriceGroupId)
			orderdetail.Status = 1
			orderdetail.SalePaymentMethodId = int64(order.PaymentTypeId)
			orderdetail.IssueRefId = int64(order.IssueId)
			orderdetail.IsFree = int64(is_free)

			res2 := db.connect.Table("order_line").Create(&orderdetail)
			if res2.RowsAffected > 0 {
				if order.PaymentTypeId != 3 {
					db.AddPayment(uint64(order_master.Id), order.CustomerId, orderdetail.LineTotal, uint64(order.CompanyId), order.BranchId, uint64(orderdetail.SalePaymentMethodId), order.UserId)
				}
				db.UpdateStock(order.RouteId, uint64(data[i].ProductId), data[i].Qty)
			}
		}
		if order_total_amt > 0 {
			db.connect.Table("orders").Where("id = ?", order_master.Id).Update("order_total_amt", order_total_amt)
		}
	}
	return order
}

type SelectedData struct {
	Id        uint64  `json:"id"`
	ProductId uint64  `json:"product_id"`
	AvlQty    float64 `json:"avl_qty"`
}

func (db *orderRepository) UpdateStock(route_id uint64, product_id uint64, qty float64) {
	var selectedData SelectedData
	//	res := db.connect.Table("order_stock").Where("route_id =?", route_id).Where("product_id = ?", product_id).Where("avl_qty >= ?", qty).Where("order_id = 202653").Select("id,product_id,avl_qty").Scan(&selectedData)
	res := db.connect.Table("order_stock").Where("route_id =?", route_id).Where("product_id = ?", product_id).Where("avl_qty >= ?", qty).Select("id,product_id,avl_qty").Scan(&selectedData)
	if res.Error == nil {
		res_update := db.connect.Table("order_stock").Where("id=?", selectedData.Id).Update("avl_qty", (selectedData.AvlQty - qty))
		if res_update.Error == nil {
			print("update stock ok")
			// print(selectedData.ProductId)
		}
	}
}

func (db *orderRepository) AddPayment(order_id uint64, customer_id uint64, amount float64, company_id uint64, branch_id uint64, payment_type_id uint64, user_id uint64) {
	var findone uint64 = 0
	var pay_amount float64 = 0
	current_date := time.Now().Local()

	recid := db.connect.Table("payment_receive").Where("customer_id = ?", customer_id).Where("date(trans_date) = ?", current_date.Format("2006-01-02")).Select("id").Take(&findone)
	if recid != nil {
		if payment_type_id == 1 {
			pay_amount = amount
		}
		print("not error but not found record")
		if findone > 0 {
			print("has old payment data")
			res_save_detail := db.connect.Table("payment_receive_line").Create(map[string]interface{}{"payment_receive_id": findone, "order_id": order_id, "payment_amount": pay_amount, "payment_channel_id": 1, "payment_method_id": payment_type_id, "status": 1, "payment_type_id": payment_type_id})
			if res_save_detail.Error == nil {
				print("create payment has old")
			}
		} else {
			//res := db.connect.Table("payment_receive").Create(map[string]interface{}{"trans_date": time.Now(), "journal_no": "xx", "status": 1, "company_id": company_id, "branch_id": branch_id})
			var payment = entity.PaymentMaster{
				Id:         0,
				TransDate:  time.Now(),
				CustomerId: customer_id,
				JournalNo:  db.GetPayLastNo(company_id, branch_id),
				Status:     1,
				CompanyId:  company_id,
				BranchId:   branch_id,
				CratedBy:   user_id,
				CreatedAt:  uint64(time.Now().Unix()),
			}
			if payment.JournalNo != "error na ja" {
				res := db.connect.Table("payment_receive").Create(&payment)
				if res.Error == nil {
					res_save_detail := db.connect.Table("payment_receive_line").Create(map[string]interface{}{"payment_receive_id": payment.Id, "order_id": order_id, "payment_amount": pay_amount, "payment_channel_id": 1, "payment_method_id": payment_type_id, "status": 1, "payment_type_id": payment_type_id})
					if res_save_detail.Error == nil {
						print("create payment")
					}
				}
			}

		}

	} else {
		print("not have old payment data")
	}
}

type MaxOrderNo struct {
	Id      uint64 `json:"id"`
	OrderNo string `json:"order_no"`
}

func (db *orderRepository) GetLastNo(company_id uint64, branch_id uint64, route_id uint64, route_code string) string {
	var max_order_no MaxOrderNo
	var pre string = "CO-" + route_code
	var prefix string = ""
	var cnum string = ""
	// var cnum2 int64 = 8
	var cnum3 int64 = 0
	current_date := time.Now().Local()

	// row := db.connect.Table("orders").Where("company_id=?", company_id).Where("branch_id=?", branch_id).Where("order_channel_id=?", route_id).Where("sale_from_mobile=1").Where("order_no LIKE ?", "CO%").Select("max(order_no)").Row()
	// err := row.Scan(&max_order_no)
	// if err != nil {
	// 	return "error na ja"
	// }
	row := db.connect.Table("orders").Where("company_id=?", company_id).Where("branch_id=?", branch_id).Where("order_channel_id=?", route_id).Where("sale_from_mobile=1").Where("date(order_date)=?", current_date.Format("2006-01-02")).Select("id,order_no").Last(&max_order_no)
	if row.Error != nil {
		//return "error na ja"
	}
	// if row.Error != nil {
	// 	print("error na")
	// }

	if max_order_no.OrderNo != "" {
		//max_order_no = "CO-VP31-230314-0009"
		var full_year string = strconv.Itoa(time.Now().Year())
		var full_month string = strconv.Itoa(int(time.Now().Month()))
		if len(full_month) == 1 {
			full_month = "0" + full_month
		}
		var full_day string = strconv.Itoa(time.Now().Day())
		if len(full_day) == 1 {
			full_day = "0" + full_day
		}
		prefix = pre + "-" + full_year[2:len(full_year)] + full_month + full_day + "-"
		cnum = max_order_no.OrderNo[15:len(max_order_no.OrderNo)]
		// cnum = "000"
		if cnumx, err := strconv.ParseInt(cnum, 10, 64); err != nil {
			panic(err)
		} else {
			//print("okk")
			cnum3 = cnumx + 1
			// cnum2 = cnumx
		}
		//cnum3 = cnum2 + 1

		var strlen int = len(cnum)
		var clen int = len(strconv.Itoa(int(cnum3)))
		var loop int = strlen - clen

		for i := 0; i <= loop-1; i++ {
			prefix = prefix + "0"
		}
		prefix = prefix + strconv.Itoa(int(cnum3))

		//return strconv.Itoa(int(cnum2))
	} else {
		var full_year string = strconv.Itoa(time.Now().Year())
		var full_month string = strconv.Itoa(int(time.Now().Month()))
		if len(full_month) == 1 {
			full_month = "0" + full_month
		}
		var full_day string = strconv.Itoa(time.Now().Day())
		if len(full_day) == 1 {
			full_day = "0" + full_day
		}
		prefix = pre + "-" + full_year[2:len(full_year)] + full_month + full_day + "-" + "0001"
	}

	return prefix
}

type MaxPayJournalNo struct {
	Id        uint64 `json:"id"`
	JournalNo string `json:"journal_no"`
}

func (db *orderRepository) GetPayLastNo(company_id uint64, branch_id uint64) string {
	var max_journal_no MaxPayJournalNo
	var pre string = "AR-"
	var prefix string = ""
	var cnum string = ""
	// var cnum2 int64 = 8
	var cnum3 int64 = 0
	// var prefix2 strings.Builder
	current_date := time.Now().Local()

	// row := db.connect.Table("payment_receive").Where("company_id=?", company_id).Where("branch_id=?", branch_id).Where("date(trans_date)=?", current_date.Format("2006-01-02")).Select("max(journal_no)").Row()
	// err := row.Scan(&max_journal_no)
	// if err != nil {
	// 	return "error na ja"
	// }
	row := db.connect.Table("payment_receive").Where("company_id=?", company_id).Where("branch_id=?", branch_id).Where("date(trans_date)=?", current_date.Format("2006-01-02")).Select("id,journal_no").Last(&max_journal_no)
	if row.Error != nil {
		//panic(row.Error)
		// if(row.Error.Error() == "record not found"){

		// }
		// return "error na ja"
	}

	if max_journal_no.JournalNo != "" {
		//max_journal_no = "CO-VP31-230314-0009"
		var full_year string = strconv.Itoa(time.Now().Year())
		var full_month string = strconv.Itoa(int(time.Now().Month()))
		if len(full_month) == 1 {
			full_month = "0" + full_month
		}
		var full_day string = strconv.Itoa(time.Now().Day())
		if len(full_day) == 1 {
			full_day = "0" + full_day
		}
		//prefix2.WriteString(pre + full_year[2:len(full_year)] + full_month + full_day + "-")
		prefix = pre + full_year[2:len(full_year)] + full_month + full_day + "-"
		cnum = max_journal_no.JournalNo[10:len(max_journal_no.JournalNo)]
		// cnum = "000"
		if cnumx, err := strconv.ParseInt(cnum, 10, 64); err != nil {
			panic(err)
		} else {
			//print("okk")
			cnum3 = cnumx + 1
			// cnum2 = cnumx
		}
		//cnum3 = cnum2 + 1

		var strlen int = len(cnum)
		var clen int = len(strconv.Itoa(int(cnum3)))
		var loop int = strlen - clen

		for i := 0; i <= loop-1; i++ {
			prefix = prefix + "0"
		}
		prefix = prefix + strconv.Itoa(int(cnum3))

		//return strconv.Itoa(int(cnum2))
	} else {
		var full_year string = strconv.Itoa(time.Now().Year())
		var full_month string = strconv.Itoa(int(time.Now().Month()))
		if len(full_month) == 1 {
			full_month = "0" + full_month
		}
		var full_day string = strconv.Itoa(time.Now().Day())
		if len(full_day) == 1 {
			full_day = "0" + full_day
		}
		//prefix2.WriteString(pre + full_year[2:len(full_year)] + full_month + full_day + "-")
		prefix = pre + full_year[2:len(full_year)] + full_month + full_day + "-" + "0001"
	}

	return prefix
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{connect: db}
}
