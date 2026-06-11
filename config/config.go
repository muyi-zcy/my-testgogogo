// Package config 负责加载多环境 YAML 配置，合并根配置与环境配置，并提供认证相关默认值。
//
// 配置文件结构：
//   - configs/config.yaml       根配置（active 环境、认证、报告、test.vars 等）
//   - configs/<env>.yaml        环境配置（base_url、用户、Token、vars 等）
//   - configs/<env>.override.yaml  可选覆盖文件（本地私密配置，不入库）
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/muyi-zcy/my-testgogogo/path"
	"gopkg.in/yaml.v3"
)

const defaultActive = "local"

// RootConfig 对应 configs/config.yaml 的根配置结构。
type RootConfig struct {
	Active string         `yaml:"active"` // 当前激活环境名，如 local、staging
	Test   TestRootConfig `yaml:"test"`
	Auth   AuthRootConfig `yaml:"auth"`
}

// TestRootConfig 测试相关全局开关。
type TestRootConfig struct {
	SkipIntegration bool           `yaml:"skip_integration"` // true 时跳过所有集成测试
	Vars            map[string]any `yaml:"vars"`             // 跨环境共享的全局变量
}

// AuthRootConfig 认证 Provider 及 Token 缓存配置。
type AuthRootConfig struct {
	Provider      string          `yaml:"provider"`        // 认证方式：login / static_token / 自定义
	CacheEnabled  bool            `yaml:"cache_enabled"`   // 是否启用 Token 本地缓存
	CacheDir      string          `yaml:"cache_dir"`       // 缓存目录，默认 .cache/auth
	CacheTTLHours int             `yaml:"cache_ttl_hours"` // 缓存有效期（小时），默认 168
	Login         LoginAuthConfig `yaml:"login"`           // login Provider 专用配置
}

// LoginAuthConfig login Provider 的登录接口与 Token 提取配置。
type LoginAuthConfig struct {
	URL         string            `yaml:"url"`          // 登录接口路径
	Method      string            `yaml:"method"`       // HTTP 方法，默认 POST
	WithoutAuth bool              `yaml:"without_auth"` // 登录请求是否不带已有 Token
	Body        map[string]string `yaml:"body"`         // 请求体模板，支持 {{user.username}} 变量
	TokenPath   string            `yaml:"token_path"`   // 响应 JSON 中 Token 的点分路径
	Validate    ValidateConfig    `yaml:"validate"`     // 可选的 Token 远程校验端点
}

// ValidateConfig Token 远程校验端点配置。
type ValidateConfig struct {
	URL    string `yaml:"url"`
	Method string `yaml:"method"`
}

// UserConfig 测试用账号信息。
type UserConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Type     string `yaml:"type"` // 用户类型，默认 app
}

// EnvConfig 对应 configs/<env>.yaml 的环境配置结构。
type EnvConfig struct {
	BaseURL        string         `yaml:"base_url"`
	TimeoutSeconds int            `yaml:"timeout_seconds"`
	User           UserConfig     `yaml:"user"`
	Token          string         `yaml:"token"`        // static_token 模式使用
	CaptchaCode    string         `yaml:"captcha_code"` // 验证码（如需要）
	Vars           map[string]any `yaml:"vars"`         // 环境级全局变量，覆盖 test.vars 同名项
}

// Config 是合并后的运行时配置，供测试框架各模块使用。
type Config struct {
	Active          string
	BaseURL         string
	Timeout         time.Duration
	User            UserConfig
	Token           string
	CaptchaCode     string
	Vars            map[string]any // 合并后的全局变量：test.vars <- env vars
	SkipIntegration bool
	Auth            AuthRootConfig
	AuthCache       AuthCacheConfig
}

// AuthCacheConfig Token 本地缓存的运行时配置（含绝对路径）。
type AuthCacheConfig struct {
	Enabled  bool
	Dir      string
	TTL      time.Duration
	RootPath string // 项目根目录
}

