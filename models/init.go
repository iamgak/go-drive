package models

import (
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Init struct {
	UsersORM UserModelORM
}

func Constructor(dbORM *gorm.DB, Logger *logrus.Logger) *Init {
	return &Init{
		UsersORM: UserModelORM{db: dbORM, logger: Logger},
	}
}
