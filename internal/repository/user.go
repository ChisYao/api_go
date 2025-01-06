package repository

import (
	"context"
	"go_web/internal/domain"
	"go_web/internal/repository/dao"
)

var (
	ErrorDuplicateEmail = dao.ErrorDuplicateEmail
	ErrorUserNotFound   = dao.ErrorRecordNotFound
)

type UserRepository struct {
	dao *dao.UserDao
}

func NewUserRepository(dao *dao.UserDao) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (repo *UserRepository) Create(context context.Context, u domain.User) error {
	return repo.dao.Insert(context, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (repo *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	u, err := repo.dao.SelectByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return repo.transferToDomain(u), nil
}

// 私有转换方法,将数据库定义内容转化为领域对象(domain)
func (repo *UserRepository) transferToDomain(u dao.User) domain.User {
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}
}
