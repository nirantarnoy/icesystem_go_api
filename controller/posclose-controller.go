package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"tarlek.com/icesystem/dto"
	"tarlek.com/icesystem/service"
)

type PoscloseController interface {
	PosClose(ctx *gin.Context)
}
type poscloseController struct {
	poscloseSevice service.PosCloseService
}

// PosClose implements PoscloseController
func (poscontl *poscloseController) PosClose(ctx *gin.Context) {
	var poscloseDto dto.PosCloseDataDto
	err := ctx.ShouldBind(&poscloseDto)
	if err != nil {
		fmt.Printf("%v", err)
		panic(err)
	}

	res := poscontl.poscloseSevice.PosClose(poscloseDto)

	ctx.JSON(http.StatusCreated, res)
	// buf := make([]byte, 1024)
	// num, _ := c.Request.Body.Read(buf)
	// reqBody := string(buf[0:num])
	// c.Request.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(reqBody))) // Write body back

	// buf2 := make([]byte, 1024)
	// num2, _ := c.Request.Body.Read(buf2)
	// reqBody2 := string(buf2[0:num2])

	// fmt.Println(reqBody, "==", reqBody2)
	// c.JSON(http.StatusOK, reqBody)

}

func NewPosCloseController(pos_service service.PosCloseService) PoscloseController {
	return &poscloseController{poscloseSevice: pos_service}
}
