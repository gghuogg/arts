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

func TestCreateGroup(t *testing.T) {

	accountNum := uint64(19566961252)
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

	data1 := &telegram.AddGroupMemberDetail{
		Initiator:  15712294113,
		GroupTitle: "测试群",
		AddMembers: []string{"8618028658256"},
	}
	req6 := addGroupMember(data1)
	r6, err := c.Connect(ctx, req6)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息：%s", r6.GetActionResult())

}

func addGroupMember(data *telegram.AddGroupMemberDetail) *protobuf.RequestMessage {

	detail := &protobuf.UintkeyStringvalue{}
	detail.Key = data.Initiator
	for _, v := range data.AddMembers {
		detail.Values = append(detail.Values, v)
	}
	req := &protobuf.RequestMessage{
		Action:  protobuf.Action_ADD_GROUP_MEMBER,
		Type:    "telegram",
		Account: detail.Key,
		ActionDetail: &protobuf.RequestMessage_AddGroupMemberDetail{
			AddGroupMemberDetail: &protobuf.AddGroupMemberDetail{
				GroupName: data.GroupTitle,
				Detail:    detail,
			},
		},
	}
	return req
}
