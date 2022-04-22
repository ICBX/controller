package rest

import (
	"github.com/ICBX/penguin/pkg/common"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/gorm"
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

	where := &common.Video{
		ID: utils.CopyString(ctx.Params("id")),
	}

	var tx *gorm.DB
	if state == "disable" {
		tx = s.db.Delete(where)
	} else if state == "enable" {
		tx = s.db.Where(where).Update("deleted_at", nil)
	} else {
		return fiber.NewError(fiber.StatusBadRequest, "invalid state (enable/disable)")
	}
	if err = tx.Error; err != nil {
		return
	}

	if tx.RowsAffected <= 0 {
		return fiber.NewError(fiber.StatusNotFound, "video not found or already in requested state")
	}
	return ctx.Status(201).SendString("video " + state + "d") // <- that's illegal! refactor later.
}
