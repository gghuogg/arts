package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/errors/gerror"
	"sync"
)

func init() {
	rpc.RegisterArtIM(&artTg{actions: make(map[protobuf.Action]rpc.ArtIMAction)})
}

var (
	notAccountActions = []protobuf.Action{protobuf.Action_SYNC_APP_INFO, protobuf.Action_LOGIN, protobuf.Action_SEND_CODE, protobuf.Action_SEND_VCARD_MESSAGE}
)

type artTg struct {
	sync.Mutex
	actions map[protobuf.Action]rpc.ArtIMAction
}

func (a *artTg) RegisterAction(actionType protobuf.Action, ac rpc.ArtIMAction) {
	a.Lock()
	defer a.Unlock()
	a.actions[actionType] = ac
}

func (a *artTg) GetType() string {
	return consts.IMTg
}

func (a *artTg) Handler(logger log.Logger, handler *rpc.Handler, in *protobuf.RequestMessage) (res *protobuf.ResponseMessage, err error) {
	_ = level.Info(logger).Log(a.GetType())
	action, ok := a.actions[in.Action]
	if !ok {
		comment := fmt.Sprintf("implement not found for im action %s, forgot register?", in.Action)
		_ = level.Error(logger).Log("err", comment)
		err = gerror.New(comment)
		return
	}
	var check = true
	for _, accountAction := range notAccountActions {
		if in.Action == accountAction {
			check = false
			break
		}
	}
	if check {
		_, ok = handler.GetTgAccountsSync().Load(in.GetAccount())
		if !ok {
			res = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: fmt.Sprintf("Account %d is not logged in", in.Account)}
		}
	}

	return action.Handler(logger, handler, in)
}
