package repository

import (
	"context"
	"github.com/MuxiKeStack/be-point/domain"
	"github.com/MuxiKeStack/be-point/pkg/logger"
	"github.com/MuxiKeStack/be-point/repository/cache"
	"github.com/MuxiKeStack/be-point/repository/dao"
	"time"
)

var ErrPointChangeNotFound = dao.ErrRecordNotFound

type PointRepository interface {
	Change(ctx context.Context, change domain.PointChange) error
	FindOldestBySource(ctx context.Context, source string) (domain.PointChange, error)
	Exists(ctx context.Context, source string) (bool, error)
	CountChangeByReasonSource(ctx context.Context, reason string, source string) (int64, error)
	GetUserPoints(ctx context.Context, uid int64) (int64, error)
}

type CachedPointRepository struct {
	dao   dao.PointDAO
	cache cache.PointCache
	l     logger.Logger
}

func NewCachedPointRepository(dao dao.PointDAO, cache cache.PointCache, l logger.Logger) PointRepository {
	return &CachedPointRepository{dao: dao, cache: cache, l: l}
}

func (repo *CachedPointRepository) Change(ctx context.Context, change domain.PointChange) error {
	// 要对点赞、评论的记录进行缓存同步
	err := repo.dao.InsertChange(ctx, repo.toEntity(change))
	if err != nil {
		return err
	}
	// 要进行缓存，缓存同步: <reason,source>数量的变化
	return repo.cache.IncrIfReasonSourcePresent(ctx, change.Reason, change.Source)
}

func (repo *CachedPointRepository) FindOldestBySource(ctx context.Context, source string) (domain.PointChange, error) {
	pc, err := repo.dao.FindOldestChaneBySource(ctx, source)
	return repo.toDomain(pc), err
}

func (repo *CachedPointRepository) Exists(ctx context.Context, source string) (bool, error) {
	// 这个必须cache
	exists, err := repo.cache.SourceExists(ctx, source)
	if err == nil {
		return exists, nil
	}
	if err != cache.ErrKeyNotExists {
		return false, err
	}
	// 去数据库
	_, err = repo.dao.FindOldestChaneBySource(ctx, source)
	switch {
	case err == nil:
		// 回写,这在这里写，一天只用写一次缓存，而且之后不存在缓存状态变更
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			er := repo.cache.SetSourceExistence(ctx, source, true, time.Hour*24)
			if er != nil {
				repo.l.Error("回写source存在性失败", logger.String("source", source))
			}
		}()
		return true, nil
	case err == dao.ErrRecordNotFound:
		return false, nil
	default:
		return false, err
	}
}

func (repo *CachedPointRepository) CountChangeByReasonSource(ctx context.Context, reason string, source string) (int64, error) {
	cnt, err := repo.cache.GetChangeCountForReasonSource(ctx, reason, source)
	if err == nil {
		return cnt, nil
	}
	if err != cache.ErrKeyNotExists {
		return 0, err
	}
	cnt, err = repo.dao.CountChangeByReasonSource(ctx, reason, source)
	if err != nil {
		return 0, err
	}
	// 进行回写
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		er := repo.cache.SetChangeCountForReasonSource(ctx, reason, source, cnt, time.Hour*24)
		if er != nil {
			repo.l.Error("回写change数量失败",
				logger.String("reason", reason),
				logger.String("source", source))
		}
	}()
	return cnt, nil
}

func (repo *CachedPointRepository) GetUserPoints(ctx context.Context, uid int64) (int64, error) {
	return repo.dao.GetUserPoints(ctx, uid)
}

func (repo *CachedPointRepository) toEntity(change domain.PointChange) dao.PointChange {
	return dao.PointChange{
		Id:           change.Id,
		Uid:          change.Uid,
		ChangeAmount: change.ChangeAmount,
		Reason:       change.Reason,
		Source:       change.Source,
	}
}

func (repo *CachedPointRepository) toDomain(change dao.PointChange) domain.PointChange {
	return domain.PointChange{
		Id:           change.Id,
		Uid:          change.Uid,
		ChangeAmount: change.ChangeAmount,
		Reason:       change.Reason,
		Source:       change.Source,
		Utime:        time.UnixMilli(change.Utime),
		Ctime:        time.UnixMilli(change.Ctime),
	}
}
