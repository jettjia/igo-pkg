package xerror

import (
	"fmt"
	"strconv"
)

type error interface {
	Error() string
}

type MyError struct {
	HttpCode int         `json:"http_code"` // http status code
	Code     int         `json:"code"`      // error code the first three digits standard http error code the middle three digits are server specific codes and the last three digits are custom codes in the service
	Message  string      `json:"message"`   // error message
	Cause    string      `json:"cause"`     // cause of the error
	Solution string      `json:"solution"`  // prompts for actions that meet the internationalization requirements for current errors
	Detail   interface{} `json:"detail"`    // error code expansion information supplementing error information generally it is incorrect stack information etc which needs to be serialized
}

func (e *MyError) Error() string {
	return fmt.Sprintf("error code：%d,error cause：%s", e.Code, e.Cause)
}

func NewError(code int, cause string, detail interface{}) *MyError {
	return &MyError{
		Code:   code,
		Cause:  cause,
		Detail: detail,
	}
}

func NewErrorOpt(code int, options ...func(*MyError)) *MyError {
	opts := &MyError{}
	opts.Code = code

	for _, option := range options {
		option(opts)
	}

	return opts
}

// obtain the error information corresponding to the error code
func GetError(lang string, path string, e *MyError) *MyError {
	msg := GetI18nValue(lang, path, e.Code)
	httpCode := getHttpCode(e.Code)
	if httpCode == 400 {
		msg = fmt.Sprintf("%s(%s)", msg, e.Cause)
	}

	return &MyError{
		HttpCode: httpCode,
		Code:     e.Code,
		Detail:   e.Detail,
		Cause:    e.Cause,
		Solution: e.Solution,
		Message:  msg,
	}
}

func getHttpCode(code int) (httpCode int) {
	codeStr := strconv.Itoa(code)
	rs := string([]rune(codeStr)[:3])

	httpCode, err := strconv.Atoi(rs)
	if err != nil {
		panic(err)
	}

	return
}
