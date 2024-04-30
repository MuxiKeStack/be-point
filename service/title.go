package service

import (
	"context"
	"errors"
	pointv1 "github.com/MuxiKeStack/be-api/gen/proto/point/v1"
	"github.com/MuxiKeStack/be-point/repository"
)

var ErrPointsNotEnough = errors.New("积分不足")

type TitleService interface {
	GetUsingTitleOfUser(ctx context.Context, uid int64) (pointv1.Title, error)
	GetOwnedTitlesOfUser(ctx context.Context, uid int64) ([]pointv1.Title, error)
	SaveUsingTitleOfUser(ctx context.Context, uid int64, title pointv1.Title) error
}

type titleService struct {
	repo           repository.TitleRepository
	pointSvc       PointService
	titlePointsMap map[pointv1.Title]int64
}

func NewTitleService(repo repository.TitleRepository, pointSvc PointService, titlePointsMap map[pointv1.Title]int64) TitleService {
	return &titleService{repo: repo, pointSvc: pointSvc, titlePointsMap: titlePointsMap}
}

func (t *titleService) GetUsingTitleOfUser(ctx context.Context, uid int64) (pointv1.Title, error) {
	return t.repo.GetUsingTitleOfUser(ctx, uid)
}

func (t *titleService) GetOwnedTitlesOfUser(ctx context.Context, uid int64) ([]pointv1.Title, error) {
	// 获取现有积分
	pInfo, err := t.pointSvc.GetPointInfoOfUser(ctx, uid)
	if err != nil {
		return nil, err
	}
	// 根据我的已有积分判断我拥有的称号
	ownedTitles := make([]pointv1.Title, 0, 3)
	for title, points := range t.titlePointsMap {
		if pInfo.Points >= points {
			ownedTitles = append(ownedTitles, title)
		}
	}
	return ownedTitles, nil
}

func (t *titleService) SaveUsingTitleOfUser(ctx context.Context, uid int64, title pointv1.Title) error {
	// 获取现有积分
	pInfo, err := t.pointSvc.GetPointInfoOfUser(ctx, uid)
	if err != nil {
		return err
	}
	points, exists := t.titlePointsMap[title]
	if !exists {
		return errors.New("非法的称号")
	}
	if pInfo.Points < points {
		return ErrPointsNotEnough
	}
	return t.repo.SaveUsingTitleOfUser(ctx, uid, title)
}
