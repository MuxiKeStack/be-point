//go:build wireinject

package main

import (
	"github.com/MuxiKeStack/be-point/events"
	"github.com/MuxiKeStack/be-point/grpc"
	"github.com/MuxiKeStack/be-point/ioc"
	"github.com/MuxiKeStack/be-point/repository"
	"github.com/MuxiKeStack/be-point/repository/cache"
	"github.com/MuxiKeStack/be-point/repository/dao"
	"github.com/MuxiKeStack/be-point/service"
	"github.com/google/wire"
)

func InitApp() *App {
	wire.Build(
		wire.Struct(new(App), "*"),
		// consumer
		ioc.InitConsumers,
		events.NewMySQLBinlogConsumer,
		service.NewPointMaintenanceService,
		// rpc client
		ioc.InitEvaluationClient,
		ioc.InitAnswerClient,
		// grpc
		ioc.InitGRPCxKratosServer,
		grpc.NewPointServiceServer,
		service.NewPointService,
		ioc.InitTitleService,
		repository.NewCachedPointRepository,
		repository.NewTitleRepository,
		dao.NewGORMPointDAO,
		dao.NewGORMTitleDAO,
		cache.NewRedisPointCache,
		// 第三方
		ioc.InitKafka,
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitEtcdClient,
		ioc.InitLogger,
	)
	return &App{}
}
