package config

import (
    "log"
	"os"

    "github.com/hftamayo/gotodo/api/v1/models"
    "gorm.io/gorm"
)

func seedData(db *gorm.DB) error {
    // Define the data to be seeded
    users := []models.User{
        {Name: "administrador", Email: "administrador@tamayo.com", Password: os.Getenv("ADMINISTRADOR_PASSWORD")},
        {Name: "supervisor", Email: "supervisor@tamayo.com", Password: os.Getenv("SUPERVISOR_PASSWORD")},
		{Name: "user01", Email: "bob@tamayo.com", Password: os.Getenv("USER01_PASSWORD")},
		{Name: "user02", Email: "mary@tamayo.com", Password: os.Getenv("USER02_PASSWORD")},		
    }

    todos := []models.Todo{
        {UserID: 1, Title: "Todo 1 for user 1"},
        {UserID: 1, Title: "Todo 2 for user 1"},
        {UserID: 2, Title: "Todo 1 for user 2"},
    }

    // Insert the data into the database
    for _, user := range users {
        if err := db.Create(&user).Error; err != nil {
            return err
        }
    }

    for _, todo := range todos {
        if err := db.Create(&todo).Error; err != nil {
            return err
        }
    }

    return nil
}