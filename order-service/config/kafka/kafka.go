package kafka

import (
	"fmt"
	"order-service/config"

	"github.com/segmentio/kafka-go"
)

func NewKafkaWriter(cfg *config.Config, topic string) *kafka.Writer {
	kafkaConfig := cfg.Kafka
	broker := fmt.Sprintf("%s:%s", kafkaConfig.Host, kafkaConfig.Port)
	return &kafka.Writer{
		Addr:                   kafka.TCP(broker),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{}, // Balancer for selecting partition
		AllowAutoTopicCreation: true,
	}
}
