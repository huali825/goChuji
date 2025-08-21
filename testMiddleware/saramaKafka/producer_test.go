package saramaKafka

import (
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"testing"
)

// var addr = []string{"localhost:9094"}
var addr = []string{"81.71.139.129:9094"}

func TestSyncProducer(t *testing.T) {
	// 创建一个新的Kafka配置对象
	cfg := sarama.NewConfig()
	// 设置生产者配置，当消息成功发送到服务器时返回确认
	cfg.Producer.Return.Successes = true
	// 使用同步模式创建生产者，并传入地址和配置
	producer, err := sarama.NewSyncProducer(addr, cfg)
	// 设置分区器为轮询分区器，确保消息均匀分布在各个分区
	cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	//cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	//cfg.Producer.Partitioner = sarama.NewHashPartitioner
	//cfg.Producer.Partitioner = sarama.NewManualPartitioner
	//cfg.Producer.Partitioner = sarama.NewConsistentCRCHashPartitioner
	//cfg.Producer.Partitioner = sarama.NewCustomPartitioner()
	assert.NoError(t, err)
	for i := 0; i < 100; i++ {
		_, _, err = producer.SendMessage(&sarama.ProducerMessage{
			Topic: "test_topic",
			Value: sarama.StringEncoder("这是一条消息"),
			// 会在生产者和消费者之间传递的
			Headers: []sarama.RecordHeader{
				{
					Key:   []byte("key1"),
					Value: []byte("value1"),
				},
			},
			Metadata: "这是 metadata",
		})
	}

}

// TestAsyncProducer 是一个测试异步生产者的函数
func TestAsyncProducer(t *testing.T) {
	// 创建一个新的 Kafka 配置
	cfg := sarama.NewConfig()
	// 设置生产者配置，表示需要返回成功消息
	cfg.Producer.Return.Successes = true
	// 设置生产者配置，表示需要返回错误消息
	cfg.Producer.Return.Errors = true
	// 使用配置创建异步生产者
	producer, err := sarama.NewAsyncProducer(addr, cfg)
	// 断言没有错误发生
	assert.NoError(t, err)
	// 获取生产者的输入通道
	msgs := producer.Input()
	// 向通道发送一条消息
	msgs <- &sarama.ProducerMessage{
		Topic: "test_topic",                   // 消息发送的主题
		Value: sarama.StringEncoder("这是一条消息"), // 消息内容
		// 会在生产者和消费者之间传递的头部信息
		Headers: []sarama.RecordHeader{
			{
				Key:   []byte("key1"),   // 头部键
				Value: []byte("value1"), // 头部值
			},
		},
		Metadata: "这是 metadata", // 消息的元数据
	}

	// 使用 select 语句等待生产者的成功或错误响应
	select {
	case msg := <-producer.Successes(): // 接收成功消息
		t.Log("发送成功", string(msg.Value.(sarama.StringEncoder)))
	case err := <-producer.Errors(): // 接收错误消息
		t.Log("发送失败", err.Err, err.Msg)
	}
}
