package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

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
    FeOrigins []string
}

func LoadEnvVars() (*EnvVars, error) {
    godotenv.Load()

    getEnv := func(key, fallback string) string {
        if value, exists := os.LookupEnv(key); exists {
            return value
        }
        return fallback
    }

    port, err := strconv.Atoi(getEnv("POSTGRES_PORT", "5432"))
    if err != nil {
        return nil, fmt.Errorf("invalid postgres port: %v", err)
    }

    appPort, err := strconv.Atoi(getEnv("GOAPP_PORT", "8001"))
    if err != nil {
        return nil, fmt.Errorf("invalid app port: %v", err)
    }

    seedDev, _ := strconv.ParseBool(getEnv("SEED_DEVELOPMENT", "false"))
    seedProd, _ := strconv.ParseBool(getEnv("SEED_PRODUCTION", "false"))

    originsStr := getEnv("FRONTEND_ORIGINS", "http://localhost:5173")
    origins := strings.Split(originsStr, ",")

    for i := range origins {
        origins[i] = strings.TrimSpace(origins[i])
    }    

    envVars := &EnvVars{
        Host:     getEnv("POSTGRES_HOST", "localhost"),
        Port:     port,
        AppPort:  appPort,
        User:     getEnv("POSTGRES_USER", "postgres"),
        Password: getEnv("POSTGRES_PASSWORD", ""),
        Dbname:   getEnv("POSTGRES_DB", "gotodo_dev"),
        Mode:     getEnv("GOAPP_MODE", "development"),
        timeOut:  30,
        seedDev:  seedDev,
        seedProd: seedProd,
        FeOrigins: origins,
    }

    if envVars.Password == "" {
        return nil, fmt.Errorf("POSTGRES_PASSWORD is required")
    }

    return envVars, nil
}
