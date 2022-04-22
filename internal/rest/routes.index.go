package rest

import "github.com/gofiber/fiber/v2"

func (s *Server) routeIndex(c *fiber.Ctx) error {
	return c.Status(fiber.StatusTeapot).SendString("Hello World 🐧!")
}
