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

func TestGetOnline(t *testing.T) {
	conn, err := grpc.Dial("127.0.0.1:50052", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	c := protobuf.NewArthasClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	data2 := &telegram.GetOnlineAccountDetail{
		Phone: []string{"344323422"},
	}

	req7 := getOnline(data2)
	r7, err := c.Connect(ctx, req7)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息：%s", r7.GetActionResult())

}

func getOnline(data *telegram.GetOnlineAccountDetail) *protobuf.RequestMessage {

	req := &protobuf.RequestMessage{
		Action: protobuf.Action_GET_ONLINE_ACCOUNTS,
		Type:   "telegram",
		ActionDetail: &protobuf.RequestMessage_GetOnlineAccountsDetail{
			GetOnlineAccountsDetail: &protobuf.GetOnlineAccountsDetail{
				Phone: data.Phone,
			},
		},
	}
	return req
}
