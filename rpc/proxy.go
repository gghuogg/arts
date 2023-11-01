package rpc

import (
	"arthas/protobuf"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	gopool "github.com/panjf2000/ants/v2"
	"golang.org/x/net/context"
	"strings"
	"sync"
)

type ProxyServer struct {
	protobuf.UnimplementedArthasServer
	logger             log.Logger
	Handler            *Handler
	cancelServerStream context.CancelFunc // 用于在必要时取消与服务器的连接
	targetAddr         string
	mu                 sync.Mutex
	wg                 sync.WaitGroup
	responseChan       chan *protobuf.ResponseMessage
	errChan            chan error
	errsChan           chan []error
	finalResponse      *protobuf.ResponseMessage
	taskSyncMaps       map[string]*sync.Map
	once               sync.Once
	currentWsIPIndex   int
	gp                 *gopool.Pool
}

func (p *ProxyServer) Connect(ctx context.Context, in *protobuf.RequestMessage) (*protobuf.ResponseMessage, error) {
	p.once.Do(p.initialize)
	p.wg.Add(1)

	p.gp.Submit(func() {
		defer p.wg.Done()
		//p.handleRequestMessage(in)
		if in != nil {
			pr, err := GetPR(in.GetType())
			if err != nil {
				level.Error(p.logger).Log("msg", "pr error ", "err", err)
				in.Action = protobuf.Action_ERR_STOP
				p.errChan <- err
			} else {
				_, err1 := pr.Handler(p.logger, p, in)
				if err1 != nil {
					in.Action = protobuf.Action_ERR_STOP
					level.Error(p.logger).Log("msg", "pr error ", "err", err)
					p.errChan <- err1

				}
			}

		}
	})
	p.wg.Wait()

	switch in.Action {
	case protobuf.Action_LOGIN:
		if in.GetType() == "telegram" {
			aErr := p.tgLogin()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		} else {
			p.whatsappLogin()
		}
	case protobuf.Action_LOGOUT:
		if in.GetType() == "telegram" {
			aErr := p.tgLogout()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		} else if in.GetType() == "whatsapp" {
			p.whatsappLogout()
		}
	case protobuf.Action_SEND_MESSAGE:
		if in.GetType() == "telegram" {
			aErr := p.tgSendMsg()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		} else if in.GetType() == "whatsapp" {
			p.whatsappSendMsg()
		}
	case protobuf.Action_SEND_VCARD_MESSAGE:
		if in.GetType() == "telegram" {
			aErr := p.tgSendVcard()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		} else if in.GetType() == "whatsapp" {
			p.whatsappSendVcardMsg()
		}
	case protobuf.Action_SEND_GROUP_MESSAGE:
		if in.GetType() == "telegram" {
			aErr := p.tgSendGroupMsg()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}
		} else if in.GetType() == "whatsapp" {
		}
	case protobuf.Action_GET_USER_HEAD_IMAGE:
		if in.GetType() == "telegram" {
		} else if in.GetType() == "whatsapp" {
			p.whatsappGetUserHeadImage()
		}
	case protobuf.Action_SYNC_ACCOUNT_KEY:
		if in.GetType() == "telegram" {
		} else if in.GetType() == "whatsapp" {
			p.whatsappSyncKey()
		}
	case protobuf.Action_SYNC_CONTACTS:
		if in.GetType() == "telegram" {
			aErr := p.tgSyncContact()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		} else if in.GetType() == "whatsapp" {
			p.whatsappSyncContact()
		}
	case protobuf.Action_SYNC_APP_INFO:
		if in.GetType() == "telegram" {
			aErr := p.tgSyncAppInfo()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		}
	case protobuf.Action_CREATE_GROUP:
		if in.GetType() == "telegram" {
			aErr := p.tgCreateGroup()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		}
	case protobuf.Action_ADD_GROUP_MEMBER:
		if in.GetType() == "telegram" {
			aErr := p.tgAddGroupMember()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		}
	case protobuf.Action_GET_GROUP_MEMBERS:
		if in.GetType() == "telegram" {
			aErr := p.tgGetGroupMembers()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		}
	case protobuf.Action_GET_CHANNEL_MEMBER:
		if in.GetType() == "telegram" {
			aErr := p.tgGetChannelMembers()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		}
	case protobuf.Action_SEND_FILE:
		if in.GetType() == "telegram" {
			aErr := p.tgSendFileMsg()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		} else if in.GetType() == "whatsapp" {
			p.whatsappSendFile()
		}
	case protobuf.Action_SEND_PHOTO:
		if in.GetType() == "telegram" {
			aErr := p.tgSendPhotoMsg()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		} else if in.GetType() == "whatsapp" {
			p.whatsappSendPhoto()
		}
	case protobuf.Action_SEND_VIDEO:
		if in.GetType() == "telegram" {
			aErr := p.tgSendVideo()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		} else if in.GetType() == "whatsapp" {
			p.whatsappSendVideo()
		}
	case protobuf.Action_SEND_CODE:
		if in.GetType() == "telegram" {
			aErr := p.tgSendCode()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		} else if in.GetType() == "whatsapp" {
		}
	case protobuf.Action_CONTACT_LIST:
		if in.GetType() == "telegram" {
			aErr := p.getContactList()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		} else if in.GetType() == "whatsapp" {
		}
	case protobuf.Action_DIALOG_LIST:
		if in.GetType() == "telegram" {
			aErr := p.getDialogList()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		} else if in.GetType() == "whatsapp" {
		}
	case protobuf.Action_Get_MSG_HISTORY:
		if in.GetType() == "telegram" {
			aErr := p.tgGetMsgHistory()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		} else if in.GetType() == "whatsapp" {
		}
	case protobuf.Action_DOWNLOAD_FILE:
		if in.GetType() == "telegram" {
			aErr := p.tgDownloadFile()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		}

	case protobuf.Action_CREATE_CHANNEL:
		if in.GetType() == "telegram" {
			aErr := p.tgCreateChannel()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		}
	case protobuf.Action_INVITE_TO_CHANNEL:
		if in.GetType() == "telegram" {
			aErr := p.tgInviteToChannel()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		}
	case protobuf.Action_JOIN_BY_LINK:
		if in.GetType() == "telegram" {
			aErr := p.tgJoinByLink()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		}
	case protobuf.Action_GET_EMOJI_GROUP:
		if in.GetType() == "telegram" {
			aErr := p.tgGetEmojiGroups()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		}
	case protobuf.Action_IMPORT_TG_SESSION:
		if in.GetType() == "telegram" {
			aErr := p.tgImportSession()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		}
	case protobuf.Action_MESSAGES_REACTION:
		if in.GetType() == "telegram" {
			aErr := p.tgMessagesReaction()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		}
	case protobuf.Action_LEAVE:
		if in.GetType() == "telegram" {
			aErr := p.tgLeave()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}

		}
	case protobuf.Action_GET_ONLINE_ACCOUNTS:
		if in.GetType() == "telegram" {
			aErr := p.tgGetOnlineAccounts()
			if len(aErr) > 0 {
				p.errsChan <- aErr
			}
		}
	}

	select {
	case msg := <-p.Handler.finalResponse:
		if msg != nil && msg.Account != "" {
			// 删除登录需要验证码的loginId，一串uuid
			p.Handler.tgLoginCodeSync.Delete(msg.Account)
		}
		return msg, nil
	case e := <-p.errChan:
		level.Error(p.logger).Log("msg", "errChan send error ", "err", e.Error())
		return &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: e.Error()}, e
	case el := <-p.errsChan:
		var errString string
		errorMessages := make([]string, len(el))
		for i, e := range el {
			errorMessages[i] = e.Error()
		}
		errString = strings.Join(errorMessages, "\n")
		return &protobuf.ResponseMessage{ActionResult: protobuf.ActionResult_ALL_FAIL, Comment: errString}, nil
	}

}

func (p *ProxyServer) initialize() {
	p.errChan = make(chan error, 100)
	p.errsChan = make(chan []error, 100)
	p.finalResponse = &protobuf.ResponseMessage{}
	p.taskSyncMaps = make(map[string]*sync.Map)
}
