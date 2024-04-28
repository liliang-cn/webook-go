package dao

import (
	"context"
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"github.com/liliang-cn/webook/internal/domain"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
	ErrEmailNotFound  = gorm.ErrRecordNotFound
	ErrorUserNotFound = gorm.ErrRecordNotFound
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{
		db: db,
	}
}

func (dao UserDAO) Insert(ctx context.Context, u User) error {
	now := time.Now().UnixMilli()
	u.UTime = now
	u.CTime = now

	err := dao.db.WithContext(ctx).Create(&u).Error

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		const uniqueConstraintViolation = 1062
		if mysqlErr.Number == uniqueConstraintViolation {
			return ErrDuplicateEmail
		}

	}

	return err
}

func (dao UserDAO) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var user User

	err := dao.db.WithContext(ctx).Where("email = ?", email).First(&user).Error

	return user, err
}

func (dao UserDAO) FindUserByID(ctx context.Context, ID int64) (User, error) {
	var user User

	err := dao.db.WithContext(ctx).First(&user, ID).Error

	return user, err
}

func (dao UserDAO) Update(ctx context.Context, u domain.User) error {
	err := dao.db.Model(&u).WithContext(ctx).Updates(User{
		ID:          u.ID,
		BirthDate:   u.BirthDate,
		NickName:    u.NickName,
		Description: u.Description,
		UTime:       time.Now().UnixMilli(),
	}).Error

	return err
}

type User struct {
	ID       int64  `gorm:"autoIncrement;primaryKey"`
	Email    string `gorm:"unique"`
	Password string

	CTime int64
	UTime int64

	NickName    string
	BirthDate   string
	Description string
}
