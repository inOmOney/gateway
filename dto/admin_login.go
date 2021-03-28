package dto

import (
	"gateway/public"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminLoginInput struct {
	UserName string `form:"username" comment:"用户名" validate:"required,validUsername"`
	PassWord string `form:"password" comment:"密码" validate:"required"`
}

func (param *AdminLoginInput) BindValidAdminLogin(c *gin.Context) error {
	return public.DefaultGetValidParams(c, param)
}

type AdminLoginOutput struct {
	Token string `json:"token" comment:"token" example:"token" validate:""` //token
}

type AdminSessionInfo struct{
	Id int `json:"id"`
	Username string	`json:"username"`
	LoginTime time.Time	`json:"login_time"`
}