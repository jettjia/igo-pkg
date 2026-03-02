package mgin

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jettjia/igo-pkg/pkg/conf"
	"github.com/jettjia/igo-pkg/pkg/hydra"
	"github.com/jettjia/igo-pkg/pkg/xerror"
	"github.com/jettjia/igo-pkg/pkg/xmiddleware/jwt"
	"github.com/jettjia/igo-pkg/pkg/xresponse"
)

// TokenAuthorization verify the token
func TokenAuthorization(dev bool, jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		if dev {
			return
		}

		auth := c.GetHeader("Authorization")
		token := strings.TrimPrefix(auth, "Bearer ")
		if token == "" {
			err := xerror.NewError(xerror.UnauthorizedErr, "Unauthorized", nil)
			xresponse.RspErr(c, err)
			c.Abort()
			return
		}

		// verify the token
		claims, err := jwt.ParseToken(token, jwtSecret)
		if err != nil {
			err = xerror.NewError(xerror.UnauthorizedErr, "Unauthorized", nil)
			xresponse.RspErr(c, err)
			c.Abort()
			return
		} else if time.Now().Unix() > claims.ExpiresAt {
			err = xerror.NewError(xerror.UnauthorizedErr, "Token expired", nil)
			xresponse.RspErr(c, err)
			c.Abort()
			return
		}
	}
}

// TokenAuthorizationHydra verify the hydra token
func TokenAuthorizationHydra(conf *conf.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if conf.Server.Dev {
			return
		}

		hydraClient := hydra.NewHydraAdmin(conf)

		auth := c.GetHeader("Authorization")
		token := strings.TrimPrefix(auth, "Bearer ")

		if token == "" {
			err := xerror.NewError(xerror.UnauthorizedErr, "Unauthorized", nil)
			xresponse.RspErr(c, err)
			c.Abort()
			return
		}

		flag := hydraClient.Introspect(c, token)
		if !flag {
			err := xerror.NewError(xerror.UnauthorizedErr, "Unauthorized", nil)
			xresponse.RspErr(c, err)
			c.Abort()
			return
		}
	}

}
