package repository

import (

	"tarlek.com/icesystem/entity"
	"gorm.io/gorm"
)

type UserRepository interface {
	FindByUserAD(userAD string) entity.User
	ProfileUser(userID string, username string) entity.User

}

type UserConnect struct {
	connect *gorm.DB
}


func NewUserRepository(db *gorm.DB) UserRepository {
	return &UserConnect{connect: db}
}

func (db *UserConnect) FindByUserAD(userAD string) entity.User {
	var user entity.User
	db.connect.Table("user").Where("dns_user = ?", userAD).Take(&user)
	return user
}

func (db *UserConnect) ProfileUser(userID string, username string) entity.User {
	var user entity.User
	//	db.connect.Table("person").Find(&user, userID)
	db.connect.Table("person").Select("person.id,person.current_team_id,person.current_safety_team_id,person.photo,query_user_emp_data.section_code").Joins("inner join query_user_emp_data on person.id=query_user_emp_data.person_id").Where("person.ad_user = ?", username).Find(&user)
	return user
}
