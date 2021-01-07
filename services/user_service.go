package services

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
	"seckill/datamodels"
	"seckill/repositories"
)

type IUserService interface {
	IsPwdSuccess(string, string) (*datamodels.User, bool)
	AddUser(*datamodels.User) (int64, error)
}

type UserService struct {
	UserRepository repositories.IUserRepository
}

func NewUserService(repository repositories.IUserRepository) IUserService {
	return &UserService{UserRepository: repository}
}

func (u *UserService) IsPwdSuccess(userName, pwd string) (user *datamodels.User, isOK bool) {
	user, err := u.UserRepository.Select(userName)
	if err != nil {
		return
	}
	isOK, _ = validatePassword(pwd, user.HashPassword)
	if !isOK {
		return &datamodels.User{}, false
	}
	return
}

func (u *UserService) AddUser(user *datamodels.User) (userId int64, err error) {
	pwdByte, errPwd := genertePassword(user.HashPassword)
	if errPwd != nil {
		return userId, errPwd
	}
	user.HashPassword = string(pwdByte)
	return u.UserRepository.Insert(user)
}

func genertePassword(userPassword string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
}

func validatePassword(userPassword string, hashed string) (isOk bool, err error) {
	if err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(userPassword)); err != nil {
		return false, errors.New("密码错误")
	}
	return true, nil
}
