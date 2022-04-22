package rest

import (
	"github.com/ICBX/penguin/pkg/common"
	"github.com/gofiber/fiber/v2"
)

// rest payload
type newBlobberPayload struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

func (s *Server) routeBlobberAdd(ctx *fiber.Ctx) (err error) {
	var req newBlobberPayload
	if err = ctx.BodyParser(&req); err != nil {
		return
	}

	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadGateway, "name required")
	}
	if req.Secret == "" {
		return fiber.NewError(fiber.StatusBadRequest, "secret required")
	}

	if err = s.db.Create(&common.BlobDownloader{
		Name:   req.Name,
		Secret: req.Secret,
	}).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.Status(201).SendString("")
}
