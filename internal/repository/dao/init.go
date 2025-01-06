package dao

import "gorm.io/gorm"

func InitTable(db *gorm.DB) error {
	// 严格来说,并不是一个好的实践
	return db.AutoMigrate(&User{})
}
