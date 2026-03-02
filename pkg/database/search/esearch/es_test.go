package esearch

import (
	"testing"

	"github.com/jettjia/go-pkg/pkg/conf"
)

func Test_EsClient(t *testing.T) {
	esCfg := conf.SearchConf{
		Addr:     "127.0.0.1",
		Username: "",
		Password: "",
	}

	var pkgConf = conf.Config{}
	pkgConf.Search = esCfg

	esClient := NewEsClient(&pkgConf)

	esClient.Conn.Indices.Create("my_index")
}
