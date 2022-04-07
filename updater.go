package main

import (
	"errors"
	"github.com/apex/log"
	"google.golang.org/api/youtube/v3"
	"gorm.io/gorm"
	"strings"
	"sync"
	"time"
)

var part = []string{
	"contentDetails",
	"id",
	"liveStreamingDetails",
	"localizations",
	"player",
	"recordingDetails",
	"snippet",
	"statistics",
	"status",
	"topicDetails",
}

var ErrVideoNotFound = errors.New("video not found")

func updateVideos(service *youtube.Service, db *gorm.DB) (err error) {
	var videos []*Video
	if err = db.Find(&videos).Error; err != nil {
		return
	}

	jobsChan := make(chan *Video)
	jobsCount := 6

	var wg sync.WaitGroup
	for i := 0; i < jobsCount; i++ {
		wg.Add(1)

		go func(n int) {
			for {
				select {
				case job, more := <-jobsChan:
					if !more {
						wg.Done()
						return
					}
					if err = updateJob(service, db, job); err != nil {
						log.WithError(err).Warnf("[Job %d] Failed to update video", n)
					}
				}
			}
		}(i)
	}

	for _, v := range videos {
		jobsChan <- v
	}
	close(jobsChan)

	wg.Wait()
	return
}

func updateJob(service *youtube.Service, db *gorm.DB, v *Video) (err error) {
	var resp *youtube.VideoListResponse
	if resp, err = service.Videos.List(part).Id(v.ID).Do(); err != nil {
		return
	}
	if len(resp.Items) < 1 {
		return ErrVideoNotFound // TODO: Video is private
	}
	r := resp.Items[0]

	t := time.Now()

	check := func(old, new, field string) error {
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
		v.VideoLength = det.Duration
		// TODO: mark video to download
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

	return
}
