package mgin

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"

	"github.com/jettjia/go-pkg/pkg/xerror"
	"github.com/jettjia/go-pkg/pkg/xresponse"
)

func CatchError() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			// catch panic errors e g panic
			if errAny := recover(); errAny != nil {
				xerror.RecoverWithStack()
				switch errAny.(type) {
				case error:
					//unified handling of my sql 1062 errors and sql content conflicts
					var mysqlErr *mysql.MySQLError
					if errors.As(errAny.(error), &mysqlErr) && mysqlErr.Number == 1062 {
						err := xerror.NewError(xerror.ConflictErr, mysqlErr.Message, nil)
						xresponse.RspErr(c, err)
						c.Abort()
						return
					}

					//unified handling of my sql 1054 errors and sql field errors
					if errors.As(errAny.(error), &mysqlErr) && mysqlErr.Number == 1054 {
						err := xerror.NewError(xerror.ForbiddenErr, mysqlErr.Message, nil)
						xresponse.RspErr(c, err)
						c.Abort()
						return
					}

					// uniformly handle my sql 1064 errors sql syntax errors such as multiple quotation marks and missing parentheses
					if errors.As(errAny.(error), &mysqlErr) && mysqlErr.Number == 1064 {
						err := xerror.NewError(xerror.ForbiddenErr, mysqlErr.Message, nil)
						xresponse.RspErr(c, err)
						c.Abort()
						return
					}

				default:
					// handle other errors in a unified manner
					err := xerror.NewError(xerror.InternalServerErr, fmt.Sprintf("%+v", errAny), nil)
					xresponse.RspErr(c, err)
					c.Abort()
					return
				}
			}

			// manually thrown errors such as ierror.New()
			if len(c.Errors) != 0 {
				xerror.RecoverWithStack()
				for _, errAny := range c.Errors {
					switch errAny.Error() {
					case "EOF":
						err := xerror.NewError(xerror.BadRequestErr, errAny.Error(), nil)
						xresponse.RspErr(c, err)
						c.Abort()
						return
					default:
						xresponse.RspErr(c, errAny)
						c.Abort()
						return
					}
				}
			}
		}()
		c.Next()
	}
}
