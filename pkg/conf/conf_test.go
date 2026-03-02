package conf

import (
	"flag"
	"fmt"
	"testing"
)

// go test -v -run Test_Conf ./
func Test_Conf(t *testing.T) {
	file := "example.yaml"
	cfg := Config{}
	if err := flag.Set("conf", file); err != nil {
		panic(err)
	}
	if err := ParseYaml(&cfg); err != nil {
		panic(err)
	}
	fmt.Println(cfg.Server.PublicPort)

	// read extra config
	fmt.Println("hydra_admin_host:", cfg.Third.Extra["hydra_admin_host"])
}
