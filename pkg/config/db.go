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
	timeOut  int
	seedDev  bool
	seedProd bool
}

func buildConnectionString(envVars *EnvVars) string {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=%d",
		envVars.Host, envVars.Port, envVars.User, envVars.Password, envVars.Dbname, envVars.timeOut)
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

	seedDevStr := os.Getenv("SEED_DEVELOPMENT")
	seedDev, _ := strconv.ParseBool(seedDevStr)

	seedProdStr := os.Getenv("SEED_PRODUCTION")
	seedProd, _ := strconv.ParseBool(seedProdStr)

	envVars := &EnvVars{
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     port,
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Dbname:   os.Getenv("POSTGRES_DB"),
		Mode:     os.Getenv("GOAPP_MODE"),
		timeOut:  30,
		seedDev:  seedDev,
		seedProd: seedProd,
	}
	return envVars, nil
}

func CheckDataLayerAvailability(envVars *EnvVars) (*gorm.DB, error) {
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

			time.Sleep(30 * time.Second)
			continue
		}
		return db, nil
	}

	return nil, errors.New("data layer is not available")
}

func DataLayerConnect(envVars *EnvVars) (*gorm.DB, error) {
	if !isTestEnviro(envVars) {
		connectionString := buildConnectionString(envVars) // Assign the returned value to the connectionString variable

		db, err := gorm.Open("postgres", connectionString)
		if err != nil {
			log.Printf("Error connecting to the database.\n%v", err)
			return nil, err
		}

		if envVars.seedDev || envVars.seedProd {
			db = db.AutoMigrate(&models.User{}, &models.Todo{})
			if db.Error != nil {
				log.Printf("Error during data seeding.\n%v", err)
				return nil, db.Error
			}
			log.Println("Data seeding successful")
		} else {
			log.Println("No data seeding required")
		}
		return db, nil
	} else {
		return nil, errors.New("no running mode specified, system halted")
	}
}
