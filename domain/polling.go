package domain

import (
	"context"
	"sync"
	"time"
)

type PollingUsecase interface {
	Polling(ctx context.Context, wg *sync.WaitGroup, interval time.Duration, stopChan <-chan struct{})
}
