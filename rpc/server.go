package rpc

import (
	"arthas/protobuf"
	"arthas/util"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"io"
	"time"
)

type server struct {
	//protobuf.UnimplementedArthasServer
	protobuf.UnimplementedArthasStreamServer
	logger  log.Logger
	Handler *Handler
	ticker  *time.Ticker
}

func (s server) Connect(connectServer protobuf.ArthasStream_ConnectServer) error {

	// 创建一个channel用于接收错误
	errChan := make(chan error, 1)

	// 接收管理端发来的ping消息
	go func() {
		for {
			in, err := connectServer.Recv()
			if err == io.EOF {
				// 流结束
				level.Error(s.logger).Log("msg", "receive stream off", "err", err)
			}
			if err != nil {
				errChan <- err
			}
			// 如果收到ping，则回应pong
			if in.GetPingMessage() == "ping" {
				//ip, _ := getIPFromContext(connectServer.Context())
				//fmt.Println("收到:", ip)
				endpoints := util.CalculateListenedEndpoints(s.Handler.options.ListenAddress)
				if err := connectServer.Send(&protobuf.ResponseMessage{PongMessage: endpoints[0].String() + "/" + "pong"}); err != nil {
					errChan <- err
				}
				continue
			}

			if in != nil {
				im, err := GetIM(in.GetType())
				if err != nil {
					err = connectServer.Send(&protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL})
					if err != nil {
						level.Error(s.logger).Log("msg", "send fail", "err", err)
					}
					continue
				}
				res, err1 := im.Handler(s.logger, s.Handler, in)
				if err1 != nil {
					err = connectServer.Send(&protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL})
					if err != nil {
						level.Error(s.logger).Log("msg", "send fail", "err", err)
					}
					continue
				} else {
					_ = connectServer.Send(res)
					continue
				}

			}
		}
	}()

	for {
		select {
		case <-connectServer.Context().Done():
			level.Info(s.logger).Log("msg", "Management disconnected")
			return connectServer.Context().Err()
		case err := <-errChan:
			// 如果接收goroutine返回错误
			level.Error(s.logger).Log("msg", "receive from management server fail", "err", err)
			return err
		}
	}

}
