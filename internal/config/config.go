// Package config 加载 Harness 配置：内置默认值 < configs/config.yaml < .env < BH_* 环境变量。
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config 是完整配置。
type Config struct {
	Server struct {
		Addr string `yaml:"addr"`
	} `yaml:"server"`

	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DB       string `yaml:"db"`
		SSLMode  string `yaml:"sslmode"`
		Driver   string `yaml:"driver"` // memory | postgres
	} `yaml:"database"`

	Redis struct {
		Addr     string `yaml:"addr"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
		Enabled  bool   `yaml:"enabled"`
	} `yaml:"redis"`

	LLM struct {
		Provider    string  `yaml:"provider"` // mock | openai
		Model       string  `yaml:"model"`
		BaseURL     string  `yaml:"base_url"`
		APIKey      string  `yaml:"api_key"`
		Temperature float64 `yaml:"temperature"`
		MaxTokens   int     `yaml:"max_tokens"`
	} `yaml:"llm"`

	Policy struct {
		AutoApprove bool `yaml:"auto_approve"`
	} `yaml:"policy"`

	Workspace struct {
		Name            string `yaml:"name"`
		Slug            string `yaml:"slug"`
		Industry        string `yaml:"industry"`
		Region          string `yaml:"region"`
		DefaultLanguage string `yaml:"default_language"`
	} `yaml:"workspace"`
}

// Default 返回内置默认配置（mock 模式，开箱即跑）。
func Default() *Config {
	c := &Config{}
	c.Server.Addr = ":8090"
	c.Database.Host, c.Database.Port = "127.0.0.1", 5432
	c.Database.User, c.Database.Password = "postgres", "postgres"
	c.Database.DB, c.Database.SSLMode, c.Database.Driver = "b2b_marketing_agent", "disable", "memory"
	c.Redis.Addr, c.Redis.DB, c.Redis.Enabled = "127.0.0.1:6379", 0, false
	c.LLM.Provider, c.LLM.Model, c.LLM.Temperature, c.LLM.MaxTokens = "mock", "mock-1", 0.3, 1024
	c.Policy.AutoApprove = true
	c.Workspace.Name = "Microchip Taiwan Demo"
	c.Workspace.Slug = "microchip-tw"
	c.Workspace.Industry = "semiconductor"
	c.Workspace.Region = "TW"
	c.Workspace.DefaultLanguage = "zh-CN"
	return c
}

// Load 读取配置。
func Load(path string) (*Config, error) {
	loadDotEnv()
	c := Default()
	if path != "" {
		if raw, err := os.ReadFile(path); err == nil {
			if err := yaml.Unmarshal(raw, c); err != nil {
				return nil, fmt.Errorf("解析配置文件失败 %s: %w", path, err)
			}
		} else if !os.IsNotExist(err) {
			return nil, err
		}
	}
	c.applyEnv()
	return c, nil
}

// DatabaseURL 返回 PostgreSQL DSN（预留，postgres store 使用）。
func (c *Config) DatabaseURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User, c.Database.Password, c.Database.Host, c.Database.Port, c.Database.DB, c.Database.SSLMode)
}

func (c *Config) applyEnv() {
	c.Server.Addr = env("BH_SERVER_ADDR", c.Server.Addr)
	c.Database.Host = env("BH_PG_HOST", c.Database.Host)
	c.Database.Port = envInt("BH_PG_PORT", c.Database.Port)
	c.Database.User = env("BH_PG_USER", c.Database.User)
	c.Database.Password = env("BH_PG_PASSWORD", c.Database.Password)
	c.Database.DB = env("BH_PG_DB", c.Database.DB)
	c.Database.Driver = env("BH_DB_DRIVER", c.Database.Driver)

	c.Redis.Addr = env("BH_REDIS_ADDR", c.Redis.Addr)
	c.Redis.Password = env("BH_REDIS_PASSWORD", c.Redis.Password)
	c.Redis.Enabled = envBool("BH_REDIS_ENABLED", c.Redis.Enabled)

	c.LLM.Provider = env("BH_LLM_PROVIDER", c.LLM.Provider)
	c.LLM.Model = env("BH_LLM_MODEL", c.LLM.Model)
	c.LLM.BaseURL = env("BH_LLM_BASE_URL", c.LLM.BaseURL)
	c.LLM.APIKey = env("BH_LLM_API_KEY", c.LLM.APIKey)

	c.Policy.AutoApprove = envBool("BH_POLICY_AUTO_APPROVE", c.Policy.AutoApprove)
	c.Workspace.DefaultLanguage = env("BH_DEFAULT_LANGUAGE", c.Workspace.DefaultLanguage)
}

// loadDotEnv 从 .env / .env.local 读取（真实环境变量优先，只加载第一个命中文件）。
func loadDotEnv() {
	for _, p := range []string{".env", ".env.local", "../.env"} {
		raw, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(raw), "\n") {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			eq := strings.Index(line, "=")
			if eq <= 0 {
				continue
			}
			k := strings.TrimSpace(line[:eq])
			v := strings.Trim(strings.TrimSpace(line[eq+1:]), "\"'")
			if k != "" && v != "" && os.Getenv(k) == "" {
				_ = os.Setenv(k, v)
			}
		}
		return
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "on":
			return true
		case "0", "false", "no", "off":
			return false
		}
	}
	return def
}
