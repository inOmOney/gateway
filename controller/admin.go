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
)

type AdminController struct{}

func AdminRegister(group *gin.RouterGroup) {
	admin := &AdminController{}
	group.GET("/admin_info", admin.getAdminInfo)
	group.POST("/change_pwd", admin.changePwd)
}

func (admin *AdminController) getAdminInfo(c *gin.Context) {

	session := sessions.Default(c)
	adminJson := session.Get(public.AdminSessionKey)
	adminSession := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(adminJson.(string)), admin); err != nil {
		middleware.ResponseError(c, middleware.InternalErrorCode, err)
		return
	}
	out := &dto.AdminInfoOutPut{
		Id:           adminSession.Id,
		UserName:     adminSession.Username,
		LoginTime:    adminSession.LoginTime,
		Avatar:       "https://raw.githubusercontent.com/inOmOney/imgbed/master/2021-02/developer-github.gif",
		Introduction: "我是超级管理员",
		Roles:        []string{"admin"},
	}

	middleware.ResponseSuccess(c, out)
}

func (controller *AdminController) changePwd(c *gin.Context) {
	params := &dto.ChangePwdInput{}
	if err := params.BindValidChangePwd(c); err != nil {
		middleware.ResponseError(c, 2000, err)
		return
	}
	session := sessions.Default(c)
	adminJson := session.Get(public.AdminSessionKey)
	adminSession := &dto.AdminSessionInfo{}
	if err := json.Unmarshal([]byte(adminJson.(string)), adminSession); err != nil {
		middleware.ResponseError(c, middleware.InternalErrorCode, err)
		return
	}
	admin := &dao.Admin{}
	db, err := lib.GetDBPool("default")
	if err != nil {
		middleware.ResponseError(c, middleware.SystemBusy, err)
		return
	}
	admin, err = admin.Find(c, db, &dao.Admin{UserName: adminSession.Username})
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	//新加盐密码
	saltPwd := public.PasswordAddSalt(params.Password, admin.Salt)
	admin.Password = saltPwd
	if err=admin.Save(db, admin);err!=nil{
		middleware.ResponseError(c, 2003, err)
		return
	}
	// 更改密码后注销
	adminLogin := &AdminLoginController{}
	adminLogin.AdminLogout(c)

}
