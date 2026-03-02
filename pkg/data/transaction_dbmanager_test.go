package data

import (
	"context"
	"testing"

	"github.com/jettjia/igo-pkg/pkg/conf"
	"github.com/jettjia/igo-pkg/pkg/database/mysqlresolver"
)

var dataManagerCli *Data

func init() {
	var pkgConf = conf.Config{}

	// db config
	dbCfg := conf.DBManagerConf{
		DataSources: []conf.DataSourceCfg{
			{
				Name:      "primary",
				MasterDSN: "root:admin123@tcp(127.0.0.1:3306)/primary_db?charset=utf8mb4&parseTime=True&loc=Local",
				SlaveDSNs: []string{"root:admin123@tcp(127.0.0.1:3306)/primary_db?charset=utf8mb4&parseTime=True&loc=Local"},
			},
		},
	}

	pkgConf.DBManager = dbCfg

	dbCli := mysqlresolver.NewDBManagerClient(&pkgConf).Manager

	dataManagerCli = NewDataOption(WithDBManager(dbCli))
}

// go test -v -run Test_TransactionForDBManager ./
func Test_TransactionForDBManager(t *testing.T) {
	err := dataManagerCli.ExecTxForDBManager(context.Background(), "primary", func(ctx context.Context) error {
		err := createM1(ctx)
		if err != nil {
			return err
		}

		err = createM2(ctx)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		t.Errorf("err:%v", err)
	}
}

func createM1(ctx context.Context) (err error) {
	err = dataManagerCli.DBForDBManager(ctx, "primary").Create(&User{Name: "jet", Ulid: "test3"}).Error

	return
}

func createM2(ctx context.Context) (err error) {
	err = dataManagerCli.DBForDBManager(ctx, "primary").Create(&User{Name: "jack", Ulid: "test4"}).Error

	return
}
