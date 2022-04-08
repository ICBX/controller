package main

import (
	"context"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/gofiber/fiber/v2"
	"github.com/robfig/cron/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"time"
)

func init() {
	log.SetHandler(cli.Default)
	log.SetLevel(log.DebugLevel)
}

func main() {
	// YouTube service
	service, err := youtube.NewService(context.Background(), option.WithAPIKey(os.Getenv("API_KEY")))
	if err != nil {
		log.WithError(err).Fatal("cannot create youtube service")
		return
	}

	// database
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		log.WithError(err).Fatal("cannot open database")
		return
	}
	log.Info("Migrating Database...")
	if err = db.AutoMigrate(
		&APIKey{},
		&User{},
		&Video{},
		&VideoHistory{},
		&BlobDownloader{},
		&BlobLocation{},
		&VideoViewCountHistory{},
		&VideoLikeCountHistory{},
		&VideoCommentCountHistory{},
	); err != nil {
		log.WithError(err).Fatal("cannot migrate db")
		return
	}
	log.Info("OK!")

	c := cron.New(cron.WithSeconds())
	if _, err = c.AddFunc("*/10 * * * * *", func() {
		log.Debug("[Meta-Update] Checking...")

		var videos []*Video
		if err = db.Find(&videos).Error; err != nil {
			log.WithError(err).Warn("[Meta-Update] cannot fetch videos from database")
			return
		}

		log.Infof("[Meta-Update] Updating %d videos...", len(videos))

		swStart := time.Now()
		if videos, err = updateVideos(service, db, videos); err != nil {
			log.WithError(err).Warn("cannot update videos")
		}
		swStop := time.Now()

		log.Infof("[Meta-Update] Done! Took %s. %d videos should be downloaded.",
			swStop.Sub(swStart).String(), len(videos))
	}); err != nil {
		log.WithError(err).Fatal("Cannot start meta updater")
		return
	}

	c.Run()

	// start REST webserver
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	app.Post("/video/add", addNewVideo)

	app.Listen(":3000")

}
