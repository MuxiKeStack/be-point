package service

import (
	"context"
	"fmt"
	answerv1 "github.com/MuxiKeStack/be-api/gen/proto/answer/v1"
	evaluationv1 "github.com/MuxiKeStack/be-api/gen/proto/evaluation/v1"
	stancev1 "github.com/MuxiKeStack/be-api/gen/proto/stance/v1"
	"github.com/MuxiKeStack/be-point/domain"
	"github.com/MuxiKeStack/be-point/repository"
	"time"
)

const (
	INSERT = "INSERT"
	UPDATE = "UPDATE"
)

type PointMaintenanceService interface {
	HandleGradeShareAgreement(ctx context.Context, typ string, uid int64, updateIsSigned bool, oldIsSigned bool) error
	HandleEvaluation(ctx context.Context, typ string, uid int64, id int64, updateContent string, updateStatus evaluationv1.EvaluationStatus,
		oldStatus evaluationv1.EvaluationStatus) error
	HandleUserBizStance(ctx context.Context, typ string, uid int64, updateStance stancev1.Stance, oldStance stancev1.Stance) error
	HandleBizStanceCount(ctx context.Context, typ string, biz stancev1.Biz, bizId int64, updateSupportCnt int64, oldSupportCnt int64) error
	HandleComment(ctx context.Context, typ string, uid int64) error
}

// 这是一个并发不安全的实现，binlog中的更改要一个一个的顺序处理，并行处理，这和使用canal有本质的关联
type pointMaintenanceService struct {
	repo       repository.PointRepository
	uidGetters map[stancev1.Biz]UIDGetter
}

func NewPointMaintenanceService(repo repository.PointRepository, evaluationClient evaluationv1.EvaluationServiceClient,
	answerClient answerv1.AnswerServiceClient) PointMaintenanceService {
	return &pointMaintenanceService{
		repo: repo,
		uidGetters: map[stancev1.Biz]UIDGetter{
			stancev1.Biz_Evaluation: &EvaluationUIDGetter{evaluationClient: evaluationClient},
			stancev1.Biz_Answer:     &AnswerUIDGetter{answerClient: answerClient},
		},
	}
}

func (p *pointMaintenanceService) HandleGradeShareAgreement(ctx context.Context, typ string, uid int64, updateIsSigned bool, oldIsSigned bool) error {
	if typ == INSERT {
		if updateIsSigned {
			// change ==> += 15
			return p.repo.Change(ctx, domain.PointChange{
				Uid:          uid,
				ChangeAmount: +15,
				Reason:       "签约",
				Source:       "grade_share_agreement",
			})
		}
	} else if typ == UPDATE {
		if updateIsSigned && !oldIsSigned {
			// change ==> += 15
			return p.repo.Change(ctx, domain.PointChange{
				Uid:          uid,
				ChangeAmount: +15,
				Reason:       "签约",
				Source:       "grade_share_agreement",
			})
		} else if !updateIsSigned && oldIsSigned {
			// change ==> -=15
			return p.repo.Change(ctx, domain.PointChange{
				Uid:          uid,
				ChangeAmount: -15,
				Reason:       "取消签约",
				Source:       "grade_share_agreement",
			})
		}
	}
	return nil
}

func (p *pointMaintenanceService) evaluationSource(id int64) string {
	return fmt.Sprintf("evaluation:%d", id)
}

func (p *pointMaintenanceService) HandleEvaluation(ctx context.Context, typ string, uid int64, id int64, updateContent string,
	updateStatus evaluationv1.EvaluationStatus, oldStatus evaluationv1.EvaluationStatus) error {
	if typ == INSERT {
		if updateStatus != evaluationv1.EvaluationStatus_Public {
			return nil
		}
		if len(updateContent) == 0 {
			// +3
			return p.repo.Change(ctx, domain.PointChange{
				Uid:          uid,
				ChangeAmount: +3,
				Reason:       "发布空内容课评",
				Source:       p.evaluationSource(id),
			})
		} else {
			// +5
			return p.repo.Change(ctx, domain.PointChange{
				Uid:          uid,
				ChangeAmount: +5,
				Reason:       "发布有内容课评",
				Source:       p.evaluationSource(id),
			})
		}
	} else if typ == UPDATE {
		if updateStatus == oldStatus {
			return nil
		}
		switch {
		case oldStatus == evaluationv1.EvaluationStatus_Public && updateStatus == evaluationv1.EvaluationStatus_Private:
			// 查库判断第一次创建时加的积分，然后减掉
			source := p.evaluationSource(id)
			epc, err := p.repo.FindOldestBySource(ctx, source)
			if err != nil {
				return err
			}
			return p.repo.Change(ctx, domain.PointChange{
				Uid:          uid,
				ChangeAmount: -epc.ChangeAmount,
				Reason:       "隐藏课评",
				Source:       source,
			})
		case (oldStatus == evaluationv1.EvaluationStatus_Public || oldStatus == evaluationv1.EvaluationStatus_Private) && updateStatus == evaluationv1.EvaluationStatus_Folded:
			// 无脑直接扣10分，被折叠（删除）只是一个附带效应，这个附带效应产生扣分忽略，其是一定小于违规扣分的
			return p.repo.Change(ctx, domain.PointChange{
				Uid:          uid,
				ChangeAmount: -10,
				Reason:       "违规课评被折叠",
				Source:       p.evaluationSource(id),
			})
		case (oldStatus == evaluationv1.EvaluationStatus_Private || oldStatus == evaluationv1.EvaluationStatus_Folded) && updateStatus == evaluationv1.EvaluationStatus_Public:
			// 根据内容加分，一定是update分数，先查出来旧的分数，然后加上
			epc, err := p.repo.FindOldestBySource(ctx, p.evaluationSource(id))
			if err != nil {
				return err
			}
			return p.repo.Change(ctx, domain.PointChange{
				Uid:          uid,
				ChangeAmount: +epc.ChangeAmount,
				Reason:       "重新公开课评",
				Source:       p.evaluationSource(id),
			})
		case oldStatus == evaluationv1.EvaluationStatus_Folded && updateStatus == evaluationv1.EvaluationStatus_Private:
			// 无操作
		}
	}
	return nil
}

