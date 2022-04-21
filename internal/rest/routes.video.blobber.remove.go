package rest

import (
	"github.com/ICBX/penguin/pkg/common"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func (s *Server) routeVideoRemoveBlobber(ctx *fiber.Ctx) (err error) {

	videoID := ctx.Params("id")
	blobberID := ctx.Params("blobberID")

	// convert blobberID to int
	blobberIDU, err := convertStringToUint(blobberID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Could not process blobber id")
	}

	var video common.Video
	if err = s.db.Find(&video, "ID = ?", videoID).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Did not find video")
	}

	var blobber common.BlobDownloader
	if err = s.db.Find(&blobber, "ID = ?", blobberIDU).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "Did not find blobber")
	}

	// remove blobber from video
	if err = s.db.Model(&video).Association("Blobbers").Delete(&blobber); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Could not remove blobber from video: "+err.Error())
	}

	// remove blobLocation for blobberID and videoID
	blobLocation := common.BlobLocation{
		VideoID:          videoID,
		BlobDownloaderID: blobberIDU,
	}
	if err = s.db.Find(&blobLocation).Delete(&blobLocation).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
	}

	// TODO: add video to blobber 'remove' queue

	return
}
