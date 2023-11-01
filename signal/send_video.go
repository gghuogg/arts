package signal

import (
	"arthas/etcd"
	"arthas/protobuf"
	"arthas/util"
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/google/uuid"
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
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

type SendVideoFileDetail struct {
	Sender   uint64
	Receiver uint64
	FileType []SendFileType
	ResChan  chan *protobuf.ResponseMessage
}

func (a *Account) sendVideo(detail SendVideoFileDetail) {
	var deletedItems []string
	for _, f := range detail.FileType {
		var (
			videoThumbnail string

			uuidStr = uuid.NewString()
		)

		// Get thumbnail video with ffmpeg
		thumbnailVideoPath := fmt.Sprintf("%s/%s", "cmd/arthtoolTG/testFile", uuidStr+".png")
		cmdThumbnail := exec.Command("ffmpeg", "-i", f.Path, "-ss", "00:00:01.000", "-vframes", "1", thumbnailVideoPath)
		err := cmdThumbnail.Run()
		if err != nil {
			level.Error(a.logger).Log("msg", "failed to create thumbnail.", "err", err)
			return
		}

		// Resize Thumbnail
		openImageBuffer, err := os.ReadFile(thumbnailVideoPath)
		buf := bytes.NewBuffer(openImageBuffer)
		thumbnailImage, err := imaging.Decode(buf)
		if err != nil {
			level.Error(a.logger).Log("msg", "failed to resize thumbnail", "err", err)
			return
		}
		thumbnailResizeVideoPath := fmt.Sprintf("%s/%s", "cmd/arthtoolTG/testFile", uuidStr+"_resize.png")
		thumbnailImage = imaging.Resize(thumbnailImage, 600, 90, imaging.Lanczos)
		err = imaging.Save(thumbnailImage, thumbnailResizeVideoPath)
		if err != nil {
			level.Error(a.logger).Log("msg", "failed to create thumbnailImage thumbnail ", "err", err)
			return
		}
		deletedItems = append(deletedItems, thumbnailVideoPath)
		deletedItems = append(deletedItems, thumbnailResizeVideoPath)
		videoThumbnail = thumbnailResizeVideoPath

		mediaKey := make([]byte, 32)
		_, _ = rand.Read(mediaKey)

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

		dataWaVideo := f.FileBytes

		encryptedData, fileSha256, fileEncSha256, err := EncryptMediaBytes(dataWaVideo, mediaKey, []byte(labelVideoKeys))
		if err != nil {
			level.Info(a.logger).Log("msg", "encryp file fail:"+err.Error(), "user", a.ClientPayLoad.Username)
			return
		}

		hostname := "media-sin6-1.cdn.whatsapp.net"
		filename := base64.URLEncoding.EncodeToString(fileEncSha256) // use url-encoding instead of standard-encoding

		req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("https://%s/optimistic/video/%s", hostname, filename), bytes.NewReader(encryptedData))

		if err != nil {
			level.Info(a.logger).Log("msg", "file new http request fail:"+err.Error(), "user", a.ClientPayLoad.Username)
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
			level.Info(a.logger).Log("msg", "file http client do send fail:"+err.Error(), "user", a.ClientPayLoad.Username)
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
			level.Info(a.logger).Log("msg", "unarshal json fail:"+err.Error(), "user", a.ClientPayLoad.Username)
		}
		msg := a.signal.StatFileRequest(respBytes, strconv.FormatInt(timestamp, 10))
		fileInfo := &FileKeyInfo{
			MediaKey:      mediaKey,
			FileSha256:    fileSha256,
			FileEncSha256: fileEncSha256,
			FileLength:    uint64(len(dataWaVideo)),
			Url:           fileUrlInfo.Url,
			DirectPath:    fileUrlInfo.DirectpPath,
			FileType:      f.FileType,
		}
		a.send(msg)
		a.SendVideMessage(detail, fileInfo, videoThumbnail)
		go func() {
			errDelete := util.RemoveFile(1, deletedItems...)
			if errDelete != nil {
				fmt.Println(errDelete)
			}
		}()
	}

}

func (a *Account) SendVideMessage(detail SendVideoFileDetail, info *FileKeyInfo, videoThumbnail string) {
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

	dataWaThumbnail, err := os.ReadFile(videoThumbnail)
	if err != nil {
		level.Info(a.logger).Log("msg", "open file fail:"+err.Error(), "user", a.ClientPayLoad.Username)
		return
	}

	timeStamp := proto.Uint64(uint64(time.Now().Unix()))

	version := proto.Int32(2)
	msg := &protobuf.Message{
		VideoMessage: &protobuf.Message_VideoMessage{
			Url:                 proto.String(info.Url),
			Mimetype:            proto.String(info.FileType),
			FileLength:          &info.FileLength,
			FileSha256:          info.FileSha256,
			FileEncSha256:       info.FileEncSha256,
			MediaKey:            info.MediaKey,
			DirectPath:          &info.DirectPath,
			ViewOnce:            proto.Bool(false),
			JpegThumbnail:       dataWaThumbnail,
			ThumbnailEncSha256:  dataWaThumbnail,
			ThumbnailSha256:     dataWaThumbnail,
			ThumbnailDirectPath: &info.DirectPath,
		},
		MessageContextInfo: &protobuf.MessageContextInfo{
			DeviceListMetadata: &protobuf.DeviceListMetadata{
				SenderTimestamp: timeStamp,
			},
			DeviceListMetadataVersion: version,
		},
	}
	msgBytes, _ := proto.Marshal(msg)
	ciphertext, _ := cipher.Encrypt(bytesWithPadding(msgBytes))
	// mediaType 分两种，视屏格式为video，动图为gif
	request := createMediaMessageRequest(recipientJid, "video", ciphertext)
	a.send(request)

	lastLoadedSession := a.SignalprotocolStore.SessionStore.LoadSession(a.RemoteAddresses[contact])

	put := etcd.PutReq{
		Key:     "/arthas/Session/" + strconv.FormatUint(accountPhone, 10) + "/" + strconv.FormatUint(contact, 10),
		Value:   string(lastLoadedSession.Serialize()),
		Options: nil,
		ResChan: make(chan etcd.PutRes),
	}
	a.signal.putChan <- put
}
