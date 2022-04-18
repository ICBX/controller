package main

import (
	"context"
	"github.com/ICBX/penguin/internal/rest"
	"github.com/ICBX/penguin/internal/tasks"
	"github.com/ICBX/penguin/pkg/common"
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/robfig/cron/v3"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func init() {
	log.SetHandler(cli.Default)
	log.SetLevel(log.DebugLevel)
}

func startCron(ctx context.Context, wg *sync.WaitGroup, service *youtube.Service, db *gorm.DB) (err error) {
	defer wg.Done()

	c := cron.New(cron.WithSeconds())
	if _, err = c.AddFunc("0 */1 * * * *", func() {
		log.Debug("[Meta-Update] Checking...")

		var videos []*common.Video
		if err = db.Where(&common.Video{Active: true}).Find(&videos).Error; err != nil {
			log.WithError(err).Warn("[Meta-Update] cannot fetch videos from database")
			return
		}

		log.Infof("[Meta-Update] Updating %d videos...", len(videos))

		swStart := time.Now()
		if videos, err = tasks.UpdateVideos(service, db, videos); err != nil {
			log.WithError(err).Warn("cannot update videos")
		}
		swStop := time.Now()

		log.Infof("[Meta-Update] Done! Took %s. %d videos should be downloaded.",
			swStop.Sub(swStart).String(), len(videos))

		// add videos to download queue
		for _, v := range videos {
			if err = addToQueue(db, v); err != nil {
				return
			}
		}
	}); err != nil {
		log.WithError(err).Fatal("Cannot create updater cronjob")
		return
	}

	go c.Run()
	<-ctx.Done()

	log.Info("[task#update] Shutting down...")
	c.Stop()

	return
}

func startRESTApi(ctx context.Context, wg *sync.WaitGroup, db *gorm.DB) error {
	// start REST webserver
	r := rest.New(db)

	go func() {
		<-ctx.Done()
		log.Info("[srv#rest] Shutting down...")
		if err := r.Shutdown(); err != nil {
			log.WithError(err).Warn("Cannot shutdown REST api")
		}
		wg.Done()
	}()

	return r.Listen(":3000")
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
	if err = db.AutoMigrate(common.TableModels...); err != nil {
		log.WithError(err).Fatal("cannot migrate db")
		return
	}
	log.Info("OK!")

	// services
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup

	log.Info("[SRV] Starting service cron#updater")
	wg.Add(1)
	go func() {
		err := startCron(ctx, &wg, service, db)
		if err != nil {
			stop()
			log.WithError(err).Warn("Cannot start cron service")
		}
	}()

	log.Info("[SRV] Starting service api#rest")
	wg.Add(1)
	go func() {
		err := startRESTApi(ctx, &wg, db)
		if err != nil {
			if err != nil {
				stop()
				log.WithError(err).Warn("Cannot start rest service")
			}
		}
	}()

	wg.Wait()
	log.Info("All Services Shut Down.")
}

func addToQueue(db *gorm.DB, v *common.Video) (err error) {
	// fetch all blobbers for the video
	if err = db.Preload("Blobbers").Where(v).First(v).Error; err != nil {
		return
	}
	// add video to queue
	for _, b := range v.Blobbers {
		log.Infof("Adding video %s to blobber-queue %d", v.ID, b.ID)
		if err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(&common.Queue{
			VideoID:   v.ID,
			BlobberID: b.ID,
		}).Error; err != nil {
			return
		}
	}
	return
}
