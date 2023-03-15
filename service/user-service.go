package service

import (
	"tarlek.com/icesystem/entity"
	"tarlek.com/icesystem/repository"
)

type UserService interface {
	Profile(userID string, userName string) entity.User
}

type userService struct {
	userRepo repository.UserRepository
}


func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{
		userRepo: userRepo,
	}
}

func (u *userService) Profile(userID string, username string) entity.User {
	return u.userRepo.ProfileUser(userID, username)
}
