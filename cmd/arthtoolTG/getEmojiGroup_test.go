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

func TestGetEmoji(t *testing.T) {
	accountNum := uint64(601164247658)

	conn, err := grpc.Dial("127.0.0.1:50052", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	c := protobuf.NewArthasClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	//
	//

	data := telegram.GetEmojiGroupsDetail{
		Sender: accountNum,
	}
	req3 := getEmojiGroup(data)
	r3, err := c.Connect(ctx, req3)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息: %s", r3.GetActionResult())

}

func getEmojiGroup(data telegram.GetEmojiGroupsDetail) *protobuf.RequestMessage {
	req := &protobuf.RequestMessage{
		Action:  protobuf.Action_GET_EMOJI_GROUP,
		Type:    "telegram",
		Account: data.Sender,
		ActionDetail: &protobuf.RequestMessage_GetEmojiGroupDetail{
			GetEmojiGroupDetail: &protobuf.GetEmojiGroupsDetail{
				Sender: data.Sender,
			},
		},
	}

	return req
}
