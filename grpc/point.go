package grpc

import (
	"context"
	pointv1 "github.com/MuxiKeStack/be-api/gen/proto/point/v1"
	"github.com/MuxiKeStack/be-point/service"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"golang.org/x/sync/errgroup"
)

type PointServiceServer struct {
	pointv1.UnimplementedPointServiceServer
	pointSvc service.PointService
	titleSvc service.TitleService
}

func NewPointServiceServer(pointSvc service.PointService, titleSvc service.TitleService) *PointServiceServer {
	return &PointServiceServer{pointSvc: pointSvc, titleSvc: titleSvc}
}

func (p *PointServiceServer) GetTitleOfUser(ctx context.Context, request *pointv1.GetTitleOfUserRequest) (*pointv1.GetTitleOfUserResponse, error) {
	var eg errgroup.Group
	var res pointv1.GetTitleOfUserResponse
	eg.Go(func() error {
		usingTitle, err := p.titleSvc.GetUsingTitleOfUser(ctx, request.GetUid())
		if err != nil {
			return err
		}
		res.UsingTitle = usingTitle
		return nil
	})
	eg.Go(func() error {
		ownedTitles, err := p.titleSvc.GetOwnedTitlesOfUser(ctx, request.GetUid())
		if err != nil {
			return err
		}
		res.OwnedTitles = ownedTitles
		return nil
	})
	err := eg.Wait()
	if err != nil {
		return &pointv1.GetTitleOfUserResponse{}, err
	}
	return &res, nil
}

func (p *PointServiceServer) SaveUsingTitleOfUser(ctx context.Context, request *pointv1.SaveUsingTitleOfUserRequest) (*pointv1.SaveUsingTitleOfUserResponse, error) {
	err := p.titleSvc.SaveUsingTitleOfUser(ctx, request.GetUid(), request.GetTitle())
	return &pointv1.SaveUsingTitleOfUserResponse{}, err
}

func (p *PointServiceServer) GetPointInfoOfUser(ctx context.Context, request *pointv1.GetPointInfoOfUserRequest) (*pointv1.GetPointInfoOfUserResponse, error) {
	pInfo, err := p.pointSvc.GetPointInfoOfUser(ctx, request.GetUid())
	return &pointv1.GetPointInfoOfUserResponse{
		Points:          pInfo.Points,
		Level:           pInfo.Level,
		NextLevelPoints: pInfo.NextLevelPoints,
	}, err
}

func (p *PointServiceServer) Register(server *grpc.Server) {
	pointv1.RegisterPointServiceServer(server, p)
}
