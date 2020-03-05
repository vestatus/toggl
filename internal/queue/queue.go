package queue

import (
	"errors"
	"log"
	"sync"
	"toggl/internal/service"
)

type inmem struct {
	mu     *sync.Mutex
	takers []service.Taker
}

func NewInmem() service.TakerQueue {
	return &inmem{
		mu: &sync.Mutex{},
	}
}

func (q *inmem) Push(taker *service.Taker) error {
	if taker == nil {
		return errors.New("taker is nil")
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	log.Printf("push %#v", taker)

	q.takers = append(q.takers, *taker)
	return nil
}

func (q *inmem) Pop() (*service.Taker, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.takers) == 0 {
		return nil, service.ErrNoTakers
	}

	taker := &q.takers[0]
	q.takers = q.takers[1:]

	log.Printf("pop %#v", taker)

	return taker, nil
}
