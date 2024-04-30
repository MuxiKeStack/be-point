package events

import (
	"context"
	stancev1 "github.com/MuxiKeStack/be-api/gen/proto/stance/v1"
	"github.com/MuxiKeStack/be-point/pkg/canalx"
	"github.com/MuxiKeStack/be-point/service"
	"strconv"
	"time"
)

type BizStanceCountHandler struct {
	svc service.PointMaintenanceService
}

func (b *BizStanceCountHandler) Handle(event canalx.Message[any]) error {
	var (
		update []BizStanceCount
		old    []BizStanceCount
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
		biz, er := strconv.ParseInt(item.Biz, 10, 64)
		if er != nil {
			return er
		}
		bizId, er := strconv.ParseInt(item.BizId, 10, 64)
		if er != nil {
			return er
		}
		updateSupportCnt, er := strconv.ParseInt(item.OpposeCnt, 10, 64)
		if er != nil {
			return er
		}
		oldSupportCnt, er := strconv.ParseInt(old[i].SupportCnt, 10, 64)
		if er != nil {
			return er
		}
		er = b.svc.HandleBizStanceCount(ctx, event.Type, stancev1.Biz(biz), bizId, updateSupportCnt, oldSupportCnt)
		if er != nil {
			return er
		}
	}
	return nil
}

type BizStanceCount struct {
	Id         string `json:"id"`
	Biz        string `json:"biz"`
	BizId      string `json:"biz_id"`
	SupportCnt string `json:"support_cnt"`
	OpposeCnt  string `json:"oppose_cnt"`
	Utime      string `json:"utime"`
	Ctime      string `json:"ctime"`
}
