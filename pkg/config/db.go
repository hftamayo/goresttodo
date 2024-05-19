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

func buildConnectionString(envVars *EnvVars) string {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", envVars.Host, envVars.Port, envVars.User, envVars.Password, envVars.Dbname)
	return connectionString
}

func isTestEnviro(envVars *EnvVars) bool {
	switch envVars.Mode {
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

func LoadEnvVars() (*EnvVars, error) {
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

func CheckDataLayerAvailability(envVars *EnvVars) bool {
	connectionString := buildConnectionString(envVars)

	for i := 0; i < 3; i++ {
		start := time.Now() // Start the timer

		db, err := gorm.Open("postgres", connectionString)
		if err != nil {
			elapsed := time.Since(start) // Calculate elapsed time

			if i == 0 {
				log.Printf("First connection attempt unsuccessful.\n%v", err)
			} else {
				log.Printf("Connection attempt %d unsuccessful. Elapsed time: %v\n%v", i+1, elapsed, err)
			}

			time.Sleep(3 * time.Second)
			continue
		}

		db.Close()
		return true
	}

	return false
}

func DataLayerConnect(envVars *EnvVars) (*gorm.DB, error) {

	if !isTestEnviro(envVars) {
		connectionString := buildConnectionString(envVars) // Assign the returned value to the connectionString variable
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
