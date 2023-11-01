package main

import (
	"arthas/protobuf"
	"arthas/signal"
	"arthas/telegram"
	tg "github.com/gotd/td/telegram"
	"github.com/iyear/tdl/pkg/consts"

	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"os"
	"strconv"
	"time"
)

type AccountDetail struct {
	PhoneNumber string
	AppId       uint64
	AppHash     string
	Device      tg.DeviceConfig
}

func main() {
	//24242552, "12adb15f288feeb37101c331fe2789ad"
	accountNum := uint64(6283835577586)
	//account1 := AccountDetail{
	//	PhoneNumber: strconv.FormatUint(accountNum, 10),
	//	AppId:       29453132,
	//	AppHash:     "60e35bbbefb9d5cb6b395f1d04b1d588",
	//	Device:      consts.Device,
	//}
	//+601164247658
	account2 := AccountDetail{
		PhoneNumber: strconv.FormatUint(6283835577586, 10),
		AppId:       29453132,
		AppHash:     "60e35bbbefb9d5cb6b395f1d04b1d588",
		Device:      consts.Device,
	}

	accounts := make([]AccountDetail, 0)
	accounts = append(accounts, account2)

	cardList := make([]*telegram.ContactCardDetail, 0)
	myCntactCard := &telegram.ContactCardDetail{
		Sender:   accountNum,
		Receiver: "8618818877128",
		ContactCards: []telegram.ContactCard{
			{
				FirstName:   "lu",
				LastName:    "xuedong",
				PhoneNumber: "8618818877128",
			},
		},
	}

	cardList = append(cardList, myCntactCard)

	//list := make([]*telegram.SendMessageDetail, 0)
	//sendText1 := &telegram.SendMessageDetail{
	//	Sender:   15712294113,
	//	Receiver: 8618028658256,
	//	SendText: []string{"今天几点"},
	//}
	list := make([]*telegram.SendMessageDetail, 0)
	//sendText1 := &telegram.SendMessageDetail{
	//	Sender:   accountNum,
	//	Receiver: 8618818877128,
	//	SendText: []string{"23232323"},
	//}
	sendText2 := &telegram.SendMessageDetail{
		Sender:   accountNum,
		Receiver: "8618028658256",
		SendText: []string{"aaaaaa", "nihaonihao", "emmmmm"},
	}
	//sendText3 := &signal.SendMessageDetail{
	//	Sender:   accountNum,
	//	Receiver: 8615077731547,
	//	SendText: []string{"牛逼", "真牛逼", "好牛逼"},
	//}
	list = append(list, sendText2)

	b, err := os.ReadFile("cmd/arthtoolTG/testFile/picture.jpg")
	if err != nil {
		fmt.Println(err)
		return
	}

	photoList := make([]*telegram.SendImageFileDetail, 0)
	SendImageFileTest := &telegram.SendImageFileDetail{
		Sender:   accountNum,
		Receiver: strconv.FormatUint(8618818877128, 10),
		FileType: []telegram.SendImageFileType{
			{
				FileType:   "image/jpeg",
				SendType:   "url",
				Path:       "cmd/arthtoolTG/testFile/picture.jpg",
				Name:       "picture.jpg",
				ImageBytes: b,
			},
		},
	}
	photoList = append(photoList, SendImageFileTest)
	//
	b2, err := os.ReadFile("cmd/arthtoolTG/testFile/test.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	FileList := make([]*telegram.SendFileDetail, 0)
	FileTest := &telegram.SendFileDetail{
		Sender:   accountNum,
		Receiver: strconv.FormatUint(8618818877128, 10),
		FileType: []telegram.SendFileType{
			{
				FileType:  signal.SIGNALTXT,
				SendType:  "url",
				Path:      "http://grata.gen-code.top/grata/attachment/2023-09-28/cvu5vsha49eopdumvs.jpg",
				Name:      "test.txt",
				FileBytes: b2,
			},
		},
	}
	FileList = append(FileList, FileTest)

	//34.124.217.133:59877
	conn, err := grpc.Dial("127.0.0.1:50052", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	c := protobuf.NewArthasClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	////
	req1 := syncAppInfo(accounts)
	r1, err := c.Connect(ctx, req1)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息: %s", r1.GetActionResult())

	//time.Sleep(5 * time.Second)
	req2 := login(accounts)
	r2, err := c.Connect(ctx, req2)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息: %s", r2.GetActionResult())

	//time.Sleep(4 * time.Second)
	//req3 := syncContact()
	//r3, err := c.Connect(ctx, req3)
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//log.Printf("返回消息：%s", r3.GetActionResult())
	////
	////time.Sleep(5 * time.Second)
	//req4 := sendMessage(list)
	//r4, err := c.Connect(ctx, req4)
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//log.Printf("返回消息：%s", r4.GetActionResult())

	// 发送图片信息
	//time.Sleep(5 * time.Second)
	//req5 := sendImage(photoList)
	//r5, err := c.Connect(ctx, req5)
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//log.Printf("返回消息：%s", r5.GetActionResult())
	//time.Sleep(4 * time.Second)

	//time.Sleep(5 * time.Second)
	//data1 := &telegram.CreateGroupDetail{
	//	Initiator:    19566961252,
	//	GroupTitle:   "ksdjfkj",
	//	GroupMembers: []uint64{8618818877128},
	//}
	//req6 := createGroup(data1)
	//r6, err := c.Connect(ctx, req6)
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//log.Printf("返回消息：%s", r6.GetActionResult())
	//time.Sleep(5 * time.Second)
	//data2 := &telegram.AddGroupMemberDetail{
	//	Initiator:  19566961252,
	//	GroupTitle: "ksdjfkj",
	//	AddMembers: []uint64{8618028658256},
	//}
	//req7 := addGroupMember(data2)
	//r7, err := c.Connect(ctx, req7)
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//log.Printf("返回消息：%s", r7.GetActionResult())

	// 发送文件
	//time.Sleep(5 * time.Second)
	//req7 := sendFile(FileList)
	//r7, err := c.Connect(ctx, req7)
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//log.Printf("返回消息：%s", r7.GetActionResult())
	//
	//// 发送名片
	//time.Sleep(5 * time.Second)
	//req8 := contactCard(cardList)
	//r8, err := c.Connect(ctx, req8)
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//log.Printf("返回消息：%s", r8.GetActionResult())

	//time.Sleep(5 * time.Second)
	//listgroup := make([]*telegram.SendGroupMessageDetail, 0)
	//sendTextgroup1 := &telegram.SendGroupMessageDetail{
	//	Sender:     accountNum,
	//	GroupTitle: "ksdjfkj",
	//	SendText:   []string{"1111111"},
	//}
	//listgroup = append(listgroup, sendTextgroup1)
	//req9 := sendGroupMessage(listgroup)
	//r9, err := c.Connect(ctx, req9)
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//log.Printf("返回消息：%s", r9.GetActionResult())

}

func login(accounts []AccountDetail) *protobuf.RequestMessage {
	loginDetail := make(map[uint64]*protobuf.LoginDetail)

	for _, account := range accounts {
		ld := &protobuf.LoginDetail{
			ProxyUrl: "",
			TgDevice: &protobuf.TgDeviceConfig{
				DeviceModel:    account.Device.DeviceModel,
				SystemVersion:  account.Device.SystemVersion,
				AppVersion:     account.Device.AppVersion,
				LangCode:       account.Device.LangCode,
				SystemLangCode: account.Device.SystemLangCode,
				LangPack:       account.Device.LangPack,
			},
		}
		user, _ := strconv.ParseUint(account.PhoneNumber, 10, 64)
		loginDetail[user] = ld
	}
	req := &protobuf.RequestMessage{
		Action: protobuf.Action_LOGIN,
		Type:   "telegram",
		ActionDetail: &protobuf.RequestMessage_OrdinaryAction{
			OrdinaryAction: &protobuf.OrdinaryAction{
				LoginDetail: loginDetail,
			},
		},
	}
	return req
}

func syncAppInfo(accounts []AccountDetail) *protobuf.RequestMessage {
	appData := make(map[uint64]*protobuf.AppData)
	for _, account := range accounts {
		user, _ := strconv.ParseUint(account.PhoneNumber, 10, 64)
		appData[user] = &protobuf.AppData{AppId: account.AppId, AppHash: account.AppHash}
	}
	req := &protobuf.RequestMessage{
		Action: protobuf.Action_SYNC_APP_INFO,
		Type:   "telegram",
		ActionDetail: &protobuf.RequestMessage_SyncAppAction{
			SyncAppAction: &protobuf.SyncAppInfoAction{
				AppData: appData,
			},
		},
	}

	return req
}

func sendMessage(accountNum uint64, contant []*telegram.SendMessageDetail) *protobuf.RequestMessage {
	list := make([]*protobuf.SendMessageAction, 0)

	for _, detail := range contant {
		tmp := &protobuf.SendMessageAction{}
		sendData := make(map[uint64]*protobuf.StringKeyStringvalue)
		sendData[detail.Sender] = &protobuf.StringKeyStringvalue{Key: detail.Receiver, Values: detail.SendText}
		tmp.SendTgData = sendData
		list = append(list, tmp)
	}
	req := &protobuf.RequestMessage{
		Action:  protobuf.Action_SEND_MESSAGE,
		Type:    "telegram",
		Account: accountNum,
		ActionDetail: &protobuf.RequestMessage_SendmessageDetail{
			SendmessageDetail: &protobuf.SendMessageDetail{
				Details: list,
			},
		},
	}

	return req
}

func createGroup(data *telegram.CreateGroupDetail) *protobuf.RequestMessage {

	detail := &protobuf.UintkeyStringvalue{}
	detail.Key = data.Initiator
	for _, v := range data.GroupMembers {
		detail.Values = append(detail.Values, v)
	}
	req := &protobuf.RequestMessage{
		Action: protobuf.Action_CREATE_GROUP,
		Type:   "telegram",
		ActionDetail: &protobuf.RequestMessage_CreateGroupDetail{
			CreateGroupDetail: &protobuf.CreateGroupDetail{
				GroupName: data.GroupTitle,
				Detail:    detail,
			},
		},
	}
	return req
}

func sendGroupMessage(data []*telegram.SendGroupMessageDetail) *protobuf.RequestMessage {
	list := make([]*protobuf.SendGroupMessageAction, 0)

	for _, detail := range data {
		tmp := &protobuf.SendGroupMessageAction{}
		sendData := make(map[uint64]*protobuf.StringKeyStringvalue)
		sendData[detail.Sender] = &protobuf.StringKeyStringvalue{Key: detail.GroupTitle, Values: detail.SendText}
		tmp.SendData = sendData
		list = append(list, tmp)
	}
	req := &protobuf.RequestMessage{
		Action: protobuf.Action_SEND_GROUP_MESSAGE,
		Type:   "telegram",
		ActionDetail: &protobuf.RequestMessage_SendGroupMessageDetail{
			SendGroupMessageDetail: &protobuf.SendGroupMessageDetail{
				Details: list,
			},
		},
	}

	return req
}
