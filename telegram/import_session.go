package telegram

import (
	"arthas/etcd"
	"arthas/protobuf"
	"context"
	"crypto/sha1"
	"encoding/json"
	"github.com/go-kit/log/level"
	"github.com/gotd/td/session"
	"strconv"
)

type ImportTgSessionMsg struct {
	Sender  uint64
	Details []ImportTgSessionDetail
	ResChan chan ImportSessionRes
}

type ImportSessionRes struct {
	ErrorAccounts []uint64
	ResMsg        *protobuf.ResponseMessage
	Comment       string
}
type ImportTgSessionDetail struct {
	Account          uint64
	DC               int32
	Addr             string
	AuthKey          []byte
	AuthKeyId        []byte
	AccountDeviceMsg AccountDevice
}

type AccountDevice struct {
	AppId   uint64
	AppHash string

	DeviceModel    string
	SystemVersion  string
	AppVersion     string
	LangCode       string
	SystemLangCode string
	LangPack       string
}

type Key [256]byte

type AuthKey struct {
	Value Key
	ID    [8]byte
}

type AccountSession struct {
	Version int
	Data    session.Data
}

func (m *Manager) ImportTgSessionToEtcd(ctx context.Context, details ImportTgSessionMsg) error {
	errorAccounts := make([]uint64, 0)
	for _, d := range details.Details {
		accountSession := AccountSession{}
		accountSession.Version = 1

		se := session.Data{}
		se.DC = int(d.DC)
		se.Addr = d.Addr
		se.AuthKey = d.AuthKey

		var key AuthKey
		copy(key.Value[:], se.AuthKey)

		se.AuthKeyID = keyGetAutoId(key.Value)
		accountSession.Data = se

		b, err := json.Marshal(accountSession)
		if err != nil {
			level.Error(m.logger).Log("msg", "Failed to convert session to JSON.", err.Error())
			errorAccounts = append(errorAccounts, d.Account)
			continue
		}

		// 导入session
		putSessionChan := m.putChan
		putSessionReq := etcd.PutReq{
			Key:     "/tg/" + strconv.FormatUint(d.Account, 10) + "/session/json",
			Value:   string(b),
			Options: nil,
			ResChan: make(chan etcd.PutRes)}
		putSessionChan <- putSessionReq

		// 导入app，证明是否登录
		loginFlag := "builtin"
		b2 := []byte(loginFlag)
		putAppChan := m.putChan
		putAppReq := etcd.PutReq{
			Key:     "/tg/" + strconv.FormatUint(d.Account, 10) + "/app/json",
			Value:   string(b2),
			Options: nil,
			ResChan: make(chan etcd.PutRes)}
		putAppChan <- putAppReq

		// 将appId,appHash,Device存入到etcd
		b3, err := json.Marshal(d.AccountDeviceMsg)
		if err != nil {
			level.Error(m.logger).Log("msg", "Failed to convert session to JSON.", err.Error())
			errorAccounts = append(errorAccounts, d.Account)
			continue
		}
		putDeviceChan := m.putChan
		putDeviceReq := etcd.PutReq{
			Key:     "/tg/" + strconv.FormatUint(d.Account, 10) + "/device/json",
			Value:   string(b3),
			Options: nil,
			ResChan: make(chan etcd.PutRes)}
		putDeviceChan <- putDeviceReq

	}
	msgRes := ImportSessionRes{ResMsg: &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}}
	details.ResChan <- msgRes
	return nil
}

func keyGetAutoId(k [256]byte) []byte {
	raw := sha1.Sum(k[:])
	// #nosec
	var id [8]byte
	copy(id[:], raw[12:])
	return id[:]
}
