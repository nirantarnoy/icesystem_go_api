package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"tarlek.com/icesystem/dto"
	"tarlek.com/icesystem/service"
)

type OrderController interface {
	CreateOrder(ctx *gin.Context)
	GetLastNo(ctx *gin.Context)
}
type orderController struct {
	OrderService service.OrderService
	jwtService   service.JWTService
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
