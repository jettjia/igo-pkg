package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/go-resty/resty/v2"
)

type Http struct {
	ApiUrl string `json:"api_url"`
	Method string `json:"method"`

	ReqParams     interface{}       `json:"req_params"`
	Headers       map[string]string `json:"headers"`
	BearerToken   string            `json:"bearer_token"`
	SetRetryCount int               `json:"set_retry_count"`
	ShowLog       bool              `json:"show_log"`

	SecretID  string `json:"secret_id"`  // for basic auth
	SecretKey string `json:"secret_key"` // for  basic auth
}

var (
	gclient *resty.Client

	httpCfg Http
)

func NewHttpClient() *resty.Client {
	gclient = resty.New()

	return gclient
}

func NewHttpClientWithBearer(apiUrl string, method string, bearerToken string, options ...func(http *Http)) *Http {
	httpCfg = Http{
		ApiUrl:        apiUrl,
		Method:        method,
		BearerToken:   bearerToken,
		SetRetryCount: 3,
		ShowLog:       true,
	}

	for _, option := range options {
		option(&httpCfg)
	}

	gclient = resty.New()
	return &httpCfg
}

func NewHttpClientWithBasicAuth(apiUrl string, method string, options ...func(http *Http)) *Http {
	httpCfg = Http{
		ApiUrl:        apiUrl,
		Method:        method,
		SetRetryCount: 3,
		ShowLog:       true,
	}

	for _, option := range options {
		option(&httpCfg)
	}

	gclient = resty.New().SetBasicAuth(httpCfg.SecretID, httpCfg.SecretKey)
	return &httpCfg
}

func (c *Http) Do(ctx context.Context) (r *resty.Response, err error) {
	if len(c.BearerToken) > 0 {
		gclient.SetHeader("Authorization", "Bearer "+c.BearerToken)
	}

	if len(c.Headers) > 0 {
		for k, v := range c.Headers {
			gclient.SetHeader(k, v)
		}
	}

	if c.SetRetryCount > 0 {
		gclient.SetRetryCount(c.SetRetryCount)
	}
	if c.ShowLog {
		log.Println(ctx, "httpClient apiUrl:", c.ApiUrl)
		log.Println(ctx, "httpClient method:", c.Method)
		log.Println(ctx, "httpClient params:", c.ReqParams)
		log.Println(ctx, "httpClient header:", c.Headers)
		log.Println(ctx, "httpClient bearerToken:", c.BearerToken)
		log.Println(ctx, "httpClient ctx:", ctx)
	}

	c.Method = strings.ToLower(c.Method)
	if c.Method == "get" {
		r, err = gclient.R().Get(c.ApiUrl)
		if err != nil {
			return nil, err
		}

		if r.StatusCode() < 200 || r.StatusCode() >= 300 {
			return nil, fmt.Errorf("HttpClient.Req.StatusCode:%d", r.StatusCode())
		}

		return r, nil
	}

	if c.Method == "post" {
		gclient.SetHeader("Content-Type", "application/json")
		bytes, _ := json.Marshal(c.ReqParams)
		r, err = gclient.R().SetBody(string(bytes)).Post(c.ApiUrl)
		if err != nil {
			return nil, err
		}

		if r.StatusCode() < 200 || r.StatusCode() >= 300 {
			return nil, fmt.Errorf("HttpClient.Req.StatusCode:%d", r.StatusCode())
		}

		return r, nil
	}

	if c.Method == "postform" || c.Method == "post_form" {
		gclient.SetHeader("Content-Type", "application/x-www-form-urlencoded")

		r, err = gclient.R().SetBody(c.ReqParams).Post(c.ApiUrl)

		if err != nil {
			return nil, err
		}

		if r.StatusCode() < 200 || r.StatusCode() >= 300 {
			return nil, fmt.Errorf("HttpClient.Req.StatusCode:%d", r.StatusCode())
		}

		return r, nil
	}

	if c.Method == "delete" || c.Method == "del" {
		gclient.SetHeader("Content-Type", "application/json")
		bytes, _ := json.Marshal(c.ReqParams)
		r, err = gclient.R().SetBody(string(bytes)).Delete(c.ApiUrl)
		if err != nil {
			return nil, err
		}

		if r.StatusCode() < 200 || r.StatusCode() >= 300 {
			return nil, fmt.Errorf("HttpClient.Req.StatusCode:%d", r.StatusCode())
		}
		return r, nil
	}

	if c.Method == "put" {
		bytes, _ := json.Marshal(c.ReqParams)
		r, err = gclient.R().SetBody(string(bytes)).Put(c.ApiUrl)
		if err != nil {
			return nil, err
		}

		if r.StatusCode() < 200 || r.StatusCode() >= 300 {
			return nil, fmt.Errorf("HttpClient.Req.StatusCode:%d", r.StatusCode())
		}

		return r, nil
	}

	return nil, errors.New("no method matching")
}
