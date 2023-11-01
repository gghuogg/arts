package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/errors/gerror"
	"strconv"
	"sync"
)

func init() {
	rpc.RegisterArtPR(&artTg{actions: make(map[protobuf.Action]rpc.ArtPRAction)})
}

// 不需要登录即可调用的接口
var (
	notAccountActions = []protobuf.Action{protobuf.Action_SYNC_APP_INFO, protobuf.Action_LOGIN, protobuf.Action_SEND_CODE, protobuf.Action_IMPORT_TG_SESSION, protobuf.Action_GET_ONLINE_ACCOUNTS}
)

type artTg struct {
	sync.Mutex
	actions map[protobuf.Action]rpc.ArtPRAction
}

func (a *artTg) RegisterAction(actionType protobuf.Action, ac rpc.ArtPRAction) {
	a.Lock()
	defer a.Unlock()
	a.actions[actionType] = ac
}

func (a *artTg) GetType() string {
	return consts.PRTg
}

func (a *artTg) Handler(logger log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (res *protobuf.ResponseMessage, err error) {
	_ = level.Info(logger).Log(a.GetType())
	action, ok := a.actions[in.Action]
	if !ok {
		comment := fmt.Sprintf("implement not found for pr action %s, forgot register?", in.Action)
		_ = level.Error(logger).Log("err", comment)
		err = gerror.New(comment)
		return
	}
	// 如果需要登录访问的
	if !containsAction(notAccountActions, in.Action) {
		// 验证是否登录了
		v1, ok1 := p.Handler.GetServerMap().Load(in.Account)
		if !ok1 {
			errRes := fmt.Sprintf("%s user not logged in ", strconv.FormatUint(in.Account, 10))

			level.Error(p.GetLogger()).Log("msg", errRes)
			err = gerror.New(errRes)
			return
		}
		_, ok2 := p.Handler.StreamMap.Load(v1.(string))
		if !ok2 {
			errRes := fmt.Sprintf("%s The management terminal and the server are disconnected", v1.(string))

			level.Error(p.GetLogger()).Log("msg", errRes)
			err = gerror.New(errRes)
			return
		}
	}

	return action.Handler(logger, p, in)
}

// 检查指定的动作是否存在于切片中
func containsAction(actions []protobuf.Action, action protobuf.Action) bool {
	for _, a := range actions {
		if a == action {
			return true
		}
	}
	return false
}
