package whats

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
	rpc.RegisterArtPR(&artWhats{actions: make(map[protobuf.Action]rpc.ArtPRAction)})
}

type artWhats struct {
	sync.Mutex
	actions map[protobuf.Action]rpc.ArtPRAction
}

func (a *artWhats) RegisterAction(actionType protobuf.Action, ac rpc.ArtPRAction) {
	a.Lock()
	defer a.Unlock()
	a.actions[actionType] = ac
}

func (a *artWhats) GetType() string {
	return consts.PRWhats
}

func (a *artWhats) Handler(logger log.Logger, p *rpc.ProxyServer, in *protobuf.RequestMessage) (res *protobuf.ResponseMessage, err error) {
	_ = level.Info(logger).Log(a.GetType())
	action, ok := a.actions[in.Action]
	if !ok {
		comment := fmt.Sprintf("implement not found for pr action %s, forgot register?", in.Action)
		_ = level.Error(logger).Log("err", comment)
		err = gerror.New(comment)
		return
	}

	return action.Handler(logger, p, in)
}
