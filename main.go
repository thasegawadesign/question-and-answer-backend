package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

type User struct {
	ID    uint   `gorm:"primaryKey"`
	Email string `gorm:"unique;not null"`
	QAs   []QAs
}
type QAs struct {
	ID       uint   `gorm:"primaryKey"`
	Question string `gorm:"not null"`
	Answer   string `gorm:"not null"`
	UserID   uint
}

func loadEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Failed to load: %v", err)
	}
}

func initDB() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_PORT"))
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	db.AutoMigrate(&User{}, &QAs{})
}

func main() {
	loadEnv()
	initDB()

	e := echo.New()

	e.GET("/api/items", getItems)
	e.POST("/api/items", createItem)

	e.Logger.Fatal(e.Start(":8080"))
}

func getItems(c echo.Context) error {
	var items []QAs
	if err := db.Preload("User").Find(&items).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch items"})
	}
	return c.JSON(http.StatusOK, items)
}

func createItem(c echo.Context) error {
	item := new(QAs)
	if err := c.Bind(item); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}

	if err := db.Create(&item).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create item"})
	}
	return c.JSON(http.StatusOK, item)
}
