package telegram

import (
	"arthas/callback"
	"arthas/callback/consts"
	"github.com/gogf/gf/v2/encoding/gjson"
)

func (m *Manager) sendTgCardMsgCallback(tgCardMsgCallback []callback.TgCardMsgCallback) {
	backChan := callback.ImCallBackChan{
		Topic: consts.CallbackTopicTgSendContactCard,
		Callback: callback.ImCallback{
			Type: consts.CallbackTypeTg,
			Data: gjson.MustEncode(tgCardMsgCallback),
		},
	}
	m.callbackChan.ImCallbackChan <- backChan
}

func (m *Manager) sendTgMsgCallback(msgCallbacks []callback.TgMsgCallback) {
	backChan := callback.ImCallBackChan{
		Topic: consts.CallbackTopicTgMsg,
		Callback: callback.ImCallback{
			Type: consts.CallbackTypeTg,
			Data: gjson.MustEncode(msgCallbacks),
		},
	}
	m.callbackChan.ImCallbackChan <- backChan
}

func (m *Manager) sendGroupMsgCallback(groupMsgCallback []callback.GroupMsgCallback) {
	backChan := callback.ImCallBackChan{
		Topic: consts.CallbackTopicTgSendGroupMsg,
		Callback: callback.ImCallback{
			Type: consts.CallbackTypeTg,
			Data: gjson.MustEncode(groupMsgCallback),
		},
	}
	m.callbackChan.ImCallbackChan <- backChan
}

func (m *Manager) createGroupCallback(createGroupCallback callback.CreateGroupCallback) {
	backChan := callback.ImCallBackChan{
		Topic: consts.CallbackTopicTgCreateGroup,
		Callback: callback.ImCallback{
			Type: consts.CallbackTypeTg,
			Data: gjson.MustEncode(createGroupCallback),
		},
	}
	m.callbackChan.ImCallbackChan <- backChan
}

func (m *Manager) addGroupMemberCallback(addGroupMemberCallback []callback.AddGroupMemberCallback) {
	backChan := callback.ImCallBackChan{
		Topic: consts.CallbackTopicTgAddGroupMember,
		Callback: callback.ImCallback{
			Type: consts.CallbackTypeTg,
			Data: gjson.MustEncode(addGroupMemberCallback),
		},
	}
	m.callbackChan.ImCallbackChan <- backChan
}

func (m *Manager) receiverCallback(receiverCallback callback.ReceiverCallback) {
	backChan := callback.ImCallBackChan{
		Topic: consts.CallbackTopicTgReceiver,
		Callback: callback.ImCallback{
			Type: consts.CallbackTypeTg,
			Data: gjson.MustEncode(receiverCallback),
		},
	}
	m.callbackChan.ImCallbackChan <- backChan
}

func (m *Manager) loginMsgCallback(tgUserCallbacks []callback.TgUserLoginCallback) {
	backChan := callback.ImCallBackChan{
		Topic: consts.CallbackTopicTgLogin,
		Callback: callback.ImCallback{
			Type: consts.CallbackTypeTg,
			Data: gjson.MustEncode(tgUserCallbacks),
		},
	}
	m.callbackChan.ImCallbackChan <- backChan
}

func (m *Manager) logoutMsgCallback(tgUserCallback callback.TgUserLogoutCallback) {
	backChan := callback.ImCallBackChan{
		Topic: consts.CallbackTopicLogoutMSG,
		Callback: callback.ImCallback{
			Type: consts.CallbackTypeTg,
			Data: gjson.MustEncode(tgUserCallback),
		},
	}
	m.callbackChan.ImCallbackChan <- backChan
}

func (m *Manager) contactsSyncCallback(tgContactsCallbacks map[uint64][]callback.TgContactsCallback) {
	backChan := callback.ImCallBackChan{
		Topic: consts.CallbackTopicTgContactsSync,
		Callback: callback.ImCallback{
			Type: consts.CallbackTypeTg,
			Data: gjson.MustEncode(tgContactsCallbacks),
		},
	}
	m.callbackChan.ImCallbackChan <- backChan
}

func (m *Manager) dialogListCallback(tgDialogsCallbacks []DialogDetail) {
	backChan := callback.ImCallBackChan{
		Topic: consts.CallbackTopicTgDialogList,
		Callback: callback.ImCallback{
			Type: consts.CallbackTypeTg,
			Data: gjson.MustEncode(tgDialogsCallbacks),
		},
	}
	m.callbackChan.ImCallbackChan <- backChan
}

func (m *Manager) readMsgCallback(tgReadMsgCallbacks []callback.TgReadMsgCallback) {
	backChan := callback.ImCallBackChan{
		Topic: consts.CallbackTopicTgReadMsg,
		Callback: callback.ImCallback{
			Type: consts.CallbackTypeTg,
			Data: gjson.MustEncode(tgReadMsgCallbacks),
		},
	}
	m.callbackChan.ImCallbackChan <- backChan
}

func (m *Manager) getMsgHistoryCallback(tgGetMsgHistoryCallBack []callback.TgGetMsgHistoryCallBack) {
	backChan := callback.ImCallBackChan{
		Topic: consts.CallbackTopicTgGetHistory,
		Callback: callback.ImCallback{
			Type: consts.CallbackTypeTg,
			Data: gjson.MustEncode(tgGetMsgHistoryCallBack),
		},
	}
	m.callbackChan.ImCallbackChan <- backChan
}
