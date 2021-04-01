package controller

import (
	"encoding/base64"
	"gateway/dao"
	"gateway/dto"
	"gateway/lib"
	"gateway/middleware"
	"gateway/public"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type OAuthController struct{}

func OAuthRegister(g *gin.RouterGroup) {
	controller := &OAuthController{}
	g.POST("/token", controller.GenericToken)
}

func (controller *OAuthController) GenericToken(c *gin.Context) {

	param := &dto.TokensInput{}
	if err := param.BindValidParam(c); err != nil {
		middleware.ResponseError(c, 2001, err)
		return
	}
	out := &dto.TokensOutput{}
	//拿到头信息中的认证信息
	// 1. base64解码拿到app_id:secret
	// 2. 验证是否真实存在
	// 3. 生成jwt令牌返回
	splits := strings.Split(c.GetHeader("Authorization"), " ")
	if len(splits) != 2 {
		middleware.ResponseError(c, 2001, errors.New("用户名或密码格式错误"))
		return
	}

	appSecret, err := base64.StdEncoding.DecodeString(splits[1])
	if err != nil {
		middleware.ResponseError(c, 2002, err)
		return
	}
	parts := strings.Split(string(appSecret), ":")
	if appInfo, exist := dao.AppManagerHandler.AppMap[parts[0]]; !exist || parts[1] != appInfo.Secret {
		middleware.ResponseError(c, 2002, errors.New("没有对应的app_id"))
		return
	} else {
		claims := &jwt.StandardClaims{
			Issuer:    appInfo.AppID,
			ExpiresAt: time.Now().Add(60 * 60 * time.Second).In(lib.TimeLocation).Unix(),
		}
		token, err := public.JwtEncode(claims)
		if err != nil {
			middleware.ResponseError(c, 2004, err)
			return
		}
		out = &dto.TokensOutput{
			ExpiresIn:   60 * 60,
			TokenType:   "Bearer",
			AccessToken: token,
			Scope:       "read_write",
		}
	}

	middleware.ResponseSuccess(c, out)
}
