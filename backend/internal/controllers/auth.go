package controllers

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/omkargeedh/taskflow/internal/database"
	"golang.org/x/crypto/bcrypt"
)

type RegisterInput struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginInput struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register handles user registration
func Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input RegisterInput

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		var userID int
		err = database.DB.QueryRow(
			`INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id`,
			input.Name, input.Email, string(hash),
		).Scan(&userID)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"user_id": userID, "email": input.Email})
	}
}

func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var (
			userID int
			email  string
			hash   string
		)
		err := database.DB.QueryRow(
			`SELECT id, email, password_hash FROM users WHERE email = $1`,
			input.Email,
		).Scan(&userID, &email, &hash)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(input.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": userID,
			"email":   email,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		})

		signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"token": signed})
	}
}
