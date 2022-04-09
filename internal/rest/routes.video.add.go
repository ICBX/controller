package rest

import (
	"github.com/ICBX/penguin/pkg/common"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm/clause"
)

// rest payloads
type newVideoPayload struct {
	UserID  uint   `json:"userID"`
	VideoID string `json:"videoID"`
}

func (s *Server) routeVideoAdd(ctx *fiber.Ctx) (err error) {
	var req newVideoPayload
	if err = ctx.BodyParser(&req); err != nil {
		return
	}

	var user common.User
	if r := s.db.Find(&user, "ID = ?", req.UserID); r.RowsAffected != 1 {
		return common.ErrUserDoesNotExist
	}

	// check if video already in database
	var video common.Video
	if r := s.db.Preload("Users").Find(&video, "ID = ?", req.VideoID); r.RowsAffected == 1 {
		// check if requesting user already added that video
		for _, u := range video.Users {
			if u.ID == req.UserID {
				return common.ErrVideoAlreadyAdded
			}
		}
	} else {
		video.ID = req.VideoID
	}

	// add user to video's user list
	video.Users = append(video.Users, &user)

	// save/update video to DB
	return s.db.Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&video).
		Error
}
