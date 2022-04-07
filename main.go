package main

import (
	"context"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/robfig/cron/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"strings"
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

	// find all michael reeves videos
	/*
		res, err := service.Search.List([]string{"snippet"}).
			ChannelId("UCtHaxi4GTYDpJgMSGy7AeSw").
			MaxResults(50).
			Do()
		if err != nil {
			log.WithError(err).Warn("cannot query videos")
		} else {
			for _, s := range res.Items {
				var count int64
				if err = db.Model(&Video{}).
					Where("id = ?", s.Id.VideoId).
					Count(&count).Error; err != nil {
					log.WithError(err).Warn("cannot get count")
				} else if count == 0 {
					log.Infof("Inserting video %s", s.Id.VideoId)
					if err = db.Create(&Video{ID: s.Id.VideoId, ChannelID: s.Id.ChannelId}).Error; err != nil {
						log.WithError(err).Warn("cannot insert video")
					}
				}
			}
		}
	*/

	c := cron.New(cron.WithSeconds())
	if _, err = c.AddFunc("*/5 * * * * *", func() {
		log.Debug("checking...")

		var videos []*Video
		if err = db.Find(&videos).Error; err != nil {
			log.WithError(err).Warn("[Meta-Update] cannot fetch videos from database")
			return
		}
		log.Infof("[Meta-Update] Updating %d videos...", len(videos))

		swStart := time.Now()
		for i, v := range videos {
			metas, err := service.Videos.List(part).
				Id(v.ID).
				Do()
			if err != nil {
				log.WithError(err).Warnf("[Meta-Update] cannot get meta for video #%d: %s", i, v.ID)
				continue
			}
			if len(metas.Items) < 1 {
				log.Warnf("[Meta-Update] Cannot find video #%d %s. Is it private?", i, v.ID)
				// TODO: private video
				continue
			}
			meta := metas.Items[0]

			// compare meta
			v.Title = meta.Snippet.Title
			v.Description = meta.Snippet.Description
			v.ViewCount = meta.Statistics.ViewCount
			v.LikeCount = meta.Statistics.LikeCount
			v.CommentCount = meta.Statistics.CommentCount
			v.Tags = strings.Join(meta.Snippet.Tags, ",")
			// TODO: Video length

			if meta.Status != nil && meta.Status.MadeForKids {
				v.Rating = KidsRating
			} else if meta.ContentDetails.ContentRating.YtRating == "2022-03-31T23:49:49Z" {
				v.Rating = AgeRestrictedRating
			} else {
				v.Rating = NormalRating
			}

			publishedAt, err := time.Parse(time.RFC3339, meta.Snippet.PublishedAt)
			if err != nil {
				log.WithError(err).Warn("[Meta-Update] cannot parse time")
			} else {
				v.PublishedAt = &publishedAt
			}

			if meta.Status != nil {
				privacy := meta.Status.PrivacyStatus
				switch privacy {
				case "public":
					v.PrivacyStatus = PublicPrivacyStatus
				case "unlisted":
					v.PrivacyStatus = UnlistedPrivacyStatus
				}
			}

			t := time.Now()
			v.LastUpdated = &t

			if err = db.Updates(v).Error; err != nil {
				log.WithError(err).Warn("[Meta-Update] cannot update video")
			}
		}
		log.Infof("[Meta-Update] Done! Took %s", time.Now().Sub(swStart).String())
	}); err != nil {
		log.WithError(err).Fatal("Cannot start meta updater")
		return
	}

	c.Run()
}
