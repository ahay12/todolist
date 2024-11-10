package router

import (
	"bytes"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func setupApp() *fiber.App {
	app, _ := Make()
	return app
}

func TestCreateTodo(t *testing.T) {
	app := setupApp()

	todoData := map[string]string{
		"title":       "Test Todo",
		"description": "This is a test todo",
		"status":      "pending",
		"due_date":    "2021-12-31",
	}
	jsonData, _ := json.Marshal(todoData)
	req, _ := http.NewRequest("POST", "/api/v1/todo", bytes.NewReader(jsonData))

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to send request: %v", err)
	}

	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}
