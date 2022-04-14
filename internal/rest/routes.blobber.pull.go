package rest

import (
	"github.com/ICBX/penguin/pkg/common"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

func (s *Server) routeBlobberPull(ctx *fiber.Ctx) (err error) {

	// get blobber id from route
	blobberID := ctx.Params("id")
	if blobberID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Blobber ID is required")
	}

	// get blobber secret from headers
	headers := ctx.GetReqHeaders()
	blobberSecret, ok := headers["Blobber-Secret"]
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "no blobberID specified")
	}

	var blobberIDUint uint
	blobberIDUint, err = convertStringToUint(blobberID)
	if err != nil {
		return
	}

	// check if blobber id exists and secret is correct
	if r := s.db.Where(&common.BlobDownloader{ID: uint(blobberIDUint), Secret: blobberSecret}).First(&common.BlobDownloader{}); r.RowsAffected != 1 {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid blobberID or secret")
	}

	// return a list of videos to download
	var videoIDS = make([]string, 0)
	var queues []*common.Queue
	s.db.Find(&queues, "blobber_id = ?", uint(blobberIDUint))

	// add all video ids
	for _, q := range queues {
		videoIDS = append(videoIDS, q.VideoID)
	}

	err = ctx.JSON(videoIDS)
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
