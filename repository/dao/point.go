package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

var ErrRecordNotFound = gorm.ErrRecordNotFound

type PointDAO interface {
	InsertChange(ctx context.Context, pc PointChange) error
	FindOldestChaneBySource(ctx context.Context, source string) (PointChange, error)
	CountChangeByReasonSource(ctx context.Context, reason string, source string) (int64, error)
	GetUserPoints(ctx context.Context, uid int64) (int64, error)
}

type GORMPointDAO struct {
	db *gorm.DB
}

func NewGORMPointDAO(db *gorm.DB) PointDAO {
	return &GORMPointDAO{db: db}
}

func (dao *GORMPointDAO) InsertChange(ctx context.Context, pc PointChange) error {
	now := time.Now().UnixMilli()
	pc.Ctime = now
	pc.Utime = now
	return dao.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		err := tx.Create(&pc).Error
		if err != nil {
			return err
		}
		// 尝试更新用户积分，如果不存在则插入新记录，todo 关与upsert和update的性能分析
		sql := "INSERT INTO user_points (uid, points, ctime, utime) VALUES (?, GREATEST(0, ?), ?, ?) " +
			"ON DUPLICATE KEY UPDATE points = GREATEST(0, points + ?), utime = VALUES(utime)"
		return tx.Exec(sql, pc.Uid, pc.ChangeAmount, now, now, pc.ChangeAmount).Error
	})
}

func (dao *GORMPointDAO) FindOldestChaneBySource(ctx context.Context, source string) (PointChange, error) {
	var pc PointChange
	err := dao.db.WithContext(ctx).
		Where("source = ?", source).
		Order("id").
		First(&pc).Error
	return pc, err
}

func (dao *GORMPointDAO) CountChangeByReasonSource(ctx context.Context, reason string, source string) (int64, error) {
	var cnt int64
	err := dao.db.WithContext(ctx).
		Model(&PointChange{}).
		Where("source = ? and reason = ?", source, reason).
		Count(&cnt).Error
	return cnt, err
}

func (dao *GORMPointDAO) GetUserPoints(ctx context.Context, uid int64) (int64, error) {
	var up UserPoints
	err := dao.db.WithContext(ctx).
		Where("uid = ?", uid).
		First(&up).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	return up.Points, err
}

type PointChange struct {
	Id           int64 `gorm:"primaryKey,autoIncrement"`
	Uid          int64
	ChangeAmount int64
	Source       string `gorm:"index:source_reason"` // todo
	Reason       string `gorm:"index:source_reason"`
	Utime        int64
	Ctime        int64
}

type UserPoints struct {
	Id     int64 `gorm:"primaryKey,autoIncrement"`
	Uid    int64 `gorm:"uniqueIndex"`
	Points int64
	Utime  int64
	Ctime  int64
}
