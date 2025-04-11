package main

import (
	"fmt"
	"log"
	"os"

	"github.com/iamgak/go-drive/models"
	"github.com/iamgak/go-drive/pkg"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// create a connection
func openDBORM() (*gorm.DB, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, pkg.ErrNoEnvFileFound
	}

	dbUser := os.Getenv("DB_USERNAME")
	dbName := os.Getenv("DB_DATABASE")
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbUser == "" || dbName == "" || dbPassword == "" {
		return nil, pkg.ErrNoEnvFileFound
	}

	dsn := fmt.Sprintf("%s:%s@/%s?parseTime=true", dbUser, dbPassword, dbName)
	return gorm.Open(mysql.Open(dsn), &gorm.Config{})
}

// create migration
func MigrateDB(DB *gorm.DB) {
	err := DB.AutoMigrate(
		&models.User{},
		&models.UsersSession{},
		&models.UserActivityLog{},
	)
	if err != nil {
		log.Fatal("Migration failed:", err)
	}
}
