package repository

import (
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"
	"tarlek.com/icesystem/entity"
)

type PosCloseRepository interface {
	CloseOrder(closData entity.PosCloseData) entity.PosBill
}

type posCloseRepository struct {
	connect *gorm.DB
}

type OrderHeader struct {
	Id              uint64
	OrderNo         string    `json:"order_no"`
	OrderDate       time.Time `json:"order_date"`
	CustomerId      uint64    `json:"customer_id"`
	SaleChannelId   uint64    `json:"sale_channel_id"`
	PaymentStatus   uint64    `json:"payment_status"`
	OrderTotalAmt   float64   `json:"order_total_amt"`
	Status          uint64    `json:"status"`
	CompanyId       uint64    `json:"company_id"`
	BranchId        uint64    `json:"branch_id"`
	PaymentMethodId uint64    `json:"payment_method_id"`
	CreatedAt       uint64    `json:"created_at"`
	CreatedBy       uint64    `json:"created_by"`
}

type OrderLine struct {
	Id           uint64  `json:"id"`
	OrderId      uint64  `json:"order_id"`
	ProductId    uint64  `json:"product_id"`
	Qty          float64 `json:"qty"`
	Price        float64 `json:"price"`
	LineTotal    float64 `json:"line_total"`
	CustomerId   uint64  `json:"customer_id"`
	PriceGroupId uint64  `json:"price_group_id"`
	Status       uint64  `json:"status"`
	CreatedAt    uint64  `json:"created_at"`
}

// CloseOrder implements PosCloseRepository
func (db *posCloseRepository) CloseOrder(closeData entity.PosCloseData) entity.PosBill {
	var posHeader OrderHeader
	var posLineList []entity.PosSaleLine = closeData.PosList
	var posBill entity.PosBill
	var posBillOrderLine []entity.PosBillLine
	var payment_success int = 0
	//var result int64 = 0
	fmt.Println("data is ")
	fmt.Println(int64(len(posLineList)))
	// fmt.Printf("order data is %v", closeData)
	//return int64(closeData.PosList[0].ProductId)

	time_start := time.Now()
	posHeader.OrderNo = db.GetOrderNo(closeData.CompanyID, closeData.BranchId)
	posHeader.OrderDate = time.Now().Local()
	posHeader.CustomerId = closeData.CustomerId
	posHeader.SaleChannelId = 2
	posHeader.PaymentStatus = 0
	posHeader.OrderTotalAmt = (closeData.SalePayAmount)
	posHeader.Status = 1
	posHeader.CompanyId = closeData.CompanyID
	posHeader.BranchId = closeData.BranchId
	posHeader.PaymentMethodId = closeData.PaymentMethodId
	posHeader.CreatedBy = closeData.UserId

	res := db.connect.Table("orders").Create(&posHeader)
	if res.RowsAffected > 0 {

		posBill.Id = posHeader.Id
		posBill.CustomerId = closeData.CustomerId
		posBill.OrderDate = posHeader.OrderDate
		posBill.OrderNo = posHeader.OrderNo

		create_issue_result := make(chan bool)
		go db.createIssue(posHeader.Id, posLineList, closeData.CompanyID, closeData.BranchId, create_issue_result)
		fmt.Println(<-create_issue_result)

		for _, x := range posLineList {
			//print("loop")
			var line_total = (x.Qty * x.Price)
			var orderLine = OrderLine{
				OrderId:      posHeader.Id,
				CustomerId:   closeData.CustomerId,
				ProductId:    x.ProductId,
				Qty:          x.Qty,
				Price:        x.Price,
				LineTotal:    line_total,
				PriceGroupId: 0,
				CreatedAt:    uint64(time.Now().Unix()),
			}

			if orderLine.OrderId > 0 {
				resline := db.connect.Table("order_line").Create(&orderLine)
				if resline.Error != nil {
					print(resline.Error.Error())
				} else {
					var bill_line entity.PosBillLine
					bill_line.Id = orderLine.Id
					bill_line.ProductId = orderLine.ProductId
					bill_line.Qty = orderLine.Qty
					bill_line.Price = orderLine.Price
					posBillOrderLine = append(posBillOrderLine, bill_line)

				}

				if resline.RowsAffected > 0 {
					create_stock_result := make(chan bool)
					go db.CreateStockTrans(posHeader.OrderNo, x.ProductId, x.Qty, closeData.WarehouseId, posHeader.Id, closeData.UserId, closeData.CompanyID, closeData.BranchId, create_stock_result)
					//fmt.Println(<-create_stock_result)

					update_result := make(chan bool)
					go db.UpdateSummary(x.ProductId, closeData.WarehouseId, x.Qty, update_result)
					//fmt.Sprintln(<-update_result)

					if payment_success == 0 { // check not pay yet
						create_payment_restult := make(chan bool)
						go db.CreatePayment(posHeader.Id, closeData.SalePayAmount, closeData.CustomerId, closeData.CompanyID, closeData.BranchId, closeData.UserId, create_payment_restult)
						if <-create_payment_restult == true {
							//fmt.Println(<-create_payment_restult)
							payment_success = 1
						}
					}

				}
			}
		}
		posBill.OrderLine = posBillOrderLine // add order detail
	}
	time_end := time.Now()
	time_used := time_end.Sub(time_start)
	posBill.TimeUse = time_used.String()

	return posBill
}

