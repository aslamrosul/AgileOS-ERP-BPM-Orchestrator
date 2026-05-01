package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"agileos-backend/logger"

	"github.com/gin-gonic/gin"
)

// IPFilter manages IP blacklist and whitelist
type IPFilter struct {
	blacklist map[string]time.Time
	whitelist map[string]bool
	mu        sync.RWMutex
}

var ipFilter *IPFilter

// InitIPFilter initializes the IP filter
func InitIPFilter() {
	ipFilter = &IPFilter{
		blacklist: make(map[string]time.Time),
		whitelist: make(map[string]bool),
	}

	// Add localhost to whitelist
	ipFilter.whitelist["127.0.0.1"] = true
	ipFilter.whitelist["::1"] = true

	// Start cleanup goroutine to remove expired blacklist entries
	go ipFilter.cleanupExpiredEntries()

	logger.Log.Info().Msg("✓ IP filter initialized")
}

// cleanupExpiredEntries removes expired blacklist entries every hour
func (f *IPFilter) cleanupExpiredEntries() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		f.mu.Lock()
		now := time.Now()
		for ip, expiry := range f.blacklist {
			if now.After(expiry) {
				delete(f.blacklist, ip)
				logger.Log.Info().
					Str("ip", ip).
					Msg("Removed expired IP from blacklist")
			}
		}
		f.mu.Unlock()
	}
}

// BlacklistIP adds an IP to the blacklist for specified duration
func (f *IPFilter) BlacklistIP(ip string, duration time.Duration) {
	f.mu.Lock()
	defer f.mu.Unlock()

	expiry := time.Now().Add(duration)
	f.blacklist[ip] = expiry

	logger.LogSecurity("ip_blacklisted", "", ip, map[string]interface{}{
		"duration": duration.String(),
		"expiry":   expiry,
	})
}

// WhitelistIP adds an IP to the whitelist
func (f *IPFilter) WhitelistIP(ip string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.whitelist[ip] = true

	logger.Log.Info().
		Str("ip", ip).
		Msg("IP added to whitelist")
}

// IsBlacklisted checks if an IP is blacklisted
func (f *IPFilter) IsBlacklisted(ip string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	expiry, exists := f.blacklist[ip]
	if !exists {
		return false
	}

	// Check if blacklist entry has expired
	if time.Now().After(expiry) {
		return false
	}

	return true
}

// IsWhitelisted checks if an IP is whitelisted
func (f *IPFilter) IsWhitelisted(ip string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.whitelist[ip]
}

// IPFilterMiddleware blocks blacklisted IPs
func IPFilterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Whitelisted IPs always pass
		if ipFilter.IsWhitelisted(ip) {
			c.Next()
			return
		}

		// Check if IP is blacklisted
		if ipFilter.IsBlacklisted(ip) {
			logger.LogSecurity("blocked_blacklisted_ip", "", ip, map[string]interface{}{
				"endpoint":   c.Request.URL.Path,
				"method":     c.Request.Method,
				"user_agent": c.Request.UserAgent(),
			})

			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"message": "Your IP address has been blocked due to suspicious activity.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TrustedProxyMiddleware validates requests are from trusted proxies
func TrustedProxyMiddleware(trustedProxies []string) gin.HandlerFunc {
	trustedNets := make([]*net.IPNet, 0, len(trustedProxies))

	for _, proxy := range trustedProxies {
		_, ipNet, err := net.ParseCIDR(proxy)
		if err != nil {
			// Try parsing as single IP
			ip := net.ParseIP(proxy)
			if ip != nil {
				mask := net.CIDRMask(32, 32)
				if ip.To4() == nil {
					mask = net.CIDRMask(128, 128)
				}
				ipNet = &net.IPNet{IP: ip, Mask: mask}
			}
		}
		if ipNet != nil {
			trustedNets = append(trustedNets, ipNet)
		}
	}

	return func(c *gin.Context) {
		// Get real client IP from X-Forwarded-For or X-Real-IP
		realIP := c.Request.Header.Get("X-Forwarded-For")
		if realIP == "" {
			realIP = c.Request.Header.Get("X-Real-IP")
		}
		if realIP == "" {
			realIP = c.ClientIP()
		}

		// Parse IP
		ip := net.ParseIP(realIP)
		if ip == nil {
			c.Next()
			return
		}

		// Check if request is from trusted proxy
		isTrusted := false
		for _, ipNet := range trustedNets {
			if ipNet.Contains(ip) {
				isTrusted = true
				break
			}
		}

		if !isTrusted && len(trustedNets) > 0 {
			logger.LogSecurity("untrusted_proxy_request", "", realIP, map[string]interface{}{
				"endpoint": c.Request.URL.Path,
			})
		}

		c.Next()
	}
}

// GetIPFilter returns the global IP filter instance
func GetIPFilter() *IPFilter {
	return ipFilter
}
