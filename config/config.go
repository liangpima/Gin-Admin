package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Log      LogConfig      `mapstructure:"log"`
	Upload   UploadConfig   `mapstructure:"upload"`
	Casbin   CasbinConfig   `mapstructure:"casbin"`
	CORS     CORSConfig     `mapstructure:"cors"`
}

type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	ExposeHeaders    []string `mapstructure:"expose_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

type ServerConfig struct {
	Port         int    `mapstructure:"port"`
	Mode         string `mapstructure:"mode"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

type DatabaseConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	Username        string `mapstructure:"username"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime int    `mapstructure:"conn_max_idle_time"`
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.Username, d.Password, d.Host, d.Port, d.DBName)
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	AccessExpire  int64  `mapstructure:"access_expire"`
	RefreshExpire int64  `mapstructure:"refresh_expire"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
	// DBRetentionDays 数据库操作日志/登录日志的保留天数，超期由定时任务清理；
	// 小于等于 0 表示不自动清理
	DBRetentionDays int `mapstructure:"db_retention_days"`
}

type UploadConfig struct {
	SavePath  string `mapstructure:"save_path"`
	MaxSize   int    `mapstructure:"max_size"`
	AllowExts string `mapstructure:"allow_exts"`
}

type CasbinConfig struct {
	ModelPath string `mapstructure:"model_path"`
}

var Cfg Config

func Init(path string) error {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	if err := viper.Unmarshal(&Cfg); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 环境变量覆盖敏感配置
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		Cfg.Database.Password = v
	}
	if v := os.Getenv("JWT_SECRET"); v != "" {
		Cfg.JWT.Secret = v
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		Cfg.Redis.Password = v
	}

	return nil
}

// GetJWTSecret returns JWT secret, preferring env var
func GetJWTSecret() string {
	if v := os.Getenv("JWT_SECRET"); v != "" {
		return v
	}
	return Cfg.JWT.Secret
}

// IsProduction returns true if mode is release
func IsProduction() bool {
	return strings.EqualFold(Cfg.Server.Mode, "release")
}

// 配置文件中的默认值。这些值公开可见，生产环境必须覆盖。
const (
	defaultJWTSecret  = "change-me-in-production"
	defaultDBPassword = "123456"
)

// ValidateSecurity 校验生产环境的关键密钥是否仍为默认值。
//
// 默认值一旦被带上生产环境，攻击者可据此伪造 JWT（等同于任意用户登录）
// 或直连数据库，因此这里直接拒绝启动，而不是仅打印告警。
// 开发环境不做限制，便于本地起步。
func ValidateSecurity() error {
	if !IsProduction() {
		return nil
	}

	var problems []string

	if secret := GetJWTSecret(); secret == "" || secret == defaultJWTSecret {
		problems = append(problems, "jwt.secret 仍为默认值（请设置环境变量 JWT_SECRET）")
	}
	if Cfg.Database.Password == defaultDBPassword {
		problems = append(problems, "database.password 仍为默认值（请设置环境变量 DB_PASSWORD）")
	}

	if len(problems) > 0 {
		return fmt.Errorf("生产环境安全检查未通过，已拒绝启动：\n  - %s", strings.Join(problems, "\n  - "))
	}
	return nil
}
