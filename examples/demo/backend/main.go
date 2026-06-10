// Package main 是 demo 示例的 Mock 后端服务，模拟商品商城 API。
//
// 默认监听端口 18080，提供系统信息、登录（v1/v2）、用户信息、登出及商品列表接口。
// 测试账号：demo / demo123
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
)

const defaultPort = "18080"

// item 商品实体。
type item struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

var (
	// 内存中的商品数据
	items = []item{
		{ID: "1", Code: "SKU-001", Name: "Demo Item Alpha"},
		{ID: "2", Code: "SKU-002", Name: "Demo Item Beta"},
		{ID: "3", Code: "SKU-003", Name: "Demo Item Gamma"},
	}
	sessions   = map[string]string{} // token -> username 会话表
	sessionsMu sync.RWMutex
)

func main() {
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = defaultPort
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/system/info", handleSystemInfo)
	mux.HandleFunc("/api/auth/login", handleLogin)       // v1 登录，响应 {token}
	mux.HandleFunc("/api/auth/v2/login", handleLoginV2)   // v2 登录，响应嵌套 JSON
	mux.HandleFunc("/api/auth/me", handleMe)
	mux.HandleFunc("/api/auth/logout", handleLogout)
	mux.HandleFunc("/api/items", handleItems)

	addr := ":" + port
	log.Printf("backend listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, withCORS(mux)))
}

// withCORS 添加跨域头，支持前端或测试客户端跨域调用。
func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// handleSystemInfo 返回系统名称与版本，无需认证。
func handleSystemInfo(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"name":    "my-testgogogo-demo",
		"version": "1.0.0",
	})
}

// handleLogin v1 登录接口，成功返回 { "token": "..." }。
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "method not allowed"})
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "invalid body"})
		return
	}
	if req.Username != "demo" || req.Password != "demo123" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "invalid credentials"})
		return
	}

	token := "demo-token-" + req.Username
	sessionsMu.Lock()
	sessions[token] = req.Username
	sessionsMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// handleLoginV2 v2 登录接口，响应嵌套 JSON（供 demoauth Provider 测试 Token 路径提取）。
func handleLoginV2(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"message": "method not allowed"})
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"code": 400, "message": "invalid body"})
		return
	}
	if req.Username != "demo" || req.Password != "demo123" {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"code": 401, "message": "invalid credentials"})
		return
	}

	token := "demo-v2-token-" + req.Username
	sessionsMu.Lock()
	sessions[token] = req.Username
	sessionsMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"code":    0,
		"message": "ok",
		"data": map[string]string{
			"accessToken": token,
			"tokenType":   "Bearer",
		},
	})
}

// handleMe 返回当前登录用户信息，需要 Authorization 头。
func handleMe(w http.ResponseWriter, r *http.Request) {
	username, ok := currentUser(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]string{
			"username": username,
			"nickName": "Demo User",
		},
		"roles":       []string{"admin"},
		"permissions": []string{"item:read", "item:write"},
	})
}

// handleLogout 登出并清除服务端会话。
func handleLogout(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}
	if token != "" {
		sessionsMu.Lock()
		delete(sessions, token)
		sessionsMu.Unlock()
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// handleItems 分页查询商品列表，支持按 code 过滤。
func handleItems(w http.ResponseWriter, r *http.Request) {
	if _, ok := currentUser(r); !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
		return
	}

	pageNum, _ := strconv.Atoi(r.URL.Query().Get("pageNum"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	code := r.URL.Query().Get("code")
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	filtered := items
	if code != "" {
		filtered = make([]item, 0)
		for _, it := range items {
			if it.Code == code {
				filtered = append(filtered, it)
			}
		}
	}

	total := len(filtered)
	start := (pageNum - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"current": pageNum,
		"size":    pageSize,
		"total":   total,
		"pages":   pages(total, pageSize),
		"records": filtered[start:end],
	})
}

// currentUser 从 Authorization 头解析 Token 并查找对应用户名。
func currentUser(r *http.Request) (string, bool) {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if token == "" {
		return "", false
	}
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	username, ok := sessions[token]
	return username, ok
}

// pages 计算总页数。
func pages(total, size int) int {
	if size <= 0 {
		return 0
	}
	p := total / size
	if total%size != 0 {
		p++
	}
	return p
}

// writeJSON 写入 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
