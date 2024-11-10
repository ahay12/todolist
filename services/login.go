package services

import (
	"database/sql"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"os"
	"time"
	"todolist/database"
	"todolist/helper"
	"todolist/models"
)

var jwtSecret = os.Getenv("API_KEY")

func Login(c *fiber.Ctx) error {
	type LoginInput struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	var input LoginInput
	if err := c.BodyParser(&input); err != nil {
		helper.RespondJSON(c, fiber.StatusBadRequest, "Cannot parse JSON", nil, err.Error())
		return err
	}

	var user models.User
	// Query the user by username
	err := database.DB.QueryRow("SELECT id, username, password FROM users WHERE username = :1", input.Username).
		Scan(&user.ID, &user.Username, &user.Password)
	if err == sql.ErrNoRows {
		helper.RespondJSON(c, fiber.StatusNotFound, "User not found", nil, nil)
		return nil
	} else if err != nil {
		helper.RespondJSON(c, fiber.StatusInternalServerError, "Database error", nil, err.Error())
		return err
	}

	// Check if the password is correct
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		helper.RespondJSON(c, fiber.StatusUnauthorized, "Invalid password", nil, nil)
		return nil
	}

	// Generate JWT token
	token, err := generateJwt(user.ID, user.Username)
	if err != nil {
		helper.RespondJSON(c, fiber.StatusInternalServerError, "Failed to generate JWT", nil, err.Error())
		return err
	}

	// Respond with the token
	helper.RespondJSON(c, fiber.StatusOK, "Login successful", map[string]string{"token": token}, nil)
	return nil
}

func generateJwt(id uint, username string) (string, error) {
	claims := jwt.MapClaims{
		"userId":   id,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}
