package config

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/hftamayo/gotodo/api/v1/models"
	"github.com/hftamayo/gotodo/pkg/seeder"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
    maxRetries := 3
    retryDelay := 30 * time.Second

    for i := 0; i < maxRetries; i++ {
        start := time.Now()
        fmt.Printf("Attempting to connect to database (attempt %d/%d)...\n", i+1, maxRetries)
        fmt.Printf("Connection string: %s\n", maskConnectionString(connectionString))

        db, err := gorm.Open(postgres.Open(connectionString), &gorm.Config{})
        elapsed := time.Since(start)

        if err != nil {
            log.Printf("Connection attempt %d failed after %v: %v\n", i+1, elapsed, err)
            
            if i < maxRetries-1 {
                fmt.Printf("Retrying in %v...\n", retryDelay)
                time.Sleep(retryDelay)
                continue
            }
            return nil, fmt.Errorf("failed to connect after %d attempts: %v", maxRetries, err)
        }

        // Test the connection
        sqlDB, err := db.DB()
        if err != nil {
            log.Printf("Error getting underlying *sql.DB: %v\n", err)
            return nil, err
        }

        err = sqlDB.Ping()
        if err != nil {
            log.Printf("Error pinging database: %v\n", err)
            return nil, err
        }

        fmt.Printf("Successfully connected to database after %v\n", elapsed)
        return db, nil
    }

    return nil, errors.New("data layer is not available")
}

// Helper function to mask sensitive information in connection string
func maskConnectionString(connStr string) string {
    // Replace password with asterisks but keep other information visible
    return regexp.MustCompile(`password=([^ ]+)`).ReplaceAllString(connStr, "password=*****")
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
		err = db.AutoMigrate(&models.User{}, &models.Task{})
		if err != nil {
			log.Printf("Error during migration.\n%v", err)
			return nil, err
		}

		if envVars.seedDev || envVars.seedProd {
			log.Println("Data seeding required")

			err = seeder.SeedData(db)
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
