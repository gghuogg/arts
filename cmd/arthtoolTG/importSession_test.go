package main

import (
	"arthas/protobuf"
	"arthas/telegram"
	"context"
	"github.com/iyear/tdl/pkg/consts"
	"strconv"

	"fmt"
	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"google.golang.org/grpc"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestImportSession(t *testing.T) {
	accountNum := uint64(8618818877128)
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

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Println("无法获取当前目录：", err)
		return
	}

	// 构建文件路径
	filePath := filepath.Join(currentDir, "testFile", "8618818877128.session")

	c := protobuf.NewArthasClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	session := sqliteSession(filePath, account)
	session.Account = accountNum
	sessionList := make([]*telegram.ImportTgSessionDetail, 0)
	sessionList = append(sessionList, session)

	req7 := importSession(accountNum, sessionList)
	r7, err := c.Connect(ctx, req7)
	if err != nil {
		log.Fatalf("请求失败: %v", err)
	}
	log.Printf("返回消息：%s", r7.GetActionResult())
}

func sqliteSession(path string, account AccountDetail) *telegram.ImportTgSessionDetail {
	tgSession := &telegram.ImportTgSessionDetail{}
	gdb.SetConfig(gdb.Config{
		"default": gdb.ConfigGroup{
			gdb.ConfigNode{
				Type:    "sqlite",
				Link:    fmt.Sprintf(`sqlite::@file(%s)`, path),
				Charset: "utf8",
			},
		},
	})
	ctx := gctx.New()
	all, err := g.DB().Ctx(ctx).GetAll(ctx, "select * from sessions")
	if err != nil {
		fmt.Println(err)
	}
	for _, item := range all.List() {
		auth_key := item["auth_key"]
		authKey, ok := auth_key.([]byte)
		if ok {
			tgSession.AuthKey = authKey
		}
		dc_id := item["dc_id"]
		dcId, ok := dc_id.(int64)
		if ok {
			tgSession.DC = int32(dcId)
		}

		server_address := item["server_address"]
		addr, ok := server_address.(string)
		if ok {
			tgSession.Addr = addr
		}

	}
	tgSession.AccountDeviceMsg = telegram.AccountDevice{
		AppId:          account.AppId,
		AppHash:        account.AppHash,
		DeviceModel:    account.Device.DeviceModel,
		SystemVersion:  account.Device.SystemVersion,
		AppVersion:     account.Device.AppVersion,
		LangCode:       account.Device.LangCode,
		SystemLangCode: account.Device.SystemLangCode,
		LangPack:       account.Device.LangPack,
	}

	return tgSession
}

func importSession(account uint64, data []*telegram.ImportTgSessionDetail) *protobuf.RequestMessage {
	sessionMap := make(map[uint64]*protobuf.ImportTgSessionMsg)

	for _, s := range data {
		sessionMap[s.Account] = &protobuf.ImportTgSessionMsg{
			DC:      s.DC,
			Addr:    s.Addr,
			AuthKey: s.AuthKey,
			DeviceMsg: &protobuf.ImportTgDeviceMsg{
				AppId:   s.AccountDeviceMsg.AppId,
				AppHash: s.AccountDeviceMsg.AppHash,

				DeviceModel:    s.AccountDeviceMsg.DeviceModel,
				AppVersion:     s.AccountDeviceMsg.AppVersion,
				SystemVersion:  s.AccountDeviceMsg.SystemVersion,
				LangCode:       s.AccountDeviceMsg.LangCode,
				LangPack:       s.AccountDeviceMsg.LangPack,
				SystemLangCode: s.AccountDeviceMsg.SystemLangCode,
			},
		}
	}

	req := &protobuf.RequestMessage{
		Action: protobuf.Action_IMPORT_TG_SESSION,
		Type:   "telegram",
		ActionDetail: &protobuf.RequestMessage_ImportTgSession{
			ImportTgSession: &protobuf.ImportTgSessionDetail{
				Account:  account,
				SendData: sessionMap,
			},
		},
	}
	return req
}
