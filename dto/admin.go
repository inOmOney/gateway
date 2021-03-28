package dto

import (
	"gateway/public"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminInfoOutPut struct{
	Id int `json:"id"`
	UserName string `json:"user_name"`
	Avatar string`json:"avatar"`
	LoginTime time.Time`json:"login_time"`
	Introduction string`json:"introduction"`
	Roles []string`json:"roles"`
}

type ChangePwdInput struct{
	Password string `form:"password" validator:"required"`
}

func (pwd *ChangePwdInput) BindValidChangePwd(c *gin.Context) error{
	return public.DefaultGetValidParams(c, pwd)
}