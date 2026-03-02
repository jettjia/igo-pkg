package main

import (
	"github.com/jettjia/igo-pkg/aipkg/mem/mem0"
)

func main() {
	opt := mem0.ClientOptions{
		APIKey:           "1ddwoehwio",
		Host:             "http://127.0.0.1:8888",
		OrganizationName: "mem0",
		ProjectName:      "mem0",
		OrganizationID:   "",
		ProjectID:        "",
	}
	client, err := mem0.NewMemoryClient(opt)
	if err != nil {
		panic("创建客户端失败: %v" + err.Error())
	}

	// 测试添加记忆 - 单字符串形式
	memInfo, err := client.Add("这是一条测试记忆", mem0.MemoryOptions{
		UserID: "alex",
		// AgentID 和 RunID 可能不是必需的，先注释掉
		// AgentID: "alex",
		// RunID:   "alex",
	})
	if err != nil {
		panic("添加记忆失败: %v" + err.Error())
	}
	println("添加记忆成功: %+v", memInfo)
}

func Test_Get() {
	opt := mem0.ClientOptions{
		APIKey:           "1ddwoehwio",
		Host:             "http://127.0.0.1:8888",
		OrganizationName: "mem0",
		ProjectName:      "mem0",
		OrganizationID:   "",
		ProjectID:        "",
	}
	client, err := mem0.NewMemoryClient(opt)
	if err != nil {
		panic("创建客户端失败: %v" + err.Error())
	}

	mem, err := client.Get("953934a0-e381-43aa-8616-095ff78102cd")
	if err != nil {
		panic("获取记忆失败: %v" + err.Error())
	}
	println("获取记忆成功: %+v", mem)
}
