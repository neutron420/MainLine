package interceptor

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type MetricsCollector struct {
	mu                  sync.RWMutex
	requestsTotal       map[string]int64
	requestDuration     map[string]float64
	activeSubscriptions int64
}

var DefaultMetrics = NewMetricsCollector()

func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{
		requestsTotal:   make(map[string]int64),
		requestDuration: make(map[string]float64),
	}
}

func MetricsInterceptor(collector *MetricsCollector) grpc.UnaryServerInterceptor {
	if collector == nil {
		collector = DefaultMetrics
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start).Seconds()

		method := info.FullMethod
		code := strconv.Itoa(int(status.Code(err)))

		collector.mu.Lock()
		collector.requestsTotal[method+":"+code]++
		collector.requestDuration[method+":"+code] += duration
		collector.mu.Unlock()

		if strings.Contains(method, "Subscribe") && err == nil {
		}

		return resp, err
	}
}

func (c *MetricsCollector) GetRequestCount(method, code string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.requestsTotal[method+":"+code]
}

func (c *MetricsCollector) GetTotalRequests() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var total int64
	for _, v := range c.requestsTotal {
		total += v
	}
	return total
}

func (c *MetricsCollector) GetAvgDuration(method, code string) float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	count := c.requestsTotal[method+":"+code]
	if count == 0 {
		return 0
	}
	return c.requestDuration[method+":"+code] / float64(count)
}
