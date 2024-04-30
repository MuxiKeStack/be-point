package events

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/MuxiKeStack/be-point/pkg/canalx"
	"github.com/MuxiKeStack/be-point/pkg/logger"
	"github.com/MuxiKeStack/be-point/pkg/saramax"
	"github.com/MuxiKeStack/be-point/service"
)

const mysqlBinlogEvent = "kstack_binlog"

type MySQLBinlogConsumer struct {
	client   sarama.Client
	l        logger.Logger
	handlers map[string]MysqlBinlogEventHandler
}

func NewMySQLBinlogConsumer(client sarama.Client, l logger.Logger, svc service.PointMaintenanceService) *MySQLBinlogConsumer {
	return &MySQLBinlogConsumer{
		client: client,
		l:      l,
		handlers: map[string]MysqlBinlogEventHandler{
			"grade_share_agreements": &GradeShareAgreementHandler{svc: svc},
			"evaluations":            &EvaluationHandler{svc: svc},
			"user_biz_stances":       &UserBizStanceHandler{svc: svc},
			"biz_stance_counts":      &BizStanceCountHandler{svc: svc},
			"comments":               &CommentHandler{svc: svc},
		},
	}
}

func (g *MySQLBinlogConsumer) Start() error {
	cg, err := sarama.NewConsumerGroupFromClient("points_maintenance", g.client)
	if err != nil {
		return err
	}
	go func() {
		er := cg.Consume(context.Background(), []string{mysqlBinlogEvent}, saramax.NewHandler(g.l, g.Consume))
		if er != nil {
			g.l.Error("退出了消费循环异常", logger.Error(er))
		}
	}()
	return nil
}

func (g *MySQLBinlogConsumer) Consume(msg *sarama.ConsumerMessage, event canalx.Message[any]) error {
	// 将不同的data路由到不同的逻辑
	table := event.Table
	if handler, ok := g.handlers[table]; ok {
		// 使用找到的处理器处理消息
		return handler.Handle(event)
	} else {
		return nil
	}
}
