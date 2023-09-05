package kafka

import (
	"fmt"

	"github.com/Shopify/sarama"
)

func CreateConsumer(brokerAddress, topic string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	return sarama.NewConsumer([]string{brokerAddress}, config)
}

func RunKafkaConsumer(kafkaClient sarama.Client, topic string, partition int32, offset int64) (message string) {
	consumer, err := sarama.NewConsumerFromClient(kafkaClient)
	if err != nil {
		fmt.Printf("Error creating consumer: %v\n", err)
		return
	}
	defer func() {
		if err := consumer.Close(); err != nil {
			fmt.Printf("Error closing consumer: %v\n", err)
		}
	}()

	partitionConsumer, err := consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		fmt.Printf("Error creating partition consumer: %v\n", err)
		return
	}
	defer func() {
		if err := partitionConsumer.Close(); err != nil {
			fmt.Printf("Error closing partition consumer: %v\n", err)
		}
	}()

	for message := range partitionConsumer.Messages() {
		return string(message.Value)

	}

	return

}

type ConsumerGroupHandler struct{}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		// Process the message
		fmt.Printf("Received message: %s\n", message.Value)

		// Mark the message as processed by committing its offset
		session.MarkMessage(message, "")
	}
	return nil
}
