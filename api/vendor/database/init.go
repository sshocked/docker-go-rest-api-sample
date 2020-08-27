package database

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"time"
)

var DB *gorm.DB
var err error

type Middleware struct {
	gorm.Model
	Title  string
	Path string
	Status int
}

type Item struct {
	gorm.Model
	Title string
	Path  string
	Body  string
	Desc  string
}

const MIDDLEWARE_STATUS_IN_PROGRESS = 1

const MIDDLEWARE_STATUS_FINISHED = 0

func addDatabase(dbname string) error {
	// create database with dbname, won't do anything if db already exists
	DB.Exec("CREATE DATABASE " + dbname)

	// connect to newly created DB (now has dbname param)
	connectionParams := "dbname=" + dbname + " user=docker password=docker sslmode=disable host=db"
	DB, err = gorm.Open("postgres", connectionParams)
	if err != nil {
		return err
	}

	return nil
}

func Init() (*gorm.DB, error) {
	// set up DB connection and then attempt to connect 5 times over 25 seconds
	connectionParams := "user=docker password=docker sslmode=disable host=localhost"
	for i := 0; i < 5; i++ {
		DB, err = gorm.Open("postgres", connectionParams) // gorm checks Ping on Open
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return DB, err
	}

	if !DB.HasTable(&Middleware{}) {
		DB.CreateTable(&Middleware{})
	}
	if !DB.HasTable(&Item{}) {
		DB.CreateTable(&Item{})
	}
	fmt.Println("Db loaded")


	return DB, err
}
