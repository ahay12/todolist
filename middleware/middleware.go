package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"strings"
	"todolist/helper"
)

var jwtSecret = os.Getenv("API_KEY")

func Auth(c *fiber.Ctx) error {

	tokenString := c.Get("Authorization")
	if tokenString == "" {
		c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Missing or invalid token"})
		return nil
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) { return []byte(jwtSecret), nil })
	if err != nil || !token.Valid {
		helper.RespondJSON(c, fiber.StatusUnauthorized, "Unauthorized", nil, err.Error())
		return nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		helper.RespondJSON(c, fiber.StatusUnauthorized, "Invalid token", nil, nil)
		return nil
	}

	userId, userIdExists := claims["userId"].(float64)
	if !userIdExists {
		helper.RespondJSON(c, fiber.StatusUnauthorized, "User ID is required", nil, nil)
		return nil
	}

	c.Locals("userId", uint(userId))
	return c.Next()
}
