package config

// Config 应用总配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Redis    RedisConfig
}

type ServerConfig struct {
	Port int
	Mode string // debug / release / test
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

type JWTConfig struct {
	Secret     string
	ExpireHour int
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}
