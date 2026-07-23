package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/schemahub/backend/internal/auth/domain"
	authHandler "github.com/schemahub/backend/internal/auth/handler"
	"github.com/schemahub/backend/internal/auth/repository/postgres"
	"github.com/schemahub/backend/internal/pkg/config"
	"github.com/schemahub/backend/internal/pkg/database"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	"github.com/schemahub/backend/internal/pkg/jwt"
	"github.com/schemahub/backend/internal/pkg/logger"
	"github.com/schemahub/backend/internal/pkg/redis"
	authv1 "github.com/schemahub/backend/proto/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	_ = godotenv.Load("../.env")

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	log := logger.New(cfg.LogLevel, cfg.LogFormat)
	log.Info("starting schemahub backend", "port", cfg.Port)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.Connect(ctx, cfg.DatabaseURL, cfg.DBPoolMin, cfg.DBPoolMax)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	log.Info("connected to database")

	if err := database.RunMigrations(ctx, db, "internal/pkg/database/migrations"); err != nil {
		log.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	log.Info("database migrations applied")

	rdb, err := redis.Connect(ctx, cfg.RedisURL)
	if err != nil {
		log.Error("failed to connect to redis", "error", err)
		os.Exit(1)
	}
	defer rdb.Close()
	log.Info("connected to redis")

	jwtManager, err := jwt.NewManager(cfg.JWTPrivateKey, cfg.JWTPublicKey)
	if err != nil {
		log.Error("failed to initialize jwt manager", "error", err)
		os.Exit(1)
	}

	// ── Auth Service ──
	userRepo := postgres.NewUserRepository(db)
	tokenRepo := postgres.NewRefreshTokenRepository(db)
	oauthRepo := postgres.NewOAuthIdentityRepository(db)
	authSvc := domain.NewAuthService(userRepo, tokenRepo, oauthRepo, jwtManager)
	authHandler := authHandler.NewAuthHandler(authSvc)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(log),
			interceptor.LoggingInterceptor(log),
			interceptor.AuthInterceptor(jwtManager),
		),
	)

	authv1.RegisterAuthServiceServer(srv, authHandler)
	reflection.Register(srv)

	go func() {
		log.Info("gRPC server listening", "addr", lis.Addr().String())
		if err := srv.Serve(lis); err != nil {
			log.Error("failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down server...")
	srv.GracefulStop()
	log.Info("server stopped")
}
