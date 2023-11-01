package telegram

import (
	"bytes"
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/tg"
	"io"
	"math"
	"math/rand"
)

func uint64ToIntSafe(u64 uint64) (int, error) {
	if u64 <= uint64(math.MaxInt64) {
		return int(u64), nil
	}
	return 0, fmt.Errorf("Value is too large to convert to int")
}

func RandInt64(randSource io.Reader) (int64, error) {
	var buf [bin.Word * 2]byte
	if _, err := io.ReadFull(randSource, buf[:]); err != nil {
		return 0, err
	}
	b := &bin.Buffer{Buf: buf[:]}
	return b.Long()
}

func generateRandomValue() int {
	return rand.Intn(1501) + 500
}

func getFileFromOSSAndConvertToBytes(url string) ([]byte, error) {

	client, err := oss.New("http://oss-ap-southeast-1.aliyuncs.com", "LTAI5t7aFWwdbZpsP5JWFVty", "wtH8LIVdNsymsuirE3wgXgcFqC3y4s")
	if err != nil {
		return nil, err
	}

	bucket, err := client.Bucket("tgcloud")
	if err != nil {
		return nil, err
	}
	obj, err := bucket.GetObject(url)
	if obj != nil {
		defer obj.Close()
	}
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	_, err = io.Copy(&buf, obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

type Filter interface {
	Apply() tg.ChannelParticipantsFilterClass
}

type NameFilter struct {
}

func (nf *NameFilter) Apply() tg.ChannelParticipantsFilterClass {
	return &tg.ChannelParticipantsSearch{}
}

type RecentFilter struct{}

func (rf *RecentFilter) Apply() tg.ChannelParticipantsFilterClass {
	return &tg.ChannelParticipantsRecent{}
}

type AdminsFilter struct{}

func (af *AdminsFilter) Apply() tg.ChannelParticipantsFilterClass {
	return &tg.ChannelParticipantsAdmins{}
}

type KickedFilter struct{}

func (kf *KickedFilter) Apply() tg.ChannelParticipantsFilterClass {
	return &tg.ChannelParticipantsKicked{}
}

type BotFilter struct{}

func (bf *BotFilter) Apply() tg.ChannelParticipantsFilterClass {
	return &tg.ChannelParticipantsBots{}
}

type BannedFilter struct{}

func (baf *BannedFilter) Apply() tg.ChannelParticipantsFilterClass {
	return &tg.ChannelParticipantsBanned{}
}

type ContactsFilter struct{}

func (cf *ContactsFilter) Apply() tg.ChannelParticipantsFilterClass {
	return &tg.ChannelParticipantsContacts{}
}

type MentionsFilter struct {
	TopMsgID int
}

func (mf *MentionsFilter) Apply() tg.ChannelParticipantsFilterClass {
	return &tg.ChannelParticipantsMentions{
		TopMsgID: mf.TopMsgID,
	}
}
