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

    for _, user := range users {
        if err := db.Create(&user).Error; err != nil {
            return err
        }
    }

    todos := []models.Todo{
        {Title: "backup the database", Body: "create the entire backup using incremental", UserID: users[0].ID},
        {Title: "test the restore process", Body: "restore the backup and test the process", UserID: users[0].ID},
        {Title: "supervise things", Body: "invent something to supervise", UserID: users[1].ID},
    }

    // Insert the data into the database

    for _, todo := range todos {
        if err := db.Create(&todo).Error; err != nil {
            return err
        }
    }

    return nil
}