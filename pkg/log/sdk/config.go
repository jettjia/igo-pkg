package sdk

import "github.com/jettjia/igo-pkg/pkg/log/enum"

// Config basic log configuration information
type Config struct {
	LogFilePath string        // storage path
	LogName     string        // Log Name like app-name
	MaxSize     int           // At What File Size Should Log Rotation Start
	MaxBackup   int           // Number of Retained Files
	MaxAge      int           // maximum number of days of storage
	LogLevel    string        // log level: panic, fatal, error, warn, info, debug,trace
	LogSendType enum.EnumType // log sending type
}

type ES struct {
	EsAddrs    []string // ES addr
	EsUser     string   // ES user
	EsPassword string   // ES password
}

type ZS struct {
	ZsAddrs    string // zincsearch addr
	ZsUser     string // zincsearch user
	ZsPassword string // zincsearch password
}

type MQ struct {
	MqType       string // nsq, kafka
	ProducerHost string // producer host
	ProducerPort int    // producer port
}

type OTEL struct {
	Endpoint string
	User     string
	Pwd      string
}
