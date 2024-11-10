package helper

import "github.com/gofiber/fiber/v2"

type ResponseData struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Task    interface{} `json:"task"`
	Error   interface{} `json:"error"`
}
type ErrorField struct {
	ID      string `json:"id"`
	Value   string `json:"value"`
	Caused  string `json:"caused"`
	Message string `json:"message"`
}

func RespondJSON(ctx *fiber.Ctx, status int, message string, payload interface{}, errors interface{}) {
	res := ResponseData{
		Success: status >= 200 && status < 300,
		Message: message,
		Task:    payload,
		Error:   errors,
	}

	// Send the JSON response and ignore the error since the response is the primary concern
	if err := ctx.Status(status).JSON(res); err != nil {
		// Log the error and send a generic internal server error response
		ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Internal Server Error"})
	}
}
