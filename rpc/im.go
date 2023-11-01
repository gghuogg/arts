package rpc

import (
	alog "arthas/log"
	"arthas/protobuf"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/errors/gerror"
	"sync"
)

// ArtIM im接口 实现该接口即可封装各IM的处理逻辑
type ArtIM interface {
	GetType() string                                                                                                     // 获取im类型的方法
	Handler(logger log.Logger, handler *Handler, in *protobuf.RequestMessage) (res *protobuf.ResponseMessage, err error) //处理消息的方法
	RegisterAction(actionType protobuf.Action, action ArtIMAction)                                                       //处理消息的方法
}

// ArtIMAction  im动作接口
type ArtIMAction interface {
	Handler(logger log.Logger, handler *Handler, in *protobuf.RequestMessage) (res *protobuf.ResponseMessage, err error) //处理动作的方法
}

// artIMManager im管理
type artIMManager struct {
	sync.Mutex
	list map[string]ArtIM // 维护的im列表
}

var artIMs = &artIMManager{
	list: make(map[string]ArtIM),
}

// RegisterArtIM 注册im
func RegisterArtIM(cs ArtIM) {
	artIMs.Lock()
	defer artIMs.Unlock()
	imType := cs.GetType()
	if _, ok := artIMs.list[imType]; ok {
		logger := alog.New(&alog.Config{})
		_ = level.Error(logger).Log("msg", fmt.Sprintf("im.RegisterArtIM imType:%s duplicate registration.", imType))
		return
	}
	artIMs.list[imType] = cs
}

func GetIM(imType string) (ArtIM, error) {
	if item, ok := artIMs.list[imType]; ok {
		return item, nil
	}
	return nil, gerror.Newf("implement not found for im %s, forgot register?", imType)
}
