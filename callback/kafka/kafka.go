package kafka

import (
	"arthas/callback"
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/go-kit/log/level"
	"github.com/gogf/gf/v2/encoding/gjson"
)

var Type = "kafka"

type (
	KManager struct {
		*callback.IManager
		addrs  []string //kafka addr
		config *sarama.Config
	}
)

func (m *KManager) ApplyConfig() (err error) {
	m.addrs = m.GetOptions().Addrs
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // 等待所有follower都回复ack，确保Kafka不会丢消息
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewHashPartitioner // 对Key进行Hash，同样的Key每次都落到一个分区，这样消息是有序的
	config.Producer.MaxMessageBytes = 512 * 1024 * 1024     //消息大小
	if m.GetOptions().Username != "" && m.GetOptions().Password != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = m.GetOptions().Username
		config.Net.SASL.Password = m.GetOptions().Password
	}
	config.Producer.Return.Successes = true
	m.config = config
	m.SetSendFunc(m.sendMessage)
	return nil
}

func (m *KManager) GetManager() *callback.IManager {
	return m.IManager
}

func (m *KManager) Run() error {
	level.Info(m.GetLogger()).Log("msg", "Running to callback kafka manager.")
	return m.GetManager().Run()
}

func (m *KManager) sendMessage(topic string, message interface{}) (partition int32, offset int64, err error) {
	// Create Kafka producer

	producer, err := sarama.NewSyncProducer(m.addrs, m.config)
	if err != nil {
		level.Error(m.GetLogger()).Log("msg", "Start Sarama producer failed.", "err", err)
		return -1, -1, err
	}

	// Check for errors and close connection at end
	defer func() {
		if err = producer.Close(); err != nil {
			level.Error(m.GetLogger()).Log("msg", "Close producer failed.", "err", err)
		}
	}()
	// Rest of code to produce messages...
	bytes, err := json.Marshal(message)
	if err != nil {
		level.Error(m.GetLogger()).Log("msg", "Json marshal failed.", "err", err)
		return -1, -1, err
	}
	msg := &sarama.ProducerMessage{Topic: topic, Key: sarama.StringEncoder(topic), Value: sarama.ByteEncoder(bytes)}
	partition, offset, err = producer.SendMessage(msg)
	if err != nil {
		level.Error(m.GetLogger()).Log("msg", "Send message failed.: ", "err", err)
		return -1, -1, err
	}
	level.Info(m.GetLogger()).Log("msg", fmt.Sprintf("> message sent to partition %d at offset %d\n data: %s", partition, offset, gjson.MustEncode(message)))
	return partition, offset, nil
}
