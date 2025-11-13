package server

import "time"

type JWT struct {
	SecretKey            string        `fig:"secret_key,required"`
	Issuer               string        `fig:"issuer,required"`
	Audience             string        `fig:"audience,required"`
	AccessTokenLifetime  time.Duration `fig:"access_token_lifetime,required"`
	RefreshTokenLifetime time.Duration `fig:"refresh_token_lifetime,required"`
}
