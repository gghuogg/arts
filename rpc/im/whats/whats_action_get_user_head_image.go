package whats

import (
	"arthas/consts"
	"arthas/protobuf"
	"arthas/rpc"
	"arthas/signal"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func init() {
	im, err := rpc.GetIM(consts.IMWhats)
	if err != nil {
		return
	}
	im.RegisterAction(protobuf.Action_GET_USER_HEAD_IMAGE, &whatsGetUserHeadImageAction{})
}

type whatsGetUserHeadImageAction struct{}

func (t *whatsGetUserHeadImageAction) Handler(l log.Logger, h *rpc.Handler, in *protobuf.RequestMessage) (result *protobuf.ResponseMessage, err error) {
	msg := in.GetGetUserHeadImage().GetHeadImage()
	if &msg != nil {
		for k, v := range msg {
			value, ok := h.GetAccountsSync().Load(k)
			if ok {
				a := value.(*signal.Account)
				fmt.Println(a)
				//if a.IsLogin {
				headImage := signal.GetUserHeadImageDetail{
					Account:            k,
					GetUserHeaderPhone: v.GetAccount(),
				}

				h.GetWhatsGetUserHeadImageChan() <- headImage
				//}
			} else {
				_ = level.Debug(l).Log("msg", "Account is exist,will be not continue sync account", "userJid", k)

			}
		}
	}

	result = &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_SUCCESS}
	return
}
