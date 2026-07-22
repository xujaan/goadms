package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	Scheduler SchedulerConfig `yaml:"scheduler"`
	Webhook  WebhookConfig  `yaml:"webhook"`
}

type ServerConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	ReadTimeout  string `yaml:"read_timeout"`
	WriteTimeout string `yaml:"write_timeout"`
	IdleTimeout  string `yaml:"idle_timeout"`
}

func (s ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (s ServerConfig) ReadTimeoutDuration() time.Duration {
	d, _ := time.ParseDuration(s.ReadTimeout)
	if d == 0 {
		d = 30 * time.Second
	}
	return d
}

func (s ServerConfig) IdleTimeoutDuration() time.Duration {
	d, _ := time.ParseDuration(s.IdleTimeout)
	if d == 0 {
		d = 60 * time.Second
	}
	return d
}

func (s ServerConfig) WriteTimeoutDuration() time.Duration {
	d, _ := time.ParseDuration(s.WriteTimeout)
	if d == 0 {
		d = 30 * time.Second
	}
	return d
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"sslmode"`
	PoolMax  int    `yaml:"pool_max"`
}

func (d DatabaseConfig) DSN() string {
	sslmode := d.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, sslmode,
	)
}

type AuthConfig struct {
	JWTSecret          string `yaml:"jwt_secret"`
	AccessTokenTTL     string `yaml:"access_token_ttl"`
	RefreshTokenTTL    string `yaml:"refresh_token_ttl"`
}

func (a AuthConfig) AccessTTLDuration() time.Duration {
	d, _ := time.ParseDuration(a.AccessTokenTTL)
	if d == 0 {
		d = 15 * time.Minute
	}
	return d
}

func (a AuthConfig) RefreshTTLDuration() time.Duration {
	d, _ := time.ParseDuration(a.RefreshTokenTTL)
	if d == 0 {
		d = 7 * 24 * time.Hour
	}
	return d
}

type SchedulerConfig struct {
	OnlineCheckInterval    string `yaml:"online_check_interval"`
	AutoPullInterval       string `yaml:"auto_pull_interval"`
	StatusPingInterval     string `yaml:"status_ping_interval"`
	WebhookRetryMax        int    `yaml:"webhook_retry_max"`
	WebhookRetryBaseDelay  string `yaml:"webhook_retry_base_delay"`
}

func (s SchedulerConfig) OnlineCheckDuration() time.Duration {
	d, _ := time.ParseDuration(s.OnlineCheckInterval)
	if d == 0 {
		d = 60 * time.Second
	}
	return d
}

func (s SchedulerConfig) AutoPullDuration() time.Duration {
	d, _ := time.ParseDuration(s.AutoPullInterval)
	if d == 0 {
		d = 5 * time.Minute
	}
	return d
}

func (s SchedulerConfig) StatusPingDuration() time.Duration {
	d, _ := time.ParseDuration(s.StatusPingInterval)
	if d == 0 {
		d = 30 * time.Second
	}
	return d
}

type WebhookConfig struct {
	RetryMaxAttempts    int    `yaml:"retry_max_attempts"`
	RetryBaseDelay      string `yaml:"retry_base_delay"`
	Timeout             string `yaml:"timeout"`
}

func (w WebhookConfig) RetryMax() int {
	if w.RetryMaxAttempts == 0 {
		return 5
	}
	return w.RetryMaxAttempts
}

func (w WebhookConfig) RetryBaseDuration() time.Duration {
	d, _ := time.ParseDuration(w.RetryBaseDelay)
	if d == 0 {
		d = 1 * time.Minute
	}
	return d
}

func (w WebhookConfig) TimeoutDuration() time.Duration {
	d, _ := time.ParseDuration(w.Timeout)
	if d == 0 {
		d = 30 * time.Second
	}
	return d
}

// Load reads config from YAML file, then overrides with env vars.
func Load(path string) (*Config, error) {
	cfg := &Config{}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	cfg.applyEnvOverrides()
	return cfg, nil
}

func (c *Config) applyEnvOverrides() {
	if v := os.Getenv("ADMS_DB_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("ADMS_DB_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &c.Database.Port)
	}
	if v := os.Getenv("ADMS_DB_USER"); v != "" {
		c.Database.User = v
	}
	if v := os.Getenv("ADMS_DB_PASSWORD"); v != "" {
		c.Database.Password = v
	}
	if v := os.Getenv("ADMS_DB_NAME"); v != "" {
		c.Database.Name = v
	}
	if v := os.Getenv("ADMS_JWT_SECRET"); v != "" {
		c.Auth.JWTSecret = v
	}
	if v := os.Getenv("ADMS_SERVER_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &c.Server.Port)
	}
}
