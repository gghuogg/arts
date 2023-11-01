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

func TestSendFile(t *testing.T) {
	// 14013986058被封手机号
	accountNum := uint64(19566963352)
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
	filePath := filepath.Join(currentDir, "testFile", "picture.jpg")

	b, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	photoList := make([]*telegram.SendImageFileDetail, 0)
	SendImageFileTest1 := &telegram.SendImageFileDetail{
		Sender:   accountNum,
		Receiver: strconv.FormatUint(8618818877128, 10),
		FileType: []telegram.SendImageFileType{
			{
				FileType:   telegram.TG_FILE_TYPE_IMAGE,
				SendType:   telegram.TG_SEND_TYPE_BYTE,
				Path:       "tgcloud/attachment/2023-10-08/cw2yu0gxl0rs23viro.png",
				Name:       "picture.png",
				ImageBytes: b,
			},
			{
				FileType:   telegram.TG_FILE_TYPE_IMAGE,
				SendType:   telegram.TG_SEND_TYPE_BYTE,
				Path:       "tgcloud/attachment/2023-10-08/cw2yu0gxl0rs23viro.png",
				Name:       "picture.png",
				ImageBytes: b,
			},
		},
	}
	photoList = append(photoList, SendImageFileTest1)
	filePath2 := filepath.Join(currentDir, "testFile", "picture.jpg")

	b2, err := os.ReadFile(filePath2)
	if err != nil {
		fmt.Println(err)
		return
	}

	FileList := make([]*telegram.SendFileDetail, 0)
	FileTest1 := &telegram.SendFileDetail{
		Sender:   accountNum,
		Receiver: strconv.FormatUint(8618818877128, 10),
		FileType: []telegram.SendFileType{
			{
				FileType:  telegram.TG_FILE_TYPE_File,
				SendType:  telegram.TG_SEND_TYPE_BYTE,
				Path:      "tgcloud/attachment/2023-10-08/cw2yu0gxl0rs23viro.png",
				Name:      "cvu5vsha49eopdumvs.jpg",
				FileBytes: b2,
			},
		},
	}
	FileList = append(FileList, FileTest1)

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
	//time.Sleep(3 * time.Second)
	//req2 := login(account)
	//r2, err := c.Connect(ctx, req2)
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//log.Printf("返回消息: %s", r2.GetActionResult())

	// 发送图片
	time.Sleep(5 * time.Second)
	req5 := sendImage(accountNum, photoList)
	r5, err := c.Connect(ctx, req5)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息：%s", r5.GetActionResult())

	//发送文件
	//time.Sleep(5 * time.Second)
	//req7 := sendFile(accountNum, FileList)
	//r7, err := c.Connect(ctx, req7)
	//if err != nil {
	//	log.Fatalf("请求失败: %v", err)
	//}
	//log.Printf("返回消息：%s", r7.GetActionResult())

}

func sendImage(accountNum uint64, photos []*telegram.SendImageFileDetail) *protobuf.RequestMessage {
	list := make([]*protobuf.SendFileAction, 0)

	for _, detail := range photos {
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
				FileByte: file.ImageBytes,
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

func sendFile(accountNum uint64, files []*telegram.SendFileDetail) *protobuf.RequestMessage {
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
