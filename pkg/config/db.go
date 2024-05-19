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

type EnvVars struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
	Mode     string
}

func loadEnvVars() (*EnvVars, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Println("Error reading context data")
		return nil, err
	}
	portStr := os.Getenv("POSTGRES_PORT")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Error converting port to integer.\n%v", err)
		return nil, err
	}

	envVars := &EnvVars{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     port,
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Dbname:   os.Getenv("POSTGRES_DB"),
		Mode:     os.Getenv("GOAPP_MODE"),
	}
	return envVars, nil
}

func buildConnectionString() (string, error) {
	host := os.Getenv("POSTGRES_HOST")
	portStr := os.Getenv("POSTGRES_PORT")
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	databaseName := os.Getenv("POSTGRES_DB")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Printf("Error converting port to integer.\n%v", err)
		return "", err
	}
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, databaseName)
	return connectionString, nil
}

func DataLayerConnect() (*gorm.DB, error) {
	if err := loadEnvVars(); err != nil {
		return nil, err
	}

	if !isTestEnviro() {
		connectionString, err := buildConnectionString()
		if err != nil {
			return nil, err
		}
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
	if err := loadEnvVars(); err != nil {
		return false
	}

	connectionString, err := buildConnectionString()
	if err != nil {
		return false
	}

	for i := 0; i < 3; i++ {
		db, err = gorm.Open("postgres", connectionString)
		if err != nil {
			log.Printf("Error connecting to the database.\n%v", err)
			time.Sleep(3 * time.Second)
			continue
		}
		db.Close()
		return true
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