func (db *posCloseRepository) UpdateSummary(product_id uint64, warehouse_id uint64, qty float64, result chan<- bool) {
	var res_update bool = false
	if warehouse_id > 0 && product_id > 0 && qty > 0 {
		var old_qty float64
		res := db.connect.Table("stock_sum").Select("qty").Where("warehouse_id=? and product_id=?", warehouse_id, product_id).Scan(&old_qty)
		if res.Error != nil {
			res_update = false
		}
		if old_qty < 0 && old_qty < qty {
			res_update = false
		} else {
			new_qty := (old_qty - qty)
			resupdate := db.connect.Table("stock_sum").Where("warehouse_id=? and product_id=?", warehouse_id, product_id).Update("qty", new_qty)
			if resupdate.RowsAffected > 0 {
				res_update = true
			}
		}
	}
	result <- res_update

}

type StockTransCreate struct {
	JournalNo      string    `json:"journal_no"`
	TransDate      time.Time `json:"trans_date"`
	ProductId      uint64    `json:"product_id"`
	Qty            float64   `json:"qty"`
	WarehouseId    uint64    `json:"warehouse_id"`
	StockType      uint64    `json:"stock_type"`
	ActivityTypeId uint64    `json:"activity_type_id"`
	TransRefId     uint64    `json:"trans_ref_id"`
	CompanyID      uint64    `json:"company_id"`
	BranchId       uint64    `json:"branch_id"`
	CreatedBy      uint64    `json:"created_by"`
}

func (db *posCloseRepository) CreateStockTrans(order_no string, product_id uint64, qty float64, warehouse_id uint64, order_id uint64, user_id uint64, company_id uint64, branch_id uint64, result chan<- bool) {
	var res_create bool = false
	// createData := map[string]interface{}{
	// 	"journal_no":       order_no,
	// 	"trans_date":       time.Now().Local(),
	// 	"product_id":       product_id,
	// 	"qty":              qty,
	// 	"warehouse_id":     warehouse_id,
	// 	"stock_type":       2,
	// 	"activity_type_id": 5,
	// 	"trans_ref_id":     order_id,
	// 	"company_id":       company_id,
	// 	"branch_id":        branch_id,
	// 	"created_by":       user_id,
	// }
	var createData = StockTransCreate{
		JournalNo:      order_no,
		TransDate:      time.Now().Local(),
		ProductId:      product_id,
		Qty:            qty,
		WarehouseId:    warehouse_id,
		StockType:      2,
		ActivityTypeId: 5,
		TransRefId:     order_id,
		CompanyID:      company_id,
		BranchId:       branch_id,
		CreatedBy:      user_id,
	}
	res := db.connect.Table("stock_trans").Create(&createData)
	if res.Error == nil {
		res_create = true
	}

	result <- res_create

}

