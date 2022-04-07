package main

import (
	"errors"
	"github.com/apex/log"
	"google.golang.org/api/youtube/v3"
	"gorm.io/gorm"
	"strings"
	"time"
)

// MetaUpdateParts contains the requested parts for the
// YouTube Data V3 API call
var MetaUpdateParts = []string{
	"contentDetails",
	"id",
	"snippet",
	"statistics",
	"status",
}

var ErrVideoNotFound = errors.New("video not found")

const (
	WorkerCount = 8
)

func updateVideos(service *youtube.Service, db *gorm.DB, videos []*Video) (dl []*Video, err error) {
	jobsChan := make(chan *Video, len(videos))
	resChan := make(chan *Video, len(videos))
	for i := 0; i < WorkerCount; i++ {
		go updateWorker(i, jobsChan, resChan, service, db)
	}

	// distribute jobs
	for _, v := range videos {
		jobsChan <- v
	}
	close(jobsChan)

	// await results and save them to the dl array
	for i := 0; i < len(videos); i++ {
		res := <-resChan
		if res != nil {
			dl = append(dl, res)
		}
	}
	close(resChan)

	return
}

func updateWorker(i int, in, out chan *Video, service *youtube.Service, db *gorm.DB) {
	for {
		select {
		case job, more := <-in:
			if !more {
				log.Infof("[Job %d] Done!", i)
				return
			}
			dl, err := updateJob(service, db, job)
			if err != nil {
				log.WithError(err).Warnf("[Job %d] Failed to update video", i)
			}
			if dl {
				log.Infof("[Job %d] [Video %s] should be downloaded.", i, job.ID)
				out <- job
			} else {
				out <- nil
			}
		}
	}
}

func updateJob(service *youtube.Service, db *gorm.DB, v *Video) (dl bool, err error) {
	var resp *youtube.VideoListResponse
	if resp, err = service.Videos.List(MetaUpdateParts).Id(v.ID).Do(); err != nil {
		return
	}
	if len(resp.Items) < 1 {
		// TODO: Video is private
		return false, ErrVideoNotFound
	}

	var (
		r     = resp.Items[0]
		t     = time.Now()
		check = func(old, new, field string) error {
			if old == new {
				return nil
			}
			return db.Create(&VideoHistory{
				VideoID:   v.ID,
				Field:     field,
				Old:       old,
				New:       new,
				UpdatedAt: t,
			}).Error
		}
	)

	// title
	if err = check(v.Title, r.Snippet.Title, "title"); err != nil {
		return
	}
	v.Title = r.Snippet.Title

	// description
	if err = check(v.Description, r.Snippet.Description, "desc"); err != nil {
		return
	}
	v.Description = r.Snippet.Description

	// tags
	tags := strings.Join(r.Snippet.Tags, ",")
	if err = check(v.Tags, tags, "tags"); err != nil {
		return
	}
	v.Tags = tags

	// video length
	if det := r.ContentDetails; det != nil {
		if err = check(v.VideoLength, det.Duration, "length"); err != nil {
			return
		}
		if v.VideoLength != det.Duration {
			dl = true
		}
		v.VideoLength = det.Duration
	}

	// view count
	if r.Statistics != nil {
		if view := r.Statistics.ViewCount; view != v.ViewCount {
			v.ViewCount = view
			if err = db.Create(&VideoViewCountHistory{
				VideoID: v.ID,
				Views:   view,
				Time:    t,
			}).Error; err != nil {
				return
			}
		}
		if like := r.Statistics.LikeCount; like != v.LikeCount {
			v.LikeCount = like
			if err = db.Create(&VideoLikeCountHistory{
				VideoID: v.ID,
				Likes:   like,
				Time:    t,
			}).Error; err != nil {
				return
			}
		}
		if comment := r.Statistics.ViewCount; comment != v.CommentCount {
			v.CommentCount = comment
			if err = db.Create(&VideoCommentCountHistory{
				VideoID:  v.ID,
				Comments: comment,
				Time:     t,
			}).Error; err != nil {
				return
			}
		}
	}

	err = db.Updates(v).Error
	return
}
