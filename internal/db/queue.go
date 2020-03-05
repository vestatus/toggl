package db

import (
	"errors"
	"log"
	"sync"
	"toggl/internal/service"
)

type Inmem struct {
	bufMu  *sync.Mutex
	takers []service.Taker

	setMu *sync.RWMutex
	set   map[int]struct{}
}

func (q *Inmem) Add(id int) error {
	q.setMu.Lock()
	defer q.setMu.Unlock()

	q.set[id] = struct{}{}
	return nil
}

func (q *Inmem) Contains(id int) (bool, error) {
	q.setMu.RLock()
	defer q.setMu.RUnlock()

	_, found := q.set[id]
	return found, nil
}

func NewInmem() *Inmem {
	return &Inmem{
		bufMu: &sync.Mutex{},
		setMu: &sync.RWMutex{},
		set:   map[int]struct{}{},
	}
}

func (q *Inmem) Push(taker *service.Taker) error {
	if taker == nil {
		return errors.New("taker is nil")
	}

	q.bufMu.Lock()
	defer q.bufMu.Unlock()

	log.Printf("push %#v", taker)

	q.takers = append(q.takers, *taker)
	return nil
}

func (q *Inmem) Pop() (*service.Taker, error) {
	q.bufMu.Lock()
	defer q.bufMu.Unlock()

	if len(q.takers) == 0 {
		return nil, service.ErrNoTakers
	}

	taker := &q.takers[0]
	q.takers = q.takers[1:]

	log.Printf("pop %#v", taker)

	return taker, nil
}
