package object

import (
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var Database *gorm.DB

func init() {
	os.MkdirAll("data", 0755)
	db, err := gorm.Open(sqlite.Open("data/account.db"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	Database = db

	// 自动迁移
	err = Database.AutoMigrate(
		&User{},
		&App{},
		&UserMetadata{},
		&OAuthCode{},
		&RefreshToken{},
	)
	if err != nil {
		panic(err)
	}
}
