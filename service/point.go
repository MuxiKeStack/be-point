package service

import (
	"context"
	"github.com/MuxiKeStack/be-point/domain"
	"github.com/MuxiKeStack/be-point/repository"
)

type PointService interface {
	GetPointInfoOfUser(ctx context.Context, uid int64) (domain.UserPointInfo, error)
}

type pointService struct {
	repo repository.PointRepository
}

func NewPointService(repo repository.PointRepository) PointService {
	return &pointService{repo: repo}
}

func (p *pointService) GetPointInfoOfUser(ctx context.Context, uid int64) (domain.UserPointInfo, error) {
	points, err := p.repo.GetUserPoints(ctx, uid)
	if err != nil {
		return domain.UserPointInfo{}, err
	}
	// 根据用户现有的积分，判断level和nextLevelPoints
	level, nextLevelPoints := p.calculateLevelAndNextPoints(points)
	return domain.UserPointInfo{
		Uid:             uid,
		Points:          points,
		Level:           level,
		NextLevelPoints: nextLevelPoints,
	}, nil
}

func (p *pointService) calculateLevelAndNextPoints(points int64) (int64, int64) {
	var level int64
	var nextLevelPoints int64

	switch {
	case points >= 300:
		level = 5
		nextLevelPoints = 300 // 最高级，不需要下一级积分
	case points >= 150:
		level = 4
		nextLevelPoints = 300
	case points >= 80:
		level = 3
		nextLevelPoints = 150
	case points >= 20:
		level = 2
		nextLevelPoints = 80
	case points >= 5:
		level = 1
		nextLevelPoints = 20
	default:
		level = 0
		nextLevelPoints = 5
	}

	return level, nextLevelPoints
}
