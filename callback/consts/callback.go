package consts

// im type
const (
	CallbackTypeWhats = 1
	CallbackTypeTg    = 2
)

// tg topic
const (
	CallbackTopicTgLogin = "tgLogin"
	CallbackTopicTgMsg   = "tgMsg"

	CallbackTopicTgCreateGroup    = "tgCreateGroup"
	CallbackTopicTgAddGroupMember = "tgAddGroupMember"

	CallbackTopicTgReceiver = "tgReceiver"

	CallbackTopicTgSendGroupMsg    = "tgSendGroupMsg"
	CallbackTopicTgSendContactCard = "tgSendContactCard"
	CallbackTopicTgSendFileMsg     = "tgSendFileMsg"
	CallbackTopicTgContactsSync    = "tgContactsSync"
	CallbackTopicTgContactList     = "tgContactList"
	CallbackTopicTgDialogList      = "tgDialogList"
	CallbackTopicTgReadMsg         = "tgReadMsg"
	CallbackTopicArtsServerDown    = "artsServerDown"
	CallbackTopicLogoutMSG         = "tgLogoutMsg"
	CallbackTopicTgGetHistory      = "tgGetHistory"
)
