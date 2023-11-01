package etcd

import (
	alog "arthas/log"
	"arthas/protobuf"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-kit/log"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/session"
	"github.com/gotd/td/tg"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/protobuf/proto"
	"testing"
)

var manager *Manager
var ctx, cancelFunc = context.WithCancel(context.TODO())

type AccountSession struct {
	Version int
	Data    session.Data
}

type Key struct {
	Prefix string
	ID     int64
}

type Hash struct {
	AccessHash int64
}

func initEtcd() {
	logger := alog.New(&alog.Config{})
	manager = New(log.With(logger, "component", "etcd"), ctx, &Options{
		EndPoint: []string{"10.8.5.21:2379"},
		Url:      "",
		Port:     "",
		Username: "",
		Password: "",
		Delay:    10,
	})
	err := manager.ApplyConfig()
	panicErr(err)
	go func() {
		err = manager.Run()
		panicErr(err)
	}()
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func TestGetAndPut(t *testing.T) {
	myJson := "/tg/15022186268/session/json"
	initEtcd()
	getChan := manager.GetChan()
	getReq := GetReq{Key: myJson, ResChan: make(chan GetRes)}
	getChan <- getReq
	res := <-getReq.ResChan
	panicErr(res.Err)
	as := &AccountSession{}
	json.Unmarshal(res.Result.Kvs[0].Value, as)

	// 创建一个256字节的切片
	randomBytes := make([]byte, 256)

	// 生成随机字节
	_, err := rand.Read(randomBytes)
	if err != nil {
		fmt.Println("生成随机字节时出错:", err)
		return
	}

	as.Data.AuthKey = randomBytes
	bb, _ := json.Marshal(as)
	putChan := manager.PutChan()
	putReq := PutReq{
		Key:     "/tg/15022186268/session/json",
		Value:   string(bb),
		Options: nil,
		ResChan: make(chan PutRes)}
	putChan <- putReq
	panicErr(res.Err)
	fmt.Println(res.Result)

}

func TestPhone(t *testing.T) {
	myJson := "/tg/16893992489/peers:phone:8618818877128"
	initEtcd()
	getChan := manager.GetChan()
	getReq := GetReq{Key: myJson, ResChan: make(chan GetRes)}
	getChan <- getReq
	res := <-getReq.ResChan
	panicErr(res.Err)
	key := Key{}
	json.Unmarshal(res.Result.Kvs[0].Value, &key)

	var phoneKey = &protobuf.PhoneStoragekey{
		Prefix: key.Prefix,
		ID:     key.ID,
	}

	bytes, err := proto.Marshal(phoneKey)
	panicErr(err)
	putChan := manager.PutChan()
	putReq := PutReq{
		Key:     myJson,
		Value:   string(bytes),
		Options: nil,
		ResChan: make(chan PutRes)}
	putChan <- putReq
	panicErr(res.Err)
	fmt.Println(res.Result)

}

func TestHash(t *testing.T) {
	//myJson := "/tg/15712294113/peers:key:users_:5824006024"
	initEtcd()
	getChan := manager.GetChan()
	getReq := GetReq{Key: "/tg/16893992489/peers:key:users_:5824006024", ResChan: make(chan GetRes)}
	getChan <- getReq
	res := <-getReq.ResChan
	panicErr(res.Err)
	key := Hash{}
	json.Unmarshal(res.Result.Kvs[0].Value, &key)

	var phoneKey = &protobuf.AccessHashValue{
		AccessHash: key.AccessHash,
	}

	bytes, err := proto.Marshal(phoneKey)
	panicErr(err)
	putChan := manager.PutChan()
	putReq := PutReq{
		Key:     "/tg/16893992489/peers:key:users_:5824006024",
		Value:   string(bytes),
		Options: nil,
		ResChan: make(chan PutRes)}
	putChan <- putReq
	panicErr(res.Err)
	fmt.Println(res.Result)

}

// 不传proto只传json
func TestGet1(t *testing.T) {
	json := "/tg/16893992489/session/json"
	// 获取
	initEtcd()
	getChan := manager.GetChan()
	getReq := GetReq{Key: json, Options: []clientv3.OpOption{clientv3.WithPrefix()}, ResChan: make(chan GetRes)}
	getChan <- getReq
	res := <-getReq.ResChan
	panicErr(res.Err)

	// 提交
	putChan := manager.PutChan()
	putReq := PutReq{
		Key:     "/tg/16893992489/session",
		Value:   string(res.Result.Kvs[0].Value),
		Options: nil,
		ResChan: make(chan PutRes)}
	putChan <- putReq
	panicErr(res.Err)
	fmt.Println(res.Result)
}

// json转proto   session
func TestGet(t *testing.T) {
	initEtcd()
	getChan := manager.GetChan()
	getReq := GetReq{Key: "/tg/15022186268/session/json", Options: []clientv3.OpOption{clientv3.WithPrefix()}, ResChan: make(chan GetRes)}
	getChan <- getReq
	res := <-getReq.ResChan
	se := &AccountSession{}
	json.Unmarshal(res.Result.Kvs[0].Value, se)
	se.Data.Config = session.Config{}

	//jsonB, err := json.Marshal(se)
	//panicErr(err)
	//putChan := manager.PutChan()
	//putReq := PutReq{
	//	Key:     "/tg/15022186268/session/json",
	//	Value:   string(jsonB),
	//	Options: nil,
	//	ResChan: make(chan PutRes)}
	//putChan <- putReq
	//panicErr(res.Err)
	//fmt.Println(res.Result)

}

func getSessionProto(session *AccountSession) *protobuf.TgAccountSession {
	entryData := session.Data
	entryConfig := entryData.Config
	tgDcos := entryConfig.DCOptions

	dcos := []*protobuf.DCOption{}
	for _, tgdco := range tgDcos {

		dcos = append(dcos, &protobuf.DCOption{
			Flags:             uint32(tgdco.Flags),
			Ipv6:              tgdco.Ipv6,
			MediaOnly:         tgdco.MediaOnly,
			TCPObfuscatedOnly: tgdco.TCPObfuscatedOnly,
			CDN:               tgdco.CDN,
			Static:            tgdco.Static,
			ThisPortOnly:      tgdco.ThisPortOnly,
			ID:                int32(tgdco.ID),
			IPAddress:         tgdco.IPAddress,
			Port:              int32(tgdco.Port),
			Secret:            tgdco.Secret,
		})
	}
	//authKey, err := utf8Decoder.String(string(entryData.AuthKey))
	//authKeyID, err := utf8Decoder.String(string(entryData.AuthKeyID))

	authKey := encodeBytesToString(entryData.AuthKey)
	authKeyID := encodeBytesToString(entryData.AuthKeyID)

	s := &protobuf.TgAccountSession{
		Version: int32(session.Version),
		Data: &protobuf.Data{
			Addr:      entryData.Addr,
			DC:        int32(entryData.DC),
			AuthKey:   string(authKey),
			AuthKeyID: string(authKeyID),
			Salt:      entryData.Salt,
			Config: &protobuf.Config{
				BlockedMode:     entryConfig.BlockedMode,
				ForceTryIpv6:    entryConfig.ForceTryIpv6,
				Date:            int32(entryConfig.Date),
				Expires:         int32(entryConfig.Expires),
				TestMode:        entryConfig.TestMode,
				ThisDC:          int32(entryConfig.ThisDC),
				DCTxtDomainName: entryConfig.DCTxtDomainName,
				TmpSessions:     int32(entryConfig.TmpSessions),
				WebfileDCID:     int32(entryConfig.WebfileDCID),
				DCOptions:       dcos,
			},
		},
	}
	return s
}

func getDataByProto(t *protobuf.TgAccountSession) *session.Data {
	protoData := t.Data
	protoConfig := protoData.Config
	dcos := protoConfig.DCOptions
	dcoptions := []tg.DCOption{}
	for _, dco := range dcos {
		dcoptions = append(dcoptions, tg.DCOption{
			Flags:             bin.Fields(dco.Flags),
			Ipv6:              dco.Ipv6,
			MediaOnly:         dco.MediaOnly,
			TCPObfuscatedOnly: dco.TCPObfuscatedOnly,
			CDN:               dco.CDN,
			Static:            dco.Static,
			ThisPortOnly:      dco.ThisPortOnly,
			ID:                int(dco.ID),
			IPAddress:         dco.IPAddress,
			Port:              int(dco.Port),
			Secret:            dco.Secret,
		})
	}
	authKey, err := decodeStringToBytes(protoData.AuthKey)
	authKeyId, err := decodeStringToBytes(protoData.AuthKeyID)
	fmt.Println(err)
	data := &session.Data{
		Config: session.Config{
			BlockedMode:     protoConfig.BlockedMode,
			ForceTryIpv6:    protoConfig.ForceTryIpv6,
			Date:            int(protoConfig.Date),
			Expires:         int(protoConfig.Expires),
			TestMode:        protoConfig.TestMode,
			ThisDC:          int(protoConfig.ThisDC),
			DCOptions:       dcoptions,
			DCTxtDomainName: protoConfig.DCTxtDomainName,
			TmpSessions:     int(protoConfig.TmpSessions),
			WebfileDCID:     int(protoConfig.WebfileDCID),
		},
		DC:        int(protoData.DC),
		Addr:      protoData.Addr,
		AuthKey:   authKey,
		AuthKeyID: authKeyId,
		Salt:      protoData.Salt,
	}
	return data

}

func encodeBytesToString(bytes []byte) string {
	encodedString := base64.StdEncoding.EncodeToString(bytes)
	return encodedString
}

func decodeStringToBytes(str string) ([]byte, error) {
	decodedBytes, err := base64.StdEncoding.DecodeString(str)
	if err != nil {
		return nil, err
	}
	return decodedBytes, nil
}

func TestGet2(t *testing.T) {
	initEtcd()
	getChan := manager.GetChan()
	getReq := GetReq{Key: "/tg/19566963352", Options: []clientv3.OpOption{clientv3.WithPrefix()}, ResChan: make(chan GetRes)}
	getChan <- getReq
	res := <-getReq.ResChan
	fmt.Println(res.Result.Count)
}
