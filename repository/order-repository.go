package repository

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"gorm.io/gorm"
	"tarlek.com/icesystem/entity"
)

type OrderRepository interface {
	CreateOrder(order entity.OrderCreate) entity.OrderCreate
	CloseOrder(order entity.OrderClose) int
	GetLastNo(company_id uint64, branch_id uint64, route_id uint64, route_code string) string
	CustomerOrder(customerOrder entity.OrderCustomer) entity.OrderList
}

type orderRepository struct {
	connect *gorm.DB
}

// CustomerOrder implements OrderRepository
func (db *orderRepository) CustomerOrder(customerOrder entity.OrderCustomer) entity.OrderList {
	var orderlist entity.OrderList
	current_date := time.Now().Local()
	if customerOrder.CarId > 0 {
		if customerOrder.SearchCustomer > 0 {
			res := db.connect.Table("query_api_order_daily_summary_new").Where("car_ref_id=? and date(order_date)=? and customer_id=? status=1 ", customerOrder.CarId, current_date.Format("2006-01-02"), customerOrder.SearchCustomer).Scan(&orderlist)
			if res.Error != nil {
				return orderlist
			}
		} else {
			res := db.connect.Table("query_api_order_daily_summary_new").Where("car_ref_id=? and date(order_date)=? and status=1 ", customerOrder.CarId, current_date.Format("2006-01-02")).Scan(&orderlist)
			if res.Error != nil {
				return orderlist
			}
		}

	}
	return orderlist
}

