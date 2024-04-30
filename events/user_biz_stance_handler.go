package events

import (
	"context"
	stancev1 "github.com/MuxiKeStack/be-api/gen/proto/stance/v1"
	"github.com/MuxiKeStack/be-point/pkg/canalx"
	"github.com/MuxiKeStack/be-point/service"
	"strconv"
	"time"
)

type UserBizStanceHandler struct {
	svc service.PointMaintenanceService
}

func (u *UserBizStanceHandler) Handle(event canalx.Message[any]) error {
	var (
		update []UserBizStance
		old    []UserBizStance
	)
	err := copyByJSON(event.Data, &update)
	if err != nil {
		return err
	}
	err = copyByJSON(event.Old, &old)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	for i, item := range update {
		uid, er := strconv.ParseInt(item.Uid, 10, 64)
		if er != nil {
			return er
		}
		updateStance, er := strconv.ParseInt(item.Stance, 10, 64)
		if er != nil {
			return er
		}
		oldStance, er := strconv.ParseInt(old[i].Stance, 10, 64)
		if er != nil {
			return er
		}
		er = u.svc.HandleUserBizStance(ctx, event.Type, uid, stancev1.Stance(updateStance), stancev1.Stance(oldStance))
		if er != nil {
			return er
		}
	}
	return nil
}

type UserBizStance struct {
	Id     string `json:"id"`
	Uid    string `json:"uid"`
	Biz    string `json:"biz"`
	BizId  string `json:"biz_id"`
	Stance string `json:"stance"`
	Utime  string `json:"utime"`
	Ctime  string `json:"ctime"`
}
