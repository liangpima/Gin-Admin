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
	// ReadHeaderTimeout 限制读取请求头的时间，用于防 Slowloris：
	// 攻击者保持连接、每次只发几个字节的头，ReadTimeout 会被不断刷新，
	// 连接可被无限占用。缺省（<=0）时由 Validate 填默认值。
	ReadHeaderTimeout int `mapstructure:"read_header_timeout"`
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

	return Validate()
}

// 请求头读取超时的默认值与上限（秒）。
//
// 默认 10s 对正常客户端绰绰有余（局域网内请求头是毫秒级），
// 上限 60s 是为了防止把它配成远大于 ReadTimeout 而失去防护意义。
const (
	defaultReadHeaderTimeout = 10
	maxReadHeaderTimeout     = 60
)

// Validate 校验关键配置项的取值范围，并补齐可缺省的字段。
//
// 为什么需要它：Init 只负责「解析」，解析成功不代表配置可用。
// 端口写成 70000、超时写成 0、JWT 有效期写成负数这类错误，
// 若不在启动时拦下，会一路带到运行期才以难以诊断的方式暴露 ——
// 端口越界表现成 bind 失败，超时为 0 表现成「请求永不超时」，
// 而 Slowloris 这类攻击恰好只需要一个没有超时上限的连接。
//
// 启动即失败，好过带病运行。
func Validate() error {
	var problems []string

	if Cfg.Server.Port < 1 || Cfg.Server.Port > 65535 {
		problems = append(problems, fmt.Sprintf(
			"server.port 非法（%d），应在 1-65535 之间", Cfg.Server.Port))
	}
	if Cfg.Server.ReadTimeout <= 0 {
		problems = append(problems,
			"server.read_timeout 必须为正数（为 0 时慢速连接可无限占用）")
	}
	if Cfg.Server.WriteTimeout <= 0 {
		problems = append(problems, "server.write_timeout 必须为正数")
	}

	// ReadHeaderTimeout 是后加字段，旧配置里没有，因此缺省时自动补默认值，
	// 而不是报错 —— 否则升级后所有既有配置文件都会导致启动失败。
	if Cfg.Server.ReadHeaderTimeout <= 0 {
		Cfg.Server.ReadHeaderTimeout = defaultReadHeaderTimeout
	}
	if Cfg.Server.ReadHeaderTimeout > maxReadHeaderTimeout {
		problems = append(problems, fmt.Sprintf(
			"server.read_header_timeout 过大（%d 秒），不应超过 %d 秒",
			Cfg.Server.ReadHeaderTimeout, maxReadHeaderTimeout))
	}

	if Cfg.Database.Host == "" {
		problems = append(problems, "database.host 不能为空")
	}
	if Cfg.Database.Port < 1 || Cfg.Database.Port > 65535 {
		problems = append(problems, fmt.Sprintf(
			"database.port 非法（%d），应在 1-65535 之间", Cfg.Database.Port))
	}
	if Cfg.Database.DBName == "" {
		problems = append(problems, "database.dbname 不能为空")
	}
	if Cfg.Database.MaxOpenConns < 1 {
		problems = append(problems, "database.max_open_conns 必须大于 0")
	}

	if Cfg.Redis.Addr == "" {
		problems = append(problems, "redis.addr 不能为空（本项目 Redis 为必需依赖）")
	}

	// access token 比 refresh token 还长（或相等）会让续期机制失去意义：
	// 客户端的 access 还没过期，refresh 已先失效，用户仍会被强制重登。
	if Cfg.JWT.AccessExpire <= 0 {
		problems = append(problems, "jwt.access_expire 必须为正数")
	}
	if Cfg.JWT.RefreshExpire <= 0 {
		problems = append(problems, "jwt.refresh_expire 必须为正数")
	}
	if Cfg.JWT.AccessExpire > 0 && Cfg.JWT.RefreshExpire > 0 &&
		Cfg.JWT.AccessExpire >= Cfg.JWT.RefreshExpire {
		problems = append(problems, fmt.Sprintf(
			"jwt.access_expire(%d) 必须小于 jwt.refresh_expire(%d)，否则 refresh token 无意义",
			Cfg.JWT.AccessExpire, Cfg.JWT.RefreshExpire))
	}

	if Cfg.Upload.SavePath == "" {
		problems = append(problems, "upload.save_path 不能为空")
	}
	if Cfg.Upload.MaxSize <= 0 {
		problems = append(problems, "upload.max_size 必须为正数（单位 MB）")
	}

	if Cfg.Casbin.ModelPath == "" {
		problems = append(problems, "casbin.model_path 不能为空")
	}

	if len(problems) > 0 {
		return fmt.Errorf("配置校验未通过：\n  - %s", strings.Join(problems, "\n  - "))
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
