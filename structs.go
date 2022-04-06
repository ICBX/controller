package controller

import "time"

type User struct {
	ID       uint
	Name     string   `gorm:"not null" gorm:"unique"`
	Email    string   `gorm:"not null" gorm:"unique"`
	Password string   `gorm:"not null"`
	Videos   []*Video `gorm:"many2many:VideosUsers"`
}

type Video struct {
	ID            string
	ChannelID     string    `gorm:"not null"`
	Title         string    `gorm:"not null"`
	Description   string    `gorm:"not null"`
	ViewCount     uint      `gorm:"not null"`
	LikeCount     uint      `gorm:"not null"`
	CommentCount  uint      `gorm:"not null"`
	Tags          string    `gorm:"not null"`
	VideoLength   uint      `gorm:"not null"`
	Rating        string    `gorm:"not null"` // TODO: investigate the actual type
	PublishedAt   time.Time `gorm:"not null"`
	PrivacyStatus string    `gorm:"not null"` // TODO: enum?
	lastUpdated   time.Time `gorm:"not null"`

	Users           []*User           `gorm:"many2many:VideosUsers"`
	BlobDownloaders []*BlobDownloader `gorm:"many2many:VideosBlobDownloader"`
}

type VideoHistory struct {
	ID        uint
	VideoId   string `gorm:"not null"`
	Action    string `gorm:"not null"` // TODO: enum?
	Old       string `gorm:"not null"`
	New       string `gorm:"not null"`
	UpdatedAt time.Time
}

type BlobDownloader struct {
	ID     uint
	Name   string   `gorm:"not null"`
	Secret string   `gorm:"not null"`
	Videos []*Video `gorm:"many2many:VideosBlobDownloader"`
}

type BlobLocation struct {
	ID               uint
	VideoID          string    `gorm:"not null"`
	BlobDownloaderID uint      `gorm:"not null"`
	Path             string    `gorm:"not null"`
	AddedAt          time.Time `gorm:"not null"`
	Type             string    `gorm:"not null"` // TODO: number?
}
