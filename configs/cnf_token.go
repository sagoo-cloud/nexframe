package configs

type TokenConfig struct {
	Key string `json:"key"`
	Exp int64  `json:"exp"`
}

func LoadTokenConfig() *TokenConfig {
	config := &TokenConfig{
		Key: EnvString(TokenKey, "CXEREHKHHP54PXKYTS2E"),
		Exp: int64(EnvInt(TokenExp, 2592000)),
	}
	return config
}
