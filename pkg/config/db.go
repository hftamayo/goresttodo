package config

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

var db *gorm.DB

func DataLayerConnect() (*gorm.DB, error) {
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

func isTestEnviro() bool {
	envMode := os.Getenv("GOAPP_MODE")
	switch envMode {
	case "development":
		fmt.Println("running in development mode")
		return false
	case "testing":
		fmt.Println("running in testing mode")
		return true
	default:
		fmt.Println("running in production mode")
		return false
	}
}