func (db *orderRepository) CreateOrder(order entity.OrderCreate) entity.OrderCreate {
	//order.RunNo = db.GetLastNo(order.CompanyId, order.BranchId, order.RouteId, order.RouteCode)
	var data []entity.OrderLineStruct = order.DataList
	var order_master entity.OrderMaster
	var shift_number = 0
	var order_no_new = db.GetLastNo(order.CompanyId, order.BranchId, order.RouteId, order.RouteCode)
	//var order_total_amt float64 = 0
	convert_shift, err := strconv.ParseUint(order.LoginShift, 10, 32)
	if err == nil {
		shift_number = int(convert_shift)
	}

	//order_master.OrderNo = order.OrderNo
	order_master.OrderNo = order_no_new
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
	order_master.OrderShift = int64(shift_number)
	order_master.DiscountAmt = order.Discount
	order_master.PaymentMethodId = int64(order.PaymentTypeId)
	order_master.OrderTotalAmt = order.OrderTotalAmount

	//tx := db.connect.Begin()

	res := db.connect.Table("orders").Create(&order_master) // save and return id
	if res.RowsAffected > 0 {
		//print(res.RowsAffected)
		//print(order_master.Id)

		for i := 0; i <= len(data)-1; i++ {
			//print(data[0].ProductId)
			if data[i].Qty <= 0 {
				continue
			}
			line_total := (data[i].Qty * data[i].Price)
			//order_total_amt += line_total

			// var line_price float64 = 0
			// var line_total_price float64 = 0
			var is_free int = 0

			if order.PaymentTypeId == 3 {
				is_free = 1
			}

			var orderdetail entity.OrderDetail
			orderdetail.OrderId = order_master.Id
			orderdetail.CustomerId = int64(order.CustomerId)
			orderdetail.ProductId = int64(data[i].ProductId)
			orderdetail.Qty = data[i].Qty
			orderdetail.Price = data[i].Price
			orderdetail.LineTotal = line_total
			orderdetail.PriceGroupId = int64(data[i].PriceGroupId)
			orderdetail.Status = 1
			orderdetail.SalePaymentMethodId = int64(order.PaymentTypeId)
			orderdetail.IssueRefId = int64(order.IssueId)
			orderdetail.IsFree = int64(is_free)

			res2 := db.connect.Table("order_line").Create(&orderdetail)
			if res2.RowsAffected > 0 {
				if order.PaymentTypeId != 3 {
					db.AddPayment(uint64(order_master.Id), order.CustomerId, orderdetail.LineTotal, uint64(order.CompanyId), order.BranchId, uint64(orderdetail.SalePaymentMethodId), order.UserId, order.Image)
				}
				db.UpdateStock(order.RouteId, uint64(data[i].ProductId), data[i].Qty)
			}
		}
		// if order_total_amt > 0 {
		// 	db.connect.Table("orders").Where("id = ?", order_master.Id).Update("order_total_amt", order_total_amt)
		// }

      /* if(order.SaleTypeError != ""){ // send notification when has sale type error
		print("send notification when has sale type error")
		params := url.Values{}
		params.Add("route_id", strconv.Itoa(int(order.RouteId)))
		params.Add("company_id", strconv.Itoa(int(order.CompanyId)))
		params.Add("branch_id", strconv.Itoa(int(order.BranchId)))
		params.Add("user_id", strconv.Itoa(int(order.UserId)))
		params.Add("message", order.SaleTypeError)
		params.Add("order_no", order.OrderNo)
		params.Add("customer_name", order.CustomerName)
		params.Add("total_amount",fmt.Sprintf("%f", order.OrderTotalAmount))

		resp, err := http.PostForm("http://141.98.19.240/icesystem/frontend/web/api/order/createnotifyerrorsaletype", params) // NKY
		if err != nil {
			//panic("api error")
		}

		defer resp.Body.Close()
	  } */

	}

    if(order.SaleTypeError != ""){ // send notification when has sale type error
		print("send notification when has sale type error")
		params := url.Values{}
		params.Add("route_id", strconv.Itoa(int(order.RouteId)))
		params.Add("company_id", strconv.Itoa(int(order.CompanyId)))
		params.Add("branch_id", strconv.Itoa(int(order.BranchId)))
		params.Add("user_id", strconv.Itoa(int(order.UserId)))
		params.Add("message", order.SaleTypeError)
		params.Add("order_no", order.OrderNo)
		params.Add("customer_name", order.CustomerName)
		params.Add("total_amount",fmt.Sprintf("%f", order.OrderTotalAmount))

		resp, err := http.PostForm("http://141.98.19.240/icesystem/frontend/web/api/order/createnotifyerrorsaletype", params) // NKY
		if err != nil {
			//panic("api error")
		}

		defer resp.Body.Close()
	} 

	// tx.Commit()

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

func (db *orderRepository) AddPayment(order_id uint64, customer_id uint64, amount float64, company_id uint64, branch_id uint64, payment_type_id uint64, user_id uint64, image string) {
	var findone uint64 = 0
	var pay_amount float64 = 0
	current_date := time.Now().Local()

	var new_file = ""

	if image != "" {
		var z = 0
		//var ostypename = "http://192.168.60.191/icesystem/backend/web/uploads/"

		// The path to the image you want to upload
		imagePath := "./uploads/"

		//fmt.Println(i, s)
		z += 1
		y := fmt.Sprintf("%v", z)

		var b64 = image
		dc, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			panic(err)
		}
		new_file = strconv.FormatInt(time.Now().Unix(), 20) + y + ".jpg"

		//f, err := os.Create("http://172.16.0.29/cicsupport/backend/web/uploads/myfilename.jpg")
		f, err := os.OpenFile(imagePath+new_file, os.O_WRONLY|os.O_CREATE, 0777) //administrator@172.16.0.240/uploads
		if err != nil {
			panic(err)
		}
		//ostype := runtime.GOOS

		//log.Print(ostype)

		// f, err := os.OpenFile(ostypename+new_file, os.O_WRONLY|os.O_CREATE, 0777) //administrator@172.16.0.240/uploads
		// if err != nil {
		// 	panic(err)
		// }

		if _, err := f.Write(dc); err != nil {
			panic(err)
		}

		defer f.Close()

		sendFileToPHPServer(new_file)

	}

	recid := db.connect.Table("payment_receive").Where("customer_id = ?", customer_id).Where("date(trans_date) = ?", current_date.Format("2006-01-02")).Select("id").Take(&findone)
	if recid != nil {
		if payment_type_id == 1 {
			pay_amount = amount
		}
		println("not error but not found record")
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
				SlipDoc:    new_file,
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

func sendFileToPHPServer(filename string) {
	file, err := os.Open(filepath.Join("./uploads/", filename))
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Prepare a multipart form file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("image", filepath.Base(filename))
	if err != nil {
		fmt.Println("Error creating form file:", err)
		return
	}

	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Println("Error copying file:", err)
		return
	}

	writer.Close()

	// Send the file to the PHP server
	req, err := http.NewRequest("POST", "http://141.98.19.240/icesystem/backend/web/index.php?r=site/uploadfromgo", body)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("PHP server responded with status:", resp.Status)
		return
	}

	fmt.Println("File sent to PHP server successfully")
}


type MaxOrderNo struct {
	Id      uint64 `json:"id"`
	OrderNo string `json:"order_no"`
}

