package rest

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Server struct {
	db  *gorm.DB
	app *fiber.App
}

const (
	VideoIDKey   = "video_id"
	BlobberIDKey = "blobber_id"
)

const (
	MediaPrefix         = "/media"
	MediaVideoPrefix    = MediaPrefix + "/video"
	SpecificVideoPrefix = MediaVideoPrefix + "/:" + VideoIDKey

	BlobberPrefix         = "/blobber"
	SpecificBlobberPrefix = BlobberPrefix + "/:" + BlobberIDKey
)

// routes
const (
	RouteAddVideo               = MediaVideoPrefix                            // POST
	RouteDeleteVideo            = SpecificVideoPrefix                         // DELETE
	RouteAddBlobberToVideo      = SpecificVideoPrefix + SpecificBlobberPrefix // POST
	RouteRemoveBlobberFromVideo = SpecificVideoPrefix + SpecificBlobberPrefix // DELETE

	RouteAddBlobber  = BlobberPrefix
	RouteBlobberPull = SpecificBlobberPrefix + "/pull"
)

func New(db *gorm.DB) (s *Server) {
	app := fiber.New(fiber.Config{})
	s = &Server{
		db:  db,
		app: app,
	}

	// TODO: Add routes below 👇
	app.Get("/", s.routeIndex)
	// video
	app.Post(RouteAddVideo, s.routeVideoAdd)                           // add video
	app.Delete(RouteDeleteVideo, s.routeVideoDisable)                  // remove video
	app.Post(RouteAddBlobberToVideo, s.routeVideoAddBlobber)           // add blobber to video
	app.Delete(RouteRemoveBlobberFromVideo, s.routeVideoRemoveBlobber) // remove blobber from video
	// blobber
	app.Post(RouteAddBlobber, s.routeBlobberAdd)  // add blobber
	app.Get(RouteBlobberPull, s.routeBlobberPull) // pull blobber queue
	// TODO: Add routes above 👆

	return
}

func (s *Server) Listen(addr string) error {
	return s.app.Listen(addr)
}

func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
