// Package auth 提供认证 Provider 注册、Token 缓存与统一认证入口。
package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/muyi-zcy/my-testgogogo/config"
)

// cachedToken 是写入本地缓存文件的 JSON 结构，包含 Token 及过期元数据。
type cachedToken struct {
	Key       string    `json:"key"`        // 缓存键，由 Config.CacheKey() 生成
	Token     string    `json:"token"`      // 认证 Token
	Username  string    `json:"username"`   // 登录用户名，便于排查
	BaseURL   string    `json:"base_url"`   // 接口基地址
	Env       string    `json:"env"`        // 当前激活环境（如 local）
	CreatedAt time.Time `json:"created_at"` // 缓存创建时间
	ExpiresAt time.Time `json:"expires_at"` // 缓存过期时间
}

// TokenStore 管理 Token 的本地文件缓存，避免每次测试都重新登录。
type TokenStore struct {
	cfg config.AuthCacheConfig
}

// NewTokenStore 根据 AuthCacheConfig 创建 TokenStore 实例。
func NewTokenStore(cfg config.AuthCacheConfig) *TokenStore {
	return &TokenStore{cfg: cfg}
}

// Load 从本地缓存文件读取 Token。缓存未启用、文件不存在、键不匹配或已过期时返回 false。
func (s *TokenStore) Load(appCfg *config.Config) (string, bool) {
	if !s.cfg.Enabled {
		return "", false
	}

	data, err := os.ReadFile(s.filePath(appCfg))
	if err != nil {
		return "", false
	}

	var cached cachedToken
	if err := json.Unmarshal(data, &cached); err != nil {
		return "", false
	}

	// 校验缓存键是否匹配当前环境/用户/基地址组合
	if cached.Key != appCfg.CacheKey() || cached.Token == "" {
		return "", false
	}
	if time.Now().After(cached.ExpiresAt) {
		_ = s.Clear(appCfg)
		return "", false
	}

	return cached.Token, true
}

// Save 将 Token 写入本地缓存文件，目录权限 0700，文件权限 0600。
func (s *TokenStore) Save(appCfg *config.Config, token string) error {
	if !s.cfg.Enabled || token == "" {
		return nil
	}

	if err := os.MkdirAll(s.cfg.Dir, 0o700); err != nil {
		return fmt.Errorf("create cache dir: %w", err)
	}

	now := time.Now()
	cached := cachedToken{
		Key:       appCfg.CacheKey(),
		Token:     token,
		Username:  appCfg.User.Username,
		BaseURL:   appCfg.BaseURL,
		Env:       appCfg.Active,
		CreatedAt: now,
		ExpiresAt: now.Add(s.cfg.TTL),
	}

	payload, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal token cache: %w", err)
	}

	path := s.filePath(appCfg)
	if err := os.WriteFile(path, payload, 0o600); err != nil {
		return fmt.Errorf("write token cache: %w", err)
	}
	return nil
}

// Clear 删除当前配置对应的缓存文件；文件不存在时不报错。
func (s *TokenStore) Clear(appCfg *config.Config) error {
	if !s.cfg.Enabled {
		return nil
	}
	err := os.Remove(s.filePath(appCfg))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove token cache: %w", err)
	}
	return nil
}

// filePath 根据 CacheKey 生成缓存文件路径，格式：<cache_dir>/<key>.token。
func (s *TokenStore) filePath(appCfg *config.Config) string {
	fileName := fmt.Sprintf("%s.token", sanitizeFileName(appCfg.CacheKey()))
	return filepath.Join(s.cfg.Dir, fileName)
}

// sanitizeFileName 将 CacheKey 中的非法文件名字符替换为下划线。
func sanitizeFileName(value string) string {
	replacer := []rune(value)
	for i, r := range replacer {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|', ' ':
			replacer[i] = '_'
		}
	}
	return string(replacer)
}
