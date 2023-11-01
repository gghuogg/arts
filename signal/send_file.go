package signal

import (
	"arthas/etcd"
	"arthas/protobuf"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"go.mau.fi/libsignal/ecc"
	"go.mau.fi/libsignal/keys/identity"
	"go.mau.fi/libsignal/keys/prekey"
	"go.mau.fi/libsignal/protocol"
	"go.mau.fi/libsignal/session"
	"go.mau.fi/libsignal/state/record"
	"go.mau.fi/libsignal/util/bytehelper"
	"go.mau.fi/libsignal/util/optional"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"
)

type FileKeyInfo struct {
	MediaKey      []byte
	FileSha256    []byte
	FileEncSha256 []byte
	FileLength    uint64
	DirectPath    string
	Url           string
	FileType      string
	FileName      string
}

type SendFileDetail struct {
	Sender   uint64
	Receiver uint64
	FileType []SendFileType
}

type SendFileType struct {
	FileType  string
	SendType  string
	Path      string
	Name      string
	FileBytes []byte
}

type FileUrlInfo struct {
	Url         string `json:"url"`
	DirectpPath string `json:"direct_path"`
}

const (
	FILE_MAX_SIZE = 2 * 1024 * 1024 * 1024
)

const (
	GET_BY_URL   = "URL"
	GET_BY_BYTES = "BYTES"
)

func (a *Account) sendFile(detail SendFileDetail) {
	fileKeyInfoList := make([]*FileKeyInfo, 0)
	for _, f := range detail.FileType {
		mediaKey := make([]byte, 32)
		rand.Read(mediaKey)
		if f.SendType == GET_BY_URL {
			if f.Path == "" {
				level.Error(a.logger).Log("msg", "file path if nil err")
				return
			}
			resp, err := g.Client().Get(gctx.New(), f.Path)
			if err != nil {
				level.Error(a.logger).Log("msg", "File acquisition fail", "err", err)
			}
			defer resp.Close()
			f.FileBytes = resp.ReadAll()

			if &f.Name == nil || f.Name == "" {
				f.Name = filepath.Base(f.Path)
			}
		}
		if len(f.FileBytes) > FILE_MAX_SIZE {
			level.Error(a.logger).Log("msg", "file size over 2G")
			return
		}

		b := f.FileBytes

		encryptedImageData, fileSha256, fileEncSha256, err := EncryptMediaBytes(b, mediaKey, []byte(labelDocumentKeys))
		if err != nil {
			level.Error(a.logger).Log("msg", "file encrypt is err", "err", err)
			return
		}

		//imageData, err := DecryptMediaBytes(encryptedImageData, mediaKey, fileSha256, fileEncSha256, []byte(labelImageKeys))
		//if err != nil {
		//}
		//
		//if bytes.Compare(imageData, b) != 0 {
		//	panic("encryption error")
		//}

		hostname := "media-sin6-1.cdn.whatsapp.net"
		filename := base64.URLEncoding.EncodeToString(fileEncSha256) // use url-encoding instead of standard-encoding

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s/optimistic/document/%s", hostname, filename), bytes.NewReader(encryptedImageData))

		if err != nil {
			return
		}

		authKey := "Ac152oRJI9vqObqPofYQPkSEayuuzt5C4rMniLZU0VSGUw"
		mediaId, _ := strconv.Atoi("46569494")

		q := req.URL.Query()
		q.Add("direct_ip", "0")
		q.Add("auth", authKey)
		q.Add("token", filename)
		q.Add("media_id", strconv.Itoa(mediaId))
		req.URL.RawQuery = q.Encode()

		proxyURL, _ := url.Parse(a.proxyUrl)

		httpClient := &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			},
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			level.Error(a.logger).Log("msg", "http client do  is err", "err", err)
			return
		}

		respBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		currentTime := time.Now().Unix()
		timestamp := currentTime / 1000

		fileUrlInfo := &FileUrlInfo{}
		// 获取转成得url 和 direct_path
		err = json.Unmarshal(respBytes, fileUrlInfo)
		if err != nil {
			println(err)
		}
		msg := a.signal.StatFileRequest(respBytes, strconv.FormatInt(timestamp, 10))
		fileInfo := &FileKeyInfo{
			MediaKey:      mediaKey,
			FileSha256:    fileSha256,
			FileEncSha256: fileEncSha256,
			FileLength:    uint64(len(b)),
			Url:           fileUrlInfo.Url,
			DirectPath:    fileUrlInfo.DirectpPath,
			FileType:      f.FileType,
			FileName:      f.Name,
		}
		fileKeyInfoList = append(fileKeyInfoList, fileInfo)
		a.send(msg)
	}

	a.SendFileMessage(detail, fileKeyInfoList)

}

