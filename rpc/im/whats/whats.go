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
	rpc.RegisterArtIM(&artWhats{actions: make(map[protobuf.Action]rpc.ArtIMAction)})
}

type artWhats struct {
	sync.Mutex
	actions map[protobuf.Action]rpc.ArtIMAction
}

func (a *artWhats) GetType() string {
	return consts.IMWhats
}

func (a *artWhats) RegisterAction(actionType protobuf.Action, ac rpc.ArtIMAction) {
	a.Lock()
	defer a.Unlock()
	a.actions[actionType] = ac
}

func (a *artWhats) Handler(logger log.Logger, handler *rpc.Handler, in *protobuf.RequestMessage) (res *protobuf.ResponseMessage, err error) {
	_ = level.Info(logger).Log(a.GetType())
	action, ok := a.actions[in.Action]
	if !ok {
		comment := fmt.Sprintf("implement not found for im action %s, forgot register?", in.Action)
		_ = level.Error(logger).Log("err", comment)
		err = gerror.New(comment)
		return
	}
	return action.Handler(logger, handler, in)
}
