package response

import "github.com/gofiber/fiber/v3"

// Envelope adalah format response JSON yang konsisten di seluruh API.
type Envelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func Success(c fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(Envelope{Success: true, Message: message, Data: data})
}

func Error(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(Envelope{Success: false, Error: message})
}
