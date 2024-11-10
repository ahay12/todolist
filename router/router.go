package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"log"
	"os"
	"time"
	"todolist/handler"
	"todolist/middleware"
	"todolist/services"
)

func Cors() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		ctx.Set("Access-Control-Allow-Origin", "*")
		ctx.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		ctx.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		return ctx.Next()
	}
}

func Make() (*fiber.App, *os.File) {
	logFile, err := os.OpenFile("server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	app := fiber.New()

	app.Use(logger.New(logger.Config{
		// store logs
		Output: logFile,
		// the log format
		Format: "[${time}] ${status} - ${method} ${path} - ${latency} ${locals:request_body} ${locals:response_body}\n",
		// time format
		TimeFormat: time.RFC3339,
		// Skip logging for health checks or specific routes if needed
		Next: func(c *fiber.Ctx) bool {
			return c.Path() == "/health"
		},
	}))

	app.Use(Cors())
	v1 := app.Group("/api/v1")
	{
		v1.Get("/todos", middleware.Auth, handler.GetAllTodosHandler)
		v1.Get("/todo/:id", middleware.Auth, handler.GetTodoByIDHandler)
		v1.Post("/todo", middleware.Auth, handler.CreateTodoHandler)
		v1.Put("/todo/:id", middleware.Auth, handler.UpdateTodoHandler)
		v1.Delete("/todo/:id", middleware.Auth, handler.DeleteTodoHandler)
		v1.Post("/login", services.Login)
		v1.Post("/register", handler.CreateUserHandler)
	}

	return app, logFile
}
