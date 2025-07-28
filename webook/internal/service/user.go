package service

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"gochuji/webook/internal/domain"
	"gochuji/webook/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail        = repository.ErrDuplicateEmail
	ErrInvalidUserOrPassword = errors.New("用户不存在或者密码不对")
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (svc *UserService) Signup(ctx context.Context, u domain.User) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return svc.repo.Create(ctx, u)
}

// Login 函数用于用户登录，接收一个context.Context、email和password作为参数，返回一个domain.User和一个error
func (svc *UserService) Login(ctx context.Context, email string, password string) (domain.User, error) {
	// 根据email查找用户
	u, err := svc.repo.FindByEmail(ctx, email)
	// 如果找不到用户，返回ErrInvalidUserOrPassword错误
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	// 如果发生其他错误，返回错误
	if err != nil {
		return domain.User{}, err
	}

	// 将用户密码和输入的密码进行比对
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	// 如果比对失败，返回ErrInvalidUserOrPassword错误
	if err != nil {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	// 返回用户
	return u, nil
}

func (svc *UserService) FindByID(ctx *gin.Context, userIDStr string) (domain.User, error) {
	return svc.repo.FindByID(ctx, userIDStr)
}

func (svc *UserService) Edit(ctx *gin.Context, userIDStr string, nickname string, birthday string, AboutMe string) (domain.User, error) {
	_, err := svc.FindByID(ctx, userIDStr)
	if err != nil {
		return domain.User{}, err
	}
	return svc.repo.Edit(ctx, userIDStr, nickname, birthday, AboutMe)
}
