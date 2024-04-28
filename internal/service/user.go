package service

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/liliang-cn/webook/internal/domain"
	"github.com/liliang-cn/webook/internal/repository"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("email or password is invalid")
	ErrNotFound              = errors.New("user not found")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) SignUp(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

func (svc *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindUserByEmail(ctx, email)

	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}

	if err != nil {
		return domain.User{}, err
	}

	// Compare the password from the request with the password from the database
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))

	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}

	return u, nil
}

func (svc *UserService) GetProfile(ctx context.Context, uid int64) (domain.User, error) {
	u, err := svc.repo.FindUserByID(ctx, uid)

	if errors.Is(err, repository.ErrorUserNotFound) {
		return domain.User{}, ErrNotFound
	}

	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		Description: u.Description,
		BirthDate:   u.BirthDate,
		NickName:    u.NickName,
	}, nil
}

func (svc *UserService) EditProfile(ctx context.Context, uid int64, user domain.User) error {
	_, err := svc.repo.FindUserByID(ctx, uid)

	if errors.Is(err, repository.ErrorUserNotFound) {
		return ErrNotFound
	}

	if err != nil {
		return err
	}

	// edit the user profile
	err = svc.repo.EditProfile(ctx, domain.User{
		ID:          uid,
		BirthDate:   user.BirthDate,
		NickName:    user.NickName,
		Description: user.Description,
	})

	if err != nil {
		return err
	}

	return nil
}
