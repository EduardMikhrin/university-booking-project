package config

import (
	"reflect"
	"time"

	"github.com/EduardMikhrin/university-booking-project/internal/server"
	"github.com/pkg/errors"
	"gitlab.com/distributed_lab/figure"
	"gitlab.com/distributed_lab/kit/comfig"
	"gitlab.com/distributed_lab/kit/kv"
)

type JWTer interface {
	JWT() server.JWT
}

const (
	jwtKey = "jwt"
)

func NewJWTer(getter kv.Getter) JWTer {
	return &jwt{getter: getter}
}

type jwtConfig struct {
	SecretKey            string        `fig:"secret_key,required"`
	Issuer               string        `fig:"issuer,required"`
	Audience             string        `fig:"audience,required"`
	AccessTokenLifetime  time.Duration `fig:"access_token_lifetime,required"`
	RefreshTokenLifetime time.Duration `fig:"refresh_token_lifetime,required"`
}

type jwt struct {
	getter kv.Getter
	once   comfig.Once
}

func (j *jwt) JWT() server.JWT {
	cfg := j.jwtConfig(jwtKey)
	return server.JWT{
		SecretKey:            cfg.SecretKey,
		Issuer:               cfg.Issuer,
		Audience:             cfg.Audience,
		AccessTokenLifetime:  cfg.AccessTokenLifetime,
		RefreshTokenLifetime: cfg.RefreshTokenLifetime,
	}
}

func (j *jwt) jwtConfig(key string) jwtConfig {
	return j.once.Do(func() interface{} {
		var cfg jwtConfig
		err := figure.
			Out(&cfg).
			With(figure.BaseHooks, jwtHooks).
			From(kv.MustGetStringMap(j.getter, key)).
			Please()
		if err != nil {
			panic(errors.Wrap(err, "failed to load jwt config"))
		}

		return cfg
	}).(jwtConfig)
}

var jwtHooks = figure.Hooks{
	"time.Duration": func(value interface{}) (reflect.Value, error) {
		switch v := value.(type) {
		case string:
			duration, err := time.ParseDuration(v)
			if err != nil {
				return reflect.Value{}, errors.Wrapf(err, "failed to parse duration: %s", v)
			}
			return reflect.ValueOf(duration), nil
		case int:
			return reflect.ValueOf(time.Duration(v) * time.Second), nil
		case int64:
			return reflect.ValueOf(time.Duration(v) * time.Second), nil
		default:
			return reflect.Value{}, errors.Errorf("unsupported conversion from %T to time.Duration", value)
		}
	},
}
