package dao

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("邮箱冲突")
	ErrRecordNotFound = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao *UserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.Ctime = now
	u.Utime = now
	err := dao.db.WithContext(ctx).Create(&u).Error
	if me, ok := err.(*mysql.MySQLError); ok {
		const duplicateErr uint16 = 1062
		if me.Number == duplicateErr {
			// 用户冲突，邮箱冲突
			return ErrDuplicateEmail
		}
	}
	return err
}

func (dao *UserDAO) FindByEmail(ctx context.Context, email string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("email=?", email).First(&u).Error
	return u, err
}

func (dao *UserDAO) FindByID(ctx *gin.Context, userIDStr string) (User, error) {
	var u User
	err := dao.db.WithContext(ctx).Where("id=?", userIDStr).First(&u).Error
	return u, err
}

type User struct {
	//Id       int64  `gorm:"primaryKey,autoIncrement"`
	//Email    string `gorm:"unique"`
	//Password string
	//
	//// 时区，UTC 0 的毫秒数
	//// 创建时间
	//Ctime int64
	//// 更新时间
	//Utime int64
	//
	//Phone    string
	//Nickname string
	//Birthday string
	//AboutMe  string
	Id       int64  `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Email    string `gorm:"unique;not null;size:255" json:"email" validate:"required,email"`
	Password string `gorm:"not null;size:255" json:"-" validate:"required,min=8"` // 密码不参与 JSON 序列化

	// 时间戳（使用 int64 存储 UTC 毫秒数）
	Ctime int64 `gorm:"autoCreateTime:milli;column:ctime" json:"ctime"` // 创建时间
	Utime int64 `gorm:"autoUpdateTime:milli;column:utime" json:"utime"` // 更新时间

	// 个人资料
	Phone    string `gorm:"size:20;uniqueIndex" json:"phone" validate:"omitempty,e164"`
	Nickname string `gorm:"size:50" json:"nickname" validate:"required,max=50"`
	Birthday string `gorm:"size:10" json:"birthday" validate:"omitempty,date"` // YYYY-MM-DD 格式
	AboutMe  string `gorm:"type:text" json:"about_me" validate:"max=500"`
}

//type Address struct {
//	Uid
//}
