package tg

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"arthas/telegram"
	"github.com/go-kit/log"
	"strconv"
)

func init() {
	im, err := rpc.GetIM(consts.IMTg)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_SEND_CODE, &TgSendCodeAction{})
}

type TgSendCodeAction struct{}

func (t *TgSendCodeAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	detail := in.GetSendCodeDetail().GetDetails()
	if detail != nil {
		if !detail.Flag {
			return &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, LoginId: detail.LoginId, Comment: detail.Comments}, nil

		} else {
			if len(detail.GetSendCode()) > 0 {
				for k, v := range detail.SendCode {
					sendCode := telegram.SendCodeDetail{Account: k, Code: v, ResChan: make(chan telegram.LoginDetailRes)}
					h.GetTgSendCodeChan() <- sendCode
					res := <-sendCode.ResChan
					if res.LoginStatus == protobuf.AccountStatus_SUCCESS {
						return &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS, LoginId: res.LoginId, Account: strconv.FormatUint(k, 10)}, nil

					} else if res.LoginStatus == protobuf.AccountStatus_LOGIN_CODE_FAIL {
						return &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_LOGIN_VERIFY_CODE_FAIL, LoginId: res.LoginId, Comment: res.Comment}, nil

					} else {
						return &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, LoginId: res.LoginId, Comment: res.Comment}, nil
					}
				}
			} else {
				return &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: "send protobuf msg is nil or having err "}, nil
			}
		}

	}
	return
}
