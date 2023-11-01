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

func TestGetGroupMember(t *testing.T) {
	accountNum := uint64(601164247658)

	conn, err := grpc.Dial("127.0.0.1:50052", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	c := protobuf.NewArthasClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	data := telegram.GetGroupMembers{
		Initiator: accountNum,
		ChatId:    int64(4061247392),
	}
	req3 := getGroupMembers(data)
	r3, err := c.Connect(ctx, req3)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息: %s", r3.GetActionResult())
}

func getGroupMembers(data telegram.GetGroupMembers) *protobuf.RequestMessage {

	req := &protobuf.RequestMessage{
		Action: protobuf.Action_GET_GROUP_MEMBERS,
		Type:   "telegram",
		ActionDetail: &protobuf.RequestMessage_GetGroupMembersDetail{
			GetGroupMembersDetail: &protobuf.GetGroupMembersDetail{
				Account: data.Initiator,
				ChatId:  data.ChatId,
			},
		},
	}
	return req
}
