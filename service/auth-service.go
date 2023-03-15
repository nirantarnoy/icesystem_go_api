package service

import (
	"tarlek.com/icesystem/config"
	"tarlek.com/icesystem/entity"
	"tarlek.com/icesystem/repository"
	"log"
)

type AuthService interface {
	FindByADUser(userAD string) entity.User
	VerifyCredential(username string, password string) bool
}

type authService struct {
	userRepo repository.UserRepository
}

func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

func (service authService) FindByADUser(userAD string) entity.User {
	return service.userRepo.FindByUserAD(userAD)
}

func (service authService) VerifyCredential(username string, password string) bool {
	subfix := "@cicnetgrp.net"
	ldapConn, err := config.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer ldapConn.Close()

	//Bind Connection
	err = ldapConn.Bind(username+subfix, password)
	if err != nil {
		return false
	}
	return true
}
