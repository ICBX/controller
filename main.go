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
		log.WithError(err).Fatal("cannot open database")
		return
	}
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
}
