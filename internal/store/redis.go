package store

import (
	"context"
	"encoding/json"

	"github.com/vestatus/toggl/internal/service"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
)

const (
	setName   = "sender_sent_set"
	queueName = "sender_takers_queue"
)

type Redis struct {
	client *redis.Client
}

func NewRedis(client *redis.Client) *Redis {
	return &Redis{client: client}
}

func (r *Redis) Push(ctx context.Context, taker *service.Taker) error {
	bts, err := json.Marshal(taker)
	if err != nil {
		return errors.Wrap(err, "failed to marshal taker")
	}

	err = r.client.WithContext(ctx).RPush(queueName, string(bts)).Err()
	if err != nil {
		return service.Fatal(err)
	}

	return nil
}

func (r *Redis) Pop(ctx context.Context) (*service.Taker, error) {
	res, err := r.client.WithContext(ctx).LPop(queueName).Result()
	if err == redis.Nil {
		return nil, service.ErrNoTakers
	}
	if err != nil {
		return nil, service.Fatal(err)
	}

	var taker service.Taker

	err = json.Unmarshal([]byte(res), &taker)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode message")
	}

	return &taker, nil
}

func (r *Redis) Add(ctx context.Context, id int) error {
	const resultInserted = 1

	res, err := r.client.WithContext(ctx).SAdd(setName, id).Result()
	if err != nil {
		return err
	}

	if res != resultInserted {
		return errors.New("the set already contains this key")
	}

	return nil
}

func (r *Redis) Contains(ctx context.Context, id int) (bool, error) {
	res, err := r.client.WithContext(ctx).SIsMember(setName, id).Result()
	if err != nil {
		return false, service.Fatal(err)
	}

	return res, nil
}
