package dao

import (
	"context"
	"errors"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrorDuplicateEmail = errors.New("邮箱冲突,请更改您的邮箱")
	ErrorRecordNotFound = gorm.ErrRecordNotFound
)

type UserDao struct {
	db *gorm.DB
}

// User 与真实项目的领域(domain)分离,定义与数据库交互领域
type User struct {
	Id       int64  `gorm:"primary_key"`
	Email    string `gorm:"unique"`
	Password string
	CTime    int64 // 创建时间
	UTime    int64 // 更新时间
}

func NewUserDao(db *gorm.DB) *UserDao {
	return &UserDao{
		db: db,
	}
}

func (dao *UserDao) Insert(context context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.UTime = now
	u.CTime = now

	err := dao.db.WithContext(context).Create(&u).Error

	if me, ok := err.(*mysql.MySQLError); ok {
		const duplicator uint16 = 1062
		if me.Number == duplicator {
			return ErrorDuplicateEmail
		}
	}
	return err
}

func (dao *UserDao) SelectByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	return u, err
}
