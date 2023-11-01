package signal

import (
	"fmt"
	"sync"
	"time"
)

type Request struct {
	currentTimestamp int64
	requestId        int64
	mux              sync.Mutex
}

func (r *Request) getNextRequestId() string {
	r.mux.Lock()
	defer r.mux.Unlock()

	timestamp := time.Now().Unix()
	if r.currentTimestamp != timestamp {
		r.requestId = 0
	}
	r.currentTimestamp = timestamp
	result := fmt.Sprintf("%d-%d", timestamp, r.requestId)
	r.requestId++
	return result
}
