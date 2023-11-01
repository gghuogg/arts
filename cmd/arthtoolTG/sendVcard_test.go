package main

import (
	"arthas/protobuf"
	"arthas/telegram"
	"fmt"
	"github.com/iyear/tdl/pkg/consts"
	"google.golang.org/grpc"
	"log"

	"context"
	"strconv"
	"testing"
	"time"
)

func TestSendVcard(t *testing.T) {

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

	cardList := make([]*telegram.ContactCardDetail, 0)
	myCntactCard1 := &telegram.ContactCardDetail{
		Sender:   accountNum,
		Receiver: "8618818877128",
		ContactCards: []telegram.ContactCard{
			{
				FirstName:   "lu",
				LastName:    "xuedong",
				PhoneNumber: "8618818877128",
			},
			{
				FirstName:   "lu",
				LastName:    "xuedong",
				PhoneNumber: "8618818877128",
			},
		},
	}

	cardList = append(cardList, myCntactCard1)

	myCntactCard2 := &telegram.ContactCardDetail{
		Sender:   accountNum,
		Receiver: "8618818877128",
		ContactCards: []telegram.ContactCard{
			{
				FirstName:   "k",
				LastName:    "kz",
				PhoneNumber: "15712294113",
			},
		},
	}

	cardList = append(cardList, myCntactCard2)

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

	//// 发送名片
	time.Sleep(5 * time.Second)
	req8 := contactCard(accountNum, cardList)
	r8, err := c.Connect(ctx, req8)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息：%s", r8.GetActionResult())
}

func contactCard(accountNum uint64, cards []*telegram.ContactCardDetail) *protobuf.RequestMessage {
	list := make([]*protobuf.SendContactCardAction, 0)

	for _, detail := range cards {
		tmp := &protobuf.SendContactCardAction{}
		sendData := make(map[uint64]*protobuf.UintSendContactCard)
		sendData[detail.Sender] = &protobuf.UintSendContactCard{
			Sender:   detail.Sender,
			Receiver: detail.Receiver,
		}
		cardList := make([]*protobuf.ContactCardValue, 0)
		for _, card := range detail.ContactCards {
			cardList = append(cardList, &protobuf.ContactCardValue{
				FirstName:   card.FirstName,
				LastName:    card.LastName,
				PhoneNumber: card.PhoneNumber,
			})
		}
		sendData[detail.Sender].Value = cardList
		tmp.SendData = sendData
		list = append(list, tmp)
	}
	req := &protobuf.RequestMessage{
		Action:  protobuf.Action_SEND_VCARD_MESSAGE,
		Type:    "telegram",
		Account: accountNum,
		ActionDetail: &protobuf.RequestMessage_SendContactCardDetail{
			SendContactCardDetail: &protobuf.SendContactCardDetail{
				Detail: list,
			},
		},
	}
	return req
}
