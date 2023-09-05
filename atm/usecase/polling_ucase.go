package usecase

import (
	"context"
	"sync"
	"time"

	"main/domain"
)

type pollingUsecase struct {
	transactionUsecase domain.TransactionUsecase
	interval           time.Duration
}

// NewAccountUsecase will create new an accountUsecase object representation of domain.AccountUsecase interface
func NewPollingUsecase(tu domain.TransactionUsecase, interval time.Duration) domain.PollingUsecase {
	return &pollingUsecase{
		transactionUsecase: tu,
		interval:           interval,
	}
}

func (p *pollingUsecase) Polling(ctx context.Context, wg *sync.WaitGroup, interval time.Duration, stopChan <-chan struct{}) {
	defer wg.Done()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// fmt.Println("Polling the database...")
			p.transactionUsecase.PollScheduledTransaction(ctx, time.Now().Truncate(5*time.Minute))

		case <-stopChan:
			return
		}
	}
}
