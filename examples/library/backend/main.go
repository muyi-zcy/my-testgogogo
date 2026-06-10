// Package main 是 library 示例的 Mock 后端服务，模拟图书管理 API。
//
// 默认监听端口 18081，提供系统信息、登录、用户信息、登出及图书列表接口。
// 测试账号：librarian / lib123
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

const defaultPort = "18081"

// book 图书实体。
type book struct {
	ID     string `json:"id"`
	ISBN   string `json:"isbn"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

var (
	books = []book{
		{ID: "1", ISBN: "978-001", Title: "Go 语言入门", Author: "张三"},
		{ID: "2", ISBN: "978-002", Title: "API 测试实践", Author: "李四"},
		{ID: "3", ISBN: "978-003", Title: "分布式系统", Author: "王五"},
	}
	sessions   = map[string]string{}
	sessionsMu sync.RWMutex
)

func main() {
	port := os.Getenv("BACKEND_PORT")
	if port == "" {
		port = defaultPort
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/system/info", handleSystemInfo)
	mux.HandleFunc("/api/auth/login", handleLogin)
	mux.HandleFunc("/api/auth/me", handleMe)
	mux.HandleFunc("/api/auth/logout", handleLogout)
	mux.HandleFunc("/api/books", handleBooks)

	addr := ":" + port
	log.Printf("library backend listening on http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// handleSystemInfo 返回系统名称与版本。
func handleSystemInfo(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"name":    "my-testgogogo-library",
		"version": "1.0.0",
	})
}

// handleLogin 标准登录接口，成功返回 { "token": "..." }，供内置 login Provider 使用。
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
	if req.Username != "librarian" || req.Password != "lib123" {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "invalid credentials"})
		return
	}

	token := "lib-token-" + req.Username
	sessionsMu.Lock()
	sessions[token] = req.Username
	sessionsMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

// handleMe 返回当前管理员信息。
func handleMe(w http.ResponseWriter, r *http.Request) {
	username, ok := currentUser(r)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user": map[string]string{
			"username": username,
			"nickName": "图书管理员",
		},
		"roles":       []string{"librarian"},
		"permissions": []string{"book:read", "book:borrow"},
	})
}

// handleLogout 登出并清除会话。
func handleLogout(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if token != "" {
		sessionsMu.Lock()
		delete(sessions, token)
		sessionsMu.Unlock()
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "ok"})
}

// handleBooks 分页查询图书列表，支持按 ISBN 过滤。
func handleBooks(w http.ResponseWriter, r *http.Request) {
	if _, ok := currentUser(r); !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "unauthorized"})
		return
	}

	pageNum, _ := strconv.Atoi(r.URL.Query().Get("pageNum"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("pageSize"))
	isbn := r.URL.Query().Get("isbn")
	if pageNum <= 0 {
		pageNum = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	filtered := books
	if isbn != "" {
		filtered = make([]book, 0)
		for _, b := range books {
			if b.ISBN == isbn {
				filtered = append(filtered, b)
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

// currentUser 从 Authorization 头解析 Token。
func currentUser(r *http.Request) (string, bool) {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if token == "" {
		return "", false
	}
	sessionsMu.RLock()
	defer sessionsMu.RUnlock()
	username, ok := sessions[token]
	return username, ok
}

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

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
