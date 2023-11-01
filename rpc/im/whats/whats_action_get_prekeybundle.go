package whats

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"github.com/go-kit/log"
)

func init() {
	im, err := rpc.GetIM(consts.IMWhats)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_SYNC_ACCOUNT_KEY, &whatsGetPreKeyBundleAction{})
}

type whatsGetPreKeyBundleAction struct{}

func (t *whatsGetPreKeyBundleAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	details := in.GetQueryPrekeybundleDetail().GetDetails()
	if len(details) == 0 {
		result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL}
		return
	} else if len(details) == 1 {
		_ = details[0].Key
		_ = details[0].Values
	} else {
		//for _, detail := range details {
		//
		//}
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}

	return
}
