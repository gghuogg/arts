package main

import (
	"arthas/protobuf"
	"arthas/telegram"
	"context"
	"google.golang.org/grpc"
	"log"
	"testing"
	"time"
)

func TestLeave(t *testing.T) {
	accountNum := uint64(601164247658)

	conn, err := grpc.Dial("127.0.0.1:50052", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	c := protobuf.NewArthasClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	data1 := &telegram.LeaveDetail{
		Sender: accountNum,
		Leaves: []string{"4020580410"},
	}

	req3 := leave(data1)
	r3, err := c.Connect(ctx, req3)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息: %s", r3.GetActionResult())
}

func leave(data *telegram.LeaveDetail) *protobuf.RequestMessage {

	detail := &protobuf.UintkeyStringvalue{}
	detail.Key = data.Sender
	for _, v := range data.Leaves {
		detail.Values = append(detail.Values, v)
	}
	req := &protobuf.RequestMessage{
		Action:  protobuf.Action_LEAVE,
		Type:    "telegram",
		Account: detail.Key,
		ActionDetail: &protobuf.RequestMessage_LeaveDetail{
			LeaveDetail: &protobuf.LeaveDetail{
				Detail: detail,
			},
		},
	}
	return req
}
