package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"tarlek.com/icesystem/dto"
	"tarlek.com/icesystem/entity"
	"tarlek.com/icesystem/service"
)

type OrderController interface {
	CreateOrder(ctx *gin.Context)
	GetLastNo(ctx *gin.Context)
	CloseOrder(ctx *gin.Context)
	CustomerOrder(ctx *gin.Context)
}
type orderController struct {
	OrderService service.OrderService
	jwtService   service.JWTService
}

type orderListResult struct {
	Status uint64           `json:"status"`
	Data   entity.OrderList `json:"data"`
}

// CustomerOrder implements OrderController
func (db *orderController) CustomerOrder(ctx *gin.Context) {
	var orderListDto dto.OrderList
	var result orderListResult
	errDro := ctx.ShouldBind(&orderListDto)
	if errDro != nil {
		fmt.Printf("%v", errDro)
		//res := "Cannot bind parameters"
		result.Status = 0
		result.Data = entity.OrderList{}
		ctx.JSON(http.StatusBadRequest, result)
	} else {
		if orderListDto.CarId < 1 {
			//res := "Not have data paremeter"
			result.Status = 0
			result.Data = entity.OrderList{}
			ctx.JSON(http.StatusBadRequest, result)
		} else {
			res := db.OrderService.CustomerOrder(orderListDto)
			result.Status = 0
			result.Data = res
			ctx.JSON(http.StatusOK, result)
		}
	}

}

// CloseOrder implements OrderController
func (db *orderController) CloseOrder(ctx *gin.Context) {
	var OrderCloseDto dto.OrderColseDto
	errDto := ctx.ShouldBind(&OrderCloseDto)
	if errDto != nil {
		fmt.Printf("%v", errDto)
		res := "Fail to process"
		ctx.JSON(http.StatusBadRequest, res)
	} else {
		// authHeader := ctx.GetHeader("Authorization")
		// userId := db.getUserIdByToken(authHeader)
		// convertUserId, err := strconv.ParseUint(userId, 10, 64)

		// if err == nil {
		// 	OrderDto.CreatedBy = convertUserId
		// }
		res := db.OrderService.CloseOrder(OrderCloseDto)
		ctx.JSON(http.StatusCreated, res)
	}
}

// GetLastNo implements OrderController
func (db *orderController) GetLastNo(ctx *gin.Context) {
	res := db.OrderService.GetLastNo()
	ctx.JSON(http.StatusOK, res)
}

// CreateOrder implements OrderController
func (db *orderController) CreateOrder(ctx *gin.Context) {
	//fmt.Print(ctx)
	var OrderDto dto.OrderCreateDto
	errDto := ctx.ShouldBind(&OrderDto)
	if errDto != nil {
		fmt.Printf("%v", errDto)
		res := "Fail to process"
		ctx.JSON(http.StatusBadRequest, res)
	} else {
		// authHeader := ctx.GetHeader("Authorization")
		// userId := db.getUserIdByToken(authHeader)
		// convertUserId, err := strconv.ParseUint(userId, 10, 64)

		// if err == nil {
		// 	OrderDto.CreatedBy = convertUserId
		// }
		res := db.OrderService.CreateOrder(OrderDto)
		ctx.JSON(http.StatusCreated, res)
	}

}

func NewOrderController(OrderService service.OrderService, jwtService service.JWTService) OrderController {
	return &orderController{OrderService: OrderService, jwtService: jwtService}
}
