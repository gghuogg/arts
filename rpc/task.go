package rpc

import (
	"arthas/protobuf"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/google/uuid"
	"strconv"
)

const (
	STREAM                 = "stream"
	WS_LOGIN               = "whatsappLogin"
	WS_LOGOUT              = "whatsappLogout"
	WS_SEND_MSG            = "whatsappSendMsg"
	WS_SYNC_KEY            = "whatsappSyncKey"
	WS_SYNC_CONTACT        = "whatsappSyncContact"
	WS_GET_USER_HEAD_IMAGE = "whatsapp_get_user_head_image"
	WS_SEND_VCARD_MSG      = "whatsapp_send_vcard_msg"
	WS_SEND_PHOTO          = "whatsapp_send_photo"
	WS_SEND_VIDEO          = "whatsapp_send_video"
	WS_SEND_FILE           = "whatsapp_send_file"
	TG_LOGIN               = "tgLogin"
	TG_LOGOUT              = "tgLogout"
	TG_SYNC_INFO           = "tgSyncAppInfo"
	TG_SEND_MSG            = "tgSendMsg"
	TG_CREATE_GROUP        = "tgCreateGroup"
	TG_CREATE_CHANNEL      = "tgCreateChannel"
	TG_SEND_GROUP_MSG      = "tgSendGroupMsg"
	TG_ADD_GROUP_MEMBER    = "tgAddGroupMember"
	TG_INVITE_TO_CHANNEL   = "tgInviteToChannel"
	TG_JOIN_BY_LINK        = "tgJoinByLink"
	TG_LEAVE               = "tgLeave"
	TG_GET_EMOJI_GROUPS    = "tgGetEmojiGroup"
	TG_MESSAGES_REACTION   = "tgMessagesReaction"
	TG_GET_GROUP_MEMBERS   = "tgGetGroupMembers"
	TG_GET_CHANNEL_MEMBERS = "tgGetChannelMembers"
	TG_SEND_PHOTO          = "tgSendPhoto"
	TG_SEND_FILE           = "tgSendFile"
	TG_SEND_VIDEO          = "tgSendVideo"
	TG_SYNC_CONTACT        = "tgSyncContact"
	TG_SEND_CODE           = "tgSendCode"
	TG_SEND_VCARD          = "tgSendVcard"
	TG_RECEIVING_MSG       = "tgReceivingMsg"
	TG_GET_CONTACT_LIST    = "tgGetContactList"
	TG_GET_DIALOG_LIST     = "tgGetDialogList"
	TG_GET_DOWNLOAD_FILE   = "tgGetDownLoadFile"
	TG_GET_MSG_HISTORY     = "tgGetMsgHistory"
	TG_IMPORT_SESSION      = "tgImportSession"
	TG_GET_ONLINE_ACCOUNTS = "tgGetOnlineAccounts"
)

func (p *ProxyServer) whatsappLogin() {

	wsLogin := p.getSyncMap(WS_LOGIN)
	wsLogin.Range(func(key, value any) bool {

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail")
		}
		streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)
		req := &protobuf.RequestMessage{
			Action: protobuf.Action_LOGIN,
			Type:   "whatsapp",
			ActionDetail: &protobuf.RequestMessage_OrdinaryAction{
				OrdinaryAction: &protobuf.OrdinaryAction{
					LoginDetail: value.(map[uint64]*protobuf.LoginDetail),
				},
			},
		}

		if err := streamConnectClient.Send(req); err != nil {
			level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
		}

		wsLogin.Delete(key.(string))

		p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		return true
	})
}

func (p *ProxyServer) tgLogout() []error {
	tgLogout := p.getSyncMap(TG_LOGOUT)

	errs := make([]error, 0)
	var err error
	tgLogout.Range(func(key, value any) bool {

		ld := value.(map[uint64]*protobuf.LogoutDetail)
		for userJid, _ := range ld {
			p.Handler.GetServerMap().Delete(userJid)
		}

		tgLogout.Delete(key.(string))

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)
			req := &protobuf.RequestMessage{
				Action: protobuf.Action_LOGOUT,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_LogoutAction{
					LogoutAction: &protobuf.LogoutAction{
						LogoutDetail: ld,
					},
				},
			}

			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}
			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) whatsappLogout() {

	wsLogout := p.getSyncMap(WS_LOGOUT)
	wsLogout.Range(func(key, value any) bool {

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail")
		}
		streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)
		req := &protobuf.RequestMessage{
			Action: protobuf.Action_LOGOUT,
			Type:   "whatsapp",
			ActionDetail: &protobuf.RequestMessage_OrdinaryAction{
				OrdinaryAction: &protobuf.OrdinaryAction{
					LoginDetail: value.(map[uint64]*protobuf.LoginDetail),
				},
			},
		}

		if err := streamConnectClient.Send(req); err != nil {
			level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
		}

		wsLogout.Delete(key.(string))
		p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		return true
	})
}

