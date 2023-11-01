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

func TestJoinByLink(t *testing.T) {
	accountNum := uint64(601164247658)

	conn, err := grpc.Dial("127.0.0.1:50052", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	c := protobuf.NewArthasClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	data1 := &telegram.JoinByLinkDetail{
		Sender: accountNum,
		Links:  []string{"https://t.me/juzheng213"},
	}

	req3 := joinByLink(data1)
	r3, err := c.Connect(ctx, req3)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息: %s", r3.GetActionResult())
}

func joinByLink(data *telegram.JoinByLinkDetail) *protobuf.RequestMessage {

	detail := &protobuf.UintkeyStringvalue{}
	detail.Key = data.Sender
	for _, v := range data.Links {
		detail.Values = append(detail.Values, v)
	}
	req := &protobuf.RequestMessage{
		Action:  protobuf.Action_JOIN_BY_LINK,
		Type:    "telegram",
		Account: detail.Key,
		ActionDetail: &protobuf.RequestMessage_JoinByLinkDetail{
			JoinByLinkDetail: &protobuf.JoinByLinkDetail{
				Detail: detail,
			},
		},
	}
	return req
}
