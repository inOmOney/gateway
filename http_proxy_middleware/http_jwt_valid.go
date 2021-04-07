package http_proxy_middleware

import (
	"errors"
	"gateway/dao"
	"gateway/middleware"
	"gateway/public"
	"github.com/gin-gonic/gin"
	"strings"
)

// header: Authorization:[Bear jwtStr]
//
func HttpAppJwtValid() gin.HandlerFunc {
	return func(c *gin.Context) {
		serverInterface, _ := c.Get("service")

		serviceDetail := serverInterface.(*dao.ServiceDetail)
		if serviceDetail.AccessControl.OpenAuth != 1 {	//没有开启权限访问 直接下一步
			c.Next()
		} else {	// 从头中获取jwt 并解析出对应的appID
			au := c.GetHeader("Authorization")
			jwtStr := strings.Replace(au, "Bearer ", "", 1)
			claim, err := public.JwtDecode(jwtStr)
			if err != nil {
				middleware.ResponseError(c, 2002, errors.New("无效APPID"))
				c.Abort()
				return
			}
			appMap := dao.AppManagerHandler.AppMap
			isMatch := false
			for appID := range appMap {
				if appID == claim.Issuer {
					isMatch = true
					c.Set("app", appID)
				}
			}
			if !isMatch {
				middleware.ResponseError(c, 2003, errors.New("没有对应的服务"))
				c.Abort()
			}
		}

	}
}