func (p *ProxyServer) tgSendVideo() []error {

	wsSendVcardMsg := p.getSyncMap(TG_SEND_VIDEO)
	errs := make([]error, 0)
	var err error

	wsSendVcardMsg.Range(func(key, value any) bool {

		wsSendVcardMsg.Delete(key.(string))

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

			sendDatas := value.([]map[uint64]*protobuf.UintTgFileDetailValue)
			list := make([]*protobuf.SendFileAction, 0)

			for _, v1 := range sendDatas {
				for k2, v2 := range v1 {
					tmp := &protobuf.SendFileAction{}
					sendData := make(map[uint64]*protobuf.UintTgFileDetailValue)
					sendData[k2] = v2
					tmp.SendTgData = sendData
					list = append(list, tmp)
				}

			}

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_SEND_VIDEO,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_SendFileDetail{
					SendFileDetail: &protobuf.SendFileDetail{
						Details: list,
					},
				},
			}
			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) whatsappSendVideo() {

	wsSendVcardMsg := p.getSyncMap(WS_SEND_VIDEO)
	wsSendVcardMsg.Range(func(key, value any) bool {
		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail")
		}
		streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

		sendDatas := value.(map[uint64]*protobuf.UintFileDetailValue)
		list := make([]*protobuf.SendFileAction, 0)
		list = append(list, &protobuf.SendFileAction{
			SendData: sendDatas,
		})

		req := &protobuf.RequestMessage{
			Action: protobuf.Action_SEND_VIDEO,
			Type:   "whatsapp",
			ActionDetail: &protobuf.RequestMessage_SendVideoDetail{
				SendVideoDetail: &protobuf.SendVideoDetail{
					Details: list,
				},
			},
		}
		if err := streamConnectClient.Send(req); err != nil {
			level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
		}

		wsSendVcardMsg.Delete(key.(string))
		p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		return true
	})
}

func (p *ProxyServer) whatsappSendFile() {

	wsSendVcardMsg := p.getSyncMap(WS_SEND_FILE)
	wsSendVcardMsg.Range(func(key, value any) bool {
		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail")
		}
		streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

		sendDatas := value.(map[uint64]*protobuf.UintFileDetailValue)
		list := make([]*protobuf.SendFileAction, 0)
		list = append(list, &protobuf.SendFileAction{
			SendData: sendDatas,
		})

		req := &protobuf.RequestMessage{
			Action: protobuf.Action_SEND_FILE,
			Type:   "whatsapp",
			ActionDetail: &protobuf.RequestMessage_SendFileDetail{
				SendFileDetail: &protobuf.SendFileDetail{
					Details: list,
				},
			},
		}
		if err := streamConnectClient.Send(req); err != nil {
			level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
		}

		wsSendVcardMsg.Delete(key.(string))
		p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		return true
	})
}

