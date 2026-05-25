package datax

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"kratos-admin/pkg/toolbox/errorx"
	"kratos-admin/pkg/toolbox/logx"
)

type Pubsub struct {
	rdb *redis.Client
}

func (p *Pubsub) Publish(ctx context.Context, topic string, data any) error {
	bs, err := json.Marshal(data)
	if err != nil {
		return errorx.WithStack(err)
	}
	_, err = p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: topic,
		MaxLen: 1024 * 16,
		Approx: true,
		Values: map[string]interface{}{"message": string(bs)},
	}).Result()
	return errorx.WithStack(err)
}

func (p *Pubsub) Subscribe(ctx context.Context, topic string, consumer func(data []byte)) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			_ = p.subscribe(ctx, topic, consumer)
			time.Sleep(1 * time.Second)
		}
	}
}

func (p *Pubsub) subscribe(ctx context.Context, topic string, consumer func(data []byte)) error {
	defer func() {
		if err := recover(); err != nil {
			logx.Error(ctx, err)
		}
	}()
	safeConsumer := func(data []byte) {
		if err := recover(); err != nil {
			logx.Error(ctx, err)
		}
		consumer(data)
	}
	lastId := "$"
	for {
		msgs, _ := p.rdb.XRead(ctx, &redis.XReadArgs{
			Streams: []string{topic, lastId},
			Count:   10,
			Block:   time.Millisecond * 100,
		}).Result()
		if len(msgs) == 0 {
			continue
		}
		for _, msg := range msgs[0].Messages {
			safeConsumer([]byte(msg.Values["message"].(string)))
			lastId = msg.ID
		}
	}
}

func NewPubsub(rdb *redis.Client) *Pubsub {
	return &Pubsub{rdb: rdb}
}
