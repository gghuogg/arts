package signal

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

//type phoneDetil struct {
//	inPhone         uint64[]
//	outPhone        uint64[]
//	phoneDetilChain chan uint64
//}

type phoneContext struct {
	inPhone  []string
	outPhone []string
}

var pct = phoneContext{
	inPhone:  []string{},
	outPhone: []string{},
}

var mu sync.Mutex
var lastCall time.Time

func (a *Account) givenSyncContact(pn []uint64) {
	//// 检查上一次调用时间与当前时间的差值
	elapsed := time.Since(lastCall)
	if elapsed < 30*time.Minute {
		fmt.Println("30分钟内只能调用一次")
		return
	}

	// 调用前把数组清空
	pct.inPhone = []string{}
	pct.outPhone = []string{}
	go func(pn []uint64) {
		if len(pn) > 1000 {
			a.sendPhone(getRandomInformation(pn, 900))
		} else {
			a.sendPhone(pn)
		}
	}(pn)
	//2.发送完所有号码之后，再发送一个信号，通知服务端，把数组的号码推送到kafka
	// 推送到kafka（推送之后把数组清0，数组记得上锁）
	//timer := time.Timer{}

	lastCall = time.Now()
}

// 随机数组
func getRandomInformation(allInformation []uint64, count int) []uint64 {
	if count >= len(allInformation) {
		return allInformation
	}

	rand.Seed(time.Now().UnixNano()) // 设置随机数种子，以当前时间作为种子

	// Fisher-Yates shuffle 算法用于随机打乱数组
	shuffledInformation := make([]uint64, len(allInformation))
	copy(shuffledInformation, allInformation)

	for i := len(shuffledInformation) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffledInformation[i], shuffledInformation[j] = shuffledInformation[j], shuffledInformation[i]
	}

	// 获取前count个信息
	return shuffledInformation[:count]
}

func (a *Account) sendPhone(pn []uint64) {
	for _, phoneNumber := range pn {
		msg := a.signal.SyncContactRequest("18165314688@s.whatsapp.net", []uint64{phoneNumber}, "1689755731-54726411-1")
		a.send(msg)
	}
}

func (account *Account) sendInPhone(phone string) {
	pct.inPhone = append(pct.inPhone, phone)
	fmt.Printf("我的in数组%v\n", pct.inPhone)

}

func (account *Account) sendOutPhone(phone string) {
	pct.outPhone = append(pct.outPhone, phone)
	fmt.Printf("我的out数组%v\n", pct.outPhone)

}
