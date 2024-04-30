package ioc

import (
	pointv1 "github.com/MuxiKeStack/be-api/gen/proto/point/v1"
	"github.com/MuxiKeStack/be-point/pkg/logger"
	"github.com/MuxiKeStack/be-point/repository"
	"github.com/MuxiKeStack/be-point/service"
)

func InitTitleService(repo repository.TitleRepository, pointSvc service.PointService, l logger.Logger) service.TitleService {
	return service.NewTitleService(repo, pointSvc,
		map[pointv1.Title]int64{
			pointv1.Title_None:           0,
			pointv1.Title_CaringSenior:   80,
			pointv1.Title_KeStackPartner: 150,
			pointv1.Title_CCNUWithMe:     300,
		}, l)
}
