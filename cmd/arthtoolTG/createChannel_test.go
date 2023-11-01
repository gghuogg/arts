package main

import (
	"arthas/protobuf"
	"arthas/telegram"
	"context"
	"fmt"
	"github.com/iyear/tdl/pkg/consts"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestCreateChannel(t *testing.T) {
	accountNum := uint64(628386440647)
	account := AccountDetail{
		PhoneNumber: strconv.FormatUint(accountNum, 10),
		AppId:       24242552,
		AppHash:     "12adb15f288feeb37101c331fe2789ad",
		Device:      consts.Device,
	}
	fmt.Println(account)

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

	data1 := &telegram.CreateChannelDetail{
		Creator:            601164247658,
		ChannelTitle:       "tdlibtest9003",
		ChannelDescription: "test",
		ChannelUserName:    "tdlibtest9003",
		IsSuperGroup:       true,
		ChannelMember:      []string{"zhuangkl"},
	}
	req6 := createChannel(data1)
	r6, err := c.Connect(ctx, req6)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息：%s", r6.GetActionResult())

}

func createChannel(data *telegram.CreateChannelDetail) *protobuf.RequestMessage {

	detail := &protobuf.UintkeyStringvalue{}
	detail.Key = data.Creator
	for _, v := range data.ChannelMember {
		detail.Values = append(detail.Values, v)
	}
	req := &protobuf.RequestMessage{
		Action:  protobuf.Action_CREATE_CHANNEL,
		Type:    "telegram",
		Account: detail.Key,
		ActionDetail: &protobuf.RequestMessage_CreateChannelDetail{
			CreateChannelDetail: &protobuf.CreateChannelDetail{
				ChannelUserName:    data.ChannelUserName,
				ChannelDescription: data.ChannelDescription,
				ChannelTitle:       data.ChannelTitle,
				Detail:             detail,
				IsChannel:          data.IsChannel,
				IsSuperGroup:       data.IsSuperGroup,
			},
		},
	}
	return req
}
