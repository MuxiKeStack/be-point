package service

import (
	"context"
	"errors"
	pointv1 "github.com/MuxiKeStack/be-api/gen/proto/point/v1"
	"github.com/MuxiKeStack/be-point/domain"
	"github.com/MuxiKeStack/be-point/pkg/logger"
	"github.com/MuxiKeStack/be-point/repository"
	"golang.org/x/sync/errgroup"
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
	l              logger.Logger
}

func NewTitleService(repo repository.TitleRepository, pointSvc PointService, titlePointsMap map[pointv1.Title]int64,
	l logger.Logger) TitleService {
	return &titleService{
		repo:           repo,
		pointSvc:       pointSvc,
		titlePointsMap: titlePointsMap,
		l:              l,
	}
}

func (t *titleService) GetUsingTitleOfUser(ctx context.Context, uid int64) (pointv1.Title, error) {
	// 判断 title 是否仍有效
	var (
		eg         errgroup.Group
		usingTitle pointv1.Title
		pInfo      domain.UserPointInfo
	)
	eg.Go(func() error {
		var er error
		usingTitle, er = t.repo.GetUsingTitleOfUser(ctx, uid)
		return er
	})
	eg.Go(func() error {
		var er error
		pInfo, er = t.pointSvc.GetPointInfoOfUser(ctx, uid)
		return er
	})
	err := eg.Wait()
	if err != nil {
		return 0, err
	}
	points := t.titlePointsMap[usingTitle]
	if pInfo.Points < points {
		// 更新为None
		go func() {
			er := t.SaveUsingTitleOfUser(ctx, uid, pointv1.Title_None)
			if er != nil {
				t.l.Error("更改用户使用中的Title失败",
					logger.Int64("uid", uid),
					logger.String("title", pointv1.Title_None.String()))
			}
		}()
		return pointv1.Title_None, nil
	}
	return usingTitle, nil
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

// 这里可以Save None
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
