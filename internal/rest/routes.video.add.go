package rest

import (
	"errors"
	"github.com/ICBX/penguin/pkg/common"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// rest payloads
type newVideoPayload struct {
	VideoID  string `json:"videoID"`
	Blobbers []uint `json:"blobbers"`
}

func (s *Server) routeVideoAdd(ctx *fiber.Ctx) (err error) {
	var req newVideoPayload
	if err = ctx.BodyParser(&req); err != nil {
		return
	}

	video := &common.Video{ID: req.VideoID}

	// check if video already in database
	if err = s.db.Where(video).First(&common.Video{}).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	// add blobbers to video
	for _, bid := range req.Blobbers {
		var blobber common.BlobDownloader
		if err = s.db.Where(&common.BlobDownloader{
			ID: bid,
		}).First(&blobber).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return fiber.NewError(fiber.StatusInternalServerError, err.Error())
			}
		}
		video.Blobbers = append(video.Blobbers, &blobber)
	}

	if err = s.db.Create(video).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return ctx.Status(fiber.StatusCreated).SendString("video created")
}
