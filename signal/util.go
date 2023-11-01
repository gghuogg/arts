package signal

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

func getNextRequestId(r int64) string {
	timestamp := time.Now().Unix()

	result := fmt.Sprintf("%d-%d", timestamp, r)
	r++
	return result
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getRandomSessionId() int32 {
	return int32(1_0000_0000) + rand.Int31n(int32(9_0000_0000))
}

func getRandomSignedPreKeyId() uint32 {
	min := uint32(0x100000)
	max := uint32(0x200000)
	return min + uint32(rand.Intn(int(max-min)))
}

func getRequestId() string {
	// 获取当前时间的秒数，并转为16进制
	now := time.Now().Unix()
	nowHex := fmt.Sprintf("%X", now)

	// 生成长度为6的随机字节，并转为16进制
	randomBytes := make([]byte, 6)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	randomBytesHex := hex.EncodeToString(randomBytes)

	// 拼接两个16进制字符串，并替换字符
	result := strings.ToUpper(nowHex + randomBytesHex)
	result = strings.ReplaceAll(result, "C", "1")
	result = strings.ReplaceAll(result, "D", "2")
	result = strings.ReplaceAll(result, "E", "A")
	result = strings.ReplaceAll(result, "F", "B")

	return result
}

func Make(length int) []byte {
	random := make([]byte, length)
	_, err := rand.Read(random)
	if err != nil {
		panic(fmt.Errorf("failed to get random bytes: %w", err))
	}
	return random
}

func bytesWithPadding(plaintext []byte) []byte {
	pad := Make(1)
	pad[0] &= 0xf
	if pad[0] == 0 {
		pad[0] = 0xf
	}
	plaintext = append(plaintext, bytes.Repeat(pad[:], int(pad[0]))...)
	return plaintext
}

//	func padStart(s string, l int, pad string) string {
//		if len(s) >= l {
//			return s
//		}
//		return strings.Repeat(pad, l-len(s)) + s
//	}
func padStart(input string, length int, padChar string) string {
	if len(input) >= length {
		return input
	}

	padLength := length - len(input)
	padStr := strings.Repeat(padChar, padLength)
	return padStr + input
}

func uint32ToBytes(value uint32) []byte {
	byteSlice := make([]byte, 4)
	binary.BigEndian.PutUint32(byteSlice, value)
	return byteSlice
}

func RandomInt64() int64 {
	rand.Seed(time.Now().UnixNano())
	return rand.Int63()
}

func uint64FromPtr(ptr *uint64) uint64 {
	return *ptr
}

func stringFromPtr(ptr *string) string {
	if ptr == nil {
		return ""
	} else {
		return *ptr
	}
}
