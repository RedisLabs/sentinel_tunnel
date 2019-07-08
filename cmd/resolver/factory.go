package resolver

import (
	"errors"
	"github.com/go-redis/redis"
	"github.com/rs/zerolog/log"
)

func CreateSentinels(addrs []string) ([]*redis.SentinelClient, error) {
	sentinels := make([]*redis.SentinelClient, 0, len(addrs))
	for _, addr := range addrs {
		sentinel, err := createSentinel(addr)
		if err != nil {
			continue
		}
		sentinels = append(sentinels, sentinel)
	}
	if len(sentinels) == 0 {
		return nil, errors.New("error connecting to all sentinels")
	}

	return sentinels, nil
}

func createSentinel(addr string) (*redis.SentinelClient, error) {
	sentinel := redis.NewSentinelClient(&redis.Options{
		Addr: addr,
	})

	log.Debug().Msgf("connected to sentinel: %s", addr)

	return sentinel, nil
}
