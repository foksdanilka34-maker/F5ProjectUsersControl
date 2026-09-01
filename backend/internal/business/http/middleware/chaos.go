package middleware

import (
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

func Chaos(next http.Handler) http.Handler {
	chaosEnabled := os.Getenv("ENABLE_CHAOS") == "true"
	if !chaosEnabled {
		return next
	}

	failureRatePercent, _ := strconv.Atoi(os.Getenv("CHAOS_FAILURE_RATE"))
	maxLatencyMs, _ := strconv.Atoi(os.Getenv("CHAOS_MAX_LATENCY_MS"))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if maxLatencyMs > 0 {
			sleepTime := time.Duration(rand.Intn(maxLatencyMs)) * time.Millisecond
			time.Sleep(sleepTime)
		}

		if failureRatePercent > 0 && rand.Intn(100) < failureRatePercent {
			log.Printf("[Chaos Middleware] Injected random failure on %s %s", r.Method, r.URL.Path)
			http.Error(w, `{"error":"chaos: simulated server error"}`, http.StatusInternalServerError)
			return
		}

		next.ServeHTTP(w, r)
	})
}
