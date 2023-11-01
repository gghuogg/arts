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

func TestGetHistory(t *testing.T) {

	accountNum := uint64(15022186268)

	conn, err := grpc.Dial("127.0.0.1:50052", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	c := protobuf.NewArthasClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	data := telegram.GetHistoryDetail{
		Self:       accountNum,
		Other:      "5824006024",
		Limit:      1000,
		OffsetDate: 0,
	}

	req4 := getHistory(data)
	r4, err := c.Connect(ctx, req4)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息：%s", r4.GetActionResult())

}

func getHistory(data telegram.GetHistoryDetail) *protobuf.RequestMessage {

	req := &protobuf.RequestMessage{
		Account: data.Self,
		Action:  protobuf.Action_Get_MSG_HISTORY,
		Type:    "telegram",
		ActionDetail: &protobuf.RequestMessage_GetMsgHistory{
			GetMsgHistory: &protobuf.GetMsgHistory{
				Self:      data.Self,
				Other:     data.Other,
				Limit:     int32(data.Limit),
				OffsetDat: int64(data.OffsetDate),
			},
		},
	}
	return req
}
