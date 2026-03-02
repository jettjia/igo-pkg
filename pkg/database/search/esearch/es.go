package esearch

import (
	"log"
	"sync"

	"github.com/jettjia/go-pkg/pkg/conf"

	"github.com/elastic/go-elasticsearch/v8"
)

type EsDB struct {
	Conn      *elasticsearch.Client
	ConnTyped *elasticsearch.TypedClient
}

type EsConfig struct {
	Addrs    []string
	Username string
	Password string
}

var (
	once   sync.Once
	esConn *EsDB
)

// NewEsClient 初始化
func NewEsClient(conf *conf.Config) *EsDB {
	once.Do(func() {
		// 转换配置
		cfg := tranCfg(conf)

		esConn = &EsDB{}

		esConn.Conn, _ = getConn(&cfg)

		esConn.ConnTyped, _ = getConnTyped(&cfg)

	})

	return esConn
}

// 转换配置
func tranCfg(conf *conf.Config) (cfg EsConfig) {
	cfg = EsConfig{
		Addrs:    []string{conf.Search.Addr},
		Username: conf.Search.Username,
		Password: conf.Search.Password,
	}

	return
}

func getConn(EsCfg *EsConfig) (*elasticsearch.Client, error) {
	cfg := elasticsearch.Config{
		Addresses: EsCfg.Addrs,
		Username:  EsCfg.Username,
		Password:  EsCfg.Password,
	}

	cli, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Println("elasticsearch.NewClient err:", err)
		return nil, err
	}

	return cli, nil
}

func getConnTyped(EsCfg *EsConfig) (*elasticsearch.TypedClient, error) {
	cfg := elasticsearch.Config{
		Addresses: EsCfg.Addrs,
		Username:  EsCfg.Username,
		Password:  EsCfg.Password,
	}

	cli, err := elasticsearch.NewTypedClient(cfg)
	if err != nil {
		log.Println("elasticsearch.NewClient err:", err)
		return nil, err
	}

	return cli, nil

}
