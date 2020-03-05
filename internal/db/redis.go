package db

import (
	"encoding/json"
	"toggl/internal/service"

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

func (r *Redis) Push(taker *service.Taker) error {
	bts, err := json.Marshal(taker)
	if err != nil {
		return errors.Wrap(err, "failed to marshal taker")
	}

	return r.client.RPush(queueName, string(bts)).Err()
}

func (r *Redis) Pop() (*service.Taker, error) {
	res, err := r.client.LPop(queueName).Result()
	if err == redis.Nil {
		return nil, service.ErrNoTakers
	}

	var taker service.Taker

	err = json.Unmarshal([]byte(res), &taker)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decode message")
	}

	return &taker, nil
}

func (r *Redis) Add(id int) error {
	return r.client.SAdd(setName, id).Err()
}

func (r *Redis) Contains(id int) (bool, error) {
	return r.client.SIsMember(setName, id).Result()
}
