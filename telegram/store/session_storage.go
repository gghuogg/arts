package store

import (
	"arthas/protobuf"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/go-faster/errors"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/session"
	"github.com/gotd/td/tg"
	"google.golang.org/protobuf/proto"
)

type ArtsTgLoader struct {
	Storage session.Storage
}

const latestVersion = 1

//var ErrNotFound = errors.New("session storage: not found")

// Load loads Data from Storage.
func (l *ArtsTgLoader) Load(ctx context.Context) (data *session.Data, err error) {
	buf, err := l.Storage.LoadSession(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "load")
	}
	if len(buf) == 0 {
		return nil, session.ErrNotFound
	}
	t := &protobuf.TgAccountSession{}
	err = proto.Unmarshal(buf, t)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}
	if t.Version != latestVersion {
		return nil, errors.Wrapf(session.ErrNotFound, "version mismatch (%d != %d)", t.Version, latestVersion)
	}
	data = getDataByProto(t)

	return data, nil
}

// Save saves Data to Storage.
func (l *ArtsTgLoader) Save(ctx context.Context, data *session.Data) error {
	tgProtoSession := getSessionToProto(data)
	buf, err := proto.Marshal(tgProtoSession)
	if err != nil {
		return errors.Wrap(err, "marshal")
	}
	if err := l.Storage.StoreSession(ctx, buf); err != nil {
		return errors.Wrap(err, "store")
	}
	return nil

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
	if err != nil {
		fmt.Println("decode authKey err:", err)
		return nil
	}
	authKeyID, err := decodeStringToBytes(protoData.AuthKeyID)
	if err != nil {
		fmt.Println("decode authKeyID err:", err)
		return nil
	}
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
		AuthKeyID: authKeyID,
		Salt:      protoData.Salt,
	}
	return data

}

func getSessionToProto(data *session.Data) *protobuf.TgAccountSession {
	entryConfig := data.Config
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

	s := &protobuf.TgAccountSession{
		Version: latestVersion,
		Data: &protobuf.Data{
			Addr:      data.Addr,
			DC:        int32(data.DC),
			AuthKey:   encodeBytesToString(data.AuthKey),
			AuthKeyID: encodeBytesToString(data.AuthKeyID),
			Salt:      int64(data.Salt),
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