func (p *pointMaintenanceService) bizStanceCountSource(biz stancev1.Biz, bizId int64) string {
	return fmt.Sprintf("biz_stance_count:<%d,%d>", biz, bizId) // todo
}
func (p *pointMaintenanceService) HandleBizStanceCount(ctx context.Context, typ string, biz stancev1.Biz, bizId int64, updateSupportCnt int64,
	oldSupportCnt int64) error {
	if typ != UPDATE && typ != INSERT {
		return nil
	}
	// updateCnt == 5 && oldCnt == 4 要去查出用户，并进行 +3
	if updateSupportCnt == 5 && oldSupportCnt == 4 {
		// 我在这里路由，根据不同的Biz路由的不同的rpc client然后GetDetail得到用户
		getter := p.uidGetters[biz]
		uid, err := getter.GetUID(ctx, bizId)
		if err != nil {
			return err
		}
		// 首次才会去加
		source := p.bizStanceCountSource(biz, bizId)
		_, err = p.repo.FindOldestBySource(ctx, source)
		if err == nil {
			return nil
		}
		if err == repository.ErrPointChangeNotFound {
			return p.repo.Change(ctx, domain.PointChange{
				Id:           0,
				Uid:          uid,
				ChangeAmount: +3,
				Reason:       "点赞数达到5",
				Source:       source,
			})
		}
		return err
	} else {
		return nil
	}
}

func (p *pointMaintenanceService) userBizStanceSource(uid int64) string {
	return fmt.Sprintf("user_biz_stance:%d:%s", uid, time.Now().Format(time.DateOnly))
}

func (p *pointMaintenanceService) HandleUserBizStance(ctx context.Context, typ string, uid int64, updateStance stancev1.Stance,
	oldStance stancev1.Stance) error {
	if typ != UPDATE && typ != INSERT {
		return nil
	}
	if updateStance != stancev1.Stance_Support {
		return nil
	}
	// 这里要判断一下是否已获得今日首次点击支持经验，没有获得的话就要加经验
	// 这个pointMaintenanceService实现是基于canal的binlog消息，必须要顺序处理
	// 这个实现的某些地方逻辑处理可能可以并行，但是总体上还是不能并行的，一部分可并行，一部分不可并行，会有些混乱，
	// 所以所有的方法通通都建立在该实现不并行的基础上来做
	src := p.userBizStanceSource(uid)
	ok, err := p.repo.Exists(ctx, src)
	if err != nil {
		return err
	}
	if ok {
		return nil
	}
	return p.repo.Change(ctx, domain.PointChange{
		Uid:          uid,
		ChangeAmount: +1,
		Reason:       "每日首次支持",
		Source:       src,
	})
}

func (p *pointMaintenanceService) commentSource(uid int64) string {
	return fmt.Sprintf("comment:%d:%s", uid, time.Now().Format(time.DateOnly))
}

func (p *pointMaintenanceService) HandleComment(ctx context.Context, typ string, uid int64) error {
	// 拿到今日评论奖励次数，<2就可以继续 +2
	if typ != INSERT {
		return nil
	}
	const PublishComment = "每日首两次发布评论"
	src := p.commentSource(uid)
	cnt, err := p.repo.CountChangeByReasonSource(ctx, PublishComment, src)
	if err != nil {
		return err
	}
	if cnt < 2 {
		// +2
		return p.repo.Change(ctx, domain.PointChange{
			Uid:          uid,
			ChangeAmount: +2,
			Reason:       PublishComment,
			Source:       src,
		})
	}
	return nil
}
