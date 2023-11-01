package sendUtils

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

func SendGetRequestToPrometheus(path string, params map[string]string) error {
	baseUrl := "http://localhost:8080"
	u, _ := url.Parse(baseUrl)
	// 生成对应url
	u.Path = path
	q := u.Query()
	if params != nil {
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
	}

	return SendGetRequest(u.String())
}

// SendGetRequest 发送一个GET请求到指定的URL
func SendGetRequest(url string) error {
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Error sending GET request to Prothemus: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

// SendCalculationTimeToPrometheus 计算API运行时间
func SendCalculationTimeToPrometheus(startTime time.Time, name string, path string) error {
	endTime := time.Now()
	execTime := endTime.Sub(startTime)
	timeMap := make(map[string]string)
	timeMap["time"] = strconv.FormatInt(execTime.Microseconds(), 10)
	timeMap["msg"] = name

	err := SendGetRequestToPrometheus(path, timeMap)
	if err != nil {
		return err
	}
	return nil
}
