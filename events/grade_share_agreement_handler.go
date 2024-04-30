package events

import (
	"context"
	"github.com/MuxiKeStack/be-point/pkg/canalx"
	"github.com/MuxiKeStack/be-point/service"
	"strconv"
	"time"
)

type GradeShareAgreementHandler struct {
	svc service.PointMaintenanceService
}

func (g *GradeShareAgreementHandler) Handle(event canalx.Message[any]) error {
	var (
		update []GradeShareAgreement
		old    []GradeShareAgreement
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
		// 根据Grade服务dao的编码，只有在签约状态的确发生改变才会有写入操作，才会产生日志
		uid, er := strconv.ParseInt(item.Uid, 10, 64)
		if er != nil {
			return er
		}
		updateIsSigned, er := strconv.ParseBool(item.IsSigned)
		if er != nil {
			return er
		}
		oldIsSigned, er := strconv.ParseBool(old[i].IsSigned)
		if er != nil {
			return er
		}
		er = g.svc.HandleGradeShareAgreement(ctx, event.Type, uid, updateIsSigned, oldIsSigned)
		if er != nil {
			return er
		}
	}
	return nil
}

type GradeShareAgreement struct {
	Id       string `json:"id"`
	Uid      string `json:"uid"`
	IsSigned string `json:"is_signed"`
	Utime    string `json:"utime"`
	Ctime    string `json:"ctime"`
}
