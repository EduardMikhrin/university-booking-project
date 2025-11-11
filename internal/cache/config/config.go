package config

import (
	"github.com/EduardMikhrin/university-booking-project/internal/cache"
	rdb "github.com/EduardMikhrin/university-booking-project/internal/cache/redis"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

const cacheConfigKey = "cache"

type Cacher interface {
	Cache() cache.CacheQ
}

func NewCacher(getter kv.Getter) Cacher {
	return &cacher{
		getter: getter,
	}
}

type cacher struct {
	getter kv.Getter
	once   comfig.Once
}

type config struct {
	URL      string `fig:"url, required"`
	Password string `fig:"password, required"`
	DB       int    `fig:"db, required"`
}

func (c *cacher) Cache() cache.CacheQ {
	config := c.Config()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.URL,
		Password: config.Password,
		DB:       config.DB,
	})

	return rdb.NewMaster(redisClient)
}

func (c *cacher) Config() *config {
	return c.once.Do(func() interface{} {
		var cfg config
		if err := figure.Out(&cfg).From(kv.MustGetStringMap(c.getter, cacheConfigKey)).Please(); err != nil {
			panic(errors.Wrap(err, "failed to figure out core observer config"))
		}
		return &cfg
	}).(*config)
}
