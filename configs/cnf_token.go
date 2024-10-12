package configs

import "time"

type TokenConfig struct {
	SigningKey         interface{}   `json:"signingKey"`         // 用于签名的密钥
	TokenLookup        string        `json:"tokenLookup"`        // 定义如何查找令牌
	Method             string        `json:"signingMethod"`      // 签名方法
	BufferTime         time.Duration `json:"bufferTime"`         // 生效时间
	ExpiresTime        time.Duration `json:"expiresTime"`        // 过期时间
	Issuer             string        `json:"issuer"`             // 签发者
	RefreshExpiresTime time.Duration `json:"refreshExpiresTime"` // 刷新令牌过期时间
	ExcludePaths       []string      `json:"excludePaths"`       // 不需要验证的路径
}

func LoadTokenConfig() *TokenConfig {
	config := &TokenConfig{
		SigningKey:         EnvString(TokenSigningKey, "CXEREHKHHP54PXKYTS2E"),
		TokenLookup:        EnvString(TokenTokenLookup, "header:Authorization"),
		Method:             EnvString(TokenSigningMethod, "HS256"),
		BufferTime:         EnvDuration(TokenBufferTime, "15m"),
		ExpiresTime:        EnvDuration(TokenExpiresTime, "24h"),
		Issuer:             EnvString(TokenIssuer, "sagoo"),
		RefreshExpiresTime: EnvDuration(TokenRefreshExpiresTime, "48h"),
		ExcludePaths:       EnvStringSlice(TokenExcludePaths),
	}
	return config
}
