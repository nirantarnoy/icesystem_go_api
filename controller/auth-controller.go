package controller

import (
	"github.com/gin-gonic/gin"
	"tarlek.com/icesystem/dto"
	"tarlek.com/icesystem/helper"
	"tarlek.com/icesystem/service"
	"net/http"
	"strconv"
)

type AuthController interface {
	Login(ctx *gin.Context)
}

type authController struct {
	authService service.AuthService
	jwtService  service.JWTService
}

func NewAuthController(authService service.AuthService, jwtService service.JWTService) AuthController {
	return &authController{authService: authService, jwtService: jwtService}
}

func (c *authController) Login(ctx *gin.Context) {
	var loginDTO dto.LoginDTO
	errDTO := ctx.ShouldBind(&loginDTO)
	if errDTO != nil {
		response := helper.BuildErrorResponse("failed to process request", errDTO.Error(), helper.EmptyObj{})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, response)
		return
	}

	authResult := c.authService.VerifyCredential(loginDTO.Username, loginDTO.Password)
	if authResult != false {
		res := c.authService.FindByADUser(loginDTO.Username)
		generatedToken := c.jwtService.GenerateToken(strconv.FormatUint(res.ID, 10))
		res.Token = generatedToken
		response := helper.BuildResponse(true, "OK", res)
		ctx.JSON(http.StatusOK, response)
		return
	}

	res := helper.BuildErrorResponse("please check username or password", "Invalid Login AD", helper.EmptyObj{})
	ctx.AbortWithStatusJSON(http.StatusUnauthorized, res)
}
