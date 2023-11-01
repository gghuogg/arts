package consumer

import (
	"arthas/callback"
	"github.com/IBM/sarama"
	"github.com/go-kit/log/level"
	"golang.org/x/net/context"
)

var Type = "consumer"

type (
	SManager struct {
		*callback.IManager
		addrs               []string //kafka addr
		gradeId             string
		context             context.Context
		config              *sarama.Config
		lastConsumedOffsets map[string]map[int32]int64
	}
)

func (m *SManager) ApplyConfig() (err error) {
	m.addrs = m.GetOptions().Addrs
	m.gradeId = m.GetOptions().GradeId
	m.context = m.GetCtx()
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // 等待所有follower都回复ack，确保Kafka不会丢消息
	config.Producer.Return.Successes = true
	config.Producer.Partitioner = sarama.NewHashPartitioner // 对Key进行Hash，同样的Key每次都落到一个分区，这样消息是有序的

	if m.GetOptions().Username != "" && m.GetOptions().Password != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = m.GetOptions().Username
		config.Net.SASL.Password = m.GetOptions().Password
	}
	config.Producer.Return.Successes = true
	m.config = config

	cGroup, err := sarama.NewConsumerGroup(m.addrs, m.gradeId, nil)
	if err != nil {
		panic(err)
	}

	go func() {
		for err := range cGroup.Errors() {
			level.Error(m.GetLogger()).Log("msg", "kafka consumer group fail", "err", err)
		}
	}()

	go func() {
		for {
			//err = cGroup.Consume(m.GetCtx(), []string{"whatsappRequestMessage", "telegramRequestMessage"}, m)
			err = cGroup.Consume(m.GetCtx(), []string{"RequestMessage1"}, m)

			if err != nil {
				level.Error(m.GetLogger()).Log("msg", "kafka new consumer fail", "err", err)
				_ = cGroup.Close()
				break
			}
		}
	}()
	return nil
}

func (m *SManager) GetManager() *callback.IManager {
	return m.IManager
}

func (m *SManager) Run() error {
	level.Info(m.GetLogger()).Log("msg", "Running to consumer kafka manager.")

	return nil
}