func (p *ProxyServer) tgSendFileMsg() []error {

	wsSendVcardMsg := p.getSyncMap(TG_SEND_FILE)
	errs := make([]error, 0)
	var err error

	wsSendVcardMsg.Range(func(key, value any) bool {

		wsSendVcardMsg.Delete(key.(string))

		v, ok := p.Handler.StreamMap.Load(key.(string))

		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

			sendDatas := value.([]map[uint64]*protobuf.UintTgFileDetailValue)
			list := make([]*protobuf.SendFileAction, 0)

			for _, v1 := range sendDatas {
				for k2, v2 := range v1 {
					tmp := &protobuf.SendFileAction{}
					sendData := make(map[uint64]*protobuf.UintTgFileDetailValue)
					sendData[k2] = v2
					tmp.SendTgData = sendData
					list = append(list, tmp)
				}

			}

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_SEND_FILE,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_SendFileDetail{
					SendFileDetail: &protobuf.SendFileDetail{
						Details: list,
					},
				},
			}
			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) whatsappSendPhoto() {

	wsSendVcardMsg := p.getSyncMap(WS_SEND_PHOTO)
	wsSendVcardMsg.Range(func(key, value any) bool {
		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail")
		}
		streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

		sendDatas := value.(map[uint64]*protobuf.UintFileDetailValue)
		list := make([]*protobuf.SendFileAction, 0)
		list = append(list, &protobuf.SendFileAction{
			SendData: sendDatas,
		})

		req := &protobuf.RequestMessage{
			Action: protobuf.Action_SEND_PHOTO,
			Type:   "whatsapp",
			ActionDetail: &protobuf.RequestMessage_SendImageFileDetail{
				SendImageFileDetail: &protobuf.SendImageFileDetail{
					Details: list,
				},
			},
		}
		if err := streamConnectClient.Send(req); err != nil {
			level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
		}

		wsSendVcardMsg.Delete(key.(string))
		p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		return true
	})

}
func (p *ProxyServer) whatsappSendMsg() {

	wsSend := p.getSyncMap(WS_SEND_MSG)
	wsSend.Range(func(key, value any) bool {
		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail")
		}
		streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

		sendDatas := value.([]map[uint64]*protobuf.UintkeyStringvalue)
		list := make([]*protobuf.SendMessageAction, 0)

		for _, v1 := range sendDatas {
			for k2, v2 := range v1 {
				tmp := &protobuf.SendMessageAction{}
				sendData := make(map[uint64]*protobuf.UintkeyStringvalue)
				sendData[k2] = &protobuf.UintkeyStringvalue{Key: v2.Key, Values: v2.Values}
				tmp.SendData = sendData
				list = append(list, tmp)
			}

		}

		req := &protobuf.RequestMessage{
			Action: protobuf.Action_SEND_MESSAGE,
			Type:   "whatsapp",
			ActionDetail: &protobuf.RequestMessage_SendmessageDetail{
				SendmessageDetail: &protobuf.SendMessageDetail{
					Details: list,
				},
			},
		}

		if err := streamConnectClient.Send(req); err != nil {
			level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
		}

		wsSend.Delete(key.(string))
		p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		return true
	})

}

func (p *ProxyServer) whatsappSyncKey() {

	wsSync := p.getSyncMap(WS_SYNC_KEY)
	wsSync.Range(func(key, value any) bool {

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail")
		}
		streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

		req := &protobuf.RequestMessage{
			Action: protobuf.Action_SYNC_ACCOUNT_KEY,
			Type:   "whatsapp",
			ActionDetail: &protobuf.RequestMessage_SyncAccountKeyAction{
				SyncAccountKeyAction: &protobuf.SyncAccountKeyAction{
					KeyData: value.(map[uint64]*protobuf.KeyData),
				},
			},
		}

		if err := streamConnectClient.Send(req); err != nil {
			level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
		}

		wsSync.Delete(key.(string))
		p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		return true
	})
}

func (p *ProxyServer) whatsappSendVcardMsg() {

	wsSendVcardMsg := p.getSyncMap(WS_SEND_VCARD_MSG)
	wsSendVcardMsg.Range(func(key, value any) bool {
		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail")
		}
		streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

		sendDatas := value.(map[uint64]*protobuf.UintSenderVcard)
		list := make([]*protobuf.SendVCardMsgDetailAction, 0)

		for k, v := range sendDatas {
			tmp := &protobuf.SendVCardMsgDetailAction{}
			sendData := make(map[uint64]*protobuf.UintSenderVcard)
			sendData[k] = v
			tmp.SendData = sendData
			list = append(list, tmp)
		}

		req := &protobuf.RequestMessage{
			Action: protobuf.Action_SEND_VCARD_MESSAGE,
			Type:   "whatsapp",
			ActionDetail: &protobuf.RequestMessage_SendVcardMessage{
				SendVcardMessage: &protobuf.SendVCardMsgDetail{
					Details: list,
				},
			},
		}

		if err := streamConnectClient.Send(req); err != nil {
			level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
		}

		wsSendVcardMsg.Delete(key.(string))
		p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		return true
	})
}

func (p *ProxyServer) whatsappGetUserHeadImage() {

	wsGetUserHeadImage := p.getSyncMap(WS_GET_USER_HEAD_IMAGE)
	wsGetUserHeadImage.Range(func(key, value any) bool {
		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail")
		}
		streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

		req := &protobuf.RequestMessage{
			Action: protobuf.Action_GET_USER_HEAD_IMAGE,
			Type:   "whatsapp",
			ActionDetail: &protobuf.RequestMessage_GetUserHeadImage{
				GetUserHeadImage: &protobuf.GetUserHeadImageAction{
					HeadImage: value.(map[uint64]*protobuf.GetUserHeadImage),
				},
			},
		}

		if err := streamConnectClient.Send(req); err != nil {
			level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
		}

		wsGetUserHeadImage.Delete(key.(string))
		p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		return true
	})
}

