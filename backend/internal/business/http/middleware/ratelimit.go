package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"
)

type clientBucket struct {
	tokens     float64
	lastRefill time.Time
}

type TokenBucketLimiter struct {
	rate       float64
	capacity   float64
	mu         sync.Mutex
	clients    map[string]*clientBucket
	cleanupDur time.Duration
}

func NewTokenBucketLimiter(rate, capacity float64) *TokenBucketLimiter {
	l := &TokenBucketLimiter{
		rate:       rate,
		capacity:   capacity,
		clients:    make(map[string]*clientBucket),
		cleanupDur: 5 * time.Minute,
	}

	go l.cleanupLoop()
	return l
}

func (l *TokenBucketLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, exists := l.clients[key]
	if !exists {
		l.clients[key] = &clientBucket{
			tokens:     l.capacity - 1,
			lastRefill: now,
		}
		return true
	}

	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * l.rate
	if b.tokens > l.capacity {
		b.tokens = l.capacity
	}
	b.lastRefill = now

	if b.tokens >= 1.0 {
		b.tokens -= 1.0
		return true
	}

	return false
}

func (l *TokenBucketLimiter) cleanupLoop() {
	ticker := time.NewTicker(l.cleanupDur)
	for range ticker.C {
		l.mu.Lock()
		now := time.Now()
		for k, b := range l.clients {
			if now.Sub(b.lastRefill) > l.cleanupDur {
				delete(l.clients, k)
			}
		}
		l.mu.Unlock()
	}
}

func RateLimiter(limiter *TokenBucketLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				ip = r.RemoteAddr
			}

			if !limiter.Allow(ip) {
				http.Error(w, `{"error":"too many requests, please slow down"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
