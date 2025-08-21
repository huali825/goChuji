package saramaKafka

import (
	"context"
	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
	"log"
	"testing"
	"time"
)

// TestConsumer 是一个测试函数，用于测试 Kafka 消费者的功能
func TestConsumer(t *testing.T) {
	// 创建 Kafka 配置对象
	cfg := sarama.NewConfig()
	// 创建 Kafka 消费者组，地址为 addr，组名为 "demo"
	consumer, err := sarama.NewConsumerGroup(addr, "demo", cfg)
	assert.NoError(t, err) // 断言创建消费者组时没有错误

	// 创建一个带有 10 分钟超时的上下文，用于控制消费者生命周期
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute*10)
	defer cancel() // 确保在函数返回前取消上下文

	// 记录开始时间，用于后续计算消费耗时
	start := time.Now()

	// 开始消费消息，主题为 "test_topic"，使用 ConsumerHandler 处理消息
	err = consumer.Consume(ctx,
		[]string{"test_topic"}, ConsumerHandler{})
	assert.NoError(t, err) // 断言消费过程中没有错误
	// 输出消费过程耗时
	t.Log(time.Since(start).String())
}

type ConsumerHandler struct {
}

func (c ConsumerHandler) Setup(session sarama.ConsumerGroupSession) error {
	log.Println("这是 Setup")
	//partitions := session.Claims()["test_topic"]
	//for _, part := range partitions {
	//	session.ResetOffset("test_topic",
	//		part, sarama.OffsetOldest, "")
	//}
	return nil
}

func (c ConsumerHandler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Println("这是 Cleanup")
	return nil
}

// ConsumeClaim 是 ConsumerHandler 接口的方法，用于处理消息声明的消费
// 它会批量处理消息，并使用错误组进行并发处理
func (c ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	// 从声明中获取消息通道
	msgs := claim.Messages()
	// 定义批处理大小为10条消息
	const batchSize = 10
	for {
		// 标记一个批次的开始
		log.Println("一个批次开始")
		// 创建一个容量为batchSize的切片用于存储消息批次
		batch := make([]*sarama.ConsumerMessage, 0, batchSize)
		// 创建一个1秒超时的上下文
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		var done = false
		// 创建错误组用于并发处理消息
		var eg errgroup.Group
		// 循环收集消息，直到达到批次大小或超时
		for i := 0; i < batchSize && !done; i++ {
			select {
			case <-ctx.Done():
				// 超时了
				done = true
			case msg, ok := <-msgs:
				if !ok {
					cancel()
					return nil
				}
				batch = append(batch, msg)
				eg.Go(func() error {
					// 并发处理
					log.Println(string(msg.Value))
					return nil
				})
			}
		}
		cancel()
		err := eg.Wait()
		if err != nil {
			log.Println(err)
			continue
		}
		// 凑够了一批，然后你就处理
		// log.Println(batch)

		for _, msg := range batch {
			session.MarkMessage(msg, "")
		}
	}
}

func (c ConsumerHandler) ConsumeClaimV1(session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim) error {
	msgs := claim.Messages()
	for msg := range msgs {
		log.Println(string(msg.Value))
		session.MarkMessage(msg, "")
	}
	return nil
}
