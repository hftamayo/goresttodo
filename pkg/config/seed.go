package config

import (
    "log"
	"os"

    "github.com/hftamayo/gotodo/api/v1/models"
    "gorm.io/gorm"
)

func seedData(db *gorm.DB) error {

    tx := db.Begin()
    if tx.Error != nil {
        return tx.Error
    }

    // Define the data to be seeded
    users := []models.User{
        {Name: "administrador", Email: "administrador@tamayo.com", Password: os.Getenv("ADMINISTRADOR_PASSWORD")},
        {Name: "supervisor", Email: "supervisor@tamayo.com", Password: os.Getenv("SUPERVISOR_PASSWORD")},
		{Name: "user01", Email: "bob@tamayo.com", Password: os.Getenv("USER01_PASSWORD")},
		{Name: "user02", Email: "mary@tamayo.com", Password: os.Getenv("USER02_PASSWORD")},		
    }

    log.Println("Starting to seed users...")
    for i := range users {
        log.Printf("Seeding user: %s\n", users[i].Name)
        if err := tx.Create(&users[i]).Error; err != nil {
            log.Printf("Error seeding user %s: %v\n", users[i].Name, err)
            tx.Rollback()
            return err
        }
    }

    log.Println("Finished seeding users.")

    todos := []models.Todo{
        {Title: "backup the database", Body: "create the entire backup using incremental", UserID: users[0].ID},
        {Title: "test the restore process", Body: "restore the backup and test the process", UserID: users[0].ID},
        {Title: "supervise things", Body: "invent something to supervise", UserID: users[1].ID},
    }

    // Insert the data into the database
    log.Println("Starting to seed todos...")
    for _, todo := range todos {
        log.Printf("Seeding todo: %s\n", todo.Title)
        if err := tx.Create(&todo).Error; err != nil {
            log.Printf("Error seeding todo %s: %v\n", todo.Title, err)
            tx.Rollback()
        }
    }
    log.Println("Finished seeding todos.")

    if err := tx.Commit().Error; err != nil {
        tx.Rollback()
        return err
    }    

    return nil
}