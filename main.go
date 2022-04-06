package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	log.SetHandler(cli.Default)
}

func main() {
	db, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	if err = db.AutoMigrate(
		&User{},
		&Video{},
		&VideoHistory{},
		&BlobDownloader{},
		&BlobLocation{},
		&VideoViewCountHistory{},
		&VideoLikeCountHistory{},
		&VideoCommentCountHistory{},
	); err != nil {
		panic(err)
	}
	log.Info("OK!")
}