type MaxOrderNoNew struct {
	Id     uint64 `json:"id"`
	LastNo string `json:"last_no"`
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

func (db *orderRepository) GetLastNoNew(company_id uint64, branch_id uint64, route_id uint64, route_code string) string {
	var max_order_no MaxOrderNoNew
	var pre string = "CO-" + route_code
	var prefix string = ""
	var cnum string = ""
	// var cnum2 int64 = 8
	var cnum3 int64 = 0
	//current_date := time.Now().Local()

	// row := db.connect.Table("orders").Where("company_id=?", company_id).Where("branch_id=?", branch_id).Where("order_channel_id=?", route_id).Where("sale_from_mobile=1").Where("order_no LIKE ?", "CO%").Select("max(order_no)").Row()
	// err := row.Scan(&max_order_no)
	// if err != nil {
	// 	return "error na ja"
	// }
	row := db.connect.Table("sequence_order_trans").Where("route_id=?", route_id).Where("order_type_id=1").Last(&max_order_no)
	if row.Error != nil {
		//return "error na ja"
	}
	// if row.Error != nil {
	// 	print("error na")
	// }

	if max_order_no.LastNo != "" {
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
		cnum = max_order_no.LastNo[15:len(max_order_no.LastNo)]
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

		// update new order no

		res_update_lastno := db.connect.Table("sequence_order_trans").Where("id=?", max_order_no.Id).Update("last_no", (prefix))
		if res_update_lastno.Error == nil {
			print("update last no ok")
			// print(selectedData.ProductId)
		}

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

		/// create new last no
		res_save_new_last_no := db.connect.Table("sequence_order_trans").Create(map[string]interface{}{"order_type_id": 1, "route_id": route_id, "last_no": prefix})
		if res_save_new_last_no.Error == nil {
			print("create new last no")
		}
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

// CloseOrder implements OrderRepository
type OrderStockQty struct {
	ProductId uint64  `json:"product_id"`
	AvlQty    float64 `json:"avl_qty"`
}

type StockTrans struct {
	JournalNo      string    `json:"journal_no"`
	TransDate      time.Time `json:"trans_date"`
	ProductId      uint64    `json:"product_id"`
	Qty            float64   `json:"qty"`
	WarehouseId    uint64    `json:"warehouse_id"`
	StockType      uint64    `json:"stock_type"`
	ActivityTypeId uint64    `json:"activity_type_id"`
	Company_id     uint64    `json:"company_id"`
	BranchId       uint64    `json:"branch_id"`
	CreatedBy      uint64    `json:"created_by"`
	TransRefId     uint64    `json:"trans_ref_id"`
	CreatedAt      uint64    `json:"created_at"`
}

func (db *orderRepository) CloseOrder(order entity.OrderClose) int {
	var resData int = 0
	var orderStockQty []OrderStockQty
	var res_update_sum bool = false
	var res_update_boot_sum bool = false

	current_date := time.Now().Local()
	res := db.connect.Table("order_stock").Where("route_id=? and date(trans_date)=?", order.RouteId, current_date.Format("2006-01-02")).Select("product_id,avl_qty").Scan(&orderStockQty)
	if res.Error != nil {
		panic(res.Error.Error())
	}

	defaultWarehouse := db.getDefaultWh(int64(order.CompanyId), int64(order.BranchId))

	if order.IsReturnStock == 1 {
		var update_count int = 0
		for i := 0; i <= len(orderStockQty)-1; i++ {
			// if orderStockQty[i].AvlQty <= 0 {
			// 	continue
			// }
			var stockTrans StockTrans
			stockTrans.JournalNo = db.GetReturnLastNo(order.CompanyId, order.BranchId)
			stockTrans.TransDate = time.Now().Local()
			stockTrans.ProductId = orderStockQty[i].ProductId
			stockTrans.Qty = orderStockQty[i].AvlQty
			stockTrans.WarehouseId = uint64(defaultWarehouse)
			stockTrans.StockType = 1
			stockTrans.ActivityTypeId = 7
			stockTrans.Company_id = order.CompanyId
			stockTrans.BranchId = order.BranchId
			stockTrans.TransRefId = order.RouteId
			stockTrans.CreatedBy = order.UserId
			stockTrans.CreatedAt = uint64(time.Now().Unix())

			if orderStockQty[i].AvlQty > 0 {
				trans := db.connect.Table("stock_trans").Create(&stockTrans)
				if trans.RowsAffected > 0 {
					if order.IsReturnStock == 1 {
						res_update_sum = db.updateSummary(orderStockQty[i].ProductId, uint64(defaultWarehouse), orderStockQty[i].AvlQty, order.CompanyId, order.BranchId)
						if res_update_sum == true {
							update_count += 1
						}
					} else {
						res_update_boot_sum = db.updateBootSummary(orderStockQty[i].ProductId, order.UserId, order.RouteId, orderStockQty[i].AvlQty, order.CompanyId, order.BranchId)
						if res_update_boot_sum == true {
							update_count += 1
						}
					}
				}
			} else {
				update_count += 1
			}

		}
		if update_count > 0 {
			// update orders
			res_order_update := db.connect.Table("orders").Where("order_channel_id=? and date(order_date) =? and sale_from_mobile=1", order.RouteId, current_date.Format("2006-01-02")).Updates(map[string]interface{}{"status": 100, "order_shift": 0})
			if res_order_update.RowsAffected > 0 {
				resData += 1
			}
			// update order stock
			res_stock_update := db.connect.Table("order_stock").Where("route_id=? and date(trans_date)=?", order.RouteId, current_date.Format("2006-01-02")).Update("avl_qty", 0)
			if res_stock_update.RowsAffected > 0 {
				resData += 1
			}
		}
	} else {
		var update_count int = 0
		for i := 0; i <= len(orderStockQty)-1; i++ {
			// if orderStockQty[i].AvlQty <= 0 {
			// 	continue
			// }
			// 	var stockTrans StockTrans
			// 	stockTrans.JournalNo = db.GetReturnLastNo(order.CompanyId, order.BranchId)
			// 	stockTrans.TransDate = time.Now().Local()
			// 	stockTrans.ProductId = orderStockQty[i].ProductId
			// 	stockTrans.Qty = orderStockQty[i].AvlQty
			// 	stockTrans.WarehouseId = uint64(defaultWarehouse)
			// 	stockTrans.StockType = 1
			// 	stockTrans.ActivityTypeId = 7
			// 	stockTrans.Company_id = order.CompanyId
			// 	stockTrans.BranchId = order.BranchId
			// 	stockTrans.TransRefId = order.RouteId

			// 	trans := db.connect.Table("stock_trans").Create(&stockTrans)
			// 	if trans.RowsAffected > 0 {

			res_update_boot_sum = db.updateBootSummary(orderStockQty[i].ProductId, order.UserId, order.RouteId, orderStockQty[i].AvlQty, order.CompanyId, order.BranchId)
			if res_update_boot_sum == true {
				update_count += 1
			}
		}
		if update_count > 0 {
			// update orders
			res_order_update := db.connect.Table("orders").Where("order_channel_id=? and date(order_date) =? and sale_from_mobile=1", order.RouteId, current_date.Format("2006-01-02")).Updates(map[string]interface{}{"status": 100, "order_shift": 0})
			if res_order_update.RowsAffected > 0 {
				resData += 1
			}
		}
		// }
	}
	//return defaultWarehouse

	// send line notify

	if resData > 0 {
		// client := resty.New()
		// var result map[string]string
		// json.Unmarshal([]byte(`{
		// 	   "message": "ทดสอบจบขายตัวใหม่",
		// 	 "stickerId": "125",
		// 	   "stickerPackageId": "1"
		// }`), &result)
		// resp, err := client.R().
		// 	SetHeader("Authorization", "Bearer NY1xHWO4Qa6EWGA25AKuQVeHwSwpeTEPpCGE3pYB5qT").
		// 	SetFormData(result).Post("https://notify-api.line.me/api/notify")
		// if err != nil {
		// 	log.Fatalf("ERROR LINE Notify API: %s", err)
		// }
		// println(resp.StatusCode())
		params := url.Values{}
		params.Add("route_id", strconv.Itoa(int(order.RouteId)))
		params.Add("company_id", strconv.Itoa(int(order.CompanyId)))
		params.Add("branch_id", strconv.Itoa(int(order.BranchId)))
		params.Add("user_id", strconv.Itoa(int(order.UserId)))

		resp, err := http.PostForm("http://141.98.19.240/icesystem/frontend/web/api/order/createnotifyclose", params) // NKY
		//resp, err := http.PostForm("http://103.253.73.108/icesystem/frontend/web/api/order/createnotifyclose", params) // NKY
		//resp, err := http.PostForm("http://141.98.16.4/icesystem/frontend/web/api/order/createnotifyclose", params) // BKT
		if err != nil {
			panic("api error")
		}

		defer resp.Body.Close()
	}

	return resData
}

func (db *orderRepository) getDefaultWh(company_id int64, branch_id int64) int {
	default_wh := 12
	res := db.connect.Table("warehouse").Where("is_reprocess = 1 and company_id = ? and branch_id = ?", company_id, branch_id).Select("id").Scan(&default_wh)
	if res.Error != nil {

	}
	return default_wh
}

type StockSumData struct {
	Id  uint64  `json:"id"`
	Qty float64 `json:"qty"`
}

type StockSumDataNew struct {
	WarehouseId uint64  `json:"warehouse_id"`
	ProductId   uint64  `json:"product_id"`
	Qty         float64 `json:"qty"`
	CompanyId   uint64  `json:"company_id"`
	BranchId    uint64  `json:"branch_id"`
}
type SaleRouteDailyClose struct {
	TransDate  uint64  `json:"trans_date"`
	ProductId  uint64  `json:"product_id"`
	Qty        float64 `json:"qty"`
	CompanyId  uint64  `json:"company_id"`
	BranchId   uint64  `json:"branch_id"`
	RouteId    uint64  `json:"route_id"`
	OrderShift uint64  `json:"order_shift"`
	CreatedBy  uint64  `json:"created_by"`
}
type SaleRouteDailyClose2 struct {
	TransDate  uint64  `json:"trans_date"`
	ProductId  uint64  `json:"product_id"`
	Qty        float64 `json:"qty"`
	CompanyId  uint64  `json:"company_id"`
	BranchId   uint64  `json:"branch_id"`
	RouteId    uint64  `json:"route_id"`
	OrderShift uint64  `json:"order_shift"`
}

func (db *orderRepository) updateSummary(product_id uint64, warehouse_id uint64, return_qty float64, company_id uint64, branch_id uint64) bool {
	var old_qty StockSumData
	var new_stock StockSumDataNew
	var is_update bool = false

	res := db.connect.Table("stock_sum").Select("id,qty").Where("warehouse_id=? and product_id =? and company_id=? and branch_id=?", warehouse_id, product_id, company_id, branch_id).Scan(&old_qty)
	if res.Error != nil {

	}
	if old_qty.Id > 0 {
		var onhand_qty float64 = 0
		var final_qty float64 = 0
		if old_qty.Qty >= 0 {
			onhand_qty = old_qty.Qty
		}
		final_qty = (return_qty + onhand_qty)
		resupdate := db.connect.Table("stock_sum").Where("id=?", old_qty.Id).Update("qty", final_qty)
		if resupdate.RowsAffected > 0 {
			is_update = true
		}
	} else {
		new_stock.WarehouseId = warehouse_id
		new_stock.ProductId = product_id
		new_stock.Qty = return_qty
		new_stock.CompanyId = company_id
		new_stock.BranchId = branch_id

		resnew := db.connect.Table("stock_sum").Create(&new_stock)
		if resnew.RowsAffected > 0 {
			is_update = true
		}
	}

	return is_update
}

func (db *orderRepository) updateBootSummary(product_id uint64, user_id uint64, route_id uint64, qty float64, company_id uint64, branch_id uint64) bool {
	var resData bool = false
	var saleDailyClose SaleRouteDailyClose2
	var findData StockSumData
	var orderShift uint64 = 0
	current_date := time.Now().Local()

	res := db.connect.Table("sale_route_daily_close").Select("id,qty").Where("route_id=? and product_id =? and order_shift=? and date(trans_date)=?", route_id, product_id, orderShift, current_date.Format("2006-01-02")).Scan(&findData)
	if res.Error != nil {

	}
	if findData.Id > 0 {
		resupdate := db.connect.Table("sale_route_daily_close").Where("id=?", findData.Id).Updates(map[string]interface{}{"qty": qty, "trans_date": time.Now().Local()})
		if resupdate.RowsAffected > 0 {
			resData = true
		}
	} else {
		saleDailyClose.RouteId = route_id
		saleDailyClose.ProductId = product_id
		saleDailyClose.Qty = qty
		saleDailyClose.CompanyId = company_id
		saleDailyClose.BranchId = branch_id
		saleDailyClose.OrderShift = orderShift

		resnew := db.connect.Table("sale_route_daily_close").Create(&saleDailyClose)
		if resnew.RowsAffected > 0 {
			resData = true
		}
	}

	return resData
}

type SeqModel struct {
	Id        int64  `json:"id"`
	Prefix    string `json:"prefix"`
	Symbol    string `json:"symbol"`
	UseYear   int64  `json:"use_year"`
	UseMonth  int64  `json:"use_month"`
	UseDay    int64  `json:"use_day"`
	MaximumNo int64  `json:"maximumn"`
}

func (db *orderRepository) GetReturnLastNo(company_id uint64, branch_id uint64) string {
	var max_journal_no MaxPayJournalNo
	var pre string = ""
	var prefix string = ""
	var cnum string = ""
	var seq_data SeqModel
	// var cnum2 int64 = 8
	var cnum3 int64 = 0
	// var prefix2 strings.Builder
	current_date := time.Now().Local()

	row_seq := db.connect.Table("sequence").Where("company_id=?", company_id).Where("branch_id=?", branch_id).Where("module_id = 7").Select("id,prefix,symbol,use_year,use_month,use_day,maximum").Scan(&seq_data)
	if row_seq.Error != nil {
		return "error na ja"
	}
	row := db.connect.Table("stock_trans").Where("company_id=?", company_id).Where("branch_id=?", branch_id).Where("date(trans_date)=?", current_date.Format("2006-01-02")).Where("activity_type_id= 7").Select("id,journal_no").Last(&max_journal_no)
	if row.Error != nil {
		//panic(row.Error)
		// if(row.Error.Error() == "record not found"){

		// }
		// return "error na ja"
	}

	if max_journal_no.JournalNo != "" {
		//max_journal_no = "CO-VP31-230314-0009"
		pre = seq_data.Prefix + seq_data.Symbol
		var full_year string = strconv.Itoa(time.Now().Year())
		var full_month string = strconv.Itoa(int(time.Now().Month()))
		if len(full_month) == 1 {
			full_month = "0" + full_month
		}
		var full_day string = strconv.Itoa(time.Now().Day())
		if len(full_day) == 1 {
			full_day = "0" + full_day
		}
		if seq_data.UseYear == 1 {
			pre = pre + full_year[2:len(full_year)]
		}
		if seq_data.UseMonth == 1 {
			pre = pre + full_month
		}
		if seq_data.UseDay == 1 {
			pre = pre + full_day
		}
		prefix = pre + "-"
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
		pre = seq_data.Prefix + seq_data.Symbol
		var full_year string = strconv.Itoa(time.Now().Year())
		var full_month string = strconv.Itoa(int(time.Now().Month()))
		if len(full_month) == 1 {
			full_month = "0" + full_month
		}
		var full_day string = strconv.Itoa(time.Now().Day())
		if len(full_day) == 1 {
			full_day = "0" + full_day
		}

		if seq_data.UseYear == 1 {
			pre = pre + full_year[2:len(full_year)]
		}
		if seq_data.UseMonth == 1 {
			pre = pre + full_month
		}
		if seq_data.UseDay == 1 {
			pre = pre + full_day
		}

		prefix = pre + "-" + "0001"
	}

	return prefix
}

func (db *UserConnect) UpdatePhoto(photo entity.SlipDoc, payment_id uint64) bool {
	//	 var photo []
	//var id int

	var z = 0
	var ostypename = ""
	var new_file = ""
	for _, s := range photo.Image {
		//fmt.Println(i, s)
		z += 1
		y := fmt.Sprintf("%v", z)

		var b64 = s
		dc, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			panic(err)
		}
		new_file = strconv.FormatInt(time.Now().Unix(), 20) + y + ".jpg"

		//f, err := os.Create("http://172.16.0.29/cicsupport/backend/web/uploads/myfilename.jpg")

		ostype := runtime.GOOS

		log.Print(ostype)

		f, err := os.OpenFile(ostypename+new_file, os.O_WRONLY|os.O_CREATE, 0777) //administrator@172.16.0.240/uploads
		if err != nil {
			panic(err)
		}

		defer f.Close()

		if _, err := f.Write(dc); err != nil {
			panic(err)
		}

		result := db.connect.Table("payment_receive").Where("id = ?", payment_id).Updates(map[string]interface{}{"slip_doc": new_file})
		// result := db.connect.Table("person").Updates(map[string]interface{}{"photo": new_file})
		if result.RowsAffected > 0 {
			return true
		} else {
			return false
		}

	}
	return true
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{connect: db}
}
