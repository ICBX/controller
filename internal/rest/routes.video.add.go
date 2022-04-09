package rest

import (
	"github.com/ICBX/penguin/pkg/common"
	"github.com/gofiber/fiber/v2"
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

	user := common.User{}
	r := s.db.Find(&user, "ID = ?", req.UserID)

	if r.RowsAffected != 1 {
		return common.ErrUserDoesntExist
	}

	video := common.Video{
		ID: req.VideoID,
	}

	// check if video already in database
	r = s.db.Find(&video, "ID = ?", req.VideoID)
	if r.RowsAffected == 1 {

		// check if requesting user already added that video
		for _, u := range video.Users {
			if u.ID == req.UserID {
				return common.ErrVideoAlreadyAdded
			}
		}

		// otherwise add requesting user to userlist
		video.Users = append(video.Users, &user)
		s.db.Updates(video)
	}

	// otherwise add empty video to db
	video.Users = append(video.Users, &user)
	s.db.Create(&video)

	return nil
}
