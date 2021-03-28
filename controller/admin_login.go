package controller

import (
	"encoding/json"
	"gateway/dao"
	"gateway/dto"
	"gateway/lib"

	"gateway/middleware"
	"gateway/public"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"time"
)

type AdminLoginController struct{}

func AdminLoginRegister(group *gin.RouterGroup) {
	adminLogin := &AdminLoginController{}
	group.POST("/login", adminLogin.AdminLogin)
	group.GET("/logout", adminLogin.AdminLogout)
}

func (adminLogin *AdminLoginController) AdminLogin(c *gin.Context) {
	params := &dto.AdminLoginInput{}
	if err := params.BindValidAdminLogin(c); err != nil {
		middleware.ResponseError(c, middleware.ValidErrorCode, err)
		return
	}
	admin:=&dao.Admin{}

	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, 2001, err)	//获取连接失败
		return
	}
	admin, err = admin.LoginCheck(c, db, params)
	if err!=nil {
		middleware.ResponseError(c, 2002, err)	// 没有登陆
		return
	}
	adminSession := &dto.AdminSessionInfo{
		Id:        admin.Id,
		Username:  admin.UserName,
		LoginTime: time.Now(),
	}
	adminJson, err := json.Marshal(adminSession)
	if err!=nil {
		middleware.ResponseError(c, 2003, err)  // 序列化出错
		return
	}
	session := sessions.Default(c)
	session.Set(public.AdminSessionKey, string(adminJson))
	session.Save()

	adminLoginOutput := &dto.AdminLoginOutput{Token: admin.Salt}
	middleware.ResponseSuccess(c, adminLoginOutput)
}

func (adminLogin *AdminLoginController) AdminLogout(c *gin.Context){
	session := sessions.Default(c)
	session.Delete(public.AdminSessionKey)
	session.Save()
	middleware.ResponseSuccess(c,"")
}

