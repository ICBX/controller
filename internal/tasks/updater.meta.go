package tasks

import (
	"database/sql"
	"github.com/ICBX/penguin/pkg/common"
	"github.com/apex/log"
	"google.golang.org/api/youtube/v3"
	"gorm.io/gorm"
	"strconv"
	"strings"
	"time"
)

// metaUpdateParts contains the requested parts for the
// YouTube Data V3 API call
var metaUpdateParts = []string{
	"contentDetails",
	"id",
	"snippet",
	"statistics",
	"status",
}

const (
	WorkerCount = 8
)

func UpdateVideos(service *youtube.Service, db *gorm.DB, videos []*common.Video) (dl []*common.Video, err error) {
	jobsChan := make(chan *common.Video, len(videos))
	resChan := make(chan *common.Video, len(videos))
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

func updateWorker(i int, in, out chan *common.Video, service *youtube.Service, db *gorm.DB) {
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

func updateJob(service *youtube.Service, db *gorm.DB, v *common.Video) (dl bool, err error) {
	var resp *youtube.VideoListResponse
	if resp, err = service.Videos.List(metaUpdateParts).Id(v.ID).Do(); err != nil {
		return
	}

	var (
		t       = time.Now()
		fetched = v.Fetched.Valid && v.Fetched.Bool
		check   = func(fetched bool, old, new, field string) error {
			if !fetched || old == new {
				return nil
			}
			return db.Create(&common.VideoHistory{
				VideoID:   v.ID,
				Field:     field,
				Old:       old,
				New:       new,
				UpdatedAt: t,
			}).Error
		}
	)

	var privacy = v.PrivacyStatus

	if len(resp.Items) > 0 {
		var r = resp.Items[0]

		// published at
		var pa time.Time
		if pa, err = time.Parse(time.RFC3339, r.Snippet.PublishedAt); err != nil {
			log.WithError(err).Warnf("[Video %s] cannot parse published at", v.ID)
		} else {
			v.PublishedAt = &pa
		}

		// load privacy state by api response
		if r.Status != nil {
			if state, ok := common.PrivateStatusByName[r.Status.PrivacyStatus]; ok {
				privacy = state
			}
		}

		if r.Snippet.ChannelId != "" {
			v.ChannelID = r.Snippet.ChannelId
		}

		// title
		if err = check(fetched, v.Title, r.Snippet.Title, "title"); err != nil {
			return
		}
		v.Title = r.Snippet.Title

		// description
		if err = check(fetched, v.Description, r.Snippet.Description, "desc"); err != nil {
			return
		}
		v.Description = r.Snippet.Description

		// tags
		tags := strings.Join(r.Snippet.Tags, ",")
		if err = check(fetched, v.Tags, tags, "tags"); err != nil {
			return
		}
		v.Tags = tags

		// video length
		if det := r.ContentDetails; det != nil {
			if err = check(fetched, v.VideoLength, det.Duration, "length"); err != nil {
				return
			}
			if v.VideoLength != det.Duration {
				dl = true
			}
			v.VideoLength = det.Duration
		}

		// rating
		var rating common.VideoRating
		if r.Status != nil && r.Status.MadeForKids {
			rating = common.KidsRating
		} else if r.ContentDetails.ContentRating.YtRating == "ytAgeRestricted" {
			rating = common.AgeRestrictedRating
		} else {
			rating = common.NormalRating
		}
		if err = check(fetched, strconv.Itoa(int(v.Rating)), strconv.Itoa(int(rating)), "rating"); err != nil {
			return
		}
		v.Rating = rating

		// view count
		if r.Statistics != nil {
			if view := r.Statistics.ViewCount; view != v.ViewCount {
				v.ViewCount = view
				if err = db.Create(&common.VideoViewCountHistory{
					VideoID: v.ID,
					Views:   view,
					Time:    t,
				}).Error; err != nil {
					return
				}
			}
			if like := r.Statistics.LikeCount; like != v.LikeCount {
				v.LikeCount = like
				if err = db.Create(&common.VideoLikeCountHistory{
					VideoID: v.ID,
					Likes:   like,
					Time:    t,
				}).Error; err != nil {
					return
				}
			}
			if comment := r.Statistics.ViewCount; comment != v.CommentCount {
				v.CommentCount = comment
				if err = db.Create(&common.VideoCommentCountHistory{
					VideoID:  v.ID,
					Comments: comment,
					Time:     t,
				}).Error; err != nil {
					return
				}
			}
		}
	} else if fetched {
		// if api doesn't return a video we fetched earlier, the video is most likely private
		privacy = common.PrivatePrivacyStatus
	}

	// video privacy state
	if privacy != v.PrivacyStatus {
		if err = check(
			fetched,
			strconv.Itoa(int(v.PrivacyStatus)),
			strconv.Itoa(int(privacy)),
			"privacy",
		); err != nil {
			return
		}
		v.PrivacyStatus = privacy
	}

	// force set dl to true if not already fetched
	if !fetched {
		dl = true
	}

	// mark video as fetched
	v.Fetched = sql.NullBool{
		Bool:  true,
		Valid: true,
	}

	// update last updated timestamp
	v.LastUpdated = &t

	err = db.Updates(v).Error
	return
}
