package xresponse

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jettjia/igo-pkg/pkg/xerror"
)

type rspErrorData struct {
	Code     int         `json:"code"`     // error code the first three digits standard http error code the middle three digits are server specific codes and the last three digits are custom codes in the service
	Message  string      `json:"message"`  // error messages compatible with old reserved message can be the same as description
	Cause    string      `json:"cause"`    // cause of the error
	Solution string      `json:"solution"` // prompts for actions that meet the internationalization requirements for current errors
	Detail   interface{} `json:"detail"`   // error code expansion information supplementing error information generally it is incorrect stack information etc which needs to be serialized
}

// RspErr an error is returned
func RspErr(c *gin.Context, err error) {
	var myErr *xerror.MyError
	if errors.As(err, &myErr) {
		lang := c.GetHeader("x-lang")
		path := ""
		myErrGet := xerror.GetError(lang, path, myErr)
		c.JSON(
			myErrGet.HttpCode,
			rspErrorData{
				Code:     myErrGet.Code,
				Message:  myErrGet.Message,
				Cause:    myErrGet.Cause,
				Solution: myErrGet.Solution,
				Detail:   myErrGet.Detail,
			},
		)
		return
	}

	c.JSON(
		http.StatusInternalServerError,
		rspErrorData{
			Code:     xerror.InternalServerErr,
			Message:  err.Error(),
			Cause:    err.Error(),
			Solution: "Please contact the API provider",
			Detail:   nil,
		},
	)
	return
}

// RspOk the return operation was successful
func RspOk(c *gin.Context, code int, any interface{}) {
	switch code {
	case http.StatusCreated:
		c.JSON(
			http.StatusCreated,
			any,
		)
	case http.StatusNoContent:
		c.AbortWithStatus(
			http.StatusNoContent,
		)

	default:
		c.JSON(
			http.StatusOK,
			any,
		)
	}
}
