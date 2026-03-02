package data

import (
	"context"
	"testing"

	"github.com/jettjia/go-pkg/pkg/conf"
	"github.com/jettjia/go-pkg/pkg/database/mysql"
)

var dataCli *Data

func init() {
	var pkgConf = conf.Config{}

	// db config
	dbCfg := conf.DBConf{
		DbHost:          "127.0.0.1",
		DbPort:          3306,
		Username:        "root",
		Password:        "admin123",
		DbName:          "primary_db",
		Charset:         "utf8mb4",
		MaxIdleConn:     10,
		MaxOpenConn:     100,
		ConnMaxLifetime: 20,
		LogMode:         1,
	}

	pkgConf.DB = dbCfg

	dbCli := mysql.NewDBClient(&pkgConf).Conn

	dataCli = NewDataOption(WithMysql(dbCli))
}

type User struct {
	Ulid string `json:"ulid"`
	Name string `json:"name"`
}

// go test -v -run Test_NewTransaction ./
func Test_NewTransaction(t *testing.T) {
	err := dataCli.ExecTx(context.Background(), func(ctx context.Context) error {
		err := create1(ctx)
		if err != nil {
			return err
		}

		err = create2(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		t.Errorf("err:%v", err)
	}
}

func create1(ctx context.Context) (err error) {
	err = dataCli.DB(ctx).Create(&User{Name: "test1", Ulid: "test1"}).Error

	return
}

func create2(ctx context.Context) (err error) {
	err = dataCli.DB(ctx).Create(&User{Name: "test2", Ulid: "test2"}).Error

	return
}
