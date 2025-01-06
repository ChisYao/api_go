package service

import (
	"context"
	"errors"
	"go_web/internal/domain"
	"go_web/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrorDuplicateEmail         = repository.ErrorDuplicateEmail
	ErrorInvalidEmailOrPassword = errors.New("邮箱名称或密码不正确")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) SignUp(context context.Context, u domain.User) error {
	// 对密码进行加密处理.
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(context, u)
}

func (svc *UserService) Login(context context.Context, email string, password string) (domain.User, error) {
	u, err := svc.repo.FindByEmail(context, email)

	if err == repository.ErrorUserNotFound {
		return domain.User{}, ErrorInvalidEmailOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}

	//校验密码是否正确
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		return domain.User{}, ErrorInvalidEmailOrPassword
	}

	return u, nil
}
