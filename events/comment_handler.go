package events

import (
	"context"
	"github.com/MuxiKeStack/be-point/pkg/canalx"
	"github.com/MuxiKeStack/be-point/service"
	"strconv"
	"time"
)

type CommentHandler struct {
	svc service.PointMaintenanceService
}

func (c *CommentHandler) Handle(event canalx.Message[any]) error {
	var update []Comment
	err := copyByJSON(event.Data, &update)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	for _, item := range update {
		uid, er := strconv.ParseInt(item.Uid, 10, 64)
		if er != nil {
			return er
		}
		er = c.svc.HandleComment(ctx, event.Type, uid)
		if er != nil {
			return er
		}
	}
	return nil
}

type Comment struct {
	Id         string `json:"id"`
	Uid        string `json:"uid"`
	Biz        string `json:"biz"`
	BizId      string `json:"bizID"`
	RootID     string `json:"rootID"`
	PID        string `json:"pid"`
	ReplyToUid string `json:"reply_to_uid"`
	Content    string `json:"content"`
	Ctime      string `json:"ctime"`
	Utime      string `json:"utime"`
}
