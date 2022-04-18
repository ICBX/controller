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

	// create blobber
	if err = s.db.Create(&common.BlobDownloader{
		Name:   req.Name,
		Secret: req.Secret,
	}).Error; err != nil {
		return
	}

	return
}
