package configs

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe/os/command/args"
)

type GormDbConfig struct {
	Driver       string
	Host         string
	Port         string
	Username     string
	Password     string
	Dbname       string
	Config       string
	Prefix       string
	Singular     bool
	Engine       string
	MaxIdleConns int
	MaxOpenConns int
	LogMode      string
	LogZap       bool
	Dsn          string
	ShowSQL      bool
}

func LoadDatabaseConfig() *GormDbConfig {
	driver := EnvString(DatabaseDriver, "mysql")
	dataSource := fmt.Sprintf(
		"%s:%s@(%s:%s)/%s"+"?%s",
		EnvString(DatabaseUserName, "root"),
		EnvString(DatabasePassword, "root"),
		EnvString(DatabaseHost, "127.0.0.1"),
		EnvString(DatabasePort, "3306"),
		EnvString(DatabaseDbName, "default"),
		EnvString(DatabaseConfig, "charset=utf8&collation=utf8_general_ci"),
	)
	show := false
	if args.Mode != "prod" {
		show = true
	}
	config := &GormDbConfig{
		Driver:       driver,
		Dsn:          dataSource,
		ShowSQL:      EnvBool(DatabaseShowSQL, show),
		MaxIdleConns: EnvInt(DatabaseMaxIdleConns, 5),
		MaxOpenConns: EnvInt(DatabaseMaxOpenConns, 50),
		Username:     EnvString(DatabaseUserName, "root"),
		Password:     EnvString(DatabasePassword, "root"),
		Host:         EnvString(DatabaseHost, "127.0.0.1"),
		Port:         EnvString(DatabasePort, "3306"),
		Dbname:       EnvString(DatabaseDbName, "default"),
		Config:       EnvString(DatabaseConfig, "charset=utf8&collation=utf8_general_ci"),
	}
	return config
}
