package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"main/domain"
	consumer "main/kafka/consumer"
	producer "main/kafka/producer"

	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
)

func (a *transactionUsecase) SaveScheduledTransaction(ctx context.Context, transaction *domain.ScheduledTransaction) (err error) {

	if err = a.transactionRepo.CreateScheduledTransaction(ctx, transaction); err != nil {
		return err
	}
	return nil
}

func (a *transactionUsecase) PollScheduledTransaction(ctx context.Context, fetchTime time.Time) (err error) {

	transactions, err := a.transactionRepo.GetScheduledTransaction(ctx, fetchTime)
	if err != nil {
		return err
	}

	if err = a.AddTransactionToQueue(ctx, transactions); err != nil {
		return err
	}
	return nil
}

func (a *transactionUsecase) AddTransactionToQueue(ctx context.Context, transaction []domain.ScheduledTransaction) (err error) {
	topic := "scheduled_transactions"
	brokerAddress := viper.GetString("kafka.broker_address")

	for i := range transaction {
		message := fmt.Sprintf("%d|%s|%.2f|%s|%s|%s",
			transaction[i].Id,
			transaction[i].Type,
			transaction[i].Amount,
			transaction[i].Account.AccountNo,
			transaction[i].Receiver.AccountNo,
			transaction[i].CreatedAt)

		producer.RunKafkaProducer(brokerAddress, topic, message)
		fmt.Println("producer", message)
	}

	return nil
}

func (a *transactionUsecase) ConsumeScheduledTransaction(ctx context.Context) (err error) {
	topic := "scheduled_transactions"
	var partition int32 = 0
	var offset int64 = sarama.OffsetNewest
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Processing transaction stop.")
			return
		default:
			message := consumer.RunKafkaConsumer(a.kafkaClient, topic, partition, offset)
			go a.processScheduleTransaction(ctx, message)
			offset += 1

			fmt.Println(offset)
		}
	}
}

func (a *transactionUsecase) processScheduleTransaction(ctx context.Context, message string) (transaction domain.Transaction, err error) {
	timeLayout := "2006-01-02 15:04:05.999999 -0700 MST"
	message_split := strings.Split(message, "|")

	if len(message_split) < 6 {
		fmt.Println("Unexpected message format:", message)
	}
	transaction.Type = message_split[1]
	transaction.Amount, _ = strconv.ParseFloat(message_split[2], 64)
	transaction.Account.AccountNo = message_split[3]
	transaction.Receiver.AccountNo = message_split[4]
	transaction.SubmittedAt, err = time.Parse(timeLayout, message_split[5])
	if err != nil {
		fmt.Println("Error parsing timestamp:", err)
	}

	if err = a.Transfer(ctx, &transaction); err != nil {
		fmt.Println("Error processing transaction:", err)
		return transaction, err
	}

	id, _ := strconv.ParseInt(message_split[0], 10, 64)

	if err = a.UpdateScheduledTransaction(ctx, int64(id)); err != nil {
		fmt.Println("Error processing transaction:", err)
	}
	return
}

func (a *transactionUsecase) UpdateScheduledTransaction(ctx context.Context, id int64) (err error) {

	var transaction domain.ScheduledTransaction
	transaction.Id = id
	transaction.Status = "processed"

	if err := a.transactionRepo.UpdateScheduledTransaction(ctx, transaction); err != nil {
		return err
	}

	return nil
}
