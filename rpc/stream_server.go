package rpc

import (
	"arthas/callback"
	"arthas/prometheus/sendUtils"
	"arthas/protobuf"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"time"
)

type streamServer struct {
	protobuf.UnimplementedArthasStreamServer
	logger  log.Logger
	Handler *Handler
}

func (s streamServer) Connect(stream protobuf.ArthasStream_ConnectServer) error {
	level.Info(s.logger).Log("msg", "Grpc stream connect.")
	startTime := time.Now()
	defer func(startTime time.Time, name string, path string) {
		err := sendUtils.SendCalculationTimeToPrometheus(startTime, name, path)
		if err != nil {
			level.Error(s.logger).Log("msg", "Send CalculationTime to prometheus failed.")
		}
	}(startTime, "rpc-stream-server-Connect", "/apiPerformance")
	for s.Handler.callbackType == "stream" {
		select {
		case callbacks := <-s.Handler.callbackChan.LoginCallbackChan:
			msg := s.sendLoginCallback(callbacks)
			return s.send(stream, msg)
		case callbacks := <-s.Handler.callbackChan.TextMsgCallbackChan:
			msg := s.sendTextMsgCallback(callbacks)
			return s.send(stream, msg)
		case callbacks := <-s.Handler.callbackChan.ReadMsgCallbackChan:
			msg := s.sendReadMsgCallback(callbacks)
			return s.send(stream, msg)
		case <-s.Handler.context.Done():
			return nil
		}
	}
	return nil
}
func (s streamServer) send(stream protobuf.ArthasStream_ConnectServer, msg *protobuf.ResponseMessage) error {
	if err := stream.Send(msg); err != nil {
		return err
	}
	return nil
}

func (s streamServer) sendLoginCallback(callbackChanMsg []callback.LoginCallback) *protobuf.ResponseMessage {
	level.Info(s.logger).Log("msg", "Send login callback.")
	callbackInfo := protobuf.LoginCallbacks{}
	for _, loginCallback := range callbackChanMsg {
		callbackInfo.Results = append(callbackInfo.Results, &protobuf.LoginCallback{
			UserJid:     loginCallback.UserJid,
			LoginStatus: loginCallback.LoginStatus,
			ProxyUrl:    loginCallback.ProxyUrl,
			Comment:     loginCallback.Comment,
		})
	}
	message := &protobuf.ResponseMessage{
		ActionResult:  0,
		AccountStatus: nil,
		CallbackInfo:  &protobuf.ResponseMessage_LoginCallbacks{LoginCallbacks: &callbackInfo},
	}
	return message
}

func (s streamServer) sendTextMsgCallback(callbacks []callback.TextMsgCallback) *protobuf.ResponseMessage {
	level.Info(s.logger).Log("msg", "Send text msg callback.")
	callbackInfo := protobuf.TextMsgCallbacks{}
	for _, msgCallback := range callbacks {
		callbackInfo.Results = append(callbackInfo.Results, &protobuf.TextMsgCallback{
			Sender:   msgCallback.Sender,
			Receiver: msgCallback.Receiver,
			SendText: msgCallback.SendText,
			SendTime: msgCallback.SendTime.Unix(),
			ReqId:    msgCallback.ReqId,
			Read:     true,
		})
	}
	message := &protobuf.ResponseMessage{
		ActionResult:  0,
		AccountStatus: nil,
		CallbackInfo:  &protobuf.ResponseMessage_TextMsgCallbacks{TextMsgCallbacks: &callbackInfo},
	}
	return message
}

func (s streamServer) sendReadMsgCallback(callbacks []callback.ReadMsgCallback) *protobuf.ResponseMessage {
	level.Info(s.logger).Log("msg", "Send read msg callback.")
	callbackInfo := protobuf.ReadMsgCallbacks{}
	for _, msgCallback := range callbacks {
		callbackInfo.Results = append(callbackInfo.Results, &protobuf.ReadMsgCallback{
			ReqId: msgCallback.ReqId,
		})
	}
	message := &protobuf.ResponseMessage{
		ActionResult:  0,
		AccountStatus: nil,
		CallbackInfo:  &protobuf.ResponseMessage_ReadMsgCallbacks{ReadMsgCallbacks: &callbackInfo},
	}
	return message
}
