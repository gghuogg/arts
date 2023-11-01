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

func TestSendCode(t *testing.T) {

	accountNum := uint64(16893992488)
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
	//////
	//code := "66150"
	////
	//time.Sleep(5 * time.Second)
	//req3 := sendCode(accountNum, code)
	//r3, err := c.Connect(ctx, req3)
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//log.Printf("返回消息: %s", r2.GetActionResult())
	//////
	code := "8465"
	//
	time.Sleep(5 * time.Second)
	req3 := sendCode(accountNum, code)
	r3, err := c.Connect(ctx, req3)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息: %s", r3.GetActionResult())

}

func sendCode(account uint64, code string) *protobuf.RequestMessage {

	sendCodeMap := make(map[uint64]string)
	sendCodeMap[account] = code

	detail := &protobuf.SendCodeAction{
		SendCode: sendCodeMap,
		LoginId:  "f12beb92-7090-499a-b037-550d49f67114",
	}

	req := &protobuf.RequestMessage{
		Action:  protobuf.Action_SEND_CODE,
		Type:    "telegram",
		Account: account,
		ActionDetail: &protobuf.RequestMessage_SendCodeDetail{
			SendCodeDetail: &protobuf.SendCodeDetail{
				Details: detail,
			},
		},
	}
	return req
}
