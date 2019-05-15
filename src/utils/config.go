package utils

import (
	"fmt"
	"github.com/webitel/acr/src/config"
	"github.com/webitel/acr/src/model"
)

//TODO
func GetSqlSettings() model.SqlSettings {

	dbDataSource := fmt.Sprintf("postgres://%s:%s@%s:%v/%s?fallback_application_name=acr&sslmode=%s&connect_timeout=10",
		config.Conf.Get("pg:user"),
		config.Conf.Get("pg:password"),
		config.Conf.Get("pg:host"),
		config.Conf.Get("pg:port"),
		config.Conf.Get("pg:dbName"),
		config.Conf.Get("pg:sslMode"),
	)

	var trace = config.Conf.Get("pg:trace") == "true"

	dbDriverName := "postgres"
	maxIdleConns := 5
	maxOpenConns := 5
	connMaxLifetimeMilliseconds := 3600000

	setteings := model.SqlSettings{
		DriverName:                  &dbDriverName,
		DataSource:                  &dbDataSource,
		MaxIdleConns:                &maxIdleConns,
		MaxOpenConns:                &maxOpenConns,
		ConnMaxLifetimeMilliseconds: &connMaxLifetimeMilliseconds,
		Trace: trace,
	}

	return setteings
}

func GetNoSqlSettings() model.NoSqlSettings {
	settings := model.NoSqlSettings{
		Uri: config.Conf.Get("mongodb:uri"),
	}

	return settings
}

func GetCallServerSettings() model.CallServerSettings {
	return model.CallServerSettings{
		Host: fmt.Sprintf("%v:%v", config.Conf.Get("server:host"), config.Conf.Get("server:ports")),
	}
}

//TODO
func LoadConfig() (*model.Config, error) {
	conf := model.Config{
		LogLevel:           config.Conf.Get("application:loglevel"),
		NoSqlSettings:      GetNoSqlSettings(),
		SqlSettings:        GetSqlSettings(),
		CallServerSettings: GetCallServerSettings(),
	}

	return &conf, nil
}
