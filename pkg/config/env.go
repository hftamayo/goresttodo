package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type EnvVars struct {
	Host     string
	Port     int
	User     string
	Password string
	Dbname   string
	AppPort  int
	Mode     string
	timeOut  int
	seedDev  bool
	seedProd bool
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

	appPortStr := os.Getenv("GOAPP_PORT")
	appPort, err := strconv.Atoi(appPortStr)
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
		AppPort:  appPort,
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