// Load 从项目 configs 目录加载并合并配置，应用默认值与校验规则。
func Load() (*Config, error) {
	root, err := path.ModuleRoot()
	if err != nil {
		return nil, err
	}

	rootCfg, err := loadRootConfig(root)
	if err != nil {
		return nil, err
	}

	active := rootCfg.Active
	if active == "" {
		active = defaultActive
	}

	envCfg, err := loadEnvConfig(root, active)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Active:          active,
		BaseURL:         envCfg.BaseURL,
		User:            envCfg.User,
		Token:           envCfg.Token,
		CaptchaCode:     envCfg.CaptchaCode,
		Vars:            mergeVars(rootCfg.Test.Vars, envCfg.Vars),
		SkipIntegration: rootCfg.Test.SkipIntegration,
		Auth:            rootCfg.Auth,
	}

	if envCfg.TimeoutSeconds <= 0 {
		cfg.Timeout = 15 * time.Second
	} else {
		cfg.Timeout = time.Duration(envCfg.TimeoutSeconds) * time.Second
	}

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base_url is required in configs/%s.yaml", active)
	}

	// 推断默认 Provider：有 Token 用 static_token，否则用 login
	provider := cfg.Auth.Provider
	if provider == "" {
		if cfg.Token != "" {
			provider = "static_token"
		} else {
			provider = "login"
		}
		cfg.Auth.Provider = provider
	}

	// 按 Provider 类型校验必填字段
	if provider == "static_token" && cfg.Token == "" {
		return nil, fmt.Errorf("token is required when auth.provider is static_token")
	}
	if provider == "login" && (cfg.User.Username == "" || cfg.User.Password == "") {
		return nil, fmt.Errorf("user.username and user.password are required when auth.provider is login")
	}
	if provider == "demoauth" && cfg.Token == "" && (cfg.User.Username == "" || cfg.User.Password == "") {
		return nil, fmt.Errorf("user.username and user.password are required when auth.provider is demoauth and token is empty")
	}
	if provider == "login" && cfg.Auth.Login.URL == "" {
		return nil, fmt.Errorf("auth.login.url is required when auth.provider is login")
	}
	if provider == "login" && cfg.Auth.Login.TokenPath == "" {
		cfg.Auth.Login.TokenPath = "token"
	}

	if cfg.User.Type == "" {
		cfg.User.Type = "app"
	}

	// 解析缓存目录为绝对路径
	cacheDir := rootCfg.Auth.CacheDir
	if cacheDir == "" {
		cacheDir = ".cache/auth"
	}
	if !filepath.IsAbs(cacheDir) {
		cacheDir = filepath.Join(root, cacheDir)
	}

	ttlHours := rootCfg.Auth.CacheTTLHours
	if ttlHours <= 0 {
		ttlHours = 168
	}

	cfg.AuthCache = AuthCacheConfig{
		Enabled:  rootCfg.Auth.CacheEnabled,
		Dir:      cacheDir,
		TTL:      time.Duration(ttlHours) * time.Hour,
		RootPath: root,
	}

	return cfg, nil
}

// loadRootConfig 读取 configs/config.yaml。
func loadRootConfig(root string) (*RootConfig, error) {
	path := filepath.Join(root, "configs", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg RootConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return &cfg, nil
}

// loadEnvConfig 读取环境配置，并尝试合并 override 文件。
func loadEnvConfig(root, active string) (*EnvConfig, error) {
	path := filepath.Join(root, "configs", active+".yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read env config %s: %w", path, err)
	}

	var cfg EnvConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse env config %s: %w", path, err)
	}

	// 本地覆盖文件（如含真实密码），存在则合并
	overridePath := filepath.Join(root, "configs", active+".override.yaml")
	if overrideData, err := os.ReadFile(overridePath); err == nil {
		if err := yaml.Unmarshal(overrideData, &cfg); err != nil {
			return nil, fmt.Errorf("parse override config %s: %w", overridePath, err)
		}
	}

	return &cfg, nil
}

// CacheKey 生成 Token 缓存的唯一键，由环境名、基地址和用户名组成。
func (c *Config) CacheKey() string {
	return fmt.Sprintf("%s_%s_%s", c.Active, c.BaseURL, c.User.Username)
}

// SkipIntegration 快捷方法：读取配置并返回是否跳过集成测试。
func SkipIntegration() bool {
	cfg, err := Load()
	if err != nil {
		return false
	}
	return cfg.SkipIntegration
}
