package model

const (
	DATABASE_DRIVER_POSTGRES = "postgres"
)

type SqlSettings struct {
	DriverName                  *string
	DataSource                  *string
	DataSourceReplicas          []string
	DataSourceSearchReplicas    []string
	MaxIdleConns                *int
	ConnMaxLifetimeMilliseconds *int
	MaxOpenConns                *int
	Trace                       bool
	AtRestEncryptKey            string
	QueryTimeout                *int
}

type NoSqlSettings struct {
	Uri string
}

type Config struct {
	LogLevel           string
	LogHttpApiDir      string
	SqlSettings        SqlSettings
	NoSqlSettings      NoSqlSettings
	CallServerSettings CallServerSettings
}

type CallServerSettings struct {
	Host string
}
