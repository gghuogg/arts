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

func TestLogoutUser(t *testing.T) {

	accountNum := uint64(601164247658)
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

	//
	time.Sleep(5 * time.Second)
	req3 := logout(account)
	r3, err := c.Connect(ctx, req3)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息: %s", r3.GetActionResult())

}

func logout(account AccountDetail) *protobuf.RequestMessage {
	logoutDetail := make(map[uint64]*protobuf.LogoutDetail)
	ld := &protobuf.LogoutDetail{
		ProxyUrl: "http://5.161.105.73:39000",
	}

	user, _ := strconv.ParseUint(account.PhoneNumber, 10, 64)
	logoutDetail[user] = ld

	req := &protobuf.RequestMessage{
		Action:  protobuf.Action_LOGOUT,
		Type:    "telegram",
		Account: user,
		ActionDetail: &protobuf.RequestMessage_LogoutAction{
			LogoutAction: &protobuf.LogoutAction{
				LogoutDetail: logoutDetail,
			},
		},
	}
	return req
}
