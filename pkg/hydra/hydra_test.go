package hydra

import (
	"context"
	"fmt"
	"testing"

	"github.com/jettjia/go-pkg/pkg/conf"
)

// go test -v -run Test_RegisterClient ./
func Test_RegisterClient(t *testing.T) {
	hydraData := Hydra{
		ClientName:   "test_client",
		ClientSecret: "1234567",
	}

	var pkgConf = conf.Config{}
	pkgConf.Third.Extra["hydra_admin_host"] = "127.0.0.1"
	pkgConf.Third.Extra["hydra_admin_port"] = 4445

	hydraClient := NewHydraAdmin(&pkgConf)

	resp, err := hydraClient.RegisterClient(context.Background(), hydraData)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(resp)
}

// go test -v -run Test_GetClientInfo ./
func Test_GetClientInfo(t *testing.T) {
	var pkgConf = conf.Config{}
	pkgConf.Third.Extra["hydra_admin_host"] = "127.0.0.1"
	pkgConf.Third.Extra["hydra_admin_port"] = 4445

	hydraClient := NewHydraAdmin(&pkgConf)

	rsp, err := hydraClient.GetClientInfo(context.Background(), "ffb6a68e-a7b5-411d-8b09-ae9531f668ce")
	if err != nil {
		t.Error(err)
	}

	fmt.Println(rsp)
}

// go test -v -run Test_GetToken ./
func Test_GetToken(t *testing.T) {
	hydraData := Hydra{
		ClientId:     "750d8dd3-a349-45fe-9949-1c7c7323b6c8",
		ClientSecret: "1234567",
	}

	var pkgConf = conf.Config{}
	pkgConf.Third.Extra["hydra_public_host"] = "127.0.0.1"
	pkgConf.Third.Extra["hydra_public_port"] = 4444

	hydraClient := NewHydraPublic(&pkgConf)

	data, err := hydraClient.GetToken(context.Background(), hydraData)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(data)
}

// go test -v -run Test_Introspect ./
func Test_Introspect(t *testing.T) {
	var pkgConf = conf.Config{}
	pkgConf.Third.Extra["hydra_admin_host"] = "127.0.0.1"
	pkgConf.Third.Extra["hydra_admin_port"] = 4445

	hydraClient := NewHydraPublic(&pkgConf)

	flag := hydraClient.Introspect(context.Background(), "ory_at_z9uihsfPdN5866fI1VWuCdv1XpN0wm_W8nSeLj0QnT4.Jvqjr44duW1HJ_Pxc3ADvUgVJRQc-V_q83EayEpuB6c")

	fmt.Println(flag)
}

// go test -v -run Test_RevokeToken ./
func Test_RevokeToken(t *testing.T) {

	hydraData := Hydra{
		ClientId:     "750d8dd3-a349-45fe-9949-1c7c7323b6c8",
		ClientSecret: "1234567",
	}

	var pkgConf = conf.Config{}
	pkgConf.Third.Extra["hydra_admin_host"] = "127.0.0.1"
	pkgConf.Third.Extra["hydra_admin_port"] = 4445

	hydraClient := NewHydraAdmin(&pkgConf)

	flag := hydraClient.RevokeToken(context.Background(), hydraData, "ory_at_z9uihsfPdN5866fI1VWuCdv1XpN0wm_W8nSeLj0QnT4.Jvqjr44duW1HJ_Pxc3ADvUgVJRQc-V_q83EayEpuB6c")

	fmt.Println(flag)
}
