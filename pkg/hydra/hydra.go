package hydra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
	"github.com/jettjia/igo-pkg/pkg/conf"
	"github.com/jettjia/igo-pkg/pkg/httpclient"
)

var (
	register_url   = "/admin/clients"
	introspect_url = "/admin/oauth2/introspect"
	token_url      = "/oauth2/token"
	revoke_url     = "/oauth2/revoke"
)

type Hydra struct {
	Scheme       string `json:"scheme"`
	ClientName   string `json:"client_name"`
	ClientSecret string `json:"client_secret"`
	ClientId     string `json:"client_id"`

	httpClient    *resty.Client
	registerUrl   string
	introspectUrl string
	tokenUrl      string
	revokeUrl     string
}

// HydraToken hydra token
type HydraToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

type HydraRegisterRsp struct {
	ClientId     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	ClientName   string `json:"client_name"`
}

func NewHydraAdmin(conf *conf.Config) Hydra {
	var scheme = conf.Third.Extra["hydra_scheme"].(string)
	if scheme == "" {
		scheme = "http"
	}

	hydra_admin_host := conf.Third.Extra["hydra_admin_host"].(string)
	if hydra_admin_host == "" {
		hydra_admin_host = "hydra"
	}

	hydra_admin_port := conf.Third.Extra["hydra_admin_port"].(int)
	if hydra_admin_port == 0 {
		hydra_admin_port = 4445
	}

	return Hydra{
		httpClient: httpclient.NewHttpClient(),

		registerUrl:   fmt.Sprintf("%s://%v:%v%s", scheme, hydra_admin_host, hydra_admin_port, register_url),
		introspectUrl: fmt.Sprintf("%s://%v:%v%s", scheme, hydra_admin_host, hydra_admin_port, introspect_url),
	}
}

// 注册 hydra client
func (h *Hydra) RegisterClient(ctx context.Context, hydra Hydra) (rsp HydraRegisterRsp, err error) {
	hydraData := make(map[string]interface{}, 0)
	hydraData["client_name"] = hydra.ClientName
	hydraData["client_secret"] = hydra.ClientSecret
	hydraData["scope"] = "openid offline all"
	hydraData["grant_types"] = []string{"authorization_code", "refresh_token", "implicit", "client_credentials"}
	hydraData["response_types"] = []string{"code", "token"}

	resp, err := h.httpClient.R().EnableTrace().SetBody(hydraData).Post(h.registerUrl)
	if err != nil {
		return
	}

	if resp.StatusCode() == 201 {
		err = json.Unmarshal(resp.Body(), &rsp)
		if err != nil {
			return
		}
	}

	return
}

// 校验 token的有效性
func (h *Hydra) Introspect(ctx context.Context, token string) (flag bool) {
	formData := map[string]string{
		"token": token,
	}

	resp, err := h.httpClient.R().EnableTrace().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(formData).
		Post(h.introspectUrl)
	if err != nil {
		fmt.Println("[common.third.hydra].Introspect.Post.Error:", err)
		return
	}

	type introspectInfo struct {
		Active bool `json:"active"`
	}

	if resp.StatusCode() == 200 {
		var info introspectInfo
		err = json.Unmarshal(resp.Body(), &info)
		if err != nil {
			return
		}
		flag = info.Active
	}

	return
}

// 根据clientId,获取hydra client 信息
func (h *Hydra) GetClientInfo(ctx context.Context, clientId string) (rsp HydraRegisterRsp, err error) {
	url := fmt.Sprintf("%s/%s", h.registerUrl, clientId)

	resp, err := h.httpClient.R().EnableTrace().Get(url)
	if err != nil {
		return
	}

	if resp.StatusCode() == 200 {
		err = json.Unmarshal(resp.Body(), &rsp)
		if err != nil {
			return
		}
	}

	return
}

func NewHydraPublic(conf *conf.Config) Hydra {
	var scheme = conf.Third.Extra["hydra_scheme"].(string)
	if scheme == "" {
		scheme = "http"
	}

	hydra_public_host := conf.Third.Extra["hydra_public_host"].(string)
	if hydra_public_host == "" {
		hydra_public_host = "hydra"
	}

	hydra_public_port := conf.Third.Extra["hydra_public_port"].(int)
	if hydra_public_port == 0 {
		hydra_public_port = 4444
	}

	return Hydra{
		httpClient: httpclient.NewHttpClient(),
		tokenUrl:   fmt.Sprintf("%s://%v:%v%s", scheme, hydra_public_host, hydra_public_port, token_url),
		revokeUrl:  fmt.Sprintf("%s://%v:%v%s", scheme, hydra_public_host, hydra_public_port, revoke_url),
	}
}

// 获取 hydra token
func (h *Hydra) GetToken(ctx context.Context, hydra Hydra) (hydraToken HydraToken, err error) {
	formData := map[string]string{
		"grant_type": "client_credentials",
		"scope":      "all",
	}

	resp, err := h.httpClient.R().EnableTrace().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(formData).
		SetBasicAuth(hydra.ClientId, hydra.ClientSecret).
		Post(h.tokenUrl)

	if err != nil {
		err = fmt.Errorf("[hydra][get token] err: %v", err)
		return
	}

	if resp.StatusCode() != 200 {
		err = fmt.Errorf("[hydra][get token] err, status code is %d, err:%s", resp.StatusCode(), resp.Body())
		return
	}

	err = json.Unmarshal(resp.Body(), &hydraToken)

	return
}

// 撤销 token
func (h *Hydra) RevokeToken(ctx context.Context, hydra Hydra, token string) (err error) {
	formData := map[string]string{
		"token":           token,
		"token_type_hint": "refresh_token",
	}
	resp, err := h.httpClient.R().EnableTrace().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetFormData(formData).
		SetBasicAuth(hydra.ClientId, hydra.ClientSecret).
		Post(h.revokeUrl)

	if err != nil {
		err = fmt.Errorf("[hydra][revoke token] err: %v", err)
		return
	}

	if resp.StatusCode() != 200 {
		err = fmt.Errorf("[hydra][revoke token] err, status code is %d", resp.StatusCode())
		return
	}

	return
}
