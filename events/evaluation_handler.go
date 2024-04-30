package events

import (
	"context"
	evaluationv1 "github.com/MuxiKeStack/be-api/gen/proto/evaluation/v1"
	"github.com/MuxiKeStack/be-point/pkg/canalx"
	"github.com/MuxiKeStack/be-point/service"
	"strconv"
	"time"
)

type EvaluationHandler struct {
	svc service.PointMaintenanceService
}

func (e *EvaluationHandler) Handle(event canalx.Message[any]) error {
	var (
		update []Evaluation
		old    []Evaluation
	)
	err := copyByJSON(event.Data, &update)
	if err != nil {
		return err
	}
	err = copyByJSON(event.Old, &old)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	for i, item := range update {
		id, er := strconv.ParseInt(item.Id, 10, 64)
		if er != nil {
			return er
		}
		uid, er := strconv.ParseInt(item.PublisherId, 10, 64)
		if er != nil {
			return er
		}
		updateStatus, er := strconv.ParseInt(item.Status, 10, 64)
		if er != nil {
			return er
		}
		var oldStatus int64
		if len(old) >= i+1 {
			oldStatus, er = strconv.ParseInt(old[i].Status, 10, 64)
			if er != nil {
				return er
			}
		}
		er = e.svc.HandleEvaluation(ctx, event.Type, uid, id, item.Content, evaluationv1.EvaluationStatus(updateStatus),
			evaluationv1.EvaluationStatus(oldStatus))
		if er != nil {
			return er
		}
	}
	return nil
}

type Evaluation struct {
	Id             string `json:"id"`
	PublisherId    string `json:"publisher_id"`
	CourseId       string `json:"course_id"`
	CourseProperty string `json:"course_property"`
	StarRating     string `json:"star_rating"`
	Content        string `json:"content"`
	Status         string `json:"status"`
	Utime          string `json:"utime"`
	Ctime          string `json:"ctime"`
}
