package controllers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"taskflow/backend/internal/config"
	"taskflow/backend/internal/database"
	"taskflow/backend/internal/security"
)

type RegisterInput struct {
	Name     string `json:"name"     binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginInput struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register handles user registration.
func Register() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input RegisterInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hash, err := security.HashPassword(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}

		var userID string
		var createdAt time.Time
		err = database.DB.QueryRow(
			`INSERT INTO users (name, email, password_hash) VALUES ($1, $2, $3) RETURNING id, created_at`,
			input.Name, input.Email, hash,
		).Scan(&userID, &createdAt)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}

		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": userID,
			"email":   input.Email,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		})

		signed, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"token": signed,
			"user": gin.H{
				"id":    userID,
				"name":  input.Name,
				"email": input.Email,
			},
		})
	}
}

// Login authenticates a user and returns a JWT.
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		var input LoginInput
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var (
			userID string
			name   string
			email  string
			hash   string
		)
		err := database.DB.QueryRow(
			`SELECT id, name, email, password_hash FROM users WHERE email = $1`,
			input.Email,
		).Scan(&userID, &name, &email, &hash)

		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}

		if err := security.VerifyPassword(hash, input.Password); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": userID,
			"email":   email,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		})

		signed, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to sign token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token": signed,
			"user": gin.H{
				"id":    userID,
				"name":  name,
				"email": email,
			},
		})
	}
}

// GetProfile returns the authenticated user's profile.
func GetProfile() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")

		var (
			id        string
			name      string
			email     string
			createdAt time.Time
		)
		err := database.DB.QueryRow(
			`SELECT id, name, email, created_at FROM users WHERE id = $1`, userID,
		).Scan(&id, &name, &email, &createdAt)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id":         id,
			"name":       name,
			"email":      email,
			"created_at": createdAt,
		})
	}
}
