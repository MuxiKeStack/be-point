package repository

import (
	"context"
	pointv1 "github.com/MuxiKeStack/be-api/gen/proto/point/v1"
	"github.com/MuxiKeStack/be-point/repository/dao"
)

type TitleRepository interface {
	GetUsingTitleOfUser(ctx context.Context, uid int64) (pointv1.Title, error)
	SaveUsingTitleOfUser(ctx context.Context, uid int64, title pointv1.Title) error
}

type titleRepository struct {
	dao dao.TitleDAO
}

func NewTitleRepository(dao dao.TitleDAO) TitleRepository {
	return &titleRepository{dao: dao}
}

func (repo *titleRepository) GetUsingTitleOfUser(ctx context.Context, uid int64) (pointv1.Title, error) {
	title, err := repo.dao.GetUsingTitleOfUser(ctx, uid)
	return pointv1.Title(title), err
}

func (repo *titleRepository) SaveUsingTitleOfUser(ctx context.Context, uid int64, title pointv1.Title) error {
	return repo.dao.SaveUsingTitleOfUser(ctx, uid, int32(title))
}
