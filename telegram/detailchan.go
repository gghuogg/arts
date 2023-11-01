package telegram

import "sync"

func (m *Manager) AccountsSync() *sync.Map {
	return m.accountsSync
}

func (m *Manager) AccountDetailChan() chan AccountDetail {
	return m.accountDetailChan
}

func (m *Manager) LoginDetailChan() chan LoginDetail {
	return m.loginDetailChan
}

func (m *Manager) SyncContactChan() chan SyncContact {
	return m.syncContactChan
}

func (m *Manager) SendMessageChan() chan SendMessageDetail {
	return m.sendMessageChan
}

func (m *Manager) SendPhotoChan() chan SendImageFileDetail {
	return m.sendPhotoChan
}

func (m *Manager) SendFileChan() chan SendFileDetail {
	return m.sendFileChan
}

func (m *Manager) CreateGroupChan() chan CreateGroupDetail {
	return m.createGroupChan
}

func (m *Manager) AddGroupMemberChan() chan AddGroupMemberDetail {
	return m.addGroupMemberChan
}

func (m *Manager) GetGroupMembersChan() chan GetGroupMembers {
	return m.getGroupMembersChan
}

func (m *Manager) SendGroupMessageChan() chan SendGroupMessageDetail {
	return m.sendGroupMessageChan
}

func (m *Manager) SendContactCardChan() chan ContactCardDetail {
	return m.senContactCardChan
}

func (m *Manager) SendCodeChan() chan SendCodeDetail {
	return m.sendCodeChan
}

func (m *Manager) ContatcListChan() chan ContactListDetail {
	return m.getContactListChan
}

func (m *Manager) DialogListChan() chan DialogListDetail {
	return m.getDialogListChan
}

func (m *Manager) GetMessageHistoryChan() chan GetHistoryDetail {
	return m.getHistoryDetailChan
}

func (m *Manager) TgLogoutChan() chan TgLogoutDetail {
	return m.tgLogoutChan
}

func (m *Manager) TgSendVideoChan() chan SendVideoDetail {
	return m.sendVideoChan
}

func (m *Manager) TgGetDownloadFileChan() chan DownChatFileDetail {
	return m.getDownloadChan
}

func (m *Manager) TgCreateChannelChan() chan CreateChannelDetail {
	return m.createChannelDetailChan
}

func (m *Manager) TgInviteToChannelChan() chan InviteToChannelDetail {
	return m.inviteToChannelDetailChan
}

func (m *Manager) TgImportSessionChan() chan ImportTgSessionMsg {
	return m.importTgSessionMsgChan
}

func (m *Manager) TgJoinByLinkChan() chan JoinByLinkDetail {
	return m.JoinByLinkDetailChan
}

func (m *Manager) TgGetEmojiGroupChan() chan GetEmojiGroupsDetail {
	return m.GetEmojiGroupsDetailChan
}

func (m *Manager) TgMessagesReactionChan() chan MessageReactionDetail {
	return m.MessageReactionDetailChan
}

func (m *Manager) TgLeaveChan() chan LeaveDetail {
	return m.LeaveDetailChan
}

func (m *Manager) TgGetChannelMemberChan() chan GetChannelMemberDetail {
	return m.GetChannelMemberDetailChan
}

func (m *Manager) TgGetOnlineAccountsChan() chan GetOnlineAccountDetail {
	return m.GetOnlineAccountDetailChan
}
