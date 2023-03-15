package middleware

import (
	"github.com/gin-gonic/gin"
	"tarlek.com/icesystem/helper"
	"tarlek.com/icesystem/service"
	"net/http"
)

func AuthorizeJWT(service service.JWTService) gin.HandlerFunc {
	return func(context *gin.Context) {
		authHeader := context.GetHeader("Authorization")
		if authHeader == "" {
			response := helper.BuildErrorResponse("Failed to process request", "No Token Found", nil)
			context.AbortWithStatusJSON(http.StatusUnauthorized, response)
		}

		token, err := service.ValidateToken(authHeader)
		if token.Valid {
		} else {
			response := helper.BuildErrorResponse("Token is not valid", err.Error(), nil)
			context.AbortWithStatusJSON(http.StatusUnauthorized, response)
		}
	}
}
