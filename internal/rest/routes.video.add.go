package rest

import (
	"github.com/gofiber/fiber/v2"
)

// rest payloads
type newVideoPayload struct {
	User    string `json:"user"`
	VideoID string `json:"video"`
}

func (s *Server) routeVideoAdd(ctx *fiber.Ctx) (err error) {
	var req newVideoPayload

	if err = ctx.BodyParser(&req); err != nil {
		return
	}

	// TODO: add new video to download queue
	return nil
}
