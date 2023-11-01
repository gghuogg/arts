package signal

import (
	"fmt"
	"strconv"
)

type GetUserHeadImageDetail struct {
	Account            uint64
	GetUserHeaderPhone []uint64
}

func (a *Account) getUserHeadImage(users []uint64) {
	for _, user := range users {
		userJid := strconv.FormatUint(user, 10) + "@s.whatsapp.net"
		msg := a.signal.ProfilePicturePreviewRequest(userJid)
		fmt.Println("获取头像")
		a.send(msg)
	}
}
