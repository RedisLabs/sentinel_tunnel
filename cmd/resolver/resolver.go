package resolver

import (
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"strings"
)

type Resolver interface {
	Resolve(dbName string) (string, error)
}

type resolver struct {
	sentinels []*redis.SentinelClient
}

func NewResolver(sentinels []*redis.SentinelClient) Resolver {
	return &resolver{sentinels: sentinels}
}

func (r *resolver) Resolve(dbName string) (string, error) {
	for _, sentinel := range r.sentinels {
		result, err := sentinel.GetMasterAddrByName(dbName).Result()
		if err != nil {
			continue
		}

		ip := strings.Join(result, ":")

		log.Info().Msgf("'%s' resolved to master: %s", dbName, ip)

		return ip, nil
	}

	return "", errors.New("all sentinels failed")
}
