package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	auditDomain "github.com/schemahub/backend/internal/audit/domain"
	auditHandler "github.com/schemahub/backend/internal/audit/handler"
	auditRepo "github.com/schemahub/backend/internal/audit/repository/postgres"
	authDomain "github.com/schemahub/backend/internal/auth/domain"
	authHandler "github.com/schemahub/backend/internal/auth/handler"
	authRepo "github.com/schemahub/backend/internal/auth/repository/postgres"
	driftDomain "github.com/schemahub/backend/internal/drift/domain"
	driftHandler "github.com/schemahub/backend/internal/drift/handler"
	driftRepo "github.com/schemahub/backend/internal/drift/repository/postgres"
	eventDomain "github.com/schemahub/backend/internal/event/domain"
	eventHandler "github.com/schemahub/backend/internal/event/handler"
	migrationDomain "github.com/schemahub/backend/internal/migration/domain"
	migrationHandler "github.com/schemahub/backend/internal/migration/handler"
	migrationRepo "github.com/schemahub/backend/internal/migration/repository/postgres"
	"github.com/schemahub/backend/internal/pkg/config"
	"github.com/schemahub/backend/internal/pkg/database"
	"github.com/schemahub/backend/internal/pkg/interceptor"
	"github.com/schemahub/backend/internal/pkg/jwt"
	"github.com/schemahub/backend/internal/pkg/logger"
	"github.com/schemahub/backend/internal/pkg/mailer"
	"github.com/schemahub/backend/internal/pkg/middleware"
	"github.com/schemahub/backend/internal/pkg/redis"
	"github.com/schemahub/backend/internal/pkg/worker"
	projectDomain "github.com/schemahub/backend/internal/project/domain"
	projectHandler "github.com/schemahub/backend/internal/project/handler"
	projectRepo "github.com/schemahub/backend/internal/project/repository/postgres"
	schemaDomain "github.com/schemahub/backend/internal/schema/domain"
	schemaHandler "github.com/schemahub/backend/internal/schema/handler"
	schemaRepo "github.com/schemahub/backend/internal/schema/repository/postgres"
	auditv1 "github.com/schemahub/backend/proto/audit/v1"
	authv1 "github.com/schemahub/backend/proto/auth/v1"
	driftv1 "github.com/schemahub/backend/proto/drift/v1"
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

	// Metrics endpoint (Prometheus). Served on a separate HTTP port so the
	// gRPC listener stays pure. Defaults to 9091; override with METRICS_PORT
	// to match the Prometheus scrape target.
	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9091"
	}
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:              ":" + metricsPort,
		Handler:           metricsMux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		log.Info("metrics server listening", "addr", metricsSrv.Addr)
		if err := metricsSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to serve metrics", "error", err)
			os.Exit(1)
		}
	}()

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
			ClientSecret: cfg.GoogleClientSecret,
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
	verifyRepo := authRepo.NewVerificationTokenRepo(db)
	authSvc := authDomain.NewAuthService(userRepo, tokenRepo, oauthRepo, verifyRepo, jwtManager, oauthCfg)
	appMailer := mailer.New(mailer.Config{
		SMTPHost:    cfg.SMTPHost,
		SMTPPort:    cfg.SMTPPort,
		Username:    cfg.SMTPUser,
		Password:    cfg.SMTPPassword,
		From:        cfg.SMTPFrom,
		FromName:    cfg.SMTPFromName,
		FrontendURL: cfg.FrontendURL,
		Log:         log,
	})
	authSvc.SetMailer(appMailer)
	authH := authHandler.NewAuthHandler(authSvc)

	// ── Project + Connection Service ──
	projRepo := projectRepo.NewProjectRepository(db)
	connRepo := projectRepo.NewConnectionRepository(db)
	projSvc := projectDomain.NewProjectService(projRepo, &projectUserLookup{userRepo: userRepo})
	projSvc.SetMailer(appMailer)
	connSvc := projectDomain.NewConnectionService(connRepo, encryptionKeyBytes(cfg.EncryptionKey))
	projH := projectHandler.NewProjectHandler(projSvc, connSvc)

	// ── Schema Service ──
	schemaRepoInstance := schemaRepo.NewSchemaRepository(db)
	schemaCache := schemaDomain.NewSchemaCache(rdb)
	schemaSvc := schemaDomain.NewSchemaService(schemaRepoInstance).WithCache(schemaCache)

	schemaDomain.SetConnector(func(ctx context.Context, connStr string) (schemaDomain.DBPool, error) {
		poolConfig, err := pgxpool.ParseConfig(connStr)
		if err != nil {
			return nil, fmt.Errorf("parsing target database config: %w", err)
		}
		// Serverless poolers (e.g. Neon) drop long-lived idle connections;
		// short lifetimes force fresh connections and avoid mid-query EOFs.
		poolConfig.MaxConnLifetime = 60 * time.Second
		poolConfig.MaxConnIdleTime = 30 * time.Second
		poolConfig.HealthCheckPeriod = 20 * time.Second
		poolConfig.MaxConns = 4
		pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
		if err != nil {
			return nil, fmt.Errorf("connecting to target database: %w", err)
		}
		return &pgxPoolWrapper{pool: pool}, nil
	})

	connStringResolver := func(ctx context.Context, connID string) (string, error) {
		return connSvc.GetConnectionString(ctx, connID)
	}

	connInfoResolver := func(ctx context.Context, connID string) (string, string, error) {
		conn, err := connSvc.GetByID(ctx, connID)
		if err != nil {
			return "", "", err
		}
		connStr, err := connSvc.GetConnectionString(ctx, connID)
		if err != nil {
			return "", "", err
		}
		return connStr, conn.ProjectID, nil
	}

	schemaH := schemaHandler.NewSchemaHandler(schemaSvc, connInfoResolver)

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

	// ── Drift Service ──
	driftRepoInstance := driftRepo.NewDriftRepository(db)
	driftComparator := &schemaDriftComparator{
		svc:      schemaSvc,
		connInfo: connInfoResolver,
	}
	driftSvc := driftDomain.NewDriftService(driftRepoInstance, driftComparator)
	driftH := driftHandler.NewDriftHandler(driftSvc, connStringResolver)

	// ── Workers ──
	workerRunner := worker.NewRunner(log)
	workerRunner.Add(worker.NewConnectionHealthWorker(connRepo, []byte(cfg.EncryptionKey)))
	workerRunner.Add(worker.NewAuditPartitionWorker(db))
	workerRunner.Add(worker.NewHardDeleteWorker(db))
	oauthRefreshWorker := worker.NewOAuthRefreshWorker(oauthRepo, userRepo, encryptionKeyBytes(cfg.EncryptionKey))
	oauthRefreshWorker.SetClientSecret("google", cfg.GoogleClientSecret)
	oauthRefreshWorker.SetClientSecret("github", cfg.GitHubClientSecret)
	oauthRefreshWorker.SetClientSecret("slack", cfg.SlackClientSecret)
	workerRunner.Add(oauthRefreshWorker)
	workerRunner.Add(worker.NewDriftAlertWorker(driftRepoInstance, eventSvc, rdb))
	workerRunner.Add(worker.NewDriftCheckWorker(connRepo, driftSvc, []byte(cfg.EncryptionKey)))
	workerRunner.Start(ctx)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	if err != nil {
		log.Error("failed to listen", "error", err)
		os.Exit(1)
	}

	rateLimiter := interceptor.NewRateLimiter(100, 200)

	rbacEnforcer := &rbacEnforcer{
		projects:    projRepo,
		connections: connRepo,
		schemas:     schemaRepoInstance,
		migrations:  migRepo,
		drifts:      driftRepoInstance,
		audits:      auditRepoInstance,
	}

	rbacCheck := func(ctx context.Context, userID, role, fullMethod string, req any) error {
		return rbacEnforcer.enforce(ctx, userID, role, fullMethod, req)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(log),
			middleware.TracingInterceptor(),
			middleware.CORSInterceptor([]string{cfg.FrontendURL}),
			interceptor.AuthInterceptor(jwtManager),
			interceptor.RBACInterceptor(rbacCheck),
			interceptor.AuditInterceptor(auditSvc.Insert, log.Error),
			interceptor.RateLimitInterceptor(rdb, cfg.RateLimit),
			rateLimiter.UnaryServerInterceptor(),
			interceptor.IdempotencyInterceptor(rdb),
			interceptor.MetricsInterceptor(nil),
			interceptor.DBRetryInterceptor(3),
		),
		grpc.ChainStreamInterceptor(
			middleware.CORSStreamInterceptor([]string{cfg.FrontendURL}),
			interceptor.StreamAuthInterceptor(jwtManager),
			interceptor.StreamRBACInterceptor(rbacCheck),
			rateLimiter.StreamServerInterceptor(),
		),
	)

	authv1.RegisterAuthServiceServer(srv, authH)
	projectv1.RegisterProjectServiceServer(srv, projH)
	schemav1.RegisterSchemaServiceServer(srv, schemaH)
	migrationv1.RegisterMigrationServiceServer(srv, migH)
	eventv1.RegisterEventServiceServer(srv, eventH)
	auditv1.RegisterAuditServiceServer(srv, auditH)
	driftv1.RegisterDriftServiceServer(srv, driftH)
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

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		log.Error("failed to stop metrics server", "error", err)
	}
	log.Info("server stopped")
}

