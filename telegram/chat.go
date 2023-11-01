package telegram

import (
	"context"
	"github.com/gotd/td/telegram"
)

func (m *Manager) SecretChat(ctx context.Context, c *telegram.Client) {
	//
	//c.API().MessagesGetDhConfig(ctx, &tg.MessagesGetDhConfigRequest{})
	//
	//c.API().MessagesRequestEncryption()
}
