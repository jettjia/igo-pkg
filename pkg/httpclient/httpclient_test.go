package httpclient

import (
	"context"
	"fmt"
	"testing"
)

func Test_NewHttpClient(t *testing.T) {
	client := NewHttpClient()

	resp, err := client.R().EnableTrace().Get("https://www.baidu.com")

	// Explore response object
	fmt.Println("Response Info:")
	fmt.Println("  Error      :", err)
	fmt.Println("  Status Code:", resp.StatusCode())
	fmt.Println("  Status     :", resp.Status())
	fmt.Println("  Proto      :", resp.Proto())
	fmt.Println("  Time       :", resp.Time())
	fmt.Println("  Received At:", resp.ReceivedAt())
	fmt.Println("  Body       :\n", resp)
	fmt.Println()
}

func Test_NewHttpClientWithBearer(t *testing.T) {
	client := NewHttpClientWithBearer("https://www.baidu.com", "get", "")

	resp, err := client.Do(context.TODO())

	// Explore response object
	fmt.Println("Response Info:")
	fmt.Println("  Error      :", err)
	fmt.Println("  Status Code:", resp.StatusCode())
	fmt.Println("  Status     :", resp.Status())
	fmt.Println("  Proto      :", resp.Proto())
	fmt.Println("  Time       :", resp.Time())
	fmt.Println("  Received At:", resp.ReceivedAt())
	fmt.Println("  Body       :\n", resp)
	fmt.Println()
}

func Test_NewHttpClientWithBasicAuth(t *testing.T) {
	apiUrl := "http://xxx.com"
	method := "get"
	client := NewHttpClientWithBasicAuth(apiUrl, method, WithSecretID("xx"), WithSecretKey("xxx"))

	resp, err := client.Do(context.TODO())
	if err != nil {
		t.Error(err)
		return
	}

	fmt.Println("body:", string(resp.Body()))
}
