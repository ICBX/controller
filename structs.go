package main

import "time"

type (
	VideoRating   string
	PrivacyStatus string
	BlobType      string
)

//goland:noinspection ALL
const (
	NormalRating        VideoRating = "normal"
	KidsRating          VideoRating = "kids"
	AgeRestrictedRating VideoRating = "age-restricted"
)

//goland:noinspection ALL
const (
	PublicPrivacyStatus   PrivacyStatus = "public"
	PrivatePrivacyStatus  PrivacyStatus = "private"
	UnlistedPrivacyStatus PrivacyStatus = "unlisted"
)

//goland:noinspection ALL
const (
	VideoBlobType     BlobType = "video"
	ThumbnailBlobType BlobType = "thumbnail"
)

////

type User struct {
	ID       uint
	Name     string `gorm:"not null" gorm:"unique"`
	Email    string `gorm:"not null" gorm:"unique"`
	Password string `gorm:"not null"`

	Videos []*Video `gorm:"many2many:VideoUsers"`
}

type Video struct {
	ID            string
	ChannelID     string
	Title         string
	Description   string
	ViewCount     uint
	LikeCount     uint
	CommentCount  uint
	Tags          string
	VideoLength   uint
	Rating        VideoRating
	PublishedAt   *time.Time
	PrivacyStatus PrivacyStatus
	LastUpdated   *time.Time

	Users    []*User           `gorm:"many2many:VideoUsers"`
	Blobbers []*BlobDownloader `gorm:"many2many:VideosBlobDownloader"`
}

type VideoHistory struct {
	ID uint

	VideoID string `gorm:"not null"`
	Video   *Video

	Field     string    `gorm:"not null"`
	Old       string    `gorm:"not null"`
	New       string    `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type BlobDownloader struct {
	ID     uint
	Name   string   `gorm:"not null"`
	Secret string   `gorm:"not null"`
	Videos []*Video `gorm:"many2many:VideosBlobDownloader"`
}

type BlobLocation struct {
	ID uint

	VideoID string `gorm:"not null"`
	Video   *Video

	BlobDownloaderID uint `gorm:"not null"`
	BlobDownloader   *BlobDownloader

	Path    string    `gorm:"not null"`
	AddedAt time.Time `gorm:"not null"`
	Type    BlobType  `gorm:"not null"`
}

type VideoViewCountHistory struct {
	ID    uint
	Views uint64 `gorm:"not null"`

	VideoID string `gorm:"not null"`
	Video   *Video
}

type VideoLikeCountHistory struct {
	ID    uint
	Likes uint64 `gorm:"not null"`

	VideoID string `gorm:"not null"`
	Video   *Video
}

type VideoCommentCountHistory struct {
	ID       uint
	Comments uint64 `gorm:"not null"`

	VideoID string `gorm:"not null"`
	Video   *Video
}
