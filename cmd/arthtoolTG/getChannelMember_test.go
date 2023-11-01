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

func TestGetChannelMember(t *testing.T) {
	accountNum := uint64(601164247658)

	conn, err := grpc.Dial("127.0.0.1:50052", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	c := protobuf.NewArthasClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	/*
				SearchType:
				name 按姓名查询参会人员
				recent 仅获取最近的参与者
				admins  仅获取管理员参与者
			    kicked 仅获取已踢出的参与者
			    bots   仅获取机器人参与者
		        banned  仅获取被禁止的参与者
				mentions 获取回复线程的参与者，目前不填写Top_msg_id可获得所有参与者
	*/

	data := telegram.GetChannelMemberDetail{
		Sender:     accountNum,
		Channel:    "1279877202",
		Offset:     0,
		Limit:      0,
		SearchType: "name",
		TopMsgId:   0,
	}
	req3 := getChannelMembers(data)
	r3, err := c.Connect(ctx, req3)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息: %s", r3.GetActionResult())
}

func getChannelMembers(data telegram.GetChannelMemberDetail) *protobuf.RequestMessage {

	req := &protobuf.RequestMessage{
		Action: protobuf.Action_GET_CHANNEL_MEMBER,
		Type:   "telegram",
		ActionDetail: &protobuf.RequestMessage_GetChannelMemberDetail{
			GetChannelMemberDetail: &protobuf.GetChannelMemberDetail{
				Sender:     data.Sender,
				Channel:    data.Channel,
				SearchType: data.SearchType,
				Offset:     int64(data.Offset),
				Limit:      int64(data.Limit),
				TopMsgId:   int64(data.TopMsgId),
			},
		},
	}
	return req
}
