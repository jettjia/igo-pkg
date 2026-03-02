# golog

## 使用说明

方式1：
    此方式，只会给指定目录的日志生成日志内容
```go
func InitLog() {
	config := &sdk.Config{
		LogFilePath: "/tmp",
		LogName:     "server",
		MaxSize:     1024,
		MaxBackup:   64,
		MaxAge:      30,
		InfoLevel:   "debug",
	}

	client, err := logv1.NewClient(config)
	if err != nil {
		panic(err)
	}

	global.Golog = client
}
```
```go
// 打印普通日志
func main() {
	initialize.InitLog()

	global.Golog.DoLogger().Debugln("tttttttttttttttt")
}
```

方式2：
    此方式，会生成日志到指定目录，同时会把日志上报到 ES中
```go
func InitLog() {
	config := &sdk.Config{
		LogFilePath: "/tmp",
		LogName:     "server",
		MaxSize:     1024,
		MaxBackup:   64,
		MaxAge:      30,
		InfoLevel:   "info",
		On:          true, // 开启发送到 es
	}

	es := &sdk.ES{
		EsAddrs:    []string{"http://127.0.0.1:9202/"},
		EsUser:     "",
		EsPassword: "",
	}

	client, err := logv1.NewClientES(config, es)
	if err != nil {
		panic(err)
	}

	global.Golog = client
}
```

```go
func main() {
	initialize.InitLog()

	global.Golog.DoLogger().Infoln("eseseseses...")
}
```

## es
查看所有索引
	get	http://127.0.0.1:9200/_cat/indices?v

​	get	http://127.0.0.1:9200/_all

查看索引下的日志

​	get	http://127.0.0.1:9200/2022-06-16

删除索引

​	delete http://127.0.0.1:9200/2022-06-16
