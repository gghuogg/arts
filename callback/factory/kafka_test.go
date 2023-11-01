package factory

import (
	"arthas/callback"
	"arthas/callback/kafka"
	alog "arthas/log"
	"context"
	"github.com/go-kit/log"
	"testing"
	"time"
)

func TestKafka(t *testing.T) {
	logger := alog.New(&alog.Config{})
	ctxCallback := context.Background()
	options := callback.Options{
		Type: kafka.Type,
		//Addrs: []string{"192.168.3.8:9192", "192.168.3.8:9292", "192.168.3.8:9392"},
		Addrs: []string{"101.35.51.132:30092"},
	}
	callbackManager := New(log.With(logger, "component", "callback"), ctxCallback, &options)
	err := callbackManager.ApplyConfig()
	if err != nil {
		return
	}
	go func() {
		err := callbackManager.Run()
		if err != nil {
			return
		}
	}()
	callbackChan := callbackManager.GetManager().CallbackChan().LoginCallbackChan
	loginCallbacks := make([]callback.LoginCallback, 0)
	loginCallbacks = append(loginCallbacks, callback.LoginCallback{
		UserJid:     2,
		LoginStatus: 1,
		ProxyUrl:    "2",
	})
	callbackChan <- loginCallbacks

	time.Sleep(20 * time.Second)
}
