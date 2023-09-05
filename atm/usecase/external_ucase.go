package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"main/domain"

	consumer "main/kafka/consumer"

	"github.com/Shopify/sarama"
)

type externalUsecase struct {
	contextTimeout time.Duration
	kafkaClient    sarama.Client
}

// NewAccountUsecase will create new an accountUsecase object representation of domain.AccountUsecase interface
func NewExternalUsecase(timeout time.Duration, kafka sarama.Client) domain.ExternalUsecase {
	return &externalUsecase{
		contextTimeout: timeout,
		kafkaClient:    kafka,
	}
}

func (a *externalUsecase) SendSms(ctx context.Context) {
	topic := "sms"
	var partition int32 = 0
	var offset int64 = sarama.OffsetNewest
	for {
		select {
		case <-ctx.Done():
			fmt.Println("SendSms stopped")
			return
		default:
			// Consume a Kafka message
			message := consumer.RunKafkaConsumer(a.kafkaClient, topic, partition, offset)
			message_split := strings.Split(message, "|")
			fmt.Printf("OTP='%s' for reset password\n%s\n",
				message_split[0],
				strings.Repeat("-", 60))

			// time.Sleep(time.Second)
		}

	}

}