type PaymentHeader struct {
	Id         uint64    `json:"id"`
	JournalNo  string    `json:"journal_no"`
	TransDate  time.Time `json:"trans_date"`
	CustomerId uint64    `json:"customer_id"`
	Status     uint64    `json:"status"`
	CompanyId  uint64    `json:"company_id"`
	BranchId   uint64    `json:"branch_id"`
	CratedBy   uint64    `json:"crated_by"`
	CreatedAt  uint64    `json:"created_at"`
}

type PaymentLine struct {
	PaymentReceiveId uint64  `json:"payment_receive_id"`
	OrderId          uint64  `json:"order_id"`
	PaymentAmount    float64 `json:"payment_amount"`
	PaymentChannelId uint64  `json:"payment_channel_id"`
	PaymentMethodId  uint64  `json:"payment_method_id"`
	PaymentTypeId    uint64  `json:"payment_type_id"`
	PaymentTermId    uint64  `json:"payment_term_id"`
	Status           uint64  `json:"status"`
}

func (db *posCloseRepository) CreatePayment(order_id uint64, payment_amount float64, customer_id uint64, company_id uint64, branch_id uint64, user_id uint64, result chan<- bool) {
	var result_payment bool = false
	createData := PaymentHeader{
		JournalNo:  db.GetPaymentNo(company_id, branch_id),
		TransDate:  time.Now().Local(),
		CustomerId: customer_id,
		Status:     1,
		CompanyId:  company_id,
		BranchId:   branch_id,
		CratedBy:   user_id,
		CreatedAt:  uint64(time.Now().Unix()),
	}
	res := db.connect.Table("payment_receive").Create(&createData)
	if res.Error == nil {
		lineData := PaymentLine{
			PaymentReceiveId: createData.Id,
			OrderId:          order_id,
			PaymentAmount:    payment_amount,
			PaymentChannelId: 0,
			PaymentMethodId:  2,
			PaymentTypeId:    2,
			PaymentTermId:    0,
			Status:           1,
		}
		res_line := db.connect.Table("payment_receive_line").Create(&lineData)
		if res_line.Error == nil {
			result_payment = true
		}
	}
	result <- result_payment
}

type IssueHeader struct {
	Id         uint64    `json:"id"`
	JournalNo  string    `json:"journal_no"`
	TransDate  time.Time `json:"trans_date"`
	Status     uint64    `json:"status"`
	ReasonId   uint64    `json:"reason_id"`
	CompanyID  uint64    `json:"company_id"`
	BranchId   uint64    `json:"branch_id"`
	OrderRefId uint64    `json:"order_ref_id"`
	CreatedAt  uint64    `json:"created_at"`
}
type IssueDetail struct {
	IssueId   uint64  `json:"issue_id"`
	ProductId uint64  `json:"product_id"`
	Qty       float64 `json:"qty"`
	AvlQty    float64 `json:"avl_qty"`
	SalePrice float64 `json:"sale_price"`
	OriginQty float64 `json:"origin_qty"`
	Status    uint64  `json:"status"`
	CreatedAt uint64  `json:"created_at"`
}

func (db *posCloseRepository) createIssue(order_id uint64, order_line_data []entity.PosSaleLine, company_id uint64, branch_id uint64, result chan<- bool) {
	var result_issue bool = false
	if order_id > 0 {
		print("issue function")
		var issue_header = IssueHeader{
			JournalNo:  db.GetIssueNo(company_id, branch_id),
			TransDate:  time.Now().Local(),
			Status:     2,
			ReasonId:   1,
			CompanyID:  company_id,
			BranchId:   branch_id,
			OrderRefId: order_id,
			CreatedAt:  uint64(time.Now().Unix()),
		}
		res_issue := db.connect.Table("journal_issue").Create(&issue_header)
		if res_issue.RowsAffected > 0 {
			if len(order_line_data) > 0 {
				for _, v := range order_line_data {
					var issue_line = IssueDetail{
						IssueId:   issue_header.Id,
						ProductId: v.ProductId,
						Qty:       v.Qty,
						AvlQty:    v.Qty,
						SalePrice: v.Price,
						OriginQty: v.Qty,
						Status:    1,
						CreatedAt: uint64(time.Now().Unix()),
					}

					res_issue_line := db.connect.Table("journal_issue_line").Create(&issue_line)
					if res_issue_line.Error != nil {
						result_issue = false
					}
					result_issue = true
				}
			}

		}

	}
	result <- result_issue
}

