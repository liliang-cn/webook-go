package repository

import (
	"context"

	"github.com/liliang-cn/webook/internal/domain"
	"github.com/liliang-cn/webook/internal/repository/dao"
)

var (
	ErrDuplicateEmail = dao.ErrDuplicateEmail
	ErrUserNotFound   = dao.ErrEmailNotFound
	ErrorUserNotFound = dao.ErrorUserNotFound
)

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *UserRepository) FindUserByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := r.dao.FindUserByEmail(ctx, email)

	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		ID:       u.ID,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) FindUserByID(ctx context.Context, ID int64) (domain.User, error) {
	u, err := r.dao.FindUserByID(ctx, ID)

	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		NickName:    u.NickName,
		Description: u.Description,
		BirthDate:   u.BirthDate,
	}, nil
}

func (r *UserRepository) EditProfile(ctx context.Context, user domain.User) error {
	err := r.dao.Update(ctx, user)

	if err != nil {
		return err
	}

	return nil
}
