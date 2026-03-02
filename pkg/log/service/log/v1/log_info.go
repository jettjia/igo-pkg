package v1

import (
	"time"

	"github.com/sirupsen/logrus"
)

// logInfoDocModel logInfo
type logInfoDocModel map[string]interface{}

func getLogInfo(e *logrus.Entry) logInfoDocModel {
	ins := map[string]interface{}{}
	for kk, vv := range e.Data {
		ins[kk] = vv
	}
	ins["time"] = time.Now().Local()
	ins["level"] = e.Level
	ins["message"] = e.Message
	ins["service_name"] = e.Context.Value("service_name")
	// ins["caller"] = fmt.Sprintf("%s:%d  %#v", e.Caller.File, e.Caller.Line, e.Caller.Func)
	return ins
}
