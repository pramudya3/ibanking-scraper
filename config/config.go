package config

import (
	"ibanking-scraper/internal/tool"
	"sync"
)

type LogLevel uint

var (
	once sync.Once

	appConfig *Config
)

type (
	Config struct {
		SkipAuth             bool
		Environment          string
		LogLevel             string
		ServerAddr           string
		ServerPort           int
		ServerTimeoutContext int
		ServerRateLimit      int
		ServerBurstLimit     int
		SecretAccessJWT      string
		SecretRefreshJWT     string

		Hashid   *HashidConfig
		Database *DatabaseConfig
		AWS      *AWSConfig
	}

	HashidConfig struct {
		Salt      string
		MinLength int
	}

	DatabaseConfig struct {
		DatabaseHost            string
		DatabaseName            string
		DatabaseUser            string
		DatabasePassword        string
		DatabasePort            int
		DatabaseConnMaxlifetime int
		DatabaseMaxIdleConns    int
		DatabaseMaxOpenConns    int
		DatabaseSSL             string
		DatabaseDebugMode       bool
	}

	AWSConfig struct {
		AccessKey     string
		SecretKey     string
		BucketName    string
		UploadTimeout int
		Region        string
		PrefixURL     string
	}
)

func Load() *Config {
	once.Do(func() {
		appConfig = &Config{
			Environment:          tool.GetStringWithDefault("environment", "APP_ENVIRONMENT", "development"),
			SkipAuth:             tool.GetBoolWithDefault("server.skipAuth", "SKIP_AUTH", false),
			LogLevel:             tool.GetStringWithDefault("server.logLevel", "LOG_LEVEL", "DEBUG"),
			ServerAddr:           tool.GetStringWithDefault("server.address", "SERVER_ADDRESS", "0.0.0.0"),
			ServerPort:           tool.GetIntWithDefault("server.port", "SERVER_PORT", 8080),
			ServerTimeoutContext: tool.GetIntWithDefault("server.timeoutContext", "SERVER_TIMEOUT_CONTEXT", 10),
			ServerRateLimit:      tool.GetIntWithDefault("server.rateLimit", "SERVER_RATE_LIMIT", 20),
			ServerBurstLimit:     tool.GetIntWithDefault("server.burstLimit", "SERVER_BURST_LIMIT", 50),
			SecretAccessJWT:      tool.GetStringWithDefault("server.jwt.accessSecret", "JWT_ACCESS_SECRET", "jwt-super-secret"),
			SecretRefreshJWT:     tool.GetStringWithDefault("server.jwt.refreshSecret", "JWT_REFRESH_SECRET", "jwt-super-secret"),

			Database: &DatabaseConfig{
				DatabaseHost:            tool.GetStringWithDefault("database.host", "DB_HOST", ""),
				DatabaseName:            tool.GetStringWithDefault("database.name", "DB_NAME", ""),
				DatabaseUser:            tool.GetStringWithDefault("database.user", "DB_USER", ""),
				DatabasePassword:        tool.GetStringWithDefault("database.password", "DB_PASSWORD", ""),
				DatabasePort:            tool.GetIntWithDefault("database.port", "DB_PORT", 5432),
				DatabaseConnMaxlifetime: tool.GetIntWithDefault("database.connMaxLifetime", "DB_MAX_CONN_LIFETIME", 120),
				DatabaseMaxIdleConns:    tool.GetIntWithDefault("database.maxIdleConns", "DB_MAX_IDLE_CONNS", 10),
				DatabaseMaxOpenConns:    tool.GetIntWithDefault("database.maxOpenConns", "DB_MAX_OPEN_CONNS", 50),
				DatabaseSSL:             tool.GetStringWithDefault("database.ssl", "DB_SSL", "disable"),
				DatabaseDebugMode:       tool.GetBoolWithDefault("database.debug", "DB_DEBUG", true),
			},

			Hashid: &HashidConfig{
				Salt:      tool.GetStringWithDefault("hashid.salt", "HASHID_SALT", "super-secret"),
				MinLength: tool.GetIntWithDefault("hashid.minLength", "HASHID_MIN_LENGTH", 10),
			},

			AWS: &AWSConfig{
				AccessKey:     tool.GetStringWithDefault("aws.accessKey", "AWS_ACCESS_KEY", ""),
				SecretKey:     tool.GetStringWithDefault("aws.secretKey", "AWS_SECRET_KEY", ""),
				BucketName:    tool.GetStringWithDefault("aws.bucketName", "AWS_BUCKET_NAME", ""),
				UploadTimeout: tool.GetIntWithDefault("aws.uploadTimeout", "AWS_UPLOAD_TIMEOUT", 60),
				Region:        tool.GetStringWithDefault("aws.region", "AWS_REGION", ""),
				PrefixURL:     tool.GetStringWithDefault("aws.prefixURL", "AWS_PREFIX_URL", "https://mpnstorage.s3.ap-southeast-3.amazonaws.com"),
			},
		}
	})
	return appConfig
}