func (db *posCloseRepository) GetIssueNo(company_id uint64, branch_id uint64) string {
	var max_journal_no MaxPayJournalNo
	var pre string = "IS-"
	var prefix string = ""
	var cnum string = ""
	var cnum3 int64 = 0
	current_date := time.Now().Local()
	row := db.connect.Table("journal_issue").Where("company_id=?", company_id).Where("branch_id=?", branch_id).Where("date(trans_date)=?", current_date.Format("2006-01-02")).Select("id,journal_no").Last(&max_journal_no)
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
			cnum3 = cnumx + 1
		}

		var strlen int = len(cnum)
		var clen int = len(strconv.Itoa(int(cnum3)))
		var loop int = strlen - clen

		for i := 0; i <= loop-1; i++ {
			prefix = prefix + "0"
		}
		prefix = prefix + strconv.Itoa(int(cnum3))
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
		prefix = pre + full_year[2:len(full_year)] + full_month + full_day + "-" + "0001"
	}

	return prefix
}
func (db *posCloseRepository) GetPaymentNo(company_id uint64, branch_id uint64) string {
	var max_journal_no MaxPayJournalNo
	var pre string = "AR-"
	var prefix string = ""
	var cnum string = ""
	var cnum3 int64 = 0
	current_date := time.Now().Local()
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
			cnum3 = cnumx + 1
		}

		var strlen int = len(cnum)
		var clen int = len(strconv.Itoa(int(cnum3)))
		var loop int = strlen - clen

		for i := 0; i <= loop-1; i++ {
			prefix = prefix + "0"
		}
		prefix = prefix + strconv.Itoa(int(cnum3))
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
		prefix = pre + full_year[2:len(full_year)] + full_month + full_day + "-" + "0001"
	}

	return prefix
}

type MaxOrderlNo struct {
	Id      uint64 `json:"id"`
	OrderNo string `json:"order_no"`
}

func (db *posCloseRepository) GetOrderNo(company_id uint64, branch_id uint64) string {
	var max_journal_no MaxOrderlNo
	var pre string = "SO-"
	var prefix string = ""
	var cnum string = ""
	var cnum3 int64 = 0
	current_date := time.Now().Local()
	row := db.connect.Table("orders").Where("company_id=? and sale_channel_id=2", company_id).Where("branch_id=?", branch_id).Where("date(order_date)=?", current_date.Format("2006-01-02")).Select("id,order_no").Last(&max_journal_no)
	if row.Error != nil {
		//panic(row.Error)
		// if(row.Error.Error() == "record not found"){

		// }
		// return "error na ja"
	}

	if max_journal_no.OrderNo != "" {
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
		cnum = max_journal_no.OrderNo[10:len(max_journal_no.OrderNo)]
		// cnum = "000"
		if cnumx, err := strconv.ParseInt(cnum, 10, 64); err != nil {
			panic(err)
		} else {
			cnum3 = cnumx + 1
		}

		var strlen int = len(cnum)
		var clen int = len(strconv.Itoa(int(cnum3)))
		var loop int = strlen - clen

		for i := 0; i <= loop-1; i++ {
			prefix = prefix + "0"
		}
		prefix = prefix + strconv.Itoa(int(cnum3))
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
		prefix = pre + full_year[2:len(full_year)] + full_month + full_day + "-" + "0001"
	}

	return prefix
}

func NewPosCloseRepository(db *gorm.DB) PosCloseRepository {
	return &posCloseRepository{connect: db}
}
