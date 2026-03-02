package mgin

import (
	"github.com/gin-gonic/gin"
	"github.com/jettjia/go-pkg/pkg/xerror"
	"github.com/jettjia/go-pkg/pkg/xresponse"
)

// Universal generic middleware part
// for example the processing of user login information
func Universal() gin.HandlerFunc {
	return func(c *gin.Context) {

		// parse the information in the extension section of the header
		userExp := c.GetHeader("User-Exp")

		if userExp != "" {
			loginRspExp, err := Base64ToLoginRspExp(userExp)
			if err != nil {
				err := xerror.NewError(xerror.UnauthorizedErr, "Unauthorized", nil)
				xresponse.RspErr(c, err)
				c.Abort()
				return
			}

			c.Set("user_id", loginRspExp.UserId)
			c.Set("nick_name", loginRspExp.Nickname)
			c.Set("login_ip", loginRspExp.LoginIp)
		}
	}
}
