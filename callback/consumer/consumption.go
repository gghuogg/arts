package consumer

import (
	"arthas/callback"
	"arthas/protobuf"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/go-kit/log/level"
	"google.golang.org/protobuf/proto"
)

// 消费kafka信息
func (m *SManager) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (m *SManager) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (m *SManager) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) (err error) {
	for msg := range claim.Messages() {

		switch msg.Topic {
		case "RequestMessage1":
			_, in, err := m.getRequestMsgToProxy(msg)
			if err != nil {
				panic(err.Error())
				continue
			}
			sendRequest := callback.SendRequestMsgToProxy{
				RequestMsg: in,
			}
			// 防止初始化消费阻塞
			fmt.Println(sendRequest)
			m.ConsumeChan().SendRequestMsgToProxyChan <- sendRequest

		}
		if err == nil {
			// 标记，sarama会自动进行提交，默认间隔1秒
			sess.MarkMessage(msg, "")
		}
	}

	return err
}

func (m *SManager) getRequestMsgToProxy(msg *sarama.ConsumerMessage) (uniqueId string, in *protobuf.RequestMessage, error error) {

	callbackRes := make([]callback.PushRequestMsgCallbackChan, 0)
	err := json.Unmarshal(msg.Value, &callbackRes)
	if err != nil {
		return "", nil, err
	}
	for _, callback := range callbackRes {
		//proxyServer := callback.ProxyServer

		var in protobuf.RequestMessage
		err := proto.Unmarshal(callback.RequestMessage, &in)
		if err != nil {
			level.Error(m.GetLogger()).Log("msg", "RequestMsgToProxy analyze protof fail", "err", err)
			return "", nil, err
		}

		uniqueId := callback.UniqueId
		return uniqueId, &in, nil
	}

	return "", nil, err
}
