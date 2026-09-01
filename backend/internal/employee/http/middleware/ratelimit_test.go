package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/foksdanilka34-maker/F5ProjectUsersControl/internal/employee/http/middleware"
)

func TestTokenBucketLimiter_Allow(t *testing.T) {
	// Rate = 5 tokens/sec, Capacity = 5
	limiter := middleware.NewTokenBucketLimiter(5, 5)

	// First 5 requests should be allowed immediately (consuming burst capacity)
	for i := 0; i < 5; i++ {
		if !limiter.Allow("127.0.0.1") {
			t.Fatalf("request %d should be allowed under burst capacity", i+1)
		}
	}

	// 6th request should be rejected immediately
	if limiter.Allow("127.0.0.1") {
		t.Fatalf("6th request should be rate limited")
	}

	// Different IP should still be allowed
	if !limiter.Allow("192.168.1.1") {
		t.Fatalf("different IP should have its own token bucket")
	}

	// Wait 250ms -> at 5 tokens/sec, should refill >1 token
	time.Sleep(250 * time.Millisecond)
	if !limiter.Allow("127.0.0.1") {
		t.Fatalf("request after refill should be allowed")
	}
}

func TestRateLimiterMiddleware_ConcurrentRequests(t *testing.T) {
	limiter := middleware.NewTokenBucketLimiter(10, 10)
	handler := middleware.RateLimiter(limiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var allowedCount int32
	var blockedCount int32

	var wg sync.WaitGroup
	totalRequests := 30
	wg.Add(totalRequests)

	for i := 0; i < totalRequests; i++ {
		go func() {
			defer wg.Done()
			req := httptest.NewRequest("GET", "/api/test", nil)
			req.RemoteAddr = "10.0.0.1:12345"
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				atomic.AddInt32(&allowedCount, 1)
			} else if w.Code == http.StatusTooManyRequests {
				atomic.AddInt32(&blockedCount, 1)
			}
		}()
	}

	wg.Wait()

	if allowedCount != 10 {
		t.Errorf("expected exactly 10 requests allowed, got %d", allowedCount)
	}
	if blockedCount != 20 {
		t.Errorf("expected 20 requests blocked (429), got %d", blockedCount)
	}
}
