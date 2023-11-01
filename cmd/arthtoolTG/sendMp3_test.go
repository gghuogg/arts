package main

import (
	"arthas/protobuf"
	"arthas/telegram"
	"context"
	"fmt"
	"github.com/iyear/tdl/pkg/consts"
	"google.golang.org/grpc"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestSendMp3(t *testing.T) {

	accountNum := uint64(15022186268)
	account := AccountDetail{
		PhoneNumber: strconv.FormatUint(accountNum, 10),
		AppId:       24242552,
		AppHash:     "12adb15f288feeb37101c331fe2789ad",
		Device:      consts.Device,
	}
	fmt.Println(account)

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("无法获取当前目录：", err)
		return
	}

	// 构建文件路径
	filePath := filepath.Join(currentDir, "testFile", "beibei.mp3")

	b3, err := os.ReadFile(filePath)

	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println(len(b3))

	//音频
	mp3List := make([]*telegram.SendFileDetail, 0)
	mp3Test1 := &telegram.SendFileDetail{
		Sender:   accountNum,
		Receiver: strconv.FormatUint(8618818877128, 10),
		FileType: []telegram.SendFileType{
			{
				FileType:  telegram.TG_FILE_TYPE_MP3,
				SendType:  telegram.TG_SEND_TYPE_BYTE,
				Path:      "cmd/arthtoolTG/testFile/Twins.mp3",
				Name:      "beibei.mp3",
				FileBytes: b3,
			},
		},
	}
	mp3List = append(mp3List, mp3Test1)

	conn, err := grpc.Dial("127.0.0.1:50052", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}
	defer conn.Close()

	c := protobuf.NewArthasClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

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

	//time.Sleep(5 * time.Second)
	//音频
	req7 := sendMp3File(accountNum, mp3List)
	r7, err := c.Connect(ctx, req7)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息：%s", r7.GetActionResult())
}

func sendMp3File(accountNum uint64, files []*telegram.SendFileDetail) *protobuf.RequestMessage {
	list := make([]*protobuf.SendFileAction, 0)

	for _, detail := range files {
		tmp := &protobuf.SendFileAction{}
		sendData := make(map[uint64]*protobuf.UintTgFileDetailValue)
		sendData[detail.Sender] = &protobuf.UintTgFileDetailValue{Key: detail.Receiver}
		fileDetail := make([]*protobuf.FileDetailValue, 0)
		for _, file := range detail.FileType {
			fileDetail = append(fileDetail, &protobuf.FileDetailValue{
				FileType: file.FileType,
				SendType: file.SendType,
				Path:     file.Path,
				Name:     file.Name,
				FileByte: file.FileBytes,
			})
		}
		sendData[detail.Sender].Value = fileDetail
		tmp.SendTgData = sendData
		list = append(list, tmp)
	}

	req := &protobuf.RequestMessage{
		Action:  protobuf.Action_SEND_FILE,
		Type:    "telegram",
		Account: accountNum,
		ActionDetail: &protobuf.RequestMessage_SendFileDetail{
			SendFileDetail: &protobuf.SendFileDetail{
				Details: list,
			},
		},
	}

	return req

}
