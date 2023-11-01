package telegram

import (
	"arthas/etcd"
	"arthas/protobuf"
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"strconv"
)

type SendCodeDetail struct {
	Account uint64
	Code    string
	ResChan chan LoginDetailRes
}

type noSignUp struct{}

func (c noSignUp) SignUp(_ context.Context) (auth.UserInfo, error) {
	return auth.UserInfo{}, errors.New("searchx don't support sign up Telegram account")
}

func (c noSignUp) AcceptTermsOfService(_ context.Context, tos tg.HelpTermsOfService) error {
	return &auth.SignUpRequired{TermsOfService: tos}
}

type tgAuth struct {
	manager *Manager
	noSignUp
	phoneNum           string
	passwd             string
	loginId            string
	watchChan          chan etcd.WatchReq
	sendCodeChan       chan SendCodeDetail
	ResChan            chan LoginDetailRes
	loginDetailResChan chan LoginDetailRes
}

func (a *tgAuth) Phone(_ context.Context) (result string, err error) {

	color.Blue("Sending Code...")
	fmt.Println("登录id（验证码需要）：" + a.loginId)
	return a.phoneNum, nil
}

func (a *tgAuth) Password(_ context.Context) (string, error) {

	return a.passwd, nil
}

func (a *tgAuth) Code(ctx context.Context, ac *tg.AuthSentCode) (string, error) {

	// 需要验证码
	res := LoginDetailRes{
		LoginStatus: protobuf.AccountStatus_NEED_SEND_CODE,
		LoginId:     a.loginId,
		Comment:     "Sending Code...",
	}
	a.loginDetailResChan <- res

	// 等待验证码
	for {
		select {
		case codeMsg := <-a.sendCodeChan:
			if strconv.FormatUint(codeMsg.Account, 10) == a.phoneNum {
				a.ResChan = codeMsg.ResChan
				return codeMsg.Code, nil
			}
		}
	}

}
