package dao

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type TitleDAO interface {
	GetUsingTitleOfUser(ctx context.Context, uid int64) (int32, error)
	SaveUsingTitleOfUser(ctx context.Context, uid int64, title int32) error
}

type GORMTitleDAO struct {
	db *gorm.DB
}

func NewGORMTitleDAO(db *gorm.DB) TitleDAO {
	return &GORMTitleDAO{db: db}
}

func (dao *GORMTitleDAO) GetUsingTitleOfUser(ctx context.Context, uid int64) (int32, error) {
	var ut UserTitle
	err := dao.db.WithContext(ctx).
		Where("uid = ?", uid).
		First(&ut).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	return ut.UsingTitle, err
}

func (dao *GORMTitleDAO) SaveUsingTitleOfUser(ctx context.Context, uid int64, title int32) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoUpdates: clause.Assignments(
			map[string]interface{}{
				"utime":       now,
				"using_title": title,
			},
		)}).Create(&UserTitle{
		Uid:        uid,
		UsingTitle: title,
		Utime:      now,
		Ctime:      now,
	}).Error

}

type UserTitle struct {
	Id         int64 `gorm:"primaryKey,autoIncrement"`
	Uid        int64 `gorm:"uniqueIndex"`
	UsingTitle int32
	Utime      int64
	Ctime      int64
}
