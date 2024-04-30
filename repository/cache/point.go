package cache

import (
	"context"
	_ "embed"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

var ErrKeyNotExists = redis.Nil

//go:embed lua/reason_source_incr_cnt.lua
var reasonSourceIncrCntLuaScript string

type PointCache interface {
	SourceExists(ctx context.Context, source string) (bool, error)
	SetSourceExistence(ctx context.Context, source string, existence bool, expiration time.Duration) error
	GetChangeCountForReasonSource(ctx context.Context, reason string, source string) (int64, error)
	SetChangeCountForReasonSource(ctx context.Context, reason string, source string, cnt int64, expiration time.Duration) error
	IncrIfReasonSourcePresent(ctx context.Context, reason string, source string) error
}

type RedisPointCache struct {
	cmd redis.Cmdable
}

func NewRedisPointCache(cmd redis.Cmdable) PointCache {
	return &RedisPointCache{cmd: cmd}
}

func (cache *RedisPointCache) SourceExists(ctx context.Context, source string) (bool, error) {
	key := cache.sourceExistsKey(source)
	return cache.cmd.Get(ctx, key).Bool()
}

func (cache *RedisPointCache) SetSourceExistence(ctx context.Context, source string, existence bool, expiration time.Duration) error {
	key := cache.sourceExistsKey(source)
	return cache.cmd.Set(ctx, key, existence, expiration).Err() // TODO val interface{} ?
}

func (cache *RedisPointCache) sourceExistsKey(source string) string {
	return fmt.Sprintf("kstack:point:source:%s:exists", source)
}

func (cache *RedisPointCache) GetChangeCountForReasonSource(ctx context.Context, reason string, source string) (int64, error) {
	key := cache.changeCountForReasonSourceKey(reason, source)
	return cache.cmd.Get(ctx, key).Int64()
}

func (cache *RedisPointCache) SetChangeCountForReasonSource(ctx context.Context, reason string, source string, cnt int64, expiration time.Duration) error {
	key := cache.changeCountForReasonSourceKey(reason, source)
	return cache.cmd.Set(ctx, key, cnt, expiration).Err()
}

func (cache *RedisPointCache) changeCountForReasonSourceKey(reason string, source string) string {
	return fmt.Sprintf("kstack:point:reason_source:<%s,%s>:change_count", reason, source)
}

func (cache *RedisPointCache) IncrIfReasonSourcePresent(ctx context.Context, reason string, source string) error {
	key := cache.changeCountForReasonSourceKey(reason, source)
	return cache.cmd.Eval(ctx, reasonSourceIncrCntLuaScript, []string{key}).Err()
}
