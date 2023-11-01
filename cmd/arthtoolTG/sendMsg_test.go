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

func TestSendMsg(t *testing.T) {

	accountNum := uint64(601164247658)
	account := AccountDetail{
		PhoneNumber: strconv.FormatUint(accountNum, 10),
		AppId:       24242552,
		AppHash:     "12adb15f288feeb37101c331fe2789ad",
		Device:      consts.Device,
	}
	fmt.Println(account)

	list := make([]*telegram.SendMessageDetail, 0)
	sendText1 := &telegram.SendMessageDetail{
		Sender:   accountNum,
		Receiver: "8618028658256",
		SendText: []string{"wa ha ha"},
	}
	list = append(list, sendText1)

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

	//
	//time.Sleep(5 * time.Second)
	req4 := sendMessage(accountNum, list)
	r4, err := c.Connect(ctx, req4)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息：%s", r4.GetActionResult())

}
