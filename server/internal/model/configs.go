package model

// Config 总配置结构体  区分成数据缓存以及数据存储层
type Config struct {
	Database DatabaseConfig `mapstructure:"database"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Server   Server         `mapstructure:"server"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Email    EmailConfig    `mapstructure:"email"`
	LLM      LLMConfig      `mapstructure:"llm"`
}
type Server struct {
	Port      int    `mapstructure:"port"`
	Domain    string `mapstructure:"domain"`
	DevDomain string `mapstructure:"dev_domain"`
}

// GetDomain 返回当前环境使用的域名
// 如果配置了 dev_domain 则使用（本地开发），否则使用 domain（线上）
func (s Server) GetDomain() string {
	if s.DevDomain != "" {
		return s.DevDomain
	}
	return s.Domain
}

type JWTConfig struct {
	Secret             string `mapstructure:"secret"`
	AccessTokenExpire  string `mapstructure:"access_token_expire"`
	RefreshTokenExpire string `mapstructure:"refresh_token_expire"`
}

// EmailConfig 全局邮件配置（不含用户凭证）
type EmailConfig struct {
	PollInterval  string `mapstructure:"poll_interval"`
	AttachmentDir string `mapstructure:"attachment_dir"`
}

// LLMConfig 大模型调用配置
type LLMConfig struct {
	Provider string `mapstructure:"provider"` // qwen | doubao | glm
	APIKey   string `mapstructure:"api_key"`
	BaseURL  string `mapstructure:"base_url"`
	Model    string `mapstructure:"model"`
}

type DatabaseConfig struct {
	Postgres PostgresConfig `mapstructure:"postgres"`
}

type PostgresConfig struct {
	Host            string `mapstructure:"host"`
	Port            int    `mapstructure:"port"`
	User            string `mapstructure:"user"`
	Password        string `mapstructure:"password"`
	DBName          string `mapstructure:"dbname"`
	SSLMode         string `mapstructure:"ssl_mode"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime string `mapstructure:"conn_max_lifetime"`
	ConnectTimeout  string `mapstructure:"connect_timeout"`
}

type CacheConfig struct {
	Redis RedisConfig `mapstructure:"redis"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Password     string `mapstructure:"password"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
	DialTimeout  string `mapstructure:"dial_timeout"`
	ReadTimeout  string `mapstructure:"read_timeout"`
	WriteTimeout string `mapstructure:"write_timeout"`
	IsCluster    bool   `mapstructure:"is_cluster"`
}
