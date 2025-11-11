package config

import (
	cacher "github.com/EduardMikhrin/university-booking-project/internal/cache/config"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
	"gitlab.com/distributed_lab/kit/pgdb"
)

type Config interface {
	comfig.Logger
	pgdb.Databaser
}

type config struct {
	getter kv.Getter

	comfig.Logger
	pgdb.Databaser
	cacher.Cacher
}

func New(getter kv.Getter) Config {
	return &config{
		getter:    getter,
		Logger:    comfig.NewLogger(getter, comfig.LoggerOpts{}),
		Databaser: pgdb.NewDatabaser(getter),
		Cacher:    cacher.NewCacher(getter),
	}
}
