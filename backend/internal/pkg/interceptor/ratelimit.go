package interceptor

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64
	lastRefill time.Time
}

func newTokenBucket(rate float64, burst int) *tokenBucket {
	return &tokenBucket{
		tokens:     float64(burst),
		maxTokens:  float64(burst),
		refillRate: rate,
		lastRefill: time.Now(),
	}
}

func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens = min(tb.maxTokens, tb.tokens+elapsed*tb.refillRate)
	tb.lastRefill = now

	if tb.tokens < 1 {
		return false
	}

	tb.tokens--
	return true
}

type RateLimiter struct {
	mu        sync.RWMutex
	buckets   map[string]*tokenBucket
	rate      float64
	burst     int
	stopCh    chan struct{}
	startOnce sync.Once
}

func NewRateLimiter(rate float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		buckets: make(map[string]*tokenBucket),
		rate:    rate,
		burst:   burst,
		stopCh:  make(chan struct{}),
	}
	rl.startCleanup()
	return rl
}

func (rl *RateLimiter) startCleanup() {
	rl.startOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					rl.mu.Lock()
					rl.buckets = make(map[string]*tokenBucket)
					rl.mu.Unlock()
				case <-rl.stopCh:
					return
				}
			}
		}()
	})
}

func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

func (rl *RateLimiter) getBucket(key string) *tokenBucket {
	rl.mu.RLock()
	bucket, ok := rl.buckets[key]
	rl.mu.RUnlock()
	if ok {
		return bucket
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	if bucket, ok := rl.buckets[key]; ok {
		return bucket
	}

	bucket = newTokenBucket(rl.rate, rl.burst)
	rl.buckets[key] = bucket
	return bucket
}

func (rl *RateLimiter) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		key := rl.resolveKey(ctx)
		bucket := rl.getBucket(key)
		if !bucket.allow() {
			return nil, status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}
		return handler(ctx, req)
	}
}

func (rl *RateLimiter) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		key := rl.resolveKey(stream.Context())
		bucket := rl.getBucket(key)
		if !bucket.allow() {
			return status.Error(codes.ResourceExhausted, "rate limit exceeded")
		}
		return handler(srv, stream)
	}
}

func (rl *RateLimiter) resolveKey(ctx context.Context) string {
	userID, err := UserIDFromContext(ctx)
	if err == nil && userID != "" {
		return fmt.Sprintf("user:%s", userID)
	}

	ip := extractIP(ctx)
	if ip != "" {
		return fmt.Sprintf("ip:%s", ip)
	}

	return "unknown"
}

func extractIP(ctx context.Context) string {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return ""
	}
	tcpAddr, ok := p.Addr.(*net.TCPAddr)
	if !ok {
		return ""
	}
	return tcpAddr.IP.String()
}
