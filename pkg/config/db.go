package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

var db *gorm.DB

func loadEnvVars() error {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error reading context data")
		return err
	}
	return nil
}

func DataLayerConnect() (*gorm.DB, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error reading context data")
	}

	if !isTestEnviro() {
		if err := godotenv.Load(); err != nil {
			fmt.Println("Error reading context data")
		}
		host := os.Getenv("POSTGRES_HOST")
		portStr := os.Getenv("POSTGRES_PORT")
		user := os.Getenv("POSTGRES_USER")
		password := os.Getenv("POSTGRES_PASSWORD")
		databaseName := os.Getenv("POSTGRES_DB")
		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Printf("Error converting port to integer.\n%v", err)
			return nil, err
		}
		connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, databaseName)
		db, err := gorm.Open("postgres", connectionString)
		if err != nil {
			log.Printf("Error connecting to the database.\n%v", err)
			return nil, err
		}
		db.AutoMigrate(&models.User{})
		return db, nil
	} else {
		return nil, errors.New("running in testing mode")
	}
}

func checkDataLayerAvailability() bool {
	var err error

	for i := 0; i < 3; i++ {
		db, err = gorm.Open("postgres", "host=localhost port=5432 user=postgres dbname=gotodo sslmode=disable")
		if err != nil {
			log.Printf("Error connecting to the database.\n%v", err)
			return true
		}
		time.Sleep(3 * time.Second)
	}
	return false
}

func isTestEnviro() bool {
	envMode := os.Getenv("GOAPP_MODE")
	switch envMode {
	case "development":
		fmt.Println("running in development mode")
		return false
	case "testing":
		fmt.Println("running in testing mode")
		return true
	case "production":
		fmt.Println("running in production mode")
		return true
	default:
		fmt.Println("no run mode specified")
		return false
	}
}
