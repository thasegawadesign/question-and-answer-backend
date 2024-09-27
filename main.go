package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"unique;not null"`
	Provider string `json:"provider"`
}
type QA struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	User      User   `json:"user" gorm:"foreignKey:UserEmail;references:Email;constraint:OnDelete:CASCADE"`
	UserEmail string `json:"user_email" gorm:"not null"`
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
	db.AutoMigrate(&User{}, &QA{})
}

func main() {
	loadEnv()
	initDB()

	e := echo.New()

	goEnv := os.Getenv("GO_ENV")
	cookieSecure := false
	if goEnv == "production" {
		cookieSecure = true
	}

	e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
		TokenLookup:    "header:X-CSRF-Token",
		CookieName:     "_csrf",
		CookiePath:     "/",
		CookieSecure:   cookieSecure,
		CookieHTTPOnly: true,
	}))

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "https://question-and-answer.gojiyuuniotorikudasai.com", "https://question-and-answer-alpha.vercel.app"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
	}))

	e.GET("/api/csrf-token", getCsrfToken)
	e.GET("/api/users", getUserByEmail)
	e.GET("/api/items", getItemsByEmail)
	e.POST("/api/users", createUser)
	e.POST("/api/items", createItem)
	e.PUT("/api/items", updateItem)
	e.DELETE("/api/items", deleteItemById)

	e.Logger.Fatal(e.Start(":8080"))
}

func getCsrfToken(c echo.Context) error {
	csrfToken := c.Get("csrf").(string)
	return c.JSON(http.StatusOK, map[string]string{
		"csrf_token": csrfToken,
	})
}

func getItemsByEmail(c echo.Context) error {
	email := c.QueryParam("email")
	if email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}
	var items []QA
	if err := db.Preload("User").Where("user_email = ?", email).Find(&items).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch items for the specified user"})
	}
	return c.JSON(http.StatusOK, items)
}

func createItem(c echo.Context) error {
	item := new(QA)
	if err := c.Bind(item); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	if err := db.Create(&item).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create item"})
	}
	return c.JSON(http.StatusOK, item)
}

func deleteItemById(c echo.Context) error {
	var request struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
	}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if request.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}
	if err := db.Where("id = ? AND user_email = ?", request.ID, request.Email).Delete(&QA{}).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to delete item"})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Item deleted successfully"})
}

func updateItem(c echo.Context) error {
	var request struct {
		ID       uint   `json:"id"`
		Email    string `json:"email"`
		Question string `json:"question"`
		Answer   string `json:"answer"`
	}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	if request.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}
	var item QA
	if err := db.Where("id = ? AND user_email = ?", request.ID, request.Email).First(&item).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Item not found"})
	}
	item.Question = request.Question
	item.Answer = request.Answer
	if err := db.Save(&item).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update item"})
	}
	return c.JSON(http.StatusOK, item)
}

func getUserByEmail(c echo.Context) error {
	email := c.QueryParam("email")
	var user User
	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return c.JSON(http.StatusNotFound, echo.Map{"message": "User not found"})
	}
	return c.JSON(http.StatusOK, user)
}

func createUser(c echo.Context) error {
	user := new(User)

	if err := c.Bind(user); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid request"})
	}
	if result := db.Create(&user); result.Error != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to create user"})
	}

	return c.JSON(http.StatusCreated, user)
}
