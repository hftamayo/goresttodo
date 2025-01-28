package config

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB


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

func CheckDataLayerAvailability(envVars *EnvVars) (*gorm.DB, error) {
	connectionString := buildConnectionString(envVars)

	for i := 0; i < 3; i++ {
		start := time.Now() // Start the timer

		db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
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

	return errors.New("data layer is not available")
}

func DataLayerConnect(envVars *EnvVars) (*gorm.DB, error) {
	if !isTestEnviro(envVars) {
		connectionString := buildConnectionString(envVars) // Assign the returned value to the connectionString variable

		db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
		if err != nil {
			log.Printf("Error connecting to the database.\n%v", err)
			return nil, err
		}

        // AutoMigrate will create the tables based on the models
        err = db.AutoMigrate(&models.User{}, &models.Todo{})
        if err != nil {
            log.Printf("Error during migration.\n%v", err)
            return nil, err
        }		

		if envVars.seedDev || envVars.seedProd {
			log.Println("Data seeding required")

			err = seedData(db)
            if err != nil {
                log.Printf("Error during data seeding.\n%v", err)
                return nil, err
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