func (p *ProxyServer) whatsappSyncContact() {

	wsSyncContact := p.getSyncMap(WS_SYNC_CONTACT)
	wsSyncContact.Range(func(key, value any) bool {
		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail")
		}
		streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

		v := value.(map[uint64][]uint64)
		list := make([]*protobuf.UintkeyUintvalue, 0)
		for user, contactList := range v {
			tmp := &protobuf.UintkeyUintvalue{}

			tmp.Key = user
			tmp.Values = contactList
			list = append(list, tmp)
		}

		req := &protobuf.RequestMessage{
			Action: protobuf.Action_SYNC_CONTACTS,
			Type:   "whatsapp",
			ActionDetail: &protobuf.RequestMessage_SyncContactDetail{
				SyncContactDetail: &protobuf.SyncContactDetail{
					Details: list,
				},
			},
		}

		if err := streamConnectClient.Send(req); err != nil {
			level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
		}

		wsSyncContact.Delete(key.(string))
		p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		return true
	})
}

func (p *ProxyServer) tgDownloadFile() []error {
	tgDownloadFile := p.getSyncMap(TG_GET_DOWNLOAD_FILE)
	errs := make([]error, 0)
	var err error

	tgDownloadFile.Range(func(key, value any) bool {

		tgDownloadFile.Delete(key.(string))
		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(map[uint64]*protobuf.DownLoadFileMsg)

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_DOWNLOAD_FILE,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_GetDownLoadFileDetail{
					GetDownLoadFileDetail: &protobuf.GetDownLoadFileDetail{
						DownloadFile: v,
					},
				},
			}

			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgSendCode() []error {
	tgSendCode := p.getSyncMap(TG_SEND_CODE)
	errs := make([]error, 0)
	var err error
	tgSendCode.Range(func(key, value any) bool {

		tgSendCode.Delete(key.(string))

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

			detail := value.(*protobuf.SendCodeAction)
			detail.Flag = true
			sendCode := detail.GetSendCode()
			if sendCode != nil {
				for account, _ := range sendCode {
					codeMap, getMapOK := p.Handler.tgLoginCodeSync.Load(strconv.FormatUint(account, 10))
					if getMapOK {
						codeMapValue := codeMap.(map[string]string)
						_, hasKey := codeMapValue[detail.GetLoginId()]
						if !hasKey {
							// 没有这个id，则认为这个请求过期
							p.Handler.tgLoginCodeSync.Delete(strconv.FormatUint(account, 10))
							detail.Flag = false
							detail.Comments = "Request expired, please obtain a new verification code..."
							level.Info(p.logger).Log("check code msg", "验证码请求ID不对，请重新登录获取ID")
						}
					} else {
						detail.Flag = false
						detail.Comments = "Request expired, please obtain a new verification code..."
						level.Info(p.logger).Log("check code msg", "验证码请求ID不对，请重新登录获取ID")
					}
				}
			}
			req := &protobuf.RequestMessage{
				Action: protobuf.Action_SEND_CODE,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_SendCodeDetail{
					SendCodeDetail: &protobuf.SendCodeDetail{
						Details: detail,
					},
				},
			}
			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgLogin() []error {

	tgLogin := p.getSyncMap(TG_LOGIN)
	errs := make([]error, 0)
	var err error
	tgLogin.Range(func(key, value any) bool {

		tgLogin.Delete(key.(string))

		p.targetAddr = key.(string)

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			detail := value.(map[uint64]*protobuf.LoginDetail)
			for k, v1 := range detail {
				// 每次登录时候，都把原来的tgLoginCodeSync账号对应的value删除
				p.Handler.tgLoginCodeSync.Delete(strconv.FormatUint(k, 10))

				loginId := uuid.New().String()
				codeIdMap := make(map[string]string)
				codeIdMap[loginId] = ""
				p.Handler.tgLoginCodeSync.Store(strconv.FormatUint(k, 10), codeIdMap)
				v1.LoginId = loginId
			}

			streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_LOGIN,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_OrdinaryAction{
					OrdinaryAction: &protobuf.OrdinaryAction{
						LoginDetail: detail,
					},
				},
			}

			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}
			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgSendPhotoMsg() []error {

	wsSendPhoneMsg := p.getSyncMap(TG_SEND_PHOTO)

	errs := make([]error, 0)
	var err error

	wsSendPhoneMsg.Range(func(key, value any) bool {

		wsSendPhoneMsg.Delete(key.(string))

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

			sendDatas := value.([]map[uint64]*protobuf.UintTgFileDetailValue)
			list := make([]*protobuf.SendFileAction, 0)

			for _, v1 := range sendDatas {
				for k2, v2 := range v1 {
					tmp := &protobuf.SendFileAction{}
					sendData := make(map[uint64]*protobuf.UintTgFileDetailValue)
					sendData[k2] = v2
					tmp.SendTgData = sendData
					list = append(list, tmp)
				}

			}

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_SEND_PHOTO,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_SendFileDetail{
					SendFileDetail: &protobuf.SendFileDetail{
						Details: list,
					},
				},
			}
			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}
func (p *ProxyServer) tgSyncAppInfo() []error {

	tgSync := p.getSyncMap(TG_SYNC_INFO)
	errs := make([]error, 0)
	var err error
	tgSync.Range(func(key, value any) bool {

		tgSync.Delete(key.(string))

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)
			req := &protobuf.RequestMessage{
				Action: protobuf.Action_SYNC_APP_INFO,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_SyncAppAction{
					SyncAppAction: &protobuf.SyncAppInfoAction{
						AppData: value.(map[uint64]*protobuf.AppData),
					},
				},
			}

			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}
			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgSendMsg() []error {
	tgSend := p.getSyncMap(TG_SEND_MSG)

	errs := make([]error, 0)
	var err error

	tgSend.Range(func(key, value any) bool {

		tgSend.Delete(key.(string))

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

			sendDatas := value.([]map[uint64]*protobuf.StringKeyStringvalue)
			list := make([]*protobuf.SendMessageAction, 0)

			for _, v1 := range sendDatas {
				for k2, v2 := range v1 {
					tmp := &protobuf.SendMessageAction{}
					sendData := make(map[uint64]*protobuf.StringKeyStringvalue)
					sendData[k2] = &protobuf.StringKeyStringvalue{Key: v2.Key, Values: v2.Values}
					tmp.SendTgData = sendData
					list = append(list, tmp)
				}

			}

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_SEND_MESSAGE,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_SendmessageDetail{
					SendmessageDetail: &protobuf.SendMessageDetail{
						Details: list,
					},
				},
			}

			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}
			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgCreateGroup() []error {
	tgCreateGroup := p.getSyncMap(TG_CREATE_GROUP)
	errs := make([]error, 0)
	var err error
	tgCreateGroup.Range(func(key, value any) bool {

		tgCreateGroup.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(map[uint64]*protobuf.CreateGroupDetail)

			for _, gd := range v {

				detail := &protobuf.UintkeyStringvalue{}
				detail.Key = gd.Detail.Key
				detail.Values = gd.Detail.Values

				req := &protobuf.RequestMessage{
					Action: protobuf.Action_CREATE_GROUP,
					Type:   "telegram",
					ActionDetail: &protobuf.RequestMessage_CreateGroupDetail{
						CreateGroupDetail: &protobuf.CreateGroupDetail{
							GroupName: gd.GroupName,
							Detail:    detail,
						},
					},
				}

				if err = streamConnectClient.Send(req); err != nil {
					level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
					errs = append(errs, err)
					return true
				}
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgSendGroupMsg() []error {
	tgSendGroupMsg := p.getSyncMap(TG_SEND_GROUP_MSG)

	errs := make([]error, 0)
	var err error
	tgSendGroupMsg.Range(func(key, value any) bool {

		tgSendGroupMsg.Delete(key.(string))

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

			sendDatas := value.([]map[uint64]*protobuf.StringKeyStringvalue)
			list := make([]*protobuf.SendGroupMessageAction, 0)
			for _, v1 := range sendDatas {
				for k2, v2 := range v1 {
					tmp := &protobuf.SendGroupMessageAction{}
					sendData := make(map[uint64]*protobuf.StringKeyStringvalue)
					sendData[k2] = &protobuf.StringKeyStringvalue{Key: v2.Key, Values: v2.Values}
					tmp.SendData = sendData
					list = append(list, tmp)
				}

			}

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_SEND_GROUP_MESSAGE,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_SendGroupMessageDetail{
					SendGroupMessageDetail: &protobuf.SendGroupMessageDetail{
						Details: list,
					},
				},
			}

			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}
			p.Handler.StreamMap.Store(key.(string), streamConnectClient)

		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgSyncContact() []error {
	tgSyncContact := p.getSyncMap(TG_SYNC_CONTACT)
	errs := make([]error, 0)
	var err error

	tgSyncContact.Range(func(key, value any) bool {

		tgSyncContact.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(map[uint64][]uint64)

			list := make([]*protobuf.UintkeyUintvalue, 0)
			for user, contactList := range v {
				tmp := &protobuf.UintkeyUintvalue{}

				tmp.Key = user
				tmp.Values = contactList
				list = append(list, tmp)
			}

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_SYNC_CONTACTS,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_SyncContactDetail{
					SyncContactDetail: &protobuf.SyncContactDetail{
						Details: list,
					},
				},
			}

			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}
			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgSendVcard() []error {
	tgSendVcardMsg := p.getSyncMap(TG_SEND_VCARD)

	errs := make([]error, 0)
	var err error

	tgSendVcardMsg.Range(func(key, value any) bool {
		tgSendVcardMsg.Delete(key.(string))

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

			sendDatas := value.([]map[uint64]*protobuf.UintSendContactCard)
			list := make([]*protobuf.SendContactCardAction, 0)

			for _, v1 := range sendDatas {
				for k2, v2 := range v1 {
					tmp := &protobuf.SendContactCardAction{}
					sendData := make(map[uint64]*protobuf.UintSendContactCard)
					sendData[k2] = v2
					tmp.SendData = sendData
					list = append(list, tmp)
				}

			}

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_SEND_VCARD_MESSAGE,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_SendContactCardDetail{
					SendContactCardDetail: &protobuf.SendContactCardDetail{
						Detail: list,
					},
				},
			}

			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}
			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) getDialogList() []error {
	tgSendVcardMsg := p.getSyncMap(TG_GET_DIALOG_LIST)
	errs := make([]error, 0)
	var err error
	tgSendVcardMsg.Range(func(key, value any) bool {

		tgSendVcardMsg.Delete(key.(string))

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

			sendDatas := value.(*protobuf.GetDialogList)

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_DIALOG_LIST,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_GetDialogList{
					GetDialogList: sendDatas,
				},
			}

			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) getContactList() []error {
	tgSendVcardMsg := p.getSyncMap(TG_GET_CONTACT_LIST)
	errs := make([]error, 0)
	var err error
	tgSendVcardMsg.Range(func(key, value any) bool {

		tgSendVcardMsg.Delete(key.(string))

		v, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v.(protobuf.ArthasStream_ConnectClient)

			sendDatas := value.(*protobuf.GetContactList)

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_CONTACT_LIST,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_GetContactList{
					GetContactList: sendDatas,
				},
			}

			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgAddGroupMember() []error {
	tgAddMember := p.getSyncMap(TG_ADD_GROUP_MEMBER)
	errs := make([]error, 0)
	var err error
	tgAddMember.Range(func(key, value any) bool {

		tgAddMember.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(map[uint64]*protobuf.AddGroupMemberDetail)

			for _, gd := range v {
				detail := &protobuf.UintkeyStringvalue{}
				detail.Key = gd.Detail.Key
				detail.Values = gd.Detail.Values

				req := &protobuf.RequestMessage{
					Action: protobuf.Action_ADD_GROUP_MEMBER,
					Type:   "telegram",
					ActionDetail: &protobuf.RequestMessage_AddGroupMemberDetail{
						AddGroupMemberDetail: &protobuf.AddGroupMemberDetail{
							GroupName: gd.GroupName,
							Detail:    detail,
						},
					},
				}

				if err = streamConnectClient.Send(req); err != nil {
					level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
					errs = append(errs, err)
					return true
				}
			}
			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgInviteToChannel() []error {
	tgInvite := p.getSyncMap(TG_INVITE_TO_CHANNEL)
	errs := make([]error, 0)
	var err error
	tgInvite.Range(func(key, value any) bool {

		tgInvite.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(map[uint64]*protobuf.InviteToChannelDetail)

			for _, id := range v {
				detail := &protobuf.UintkeyStringvalue{}
				detail.Key = id.Detail.Key
				detail.Values = id.Detail.Values

				req := &protobuf.RequestMessage{
					Action: protobuf.Action_INVITE_TO_CHANNEL,
					Type:   "telegram",
					ActionDetail: &protobuf.RequestMessage_InviteToChannelDetail{
						InviteToChannelDetail: &protobuf.InviteToChannelDetail{
							Channel: id.Channel,
							Detail:  detail,
						},
					},
				}

				if err = streamConnectClient.Send(req); err != nil {
					level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
					errs = append(errs, err)
					return true
				}
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgImportSession() []error {
	tgImport := p.getSyncMap(TG_IMPORT_SESSION)
	errs := make([]error, 0)
	var err error
	tgImport.Range(func(key, value any) bool {

		tgImport.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(map[uint64]*protobuf.ImportTgSessionMsg)

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_IMPORT_TG_SESSION,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_ImportTgSession{
					ImportTgSession: &protobuf.ImportTgSessionDetail{
						SendData: v,
					},
				},
			}

			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true

			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}
func (p *ProxyServer) tgGetGroupMembers() []error {
	getMembers := p.getSyncMap(TG_GET_GROUP_MEMBERS)

	errs := make([]error, 0)
	var err error
	getMembers.Range(func(key, value any) bool {

		getMembers.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(*protobuf.GetGroupMembersDetail)

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_GET_GROUP_MEMBERS,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_GetGroupMembersDetail{
					GetGroupMembersDetail: v,
				},
			}

			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgGetChannelMembers() []error {
	getMembers := p.getSyncMap(TG_GET_CHANNEL_MEMBERS)
	errs := make([]error, 0)
	var err error
	getMembers.Range(func(key, value any) bool {

		getMembers.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(*protobuf.GetChannelMemberDetail)

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_GET_CHANNEL_MEMBER,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_GetChannelMemberDetail{
					GetChannelMemberDetail: v,
				},
			}

			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgGetMsgHistory() []error {
	tgGetMsgHistory := p.getSyncMap(TG_GET_MSG_HISTORY)
	errs := make([]error, 0)
	var err error
	tgGetMsgHistory.Range(func(key, value any) bool {

		tgGetMsgHistory.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.([]*protobuf.GetMsgHistory)
			for _, gh := range v {
				req := &protobuf.RequestMessage{
					Action: protobuf.Action_Get_MSG_HISTORY,
					Type:   "telegram",
					ActionDetail: &protobuf.RequestMessage_GetMsgHistory{
						GetMsgHistory: &protobuf.GetMsgHistory{
							Self:  gh.Self,
							Other: gh.Other,
							Limit: gh.Limit,
						},
					},
				}

				if err = streamConnectClient.Send(req); err != nil {
					level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
					errs = append(errs, err)
					return true
				}
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgCreateChannel() []error {
	tgCreateChannel := p.getSyncMap(TG_CREATE_CHANNEL)
	errs := make([]error, 0)
	var err error
	tgCreateChannel.Range(func(key, value any) bool {

		tgCreateChannel.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(map[uint64]*protobuf.CreateChannelDetail)

			for _, cd := range v {

				detail := &protobuf.UintkeyStringvalue{}
				detail.Key = cd.Detail.Key
				detail.Values = cd.Detail.Values

				req := &protobuf.RequestMessage{
					Action: protobuf.Action_CREATE_CHANNEL,
					Type:   "telegram",
					ActionDetail: &protobuf.RequestMessage_CreateChannelDetail{
						CreateChannelDetail: &protobuf.CreateChannelDetail{
							ChannelUserName:    cd.ChannelUserName,
							ChannelDescription: cd.ChannelDescription,
							ChannelTitle:       cd.ChannelTitle,
							Detail:             detail,
							IsChannel:          cd.IsChannel,
							IsSuperGroup:       cd.IsSuperGroup,
						},
					},
				}
				if err = streamConnectClient.Send(req); err != nil {
					level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
					errs = append(errs, err)
					return true
				}
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgJoinByLink() []error {
	tgJoinByLink := p.getSyncMap(TG_JOIN_BY_LINK)

	errs := make([]error, 0)
	var err error
	tgJoinByLink.Range(func(key, value any) bool {

		tgJoinByLink.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(map[uint64]*protobuf.JoinByLinkDetail)

			for _, cd := range v {

				detail := &protobuf.UintkeyStringvalue{}
				detail.Key = cd.Detail.Key
				detail.Values = cd.Detail.Values

				req := &protobuf.RequestMessage{
					Action: protobuf.Action_JOIN_BY_LINK,
					Type:   "telegram",
					ActionDetail: &protobuf.RequestMessage_JoinByLinkDetail{
						JoinByLinkDetail: &protobuf.JoinByLinkDetail{
							Detail: detail,
						},
					},
				}
				if err = streamConnectClient.Send(req); err != nil {
					level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
					errs = append(errs, err)
					return true
				}
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgGetEmojiGroups() []error {
	tgGetEmoji := p.getSyncMap(TG_GET_EMOJI_GROUPS)
	errs := make([]error, 0)
	var err error
	tgGetEmoji.Range(func(key, value any) bool {

		tgGetEmoji.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(map[uint64]*protobuf.GetEmojiGroupsDetail)

			for _, cd := range v {

				req := &protobuf.RequestMessage{
					Action: protobuf.Action_GET_EMOJI_GROUP,
					Type:   "telegram",
					ActionDetail: &protobuf.RequestMessage_GetEmojiGroupDetail{
						GetEmojiGroupDetail: &protobuf.GetEmojiGroupsDetail{
							Sender: cd.Sender,
						},
					},
				}
				if err = streamConnectClient.Send(req); err != nil {
					level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
					errs = append(errs, err)
					return true
				}
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgMessagesReaction() []error {
	tgMsgReaction := p.getSyncMap(TG_MESSAGES_REACTION)
	errs := make([]error, 0)
	var err error
	tgMsgReaction.Range(func(key, value any) bool {

		tgMsgReaction.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(map[uint64]*protobuf.MessagesReactionDetail)

			for _, mr := range v {

				req := &protobuf.RequestMessage{
					Action: protobuf.Action_MESSAGES_REACTION,
					Type:   "telegram",
					ActionDetail: &protobuf.RequestMessage_MessagesReactionDetail{
						MessagesReactionDetail: &protobuf.MessagesReactionDetail{
							Emotion:  mr.Emotion,
							Detail:   mr.Detail,
							Receiver: mr.Receiver,
						},
					},
				}
				if err = streamConnectClient.Send(req); err != nil {
					level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
					errs = append(errs, err)
					return true
				}
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgLeave() []error {
	tgLeave := p.getSyncMap(TG_LEAVE)
	errs := make([]error, 0)
	var err error
	tgLeave.Range(func(key, value any) bool {

		tgLeave.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(map[uint64]*protobuf.LeaveDetail)

			for _, ld := range v {

				detail := &protobuf.UintkeyStringvalue{}
				detail.Key = ld.Detail.Key
				detail.Values = ld.Detail.Values

				req := &protobuf.RequestMessage{
					Action: protobuf.Action_LEAVE,
					Type:   "telegram",
					ActionDetail: &protobuf.RequestMessage_LeaveDetail{
						LeaveDetail: &protobuf.LeaveDetail{
							Detail: detail,
						},
					},
				}
				if err = streamConnectClient.Send(req); err != nil {
					level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
					errs = append(errs, err)
					return true
				}
			}

			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}

func (p *ProxyServer) tgGetOnlineAccounts() []error {
	tgOnline := p.getSyncMap(TG_GET_ONLINE_ACCOUNTS)
	errs := make([]error, 0)
	var err error
	tgOnline.Range(func(key, value any) bool {

		tgOnline.Delete(key.(string))

		v1, ok := p.Handler.StreamMap.Load(key.(string))
		if !ok {
			level.Info(p.logger).Log("msg", "get stream fail", "ip", key.(string))
			err = gerror.New("get stream fail,ip:" + key.(string))
			errs = append(errs, err)
		} else {
			streamConnectClient := v1.(protobuf.ArthasStream_ConnectClient)

			v := value.(*protobuf.GetOnlineAccountsDetail)

			req := &protobuf.RequestMessage{
				Action: protobuf.Action_GET_ONLINE_ACCOUNTS,
				Type:   "telegram",
				ActionDetail: &protobuf.RequestMessage_GetOnlineAccountsDetail{
					GetOnlineAccountsDetail: &protobuf.GetOnlineAccountsDetail{
						Phone: v.Phone,
					},
				},
			}
			if err = streamConnectClient.Send(req); err != nil {
				level.Error(p.logger).Log("msg", "streamConnectClient send", "err", err)
				errs = append(errs, err)
				return true
			}
			p.Handler.StreamMap.Store(key.(string), streamConnectClient)
		}

		return true
	})
	return errs
}
