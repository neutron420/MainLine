package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	authDomain "github.com/schemahub/backend/internal/auth/domain"
	authHandler "github.com/schemahub/backend/internal/auth/handler"
	authRepo "github.com/schemahub/backend/internal/auth/repository/postgres"
	projectDomain "github.com/schemahub/backend/internal/project/domain"
	projectHandler "github.com/schemahub/backend/internal/project/handler"
	projectRepo "github.com/schemahub/backend/internal/project/repository/postgres"
	"github.com/schemahub/backend/internal/pkg/config"
	"github.com/schemahub/backend/internal/pkg/database"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	"github.com/schemahub/backend/internal/pkg/jwt"
	"github.com/schemahub/backend/internal/pkg/logger"
	"github.com/schemahub/backend/internal/pkg/redis"
	auditDomain "github.com/schemahub/backend/internal/audit/domain"
	auditHandler "github.com/schemahub/backend/internal/audit/handler"
	auditRepo "github.com/schemahub/backend/internal/audit/repository/postgres"
	eventDomain "github.com/schemahub/backend/internal/event/domain"
	eventHandler "github.com/schemahub/backend/internal/event/handler"
	migrationDomain "github.com/schemahub/backend/internal/migration/domain"
	migrationHandler "github.com/schemahub/backend/internal/migration/handler"
	migrationRepo "github.com/schemahub/backend/internal/migration/repository/postgres"
	schemaDomain "github.com/schemahub/backend/internal/schema/domain"
	schemaHandler "github.com/schemahub/backend/internal/schema/handler"
	schemaRepo "github.com/schemahub/backend/internal/schema/repository/postgres"
	auditv1 "github.com/schemahub/backend/proto/audit/v1"
	authv1 "github.com/schemahub/backend/proto/auth/v1"
	eventv1 "github.com/schemahub/backend/proto/event/v1"
	migrationv1 "github.com/schemahub/backend/proto/migration/v1"
	projectv1 "github.com/schemahub/backend/proto/project/v1"
	schemav1 "github.com/schemahub/backend/proto/schema/v1"
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
	userRepo := authRepo.NewUserRepository(db)
	tokenRepo := authRepo.NewRefreshTokenRepository(db)
	oauthRepo := authRepo.NewOAuthIdentityRepository(db)
	oauthCfg := &authDomain.OAuthProviderConfig{
		Google: authDomain.OAuthConfig{
			ClientID:     cfg.GoogleClientID,
			AuthURL:      "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:     "https://oauth2.googleapis.com/token",
			CallbackURL:  cfg.GoogleCallbackURL,
			Scopes:       "openid profile email",
		},
		GitHub: authDomain.OAuthConfig{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			CallbackURL:  cfg.GitHubCallbackURL,
			AuthURL:      "https://github.com/login/oauth/authorize",
			TokenURL:     "https://github.com/login/oauth/access_token",
			Scopes:       "read:user user:email",
		},
		Slack: authDomain.OAuthConfig{
			ClientID:     cfg.SlackClientID,
			ClientSecret: cfg.SlackClientSecret,
			CallbackURL:  cfg.SlackCallbackURL,
			AuthURL:      "https://slack.com/openid/connect/authorize",
			TokenURL:     "https://slack.com/api/openid.connect.token",
			Scopes:       "openid profile email",
		},
		StateSigningKey: []byte(cfg.EncryptionKey[:32]),
	}
	authSvc := authDomain.NewAuthService(userRepo, tokenRepo, oauthRepo, jwtManager, oauthCfg)
	authH := authHandler.NewAuthHandler(authSvc)

	// ── Project + Connection Service ──
	projRepo := projectRepo.NewProjectRepository(db)
	connRepo := projectRepo.NewConnectionRepository(db)
	projSvc := projectDomain.NewProjectService(projRepo)
	connSvc := projectDomain.NewConnectionService(connRepo, []byte(cfg.EncryptionKey))
	projH := projectHandler.NewProjectHandler(projSvc, connSvc)

	// ── Schema Service ──
	schemaRepoInstance := schemaRepo.NewSchemaRepository(db)
	schemaSvc := schemaDomain.NewSchemaService(schemaRepoInstance)

	schemaDomain.SetConnector(func(ctx context.Context, connStr string) (schemaDomain.DBPool, error) {
		pool, err := pgxpool.New(ctx, connStr)
		if err != nil {
			return nil, fmt.Errorf("connecting to target database: %w", err)
		}
		return &pgxPoolWrapper{pool: pool}, nil
	})

	connStringResolver := func(ctx context.Context, connID string) (string, error) {
		return connSvc.GetConnectionString(ctx, connID)
	}

	schemaH := schemaHandler.NewSchemaHandler(schemaSvc, connStringResolver)

	// ── Migration Service ──
	migRepo := migrationRepo.NewMigrationRepository(db)
	migSvc := migrationDomain.NewMigrationService(migRepo, connStringResolver)
	migH := migrationHandler.NewMigrationHandler(migSvc)

	// ── Audit Service ──
	auditRepoInstance := auditRepo.NewAuditRepository(db)
	auditSvc := auditDomain.NewAuditService(auditRepoInstance)
	auditH := auditHandler.NewAuditHandler(auditSvc)

	// ── Event Service ──
	eventSvc := eventDomain.NewEventService(rdb, auditRepoInstance)
	eventH := eventHandler.NewEventHandler(eventSvc)

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
			interceptor.RateLimitInterceptor(rdb, cfg.RateLimit),
			interceptor.ValidationInterceptor(),
		),
	)

	authv1.RegisterAuthServiceServer(srv, authH)
	projectv1.RegisterProjectServiceServer(srv, projH)
	schemav1.RegisterSchemaServiceServer(srv, schemaH)
	migrationv1.RegisterMigrationServiceServer(srv, migH)
	eventv1.RegisterEventServiceServer(srv, eventH)
	auditv1.RegisterAuditServiceServer(srv, auditH)
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

// pgxPoolWrapper adapts *pgxpool.Pool to schemaDomain.DBPool interface.
type pgxPoolWrapper struct {
	pool *pgxpool.Pool
}

func (w *pgxPoolWrapper) Query(ctx context.Context, sql string, args ...any) (schemaDomain.Rows, error) {
	return w.pool.Query(ctx, sql, args...)
}

func (w *pgxPoolWrapper) Close() {
	w.pool.Close()
}

var _ schemaDomain.DBPool = (*pgxPoolWrapper)(nil)
