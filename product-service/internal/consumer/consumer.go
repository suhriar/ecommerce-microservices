package consumer

import (
	"context"
	"encoding/json"
	"product-service/domain"
	"product-service/internal/usecase"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

type Consumer struct {
	productUsecase usecase.ProductUsecase
}

func NewConsumer(productUsecase usecase.ProductUsecase) *Consumer {
	return &Consumer{productUsecase: productUsecase}
}

// StartKafkaConsumer starts a Kafka consumer to listen for order events
func (c *Consumer) StartKafkaConsumer() {
	// Create Kafka reader for order topic
	orderReader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{"localhost:9092"},
		Topic:    "order-topic",
		GroupID:  "product-service-group",
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})

	for {
		// Read message from order topic
		ctx := context.Background()
		msg, err := orderReader.ReadMessage(ctx)
		if err != nil {
			log.Error().Msgf("Error reading message: %v", err)
			continue
		}

		// Process message
		c.processMessage(ctx, msg)
	}
}

// processMessage processes the message received from the Kafka topic
func (c *Consumer) processMessage(ctx context.Context, msg kafka.Message) {
	// Unmarshal the message payload
	var orderEvent domain.Order

	err := json.Unmarshal(msg.Value, &orderEvent)
	if err != nil {
		log.Error().Msgf("Error unmarshalling message: %v", err)
		return
	}

	// key -> "order.created.orderID" or "order.cancelled.orderID"
	key := string(msg.Key)
	listKey := strings.Split(key, ".")
	eventType := listKey[1]

	// Process the order event based on the status
	switch eventType {
	case "created":
		// Process order created event
		for _, item := range orderEvent.ProductRequests {
			err := c.productUsecase.ReserveProductStock(ctx, item.ProductID, item.Quantity)
			if err != nil {
				log.Error().Msgf("Error updating stock for product %d: %v", item.ProductID, err)
			}
		}
	case "cancelled":
		// Process order cancelled event
		for _, item := range orderEvent.ProductRequests {
			err := c.productUsecase.ReleaseProductStock(ctx, item.ProductID, item.Quantity)
			if err != nil {
				log.Error().Msgf("Error updating stock for product %d: %v", item.ProductID, err)
			}
		}
	default:
		log.Error().Msgf("Unknown order status: %s", orderEvent.Status)
	}
}
