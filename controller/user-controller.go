package controller

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"tarlek.com/icesystem/helper"
	"tarlek.com/icesystem/service"
)

type UserController interface {
	Profile(ctx *gin.Context)

}

type userController struct {
	userService service.UserService
	jwtService  service.JWTService
}


func NewUserController(userService service.UserService, jwtService service.JWTService) UserController {
	return &userController{userService: userService, jwtService: jwtService}
}

func (u *userController) Profile(ctx *gin.Context) {
	ad_user := ctx.Param("id")
	if ad_user == "" {
		res := "No param"
		ctx.JSON(http.StatusBadRequest, res)
	}

	authHeader := ctx.GetHeader("Authorization")
	token, err := u.jwtService.ValidateToken(authHeader)
	if err != nil {
		panic(err.Error())
	}
	claims := token.Claims.(jwt.MapClaims)
	user := u.userService.Profile(fmt.Sprintf("%v", claims["user_id"]), ad_user)
	res := helper.BuildResponse(true, "OK", user)
	ctx.JSON(http.StatusOK, res)
}
