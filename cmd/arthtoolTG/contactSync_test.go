package main

import (
	"arthas/protobuf"
	"context"
	"fmt"
	"github.com/iyear/tdl/pkg/consts"
	"google.golang.org/grpc"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestContactSync(t *testing.T) {

	accountNum := uint64(15022186268)
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
	//req1 := syncAppInfo(account)
	//r1, err := c.Connect(ctx, req1)
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//log.Printf("返回消息: %s", r1.GetActionResult())
	//
	//time.Sleep(5 * time.Second)
	//req2 := login(account)
	//r2, err := c.Connect(ctx, req2)
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//log.Printf("返回消息: %s", r2.GetActionResult())

	//time.Sleep(4 * time.Second)
	req3 := syncContact(accountNum)
	r3, err := c.Connect(ctx, req3)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息：%s", r3.GetActionResult())

}

func syncContact(account uint64) *protobuf.RequestMessage {
	req := &protobuf.RequestMessage{
		Account: account,
		Action:  protobuf.Action_SYNC_CONTACTS,
		Type:    "telegram",
		ActionDetail: &protobuf.RequestMessage_SyncContactDetail{
			SyncContactDetail: &protobuf.SyncContactDetail{
				Details: []*protobuf.UintkeyUintvalue{
					{Key: account, Values: []uint64{8618818877128}},
				},
			},
		},
	}
	return req
}
