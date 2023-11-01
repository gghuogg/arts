package rpc

import (
	"arthas/callback"
	"arthas/etcd"
	"arthas/signal"
	"arthas/telegram"
	"sync"
)

func (h *Handler) EtcdGetChan() chan etcd.GetReq {
	return h.getChan
}

func (h *Handler) GetTgAccountsSync() *sync.Map {
	return h.tgAccountsSync
}

func (h *Handler) GetTgLoginDetailChan() chan telegram.LoginDetail {
	return h.tgLoginDetailChan
}

func (h *Handler) GetCallbackChan() *callback.CallbackChan {
	return h.callbackChan
}

func (h *Handler) GetAccountsSync() *sync.Map {
	return h.accountsSync
}

func (h *Handler) GetBlockedAccountSet() map[uint64]struct{} {
	return h.blockedAccountSet
}

func (h *Handler) GetLoginDetailChan() chan signal.LoginDetail {
	return h.loginDetailChan
}

func (h *Handler) GetTgAccountDetailChan() chan telegram.AccountDetail {
	return h.tgAccountDetailChan
}

func (h *Handler) GetLogoutDetailChan() chan signal.LogoutDetail {
	return h.logoutDetailChan
}

func (h *Handler) GetTgLogoutChan() chan telegram.TgLogoutDetail {
	return h.tgLogoutChan
}

func (h *Handler) GetTgSendMessageChan() chan telegram.SendMessageDetail {
	return h.tgSendMessageChan
}

func (h *Handler) GetWhatsSendMessageChan() chan signal.SendMessageDetail {
	return h.sendMessageChan
}

func (h *Handler) GetTgSendGroupMessageChan() chan telegram.SendGroupMessageDetail {
	return h.tgSendGroupMessageChan
}

func (h *Handler) GetWhatsSyncAccountKeyChan() chan signal.AccountDetail {
	return h.accountDetailChan
}

func (h *Handler) GetTgSyncContactsChan() chan telegram.SyncContact {
	return h.tgSyncContactChan
}

func (h *Handler) GetWhatsSyncContactsChan() chan signal.SyncContact {
	return h.syncContactChan
}

func (h *Handler) GetWhatsGetUserHeadImageChan() chan signal.GetUserHeadImageDetail {
	return h.GetUserHeadImageChan
}

func (h *Handler) GetTgCreateGroupChan() chan telegram.CreateGroupDetail {
	return h.tgCreateGroupChan
}

func (h *Handler) GetMessagesReactionChan() chan telegram.MessageReactionDetail {
	return h.tgMessagesReactionChan
}

func (h *Handler) GetTgCreateChannel() chan telegram.CreateChannelDetail {
	return h.tgCreateChannelChan
}

func (h *Handler) GetTgAddGroupMemberChan() chan telegram.AddGroupMemberDetail {
	return h.tgAddGroupMemberChan
}

func (h *Handler) GetTgInviteToChannelChan() chan telegram.InviteToChannelDetail {
	return h.tgInviteToChannelChan
}

func (h *Handler) GetTgJoinByLinkChan() chan telegram.JoinByLinkDetail {
	return h.tgJoinByLinkChan
}

func (h *Handler) GetTgLeaveChan() chan telegram.LeaveDetail {
	return h.tgLeaveChan
}

func (h *Handler) GetTgGetGroupMembersChan() chan telegram.GetGroupMembers {
	return h.tgGetGroupMembers
}

func (h *Handler) GetTgGetOnlineAccountsChan() chan telegram.GetOnlineAccountDetail {
	return h.tgGetOnlineAccountsChan
}

func (h *Handler) GetTgGetChannelMembersChan() chan telegram.GetChannelMemberDetail {
	return h.tgGetChannelMemberChan
}

func (h *Handler) GetTgSendPhotoChan() chan telegram.SendImageFileDetail {
	return h.tgSendPhotoChan
}

func (h *Handler) GetTgSendVideoChan() chan telegram.SendVideoDetail {
	return h.tgSendVideoChan
}

func (h *Handler) GetWhatsSendPhotoChan() chan signal.SendImageFileDetail {
	return h.sendImageFileChan
}

func (h *Handler) GetWhatsSendVideoChan() chan signal.SendVideoFileDetail {
	return h.sendVideoChan
}

func (h *Handler) GetTgSendFileChan() chan telegram.SendFileDetail {
	return h.tgSendFileChan
}

func (h *Handler) GetWhatsSendFileChan() chan signal.SendFileDetail {
	return h.sendFileChan
}

func (h *Handler) GetTgSendCodeChan() chan telegram.SendCodeDetail {
	return h.tgSendCodeChan
}

func (h *Handler) GetTgGetContactListChan() chan telegram.ContactListDetail {
	return h.tgGetContactListChan
}

func (h *Handler) GetTgGetDialogListChan() chan telegram.DialogListDetail {
	return h.tgGetDialogListChan
}

func (h *Handler) GetTgMsgHistoryChan() chan telegram.GetHistoryDetail {
	return h.tgGetMsgHistoryChan
}

func (h *Handler) GetTgSendVcardMsgChan() chan telegram.ContactCardDetail {
	return h.tgSendContactCardChan
}

func (h *Handler) GetWhatsSendVcardMsgChan() chan signal.SendVcardMessageDetail {
	return h.sendVCardMessageChan
}

func (h *Handler) GetTgDownloadFileChan() chan telegram.DownChatFileDetail {
	return h.tgDownloadFileChan
}

func (h *Handler) GeTgImportSessionChan() chan telegram.ImportTgSessionMsg {
	return h.tgImportSessionMsgChan
}

func (h *Handler) GetTgGetEmojiGroupChan() chan telegram.GetEmojiGroupsDetail {
	return h.tgGetEmojiGroupsChan
}

func (h *Handler) GetServerMap() *sync.Map {
	return &h.serverMap
}
