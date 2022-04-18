package rest

import (
	"github.com/ICBX/penguin/pkg/common"
	"github.com/gofiber/fiber/v2"
	"net/http"
	"strconv"
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
	var count int64
	if err = s.db.Model(video).Where(video).Count(&count).Error; err != nil {
		return
	}
	if count > 0 {
		return fiber.NewError(http.StatusConflict, "video already added")
	}

	// add blobbers to video
	for _, bid := range req.Blobbers {
		var blobber *common.BlobDownloader
		if err = s.db.Where(&common.BlobDownloader{ID: bid}).First(&blobber).Error; err != nil {
			return
		}
		if blobber == nil {
			return fiber.NewError(404, "blobber with id "+strconv.Itoa(int(bid))+" not found")
		}
		video.Blobbers = append(video.Blobbers, blobber)
	}

	video.Active = true

	// save/update video to DB
	return s.db.Create(video).Error
}