func (a *Account) SendFileMessage(detail SendFileDetail, fileInfolist []*FileKeyInfo) {
	//a.signal.mutex.Lock()
	//
	//defer a.signal.mutex.Unlock()
	contact := detail.Receiver
	accountPhone := uint64FromPtr(a.ClientPayLoad.Username)
	recipientJid := fmt.Sprintf("%d@s.whatsapp.net", contact)
	if a.RemoteAddresses[contact] == nil {
		a.RemoteAddresses[contact] = protocol.NewSignalAddress(recipientJid, 0)
	}
	if a.SessionStore[contact] == nil {
		a.SessionStore[contact] = session.NewBuilderFromSignal(a.SignalprotocolStore, a.RemoteAddresses[contact], a.Serializer)
	}

	current := make([]byte, 0)

	exist := a.SignalprotocolStore.SessionStore.ContainsSession(a.RemoteAddresses[contact])
	if exist == true {
		current = a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[contact]).Serialize()
	} else {

		getReq := etcd.GetReq{
			Key:     "/arthas/Session/" + strconv.FormatUint(accountPhone, 10) + "/" + strconv.FormatUint(contact, 10),
			Options: nil,
			ResChan: make(chan etcd.GetRes),
		}
		a.signal.getChan <- getReq
		res := <-getReq.ResChan

		if res.Result.Kvs == nil {
			fmt.Println("初始化")
			//初始化
			requestId := fmt.Sprintf("%s-%d", a.timestamp, a.id)
			a.id++
			a.getPrekeyBundleWithPhoneNumberJid(requestId, recipientJid)

			a.PrekeyChan[contact] = make(chan struct{})
			select {
			case <-a.PrekeyChan[contact]:
				fmt.Println("接收prekey")

				preKeyID := optional.NewOptionalUint32(a.PrekeyBundles[contact].OneTimePreKeyId)
				preKeyPublic := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(a.PrekeyBundles[contact].OneTimePreKey))

				signedPreKeyPublic := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(a.PrekeyBundles[contact].SignedPreKey))
				identitykeyPublic := ecc.NewDjbECPublicKey(bytehelper.SliceToArray(a.PrekeyBundles[contact].IdentityKey))
				keys := identity.NewKey(identitykeyPublic)

				bundle := prekey.NewBundle(a.PrekeyBundles[contact].RegistrationId, 0, preKeyID,
					a.PrekeyBundles[contact].SignedPreKeyId, preKeyPublic, signedPreKeyPublic, *a.PrekeyBundles[contact].SignedPreKeySignature, keys)

				a.SessionStore[contact].ProcessBundle(bundle)

			case <-time.After(time.Second * 90):
				level.Info(a.logger).Log("msg", "Time up.exit", "user", a.ClientPayLoad.Username, "contact", contact)
				return
			}

			loadedSession := a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[contact])

			insert := etcd.TxnInsertReq{
				Key:     "/arthas/Session/" + strconv.FormatUint(accountPhone, 10) + "/" + strconv.FormatUint(contact, 10),
				Value:   string(loadedSession.Serialize()),
				Options: nil,
				ResChan: make(chan etcd.TxnInsertRes),
			}
			a.signal.txnInsertChan <- insert

			current = loadedSession.Serialize()

		} else {
			fmt.Println("读session数据")
			//从etcd获取session
			current = res.Result.Kvs[0].Value

		}
	}

	//反序列化session
	deserializedSession, _ := record.NewSessionFromBytes(current,
		a.SignalprotocolStore.SessionStore.serializer.Session,
		a.SignalprotocolStore.SessionStore.serializer.State)

	//加载session相关
	a.SignalprotocolStore.SessionStore.StoreSession(a.RemoteAddresses[contact], deserializedSession)
	a.SignalprotocolStore.IdentityKeyStore.SaveIdentity(a.RemoteAddresses[contact], deserializedSession.SessionState().RemoteIdentityKey())

	cipher := session.NewCipherFromSession(a.RemoteAddresses[contact], a.SignalprotocolStore.SessionStore,
		a.SignalprotocolStore.PreKeyStore, a.SignalprotocolStore.IdentityKeyStore, a.Serializer.PreKeySignalMessage, a.Serializer.SignalMessage)

	if a.Cipher[contact] == nil {
		a.Cipher[contact] = cipher
	}

	timeStamp := proto.Uint64(uint64(time.Now().Unix()))

	//mediaKey, _ := hex.DecodeString("3b61de5dfe9ad2b4adc48a467c84c2e8767aac97f255e5a94915a80c9476660f")
	//fileSha256, _ := hex.DecodeString("84c40e4272fae70e0244a0e27ebc791c7f5b0e89fd1ebd8b0d7ff4c365922137")
	//fileEncSha256, _ := hex.DecodeString("76a4507c23ce14c032cb85a286908f69116bded49fe45b43127d7fe5fa431fb7")
	//fileLength := uint64(38997)
	//directPath := "/v/t62.7118-24/30927259_1125657852157689_1631449215857950242_n.enc?ccb=11-4&oh=01_AdT2tsq-7QNyALaahBKNsuO3Il5rK8h8OfF_YkWEXqbeXQ&oe=64A62411"
	//url := "https://mmg.whatsapp.net/v/t62.7118-24/30927259_1125657852157689_1631449215857950242_n.enc?ccb=11-4&oh=01_AdT2tsq-7QNyALaahBKNsuO3Il5rK8h8OfF_YkWEXqbeXQ&oe=64A62411&mms3=true"

	version := proto.Int32(2)
	for _, info := range fileInfolist {
		msg := &protobuf.Message{
			DocumentMessage: &protobuf.Message_DocumentMessage{
				Url:           proto.String(info.Url),
				Mimetype:      proto.String(info.FileType),
				Title:         proto.String(info.FileName),
				FileSha256:    info.FileSha256,
				FileLength:    &info.FileLength,
				MediaKey:      info.MediaKey,
				FileName:      proto.String(info.FileName),
				FileEncSha256: info.FileEncSha256,
				DirectPath:    &info.DirectPath,
			},
			MessageContextInfo: &protobuf.MessageContextInfo{
				DeviceListMetadata: &protobuf.DeviceListMetadata{
					SenderTimestamp: timeStamp,
				},
				DeviceListMetadataVersion: version,
			},
		}
		//???
		msg.MessageContextInfo.DeviceListMetadata.SenderTimestamp = timeStamp

		msgbytes, _ := proto.Marshal(msg)
		ciphertext, _ := cipher.Encrypt(bytesWithPadding(msgbytes))
		//fmt.Printf("    Cipher:%v  ", base64.StdEncoding.EncodeToString(ciphertext.Serialize()))
		request := createMediaMessageRequest(recipientJid, "document", ciphertext)
		//fmt.Printf("   messageRequest:%v", request)
		a.send(request)

		lastLoadedSession := a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[contact])

		put := etcd.PutReq{
			Key:     "/arthas/Session/" + strconv.FormatUint(accountPhone, 10) + "/" + strconv.FormatUint(contact, 10),
			Value:   string(lastLoadedSession.Serialize()),
			Options: nil,
			ResChan: make(chan etcd.PutRes),
		}
		a.signal.putChan <- put
		time.Sleep(1 * time.Second)
	}
}
