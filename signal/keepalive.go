package signal

import (
	"context"
	"fmt"
	"go.mau.fi/util/random"
	"math/rand"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	// KeepAliveResponseDeadline specifies the duration to wait for a response to websocket keepalive pings.
	KeepAliveResponseDeadline = 10 * time.Second
	// KeepAliveIntervalMin specifies the minimum interval for websocket keepalive pings.
	KeepAliveIntervalMin = 20 * time.Second
	// KeepAliveIntervalMax specifies the maximum interval for websocket keepalive pings.
	KeepAliveIntervalMax = 30 * time.Second

	// KeepAliveMaxFailTime specifies the maximum time to wait before forcing a reconnect if keepalives fail repeatedly.
	KeepAliveMaxFailTime = 3 * time.Minute
)

func (a *Account) keepAliveLoop(ctx context.Context) {
	uniqueIDPrefix := random.Bytes(2)
	a.uniqueID = fmt.Sprintf("%d.%d-", uniqueIDPrefix[0], uniqueIDPrefix[1])
	reqId := a.uniqueID + strconv.FormatUint(uint64(atomic.AddUint32(&a.idCounter, 1)), 10)
	for {
		interval := rand.Int63n(KeepAliveIntervalMax.Milliseconds()-KeepAliveIntervalMin.Milliseconds()) + KeepAliveIntervalMin.Milliseconds()
		select {
		case <-time.After(time.Duration(interval) * time.Millisecond):
			a.sendKeepAlive(reqId)
		case <-ctx.Done():
			return
		}
	}
}

func (a *Account) sendKeepAlive(reqId string) {
	a.ping(reqId)
}
