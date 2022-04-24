package rest

import (
	"errors"
	"github.com/ICBX/penguin/pkg/common"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/utils"
	"gorm.io/gorm"
	"strconv"
)

type BlobberPullResponse struct {
	Download []string `json:"download"`
	Remove   []string `json:"remove"`
}

func (s *Server) routeBlobberPull(ctx *fiber.Ctx) (err error) {
	// get blobber id from route
	blobberID := utils.CopyString(ctx.Params(BlobberIDKey))
	if blobberID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Blobber ID is required")
	}
	var blobberIDUint uint
	if blobberIDUint, err = convertStringToUint(blobberID); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// get blobber secret from headers
	blobberSecret := ctx.Get("Blobber-Secret")
	if blobberSecret == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "no blobberID specified")
	}

	// check if blobber id exists and secret is correct
	if err = s.db.Where(&common.BlobDownloader{
		ID:     blobberIDUint,
		Secret: blobberSecret,
	}).First(&common.BlobDownloader{}).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid blobberID or secret")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// return a list of videos to download
	var download []*common.Queue
	if err = s.db.Where(&common.Queue{BlobberID: blobberIDUint, Action: common.GetBlob}).Find(&download).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// return a list of videos to remove
	var remove []*common.Queue
	if err = s.db.Where(&common.Queue{BlobberID: blobberIDUint, Action: common.RemoveBlob}).Find(&remove).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// collect video ids to download
	var videoIDsDownload = make([]string, len(download))
	for i, q := range download {
		videoIDsDownload[i] = q.VideoID
	}

	// collect video ids to remove
	var videoIDsRemove = make([]string, len(remove))
	for i, q := range remove {
		videoIDsRemove[i] = q.VideoID
	}

	err = ctx.Status(fiber.StatusOK).JSON(BlobberPullResponse{
		Download: videoIDsDownload,
		Remove:   videoIDsRemove,
	})
	return
}

// TODO: move to util
func convertStringToUint(s string) (uint, error) {
	u, err := strconv.ParseUint(s, 10, 0)
	if err != nil {
		return 0, err
	}
	return uint(u), nil
}
