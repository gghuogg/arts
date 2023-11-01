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

// ArtPR pr接口 实现该接口即可封装各PR的处理逻辑
type ArtPR interface {
	GetType() string                                                                                                   // 获取PR类型的方法
	Handler(logger log.Logger, p *ProxyServer, in *protobuf.RequestMessage) (res *protobuf.ResponseMessage, err error) //处理消息的方法
	RegisterAction(actionType protobuf.Action, action ArtPRAction)                                                     //处理消息的方法
}

// ArtPRAction  PR动作接口
type ArtPRAction interface {
	Handler(logger log.Logger, p *ProxyServer, in *protobuf.RequestMessage) (res *protobuf.ResponseMessage, err error) //处理动作的方法
}

// artPRManager PR管理
type artPRManager struct {
	sync.Mutex
	list map[string]ArtPR // 维护的PR列表
}

var artPRs = &artPRManager{
	list: make(map[string]ArtPR),
}

// RegisterArtPR 注册PR
func RegisterArtPR(cs ArtPR) {
	artPRs.Lock()
	defer artPRs.Unlock()
	prType := cs.GetType()
	if _, ok := artPRs.list[prType]; ok {
		logger := alog.New(&alog.Config{})
		_ = level.Error(logger).Log("msg", fmt.Sprintf("pr.RegisterArtPR prType:%s duplicate registration.", prType))
		return
	}
	artPRs.list[prType] = cs
}

func GetPR(prType string) (ArtPR, error) {
	if item, ok := artPRs.list[prType]; ok {
		return item, nil
	}
	return nil, gerror.Newf("implement not found for pr %s, forgot register?", prType)
}
