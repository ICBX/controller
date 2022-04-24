package rest

import (
	"errors"
	"github.com/ICBX/penguin/pkg/common"
	"github.com/apex/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/gorm"
	"net/http"
)

// rest payloads
type newVideoBlobberPayload struct { // TODO: parameter should be enough
	BlobberID uint `json:"blobberID"`
}

func (s *Server) routeVideoAddBlobber(ctx *fiber.Ctx) (err error) {
	var req newVideoBlobberPayload
	if err = ctx.BodyParser(&req); err != nil {
		return
	}

	videoId := utils.CopyString(ctx.Params(VideoIDKey))
	if videoId == "" {
		return fiber.NewError(http.StatusBadRequest, "videoID missing")
	}

	// get video
	var video *common.Video
	if err = s.db.Where(&common.Video{
		ID: videoId,
	}).First(&video).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "Video doesn't exists")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// get blobber
	var blobber common.BlobDownloader
	if err = s.db.Where(&common.BlobDownloader{
		ID: req.BlobberID,
	}).First(&blobber).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "blobber doesn't exists")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// add initial download queue entry for video and blobber
	if err = s.db.Create(&common.Queue{
		VideoID:   videoId,
		BlobberID: req.BlobberID,
		Action:    common.GetBlob,
	}).Error; err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	// add blobber to video
	video.Blobbers = append(video.Blobbers, &blobber)
	s.db.Updates(video)

	log.Infof("Added blobber '%n' (%s) for video '%s' (%s)", blobber.Name, blobber.ID, video.Title, video.ID)

	return ctx.Status(fiber.StatusCreated).SendString("blobber added for video")
}
