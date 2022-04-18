package rest

import (
	"github.com/ICBX/penguin/pkg/common"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
)

// Disable Video:
// DELETE /media/videos/:id?state=disable
// or
// DELETE /media/videos/:id
// ---
// Enable Video:
// DELETE /media/videos/:id?state=enable
func (s *Server) routeVideoDisable(ctx *fiber.Ctx) (err error) {
	state := ctx.Query("state", "disable")
	id := utils.CopyString(ctx.Params("id"))

	var ac bool
	switch state {
	case "disable":
		ac = true
	case "enable":
		ac = false
	default:
		return fiber.NewError(400, "invalid state. available: enable/disable")
	}

	tx := s.db.Where(&common.Video{
		ID:     id,
		Active: ac,
	}).Updates(&common.Video{
		Active: !ac,
	})
	if err = tx.Error; err != nil {
		return
	}
	if tx.RowsAffected <= 0 {
		return fiber.NewError(404, "video not found or already in requested state")
	}
	return ctx.SendStatus(201)
}
