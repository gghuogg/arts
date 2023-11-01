package telegram

import (
	"arthas/callback"
	"context"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
)

func (m *Manager) getPeer(ctx context.Context, c *telegram.Client, search string) (peer peers.Peer, err error) {
	manager := peers.Options{Storage: storage.NewPeers(m.kvd)}.Build(c.API())
	peer, err = utils.Telegram.GetInputPeer(ctx, manager, search)
	if err == nil {
		return
	}

	// 发生错误说明没有存储到etcd，可能是新的会话，此时从好友列表查找
	contactsSearch, err := c.API().ContactsSearch(ctx, &tg.ContactsSearchRequest{
		Q:     search,
		Limit: 1,
	})
	if err != nil {
		_ = level.Error(m.logger).Log("msg", "get contacts fail", err)
		return
	}
	_ = manager.Apply(ctx, contactsSearch.GetUsers(), contactsSearch.GetChats())
	for _, user := range contactsSearch.Users {
		t, b := user.AsNotEmpty()
		if !b {
			continue
		}
		if t.Username == search || gconv.String(t.ID) == search || t.Phone == search {
			peer, err = manager.GetUser(ctx, t.AsInput())
			if tgerr.Is(err, tg.ErrUserIDInvalid) {
				return peers.User{}, &peers.PeerNotFoundError{
					Peer: &tg.PeerUser{UserID: t.ID},
				}
			}
			return peer, err
		}
	}
	//联系人找不到时，尝试添加好友
	if peer == nil {
		if gstr.IsNumeric(search) {
			importContacts, err := c.API().ContactsImportContacts(ctx, []tg.InputPhoneContact{{Phone: search}})
			if err != nil {
				level.Error(m.logger).Log("msg", "import contact fail", "err", err)
				return nil, err
			}
			for _, user := range importContacts.Users {
				t, b := user.AsNotEmpty()
				if !b {
					continue
				}
				if t.Username == search || gconv.String(t.ID) == search || t.Phone == search {
					peer, err = manager.GetUser(ctx, t.AsInput())
					if tgerr.Is(err, tg.ErrUserIDInvalid) {
						return peers.User{}, &peers.PeerNotFoundError{
							Peer: &tg.PeerUser{UserID: t.ID},
						}
					}
					syncCallbackMap := make(map[uint64][]callback.TgContactsCallback)
					self, _ := c.Self(ctx)
					item := callback.TgContactsCallback{
						TgId:  peer.ID(),
						Phone: search,
						Type:  1,
					}
					if username, ok := peer.Username(); ok {
						item.Username = username
					}
					syncCallbackMap[gconv.Uint64(self.Phone)] = []callback.TgContactsCallback{}
					m.contactsSyncCallback(syncCallbackMap)
					return peer, err
				}
			}
		}
	}
	if peer == nil {
		err = gerror.New("No contact found.")
	}
	return
}
