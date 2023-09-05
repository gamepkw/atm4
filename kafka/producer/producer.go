package kafka

import (
	"fmt"
	"log"

	"github.com/Shopify/sarama"
)

func CreateProducer(brokerAddress string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	return sarama.NewSyncProducer([]string{brokerAddress}, config)
}

func RunKafkaProducer(brokerAddress string, topic string, message string) {
	producer, err := CreateProducer(brokerAddress)
	if err != nil {
		fmt.Printf("Error creating producer: %v\n", err)
		return
	}
	defer func() {
		if err := producer.Close(); err != nil {
			fmt.Printf("Error closing producer: %v\n", err)
		}
	}()

	// fmt.Println("Kafka producer started.")

	kafkaMessage := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	// partition, offset, err := producer.SendMessage(kafkaMessage)
	// if err != nil {
	// 	log.Fatal("Error sending Kafka message:", err)
	// }

	_, _, err = producer.SendMessage(kafkaMessage)
	if err != nil {
		log.Fatal("Error sending Kafka message:", err)
	}

	// fmt.Printf("Message sent to partition %d at offset %d\n", partition, offset)

}
