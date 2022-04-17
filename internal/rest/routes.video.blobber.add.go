package rest

import (
	"github.com/ICBX/penguin/pkg/common"
	"github.com/apex/log"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"net/http"
)

// rest payloads
type newVideoBlobberPayload struct {
	BlobberID uint `json:"blobberId"`
}

func (s *Server) routeVideoAddBlobber(ctx *fiber.Ctx) (err error) {
	var req newVideoBlobberPayload
	if err = ctx.BodyParser(&req); err != nil {
		return
	}
	videoId := ctx.Params("id")
	if videoId == "" {
		return fiber.NewError(http.StatusBadRequest, "videoID missing")
	}

	// get video
	var video *common.Video
	if err = s.db.Where(&common.Video{
		ID: videoId,
	}).First(&video).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusConflict, "Video doesn't exists")
		}
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	// get blobber
	var blobber *common.BlobDownloader
	if err = s.db.Where(&common.BlobDownloader{
		ID: req.BlobberID,
	}).First(&blobber).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusConflict, "blobber doesn't exists")
		}
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	// check if queue record already exists
	var q *common.Queue
	if err = s.db.Where(&common.Queue{
		VideoID:   videoId,
		BlobberID: req.BlobberID,
	}).First(&q).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusConflict, "Queue record already exists")
		}
	}

	// check if BlobLocation record already exists
	var bl *common.BlobLocation
	if err = s.db.Where(&common.BlobLocation{
		VideoID:          videoId,
		BlobDownloaderID: req.BlobberID,
		Type:             common.VideoBlobType,
	}).First(&bl).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusConflict, "BlobLocation record already exists")
		}
	}

	// add queue entry for video and blobber
	if err = s.db.Create(&common.Queue{
		VideoID:   videoId,
		BlobberID: req.BlobberID,
	}).Error; err != nil {
		return fiber.NewError(fiber.StatusConflict, err.Error())
	}

	// add blobber to video
	video.Blobbers = append(video.Blobbers, blobber)
	s.db.Updates(video)

	log.Infof("Added blobber '%n' (%s) for video '%s' (%s)", blobber.Name, blobber.ID, video.Title, video.ID)

	return nil
}
