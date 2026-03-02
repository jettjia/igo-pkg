package main

import (
	"io/ioutil"
	"log"

	resolver "github.com/jettjia/go-pkg/pkg/database/mysqlresolver"
	"gopkg.in/yaml.v3"
)

type User struct {
	Ulid string `json:"ulid"`
	Name string `json:"name"`
}

func main() {
	cfgs, err := loadConfig("mysql.yaml")
	if err != nil {
		panic(err)
	}

	manager := resolver.NewDBManager(cfgs)

	if err != nil {
		log.Fatalf("Failed to initialize DB Manager: %v", err)
	}

	primaryDB := manager.Sources["primary"]
	secondaryDB := manager.Sources["secondary"]

	user := User{Ulid: "2", Name: "jack"}
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

	user2 := User{Ulid: "2", Name: "jack2"}
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

func loadConfig(filename string) ([]resolver.DataSourceCfg, error) {
	yamlData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	type Config struct {
		DataSources []resolver.DataSourceCfg `yaml:"data_sources"`
	}

	var config Config
	err = yaml.Unmarshal(yamlData, &config)
	if err != nil {
		return nil, err
	}

	return config.DataSources, nil
}