// projectUserLookup adapts the auth user repository to the project domain's
// UserLookup contract so members can be invited by email address.
type projectUserLookup struct {
	userRepo interface {
		GetByEmail(context.Context, string) (*authDomain.User, error)
	}
}

func (l *projectUserLookup) GetByEmail(ctx context.Context, email string) (string, error) {
	u, err := l.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, authDomain.ErrUserNotFound) {
			return "", projectDomain.ErrUserNotFoundByEmail{Email: email}
		}
		return "", err
	}
	return u.ID, nil
}

// encryptionKeyBytes derives a fixed-size AES key (32 bytes) from the
// master key string via SHA-256, so any secret length is supported.
func encryptionKeyBytes(masterKey string) []byte {
	sum := sha256.Sum256([]byte(masterKey))
	return sum[:]
}

// schemaDriftComparator implements driftDomain.SchemaComparator using the schema service.
type schemaDriftComparator struct {
	svc       *schemaDomain.SchemaService
	connInfo  func(ctx context.Context, connID string) (connStr string, projectID string, err error)
}

func (c *schemaDriftComparator) CompareLiveWithVersion(ctx context.Context, connStr, connectionID string, schemaNames []string) ([]*driftDomain.DriftEvent, error) {
	schemaName := "public"
	if len(schemaNames) > 0 {
		schemaName = schemaNames[0]
	}

	// Capture the tracked baseline BEFORE introspection, since introspection
	// bumps the schema's current version to the live snapshot.
	expectedVersionID := ""
	if tracked, err := c.svc.GetSchemaByConnection(ctx, connectionID, schemaName); err == nil && tracked != nil && tracked.CurrentVersionID != nil {
		expectedVersionID = *tracked.CurrentVersionID
	}

	// Re-introspect live DB (passing the real connection ID).
	_, projectID, err := c.connInfo(ctx, connectionID)
	if err != nil {
		return nil, fmt.Errorf("resolving connection: %w", err)
	}
	schema, version, err := c.svc.Introspect(ctx, connStr, connectionID, projectID, schemaNames, "")
	if err != nil {
		return nil, fmt.Errorf("introspecting live schema: %w", err)
	}

	if version == nil {
		return nil, nil
	}

	// No baseline yet (first introspection) or schema unchanged: no drift to report.
	if expectedVersionID == "" || expectedVersionID == version.ID {
		return nil, nil
	}

	// Compare live (A) against the expected baseline (B): objects added in B but
	// missing from A are missing from live; objects removed from B are extra in live.
	diff, err := c.svc.CompareVersions(ctx, version.ID, expectedVersionID)
	if err != nil {
		return nil, fmt.Errorf("comparing versions: %w", err)
	}

	var events []*driftDomain.DriftEvent
	defToString := func(def json.RawMessage) string {
		if def == nil {
			return ""
		}
		return string(def)
	}

	for _, o := range diff.AddedObjects {
		events = append(events, &driftDomain.DriftEvent{
			SchemaID:           schema.ID,
			ExpectedVersionID:  expectedVersionID,
			DriftType:          driftDomain.DriftTypeMissingObject,
			ObjectType:         o.Type,
			ObjectName:         o.Name,
			ExpectedDefinition: defToString(o.Definition),
			Severity:           classifySeverity(o.Type),
			Status:             driftDomain.DriftStatusOpen,
		})
	}
	for _, o := range diff.RemovedObjects {
		events = append(events, &driftDomain.DriftEvent{
			SchemaID:          schema.ID,
			ExpectedVersionID: expectedVersionID,
			DriftType:         driftDomain.DriftTypeExtraObject,
			ObjectType:        o.Type,
			ObjectName:        o.Name,
			ActualDefinition:  defToString(o.Definition),
			Severity:          classifySeverity(o.Type),
			Status:            driftDomain.DriftStatusOpen,
		})
	}
	for _, o := range diff.ModifiedObjects {
		events = append(events, &driftDomain.DriftEvent{
			SchemaID:           schema.ID,
			ExpectedVersionID:  expectedVersionID,
			DriftType:          driftDomain.DriftTypeModifiedObject,
			ObjectType:         o.Type,
			ObjectName:         o.Name,
			ExpectedDefinition: defToString(o.Definition),
			DiffSummary:        summarizeChanges(o.Changes),
			Severity:           classifySeverity(o.Type),
			Status:             driftDomain.DriftStatusOpen,
		})
	}

	return events, nil
}

func classifySeverity(objType string) driftDomain.Severity {
	switch objType {
	case "table", "column", "primary_key", "foreign_key":
		return driftDomain.SeverityCritical
	case "index", "unique_constraint":
		return driftDomain.SeverityWarning
	default:
		return driftDomain.SeverityInfo
	}
}

func summarizeChanges(changes []schemaDomain.FieldChange) string {
	summary := ""
	for _, c := range changes {
		if summary != "" {
			summary += "; "
		}
		summary += fmt.Sprintf("%s: %v → %v", c.Field, c.Before, c.After)
	}
	return summary
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
