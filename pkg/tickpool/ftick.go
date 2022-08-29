package tickpool

import (
	"sync"
	"time"
)

func init() {
	pool = &sync.Pool{
		New: func() interface{} {
			return &time.Ticker{}
		},
	}
	for i := 0; i < 12; i++ {
		pool.Put(&time.Ticker{})
	}
}

var (
	pool *sync.Pool
)

func Get(interval time.Duration) (ticker *time.Ticker) {
	ticker = pool.Get().(*time.Ticker)

	if ticker.C == nil {
		ticker = time.NewTicker(interval)
	} else {
		ticker.Reset(interval)
	}

	return ticker
}

func Put(ticker *time.Ticker) {
	ticker.Stop()
	pool.Put(ticker)
}
