package redis

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"
	"strings"

	"github.com/redis/go-redis/v9"
)

func Connect(ctx context.Context, rawURL string) (*redis.Client, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parsing redis url: %w", err)
	}

	password, _ := u.User.Password()

	host := u.Host
	if strings.Contains(host, ",") {
		host = strings.Split(host, ",")[0]
	}

	opts := &redis.Options{
		Addr:     host,
		Password: password,
	}

	if u.Scheme == "rediss" {
		opts.TLSConfig = &tls.Config{ServerName: u.Hostname()}
	}

	rdb := redis.NewClient(opts)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("pinging redis: %w", err)
	}

	return rdb, nil
}
