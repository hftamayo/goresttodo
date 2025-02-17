package seeder

import (
    "log"
	"os"

    "github.com/hftamayo/gotodo/api/v1/models"
    "github.com/hftamayo/gotodo/pkg/utils"
    "gorm.io/gorm"
)

func SeedData(db *gorm.DB) error {

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
        hashedPassword, err := utils.HashPassword(users[i].Password)
        if err != nil {
            log.Printf("Error hashing password for user %s: %v\n", users[i].Name, err)
            tx.Rollback()
            return err
        }
        users[i].Password = hashedPassword

        log.Printf("Seeding user: %s\n", users[i].Name)
        if err := tx.Create(&users[i]).Error; err != nil {
            log.Printf("Error seeding user %s: %v\n", users[i].Name, err)
            tx.Rollback()
            return err
        }
    }

    log.Println("Finished seeding users.")

    tasks := []models.Task{
        {Title: "backup the database", Description: "create the entire backup using incremental", Owner: users[0].ID},
        {Title: "test the restore process", Description: "restore the backup and test the process", Owner: users[0].ID},
        {Title: "supervise things", Description: "invent something to supervise", Owner: users[1].ID},
    }

    // Insert the data into the database
    log.Println("Starting to seed tasks...")
    for _, task := range tasks {
        log.Printf("Seeding task: %s\n", task.Title)
        if err := tx.Create(&task).Error; err != nil {
            log.Printf("Error seeding todo %s: %v\n", task.Title, err)
            tx.Rollback()
        }
    }
    log.Println("Finished seeding tasks.")

    if err := tx.Commit().Error; err != nil {
        tx.Rollback()
        return err
    }    

    return nil
}