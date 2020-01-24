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

type DiscoverySettings struct {
	Url string
}

type Config struct {
	LogLevel           string
	SqlSettings        SqlSettings
	NoSqlSettings      NoSqlSettings
	DiscoverySettings  DiscoverySettings
	CallServerSettings CallServerSettings
}

type CallServerSettings struct {
	Host string
}
