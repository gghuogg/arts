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

func TestInvite(t *testing.T) {
	//accountNum := uint64(19566961252)

	conn, err := grpc.Dial("127.0.0.1:50052", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	c := protobuf.NewArthasClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	data2 := &telegram.InviteToChannelDetail{
		Invite:  16893992489,
		Channel: "1774453281",
		Invited: []string{"xuedng"},
	}
	req7 := invite(data2)
	r7, err := c.Connect(ctx, req7)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息：%s", r7.GetActionResult())
}

func invite(data *telegram.InviteToChannelDetail) *protobuf.RequestMessage {

	detail := &protobuf.UintkeyStringvalue{}
	detail.Key = data.Invite
	for _, v := range data.Invited {
		detail.Values = append(detail.Values, v)
	}
	req := &protobuf.RequestMessage{
		Action:  protobuf.Action_INVITE_TO_CHANNEL,
		Type:    "telegram",
		Account: detail.Key,
		ActionDetail: &protobuf.RequestMessage_InviteToChannelDetail{
			InviteToChannelDetail: &protobuf.InviteToChannelDetail{
				Channel: data.Channel,
				Detail:  detail,
			},
		},
	}
	return req
}
