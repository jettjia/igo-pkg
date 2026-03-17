package mysql

import (
	"fmt"
	"testing"

	"gorm.io/gorm"

	"github.com/jettjia/igo-pkg/pkg/conf"
)

type User struct {
	Id     int    `sql_gorm:"primary_key" json:"id"`
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Gender int    `json:"gender"`
}

var db *gorm.DB

func init() {
	dbCfg := conf.DBConf{
		DbType:          "postgres",
		DbHost:          "127.0.0.1",
		DbPort:          5432,
		Username:        "root",
		Password:        "admin123",
		DbName:          "ddddemo",
		Charset:         "utf8mb4",
		MaxIdleConn:     10,
		MaxOpenConn:     100,
		ConnMaxLifetime: 20,
		LogMode:         1,
	}

	var pkgConf = conf.Config{}

	pkgConf.DB = dbCfg

	db = NewDBClient(&pkgConf).Conn
}

// go test -v -run=Test_AutoMigrate .
func Test_AutoMigrate(t *testing.T) {
	// 初始化表
	err := db.AutoMigrate(&User{})

	if err != nil {
		panic(any(err))
	}
}

// go test -v -run Test_NewDBClient_Find ./
func Test_NewDBClient_Find(t *testing.T) {
	var user User
	err := db.Limit(1).Find(&user).Where("id = ?", 2).Error
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(user)
}

func mysqlAutoMigrate(db *gorm.DB) {
	err := db.AutoMigrate(&User{})

	if err != nil {
		panic(any(err))
	}
}

func createUserTx(tx *gorm.DB, user *User) (err error) {
	err = tx.Create(user).Error
	return
}

func createUserSecondTx(tx *gorm.DB, user *User) (err error) {
	err = tx.Create(user).Error
	return
}
