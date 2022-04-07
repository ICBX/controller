package main

import "time"

type (
	VideoRating   uint
	PrivacyStatus uint
	BlobType      uint
)

//goland:noinspection ALL
const (
	NormalRating VideoRating = iota + 1
	KidsRating
	AgeRestrictedRating
)

//goland:noinspection ALL
const (
	PublicPrivacyStatus PrivacyStatus = iota + 1
	PrivatePrivacyStatus
	UnlistedPrivacyStatus
)

//goland:noinspection ALL
const (
	VideoBlobType BlobType = iota + 1
	ThumbnailBlobType
)

////

type APIKey struct {
	ID      uint   `gorm:"primaryKey,autoIncrement"`
	Key     string `gorm:"not null"`
	Comment string
}

type User struct {
	ID       uint   `gorm:"primaryKey,autoIncrement"`
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
	ViewCount     uint64
	LikeCount     uint64
	CommentCount  uint64
	Tags          string
	VideoLength   string
	Rating        VideoRating
	PublishedAt   *time.Time
	PrivacyStatus PrivacyStatus
	LastUpdated   *time.Time

	Users    []*User           `gorm:"many2many:VideoUsers"`
	Blobbers []*BlobDownloader `gorm:"many2many:VideosBlobDownloader"`
}

type VideoHistory struct {
	ID uint `gorm:"primaryKey,autoIncrement"`

	VideoID string `gorm:"not null"`
	Video   *Video

	Field     string    `gorm:"not null"`
	Old       string    `gorm:"not null"`
	New       string    `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

type BlobDownloader struct {
	ID     uint     `gorm:"primaryKey,autoIncrement"`
	Name   string   `gorm:"not null"`
	Secret string   `gorm:"not null"`
	Videos []*Video `gorm:"many2many:VideosBlobDownloader"`
}

type BlobLocation struct {
	ID uint `gorm:"primaryKey,autoIncrement"`

	VideoID string `gorm:"not null"`
	Video   *Video

	BlobDownloaderID uint `gorm:"not null"`
	BlobDownloader   *BlobDownloader

	Path    string    `gorm:"not null"`
	AddedAt time.Time `gorm:"not null"`
	Type    BlobType  `gorm:"not null"`
}

type VideoViewCountHistory struct {
	ID      uint   `gorm:"primaryKey,autoIncrement"`
	VideoID string `gorm:"not null"`
	Video   *Video
	Views   uint64    `gorm:"not null"`
	Time    time.Time `gorm:"not null"`
}

type VideoLikeCountHistory struct {
	ID      uint   `gorm:"primaryKey,autoIncrement"`
	VideoID string `gorm:"not null"`
	Video   *Video
	Likes   uint64    `gorm:"not null"`
	Time    time.Time `gorm:"not null"`
}

type VideoCommentCountHistory struct {
	ID       uint   `gorm:"primaryKey,autoIncrement"`
	VideoID  string `gorm:"not null"`
	Video    *Video
	Comments uint64    `gorm:"not null"`
	Time     time.Time `gorm:"not null"`
}
