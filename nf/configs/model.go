package configs

import (
	"fmt"
	"github.com/ServiceWeaver/weaver/runtime/codegen"
)

type ModSwaConfig struct {
	Jwt     ModJwt     `mapstructure:"jwt" json:"jwt" toml:"jwt"`
	Redis   ModRedis   `mapstructure:"redis" json:"redis" toml:"redis"`
	Captcha ModCaptcha `mapstructure:"captcha" json:"captcha" toml:"captcha"`
	GormDB  ModGormDb  `mapstructure:"gormDB" json:"gormDB" toml:"gormDB"`
	System  ModSystem  `mapstructure:"system" json:"system" toml:"system"`
	Local   Local      `mapstructure:"local" json:"local" toml:"local"`
	Log     LogConfig  `mapstructure:"log" json:"log" toml:"log"`
}

func (ModSwaConfig) WeaverMarshal(*codegen.Encoder)   {}
func (ModSwaConfig) WeaverUnmarshal(*codegen.Decoder) {}

type ModCaptcha struct {
	KeyLong            int `mapstructure:"keyLong" json:"keyLong" yaml:"keyLong"`
	ImgWidth           int `mapstructure:"imgWidth" json:"imgWidth" yaml:"imgWidth"`
	ImgHeight          int `mapstructure:"imgHeight" json:"imgHeight" yaml:"imgHeight"`
	OpenCaptcha        int `mapstructure:"openCaptcha" json:"openCaptcha" yaml:"openCaptcha"`
	OpenCaptchaTimeOut int `mapstructure:"openCaptchaTimeout" json:"openCaptchaTimeout" yaml:"openCaptchaTimeout"`
}

func (ModCaptcha) WeaverMarshal(*codegen.Encoder)   {}
func (ModCaptcha) WeaverUnmarshal(*codegen.Decoder) {}

type ModGormDb struct {
	DbType       string `mapstructure:"dbType" json:"dbType" yaml:"dbType"`
	Ip           string `mapstructure:"ip" json:"ip" yaml:"ip"`
	Port         string `mapstructure:"port" json:"port" yaml:"port"`
	Username     string `mapstructure:"username" json:"username" yaml:"username"`
	Password     string `mapstructure:"password" json:"password" yaml:"password"`
	Dbname       string `mapstructure:"dbName" json:"dbName" yaml:"dbName"`
	Config       string `mapstructure:"config" json:"config" yaml:"config"`
	Prefix       string `mapstructure:"prefix" json:"prefix" yaml:"prefix"`
	Singular     bool   `mapstructure:"singular" json:"singular" yaml:"singular"`
	Engine       string `mapstructure:"engine" json:"engine" yaml:"engine" default:"InnoDB"`
	MaxIdleConns int    `mapstructure:"maxIdleConns" json:"maxIdleConns" yaml:"maxIdleConns"`
	MaxOpenConns int    `mapstructure:"maxOpenConns" json:"maxOpenConns" yaml:"maxOpenConns"`
	LogMode      string `mapstructure:"logMode" json:"logMode" yaml:"logMode"`
	LogZap       bool   `mapstructure:"logZap" json:"logZap" yaml:"logZap"`
	Dsn          string
}

func (ModGormDb) WeaverMarshal(*codegen.Encoder)   {}
func (ModGormDb) WeaverUnmarshal(*codegen.Decoder) {}

func (m *ModGormDb) SetDsn() {
	m.Dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s",
		m.Username,
		m.Password,
		m.Ip,
		m.Port,
		m.Dbname,
		m.Config,
	)
}

type ModJwt struct {
	SigningKey  string `mapstructure:"signingKey" json:"signingKey" yaml:"signingKey"`    // --jwt签名
	ExpiresTime string `mapstructure:"expiresTime" json:"expiresTime" yaml:"expiresTime"` // --过期时间
	BufferTime  string `mapstructure:"bufferTime" json:"bufferTime" yaml:"bufferTime"`    // --缓冲时间
	Issuer      string `mapstructure:"issuer" json:"issuer" yaml:"issuer"`                // --签发者
}

func (ModJwt) WeaverMarshal(*codegen.Encoder)   {}
func (ModJwt) WeaverUnmarshal(*codegen.Decoder) {}

type Local struct {
	Path      string `mapstructure:"path" json:"path" yaml:"path"`
	StorePath string `mapstructure:"storePath" json:"storePath" yaml:"storePath"`
}

func (Local) WeaverMarshal(*codegen.Encoder)   {}
func (Local) WeaverUnmarshal(*codegen.Decoder) {}

type LogConfig struct {
	Level     string    `mapstructure:"level" json:"level" yaml:"level"`
	Pattern   string    `mapstructure:"pattern" json:"pattern" yaml:"pattern"`
	Output    string    `mapstructure:"output" json:"output" yaml:"output"`
	LogRotate LogRotate `mapstructure:"logRotate" json:"logRotate" yaml:"logRotate"`
}

type LogRotate struct {
	Filename   string `mapstructure:"filename" json:"filename" yaml:"filename"`
	MaxSize    int    `mapstructure:"maxSize" json:"maxSize" yaml:"maxSize"`
	MaxBackups int    `mapstructure:"maxBackups" json:"maxBackups" yaml:"maxBackups"`
	MaxAge     int    `mapstructure:"maxAge" json:"maxAge" yaml:"maxAge"`
	Compress   bool   `mapstructure:"compress" json:"compress" yaml:"compress"`
}

func (LogConfig) WeaverMarshal(*codegen.Encoder)   {}
func (LogConfig) WeaverUnmarshal(*codegen.Decoder) {}

type ModRedis struct {
	DB       int    `mapstructure:"db" json:"db" yaml:"db"`
	Addr     string `mapstructure:"addr" json:"addr" yaml:"addr"`
	Password string `mapstructure:"password" json:"password" yaml:"password"`
}

func (ModRedis) WeaverMarshal(*codegen.Encoder)   {}
func (ModRedis) WeaverUnmarshal(*codegen.Decoder) {}

type ModSystem struct {
	Debug                  bool   `mapstructure:"debug" json:"debug" yaml:"debug"`
	EnablePProf            bool   `mapstructure:"enablePProf" json:"enablePProf" yaml:"enablePProf"`
	PprofPort              string `mapstructure:"pprofPort" json:"pprofPort" yaml:"pprofPort"`
	Env                    string `mapstructure:"env" json:"env" yaml:"env"`
	Addr                   int    `mapstructure:"addr" json:"addr" yaml:"addr"`
	DbType                 string `mapstructure:"dbType" json:"dbType" yaml:"dbType"`
	OssType                string `mapstructure:"ossType" json:"ossType" yaml:"ossType"`
	UseMultipoint          bool   `mapstructure:"useMultipoint" json:"useMultipoint" yaml:"useMultipoint"`
	UseRedis               bool   `mapstructure:"useRedis" json:"useRedis" yaml:"useRedis"`
	LimitCountIP           int    `mapstructure:"ipLimitCount" json:"ipLimitCount" yaml:"ipLimitCount"`
	LimitTimeIP            int    `mapstructure:"ipLimitTime" json:"ipLimitTime" yaml:"ipLimitTime"`
	RouterPrefix           string `mapstructure:"routerPrefix" json:"routerPrefix" yaml:"routerPrefix"`
	ResourceCachingEnabled bool   `mapstructure:"resourceCachingEnabled" json:"resourceCachingEnabled" yaml:"resourceCachingEnabled"`
}

func (ModSystem) WeaverMarshal(*codegen.Encoder)   {}
func (ModSystem) WeaverUnmarshal(*codegen.Decoder) {}
