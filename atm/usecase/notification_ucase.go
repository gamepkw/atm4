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

type notificationUsecase struct {
	transactionUsecase domain.TransactionUsecase
	contextTimeout     time.Duration
	kafkaClient        sarama.Client
}

// NewAccountUsecase will create new an accountUsecase object representation of domain.AccountUsecase interface
func NewNotificationUsecase(tu domain.TransactionUsecase, timeout time.Duration, kafka sarama.Client) domain.NotificationUsecase {
	return &notificationUsecase{
		transactionUsecase: tu,
		contextTimeout:     timeout,
		kafkaClient:        kafka,
	}
}

func (n *notificationUsecase) SendTransactionNotification(ctx context.Context) {
	topic := "sms_transaction"
	var partition int32 = 0
	var offset int64 = sarama.OffsetNewest
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Send Notification stopped")
			return
		default:

			// Consume a Kafka message
			message := consumer.RunKafkaConsumer(n.kafkaClient, topic, partition, offset)
			message_split := strings.Split(message, "|")
			if message_split[0] == "withdraw" {
				fmt.Printf("%s Baht has been withdrawn from account no: %s at %s\nRemaining balance: %s\n%s\n",
					message_split[1],
					message_split[2],
					message_split[3],
					message_split[4],
					strings.Repeat("-", 120))

			} else if message_split[0] == "deposit" {
				fmt.Printf("%s Baht has been deposited into account no: %s at %s\nRemaining balance: %s\n%s\n",
					message_split[1],
					message_split[2],
					message_split[3],
					message_split[4],
					strings.Repeat("-", 120))
			} else if message_split[0] == "transfer" {
				fmt.Printf("%s Baht has been transferred from account no: %s to account no: %s at %s\n%s\n",
					message_split[1],
					message_split[2],
					message_split[5],
					message_split[3],
					strings.Repeat("-", 120))
			}

			// offset += 1

			// time.Sleep(time.Second)
		}

	}

}
