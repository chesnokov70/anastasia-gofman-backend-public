package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"anastasia_gofman_backend/pkg/config"
)

type attemptInfo struct {
	count        int
	blockedUntil time.Time
	lastAttempt  time.Time
}

var (
	attemptsMu   sync.Mutex
	attemptsByIP = make(map[string]*attemptInfo)

	maxAttempts   = 10
	window        = 10 * time.Minute
	blockDuration = 600 * time.Minute
)

func checkAdminToken(c *gin.Context) bool {
	cfg := config.GetConfig()
	expected := ""
	if cfg != nil {
		expected = cfg.Server.AdminToken
	}
	if expected == "" {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "admin only"})
		return false
	}

	ip := c.ClientIP()
	now := time.Now()

	attemptsMu.Lock()
	info, ok := attemptsByIP[ip]
	if !ok {
		info = &attemptInfo{}
		attemptsByIP[ip] = info
	}
	if now.Before(info.blockedUntil) {
		attemptsMu.Unlock()
		c.AbortWithStatus(http.StatusTooManyRequests)
		return false
	}
	attemptsMu.Unlock()

	provided := c.GetHeader("X-Auth")
	if provided == expected {
		attemptsMu.Lock()
		info.count = 0
		info.lastAttempt = now
		attemptsMu.Unlock()
		return true
	}

	attemptsMu.Lock()
	if now.Sub(info.lastAttempt) > window {
		info.count = 0
	}
	info.count++
	info.lastAttempt = now
	if info.count >= maxAttempts {
		info.blockedUntil = now.Add(blockDuration)
		info.count = 0
	}
	attemptsMu.Unlock()

	c.AbortWithStatus(http.StatusUnauthorized)
	return false
}

func RequireAdminToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !checkAdminToken(c) {
			return
		}
		c.Next()
	}
}

func RequireAdminTokenForUnsafeMethods() gin.HandlerFunc {
	return func(c *gin.Context) {
		m := c.Request.Method
		if m == http.MethodDelete || m == http.MethodPut || m == http.MethodPatch {
			if !checkAdminToken(c) {
				return
			}
		}
		c.Next()
	}
}
