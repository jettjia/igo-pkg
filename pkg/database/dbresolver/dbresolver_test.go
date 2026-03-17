package dbresolver

import (
	"log"
	"testing"

	"github.com/jettjia/igo-pkg/pkg/conf"
	"gorm.io/plugin/dbresolver"
)

/*
*
CREATE TABLE `user` (

	`ulid` varchar(128) COLLATE utf8mb4_bin NOT NULL COMMENT 'ulid',
	`name` varchar(255) COLLATE utf8mb4_bin DEFAULT NULL COMMENT 'name',
	PRIMARY KEY (`ulid`)

) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;
*/
type User struct {
	Ulid string `json:"ulid"`
	Name string `json:"name"`
}

// go test -v -run Test_NewDBManager ./
func Test_NewDBManager(t *testing.T) {
	cfgs := []DataSourceCfg{
		{
			Name:         "primary",
			MasterDSN:    "root:admin123@tcp(127.0.0.1:3306)/primary_db?charset=utf8mb4&parseTime=True&loc=Local",
			SlaveDSNs:    []string{"root:admin123@tcp(127.0.0.1:3306)/primary_db?charset=utf8mb4&parseTime=True&loc=Local"},
			ResolverConf: dbresolver.Config{Policy: dbresolver.RandomPolicy{}},
		},
		{
			Name:         "secondary",
			MasterDSN:    "root:admin123@tcp(127.0.0.1:3306)/secondary_db?charset=utf8mb4&parseTime=True&loc=Local",
			SlaveDSNs:    []string{"root:admin123@tcp(127.0.0.1:3306)/secondary_db?charset=utf8mb4&parseTime=True&loc=Local"},
			ResolverConf: dbresolver.Config{Policy: dbresolver.RandomPolicy{}},
		},
	}

	manager := NewDBManager(cfgs)

	primaryDB := manager.Sources["primary"]
	secondaryDB := manager.Sources["secondary"]

	user := User{Ulid: "1", Name: "Bob"}
	result := primaryDB.Create(&user)
	if result.Error != nil {
		log.Fatalf("Create failed: %v", result.Error)
	}

	var users []User
	result = primaryDB.Find(&users)
	if result.Error != nil {
		log.Fatalf("Find failed: %v", result.Error)
	}
	for _, u := range users {
		log.Printf("User: %v", u.Name)
	}

	/////////////////////////////

	user2 := User{Ulid: "1", Name: "Bob2"}
	result2 := secondaryDB.Create(&user2)
	if result.Error != nil {
		log.Fatalf("Create failed: %v", result2.Error)
	}

	var users2 []User
	result2 = primaryDB.Find(&users2)
	if result.Error != nil {
		log.Fatalf("Find failed: %v", result2.Error)
	}
	for _, u := range users2 {
		log.Printf("User: %v", u.Name)
	}
}

// go test -v -run Test_NewDBManagerClient ./
func Test_NewDBManagerClient(t *testing.T) {
	dbCfg := conf.DBManagerConf{
		DataSources: []conf.DataSourceCfg{
			{
				Name:      "primary",
				MasterDSN: "root:admin123@tcp(127.0.0.1:3306)/primary_db?charset=utf8mb4&parseTime=True&loc=Local",
				SlaveDSNs: []string{"root:admin123@tcp(127.0.0.1:3306)/primary_db?charset=utf8mb4&parseTime=True&loc=Local"},
			},
		},
	}

	var pkgConf = conf.Config{}
	pkgConf.DBManager = dbCfg
	DBManager := NewDBManagerClient(&pkgConf)
	primaryDB := DBManager.Manager.Sources["primary"]

	user := User{Ulid: "2", Name: "BobLi"}
	result := primaryDB.Create(&user)
	if result.Error != nil {
		log.Fatalf("Create failed: %v", result.Error)
	}

	var users []User
	result = primaryDB.Find(&users)
	if result.Error != nil {
		log.Fatalf("Find failed: %v", result.Error)
	}
	for _, u := range users {
		log.Printf("User: %v", u.Name)
	}
}

// go test -v -run Test_NewDBManager_pgsql ./
func Test_NewDBManager_pgsql(t *testing.T) {
	cfgs := []DataSourceCfg{
		{
			Name:         "primary",
			MasterDSN:    "host=127.0.0.1 user=root password=admin123 dbname=xtext port=5432 sslmode=disable",
			SlaveDSNs:    []string{"host=127.0.0.1 user=root password=admin123 dbname=xtext port=5432 sslmode=disable"},
			ResolverConf: dbresolver.Config{Policy: dbresolver.RandomPolicy{}},
		},
		{
			Name:         "secondary",
			MasterDSN:    "host=127.0.0.1 user=root password=admin123 dbname=xtext_deploy port=5432 sslmode=disable",
			SlaveDSNs:    []string{"host=127.0.0.1 user=root password=admin123 dbname=xtext_deploy port=5432 sslmode=disable"},
			ResolverConf: dbresolver.Config{Policy: dbresolver.RandomPolicy{}},
		},
	}

	manager := NewDBManager(cfgs)

	primaryDB := manager.Sources["primary"]
	secondaryDB := manager.Sources["secondary"]

	user := User{Ulid: "1", Name: "Bob"}
	result := primaryDB.Create(&user)
	if result.Error != nil {
		log.Fatalf("Create failed: %v", result.Error)
	}

	var users []User
	result = primaryDB.Find(&users)
	if result.Error != nil {
		log.Fatalf("Find failed: %v", result.Error)
	}
	for _, u := range users {
		log.Printf("User: %v", u.Name)
	}

	/////////////////////////////

	user2 := User{Ulid: "1", Name: "Bob2"}
	result2 := secondaryDB.Create(&user2)
	if result.Error != nil {
		log.Fatalf("Create failed: %v", result2.Error)
	}

	var users2 []User
	result2 = primaryDB.Find(&users2)
	if result.Error != nil {
		log.Fatalf("Find failed: %v", result2.Error)
	}
	for _, u := range users2 {
		log.Printf("User: %v", u.Name)
	}
}
