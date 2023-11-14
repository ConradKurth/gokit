package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	iaws "github.com/ConradKurth/gokit/aws"
	"github.com/ConradKurth/gokit/config"
	"github.com/ConradKurth/gokit/instrument"
	"github.com/ConradKurth/gokit/logger"
	loggerMiddleware "github.com/ConradKurth/gokit/middleware/logger"
	sentryMiddleware "github.com/ConradKurth/gokit/middleware/sentry"
	"github.com/ConradKurth/gokit/middleware/ssl"
	"github.com/ConradKurth/gokit/middleware/userinfo"
	"github.com/ConradKurth/gokit/secrets"
	"github.com/andybalholm/brotli"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/hashicorp/go-multierror"
	"github.com/riandyrn/otelchi"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	brotliCompressionLevel = 4
	shutdownTimeout        = time.Second * 5
)

// Service implements common service functionalities for all services.
type Service struct {
	cfg         *config.Config
	serviceName string
	logger      logger.Logger

	tracer *trace.TracerProvider

	temporalWorker worker.Worker
	temporalClient client.Client

	router     chi.Router
	webserver  *http.Server
	grpcServer *grpc.Server
}

type CronRegister interface {
	InitCron() error
}

type RouteRegistration interface {
	RegisterRoutes(router chi.Router, middlewares ...func(http.Handler) http.Handler)
	RegisterPublicRoutes(router chi.Router, middlewares ...func(http.Handler) http.Handler)
}
type WorkerRegistration interface {
	RegisterWithWorker(worker worker.Worker)
}

// New returns a new service instance.
func New(ctx context.Context, configPath, sentryDSN string, opts ...func(*options)) (*Service, error) {
	opt := options{}
	for _, o := range opts {
		o(&opt)
	}

	cfg := config.LoadConfig(config.WithPath(configPath))

	var err error
	if err = secrets.New(iaws.New()).LoadSecrets(cfg); err != nil {
		return nil, fmt.Errorf("loading secrets: %w", err)
	}

	svc := &Service{
		cfg:         cfg,
		serviceName: cfg.GetString("serviceName"),
		logger:      logger.NewV2(cfg),
	}

	if err = sentry.Init(sentry.ClientOptions{
		Dsn:                sentryDSN,
		ServerName:         svc.serviceName,
		Environment:        config.GetEnvironment().String(),
		AttachStacktrace:   true,
		TracesSampleRate:   opt.traceSampleRate,
		ProfilesSampleRate: opt.profileSameplRate,
	}); err != nil {
		return nil, fmt.Errorf("initializing sentry: %w", err)
	}

	svc.tracer, err = instrument.GetTracingProvider(ctx, cfg, svc.serviceName)
	if err != nil {
		return nil, fmt.Errorf("initializing tracer: %w", err)
	}

	if opt.temporalService {
		svc.temporalClient, err = instrument.NewTemporalClient(ctx, svc.logger, cfg, svc.serviceName)
		if err != nil {
			return nil, fmt.Errorf("initializing temporal: %w", err)
		}
		svc.temporalWorker = worker.New(svc.temporalClient, cfg.GetString("temporal.taskQueue"), worker.Options{})
	}

	if opt.grpcService {
		svc.grpcServer = grpc.NewServer(
			grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor()),
			grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()),
			grpc.Creds(insecure.NewCredentials()),
		)
	}

	if opt.httpService {
		svc.router = svc.initializeRouter(cfg)
		svc.webserver = &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.GetInt("api.port")),
			Handler: svc.router,
		}
	}

	return svc, nil
}

var grpcClients = make(map[string]*grpc.ClientConn)
var gprcClientLock sync.Mutex

// GetGRPCClient will create a proper grpc client
func (svc *Service) GetGRPCClient(host string) (*grpc.ClientConn, error) {
	gprcClientLock.Lock()
	defer gprcClientLock.Unlock()

	if c, ok := grpcClients[host]; ok {
		return c, nil
	}

	c, err := grpc.Dial(host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
		grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
	grpcClients[host] = c
	return c, err
}

// Config returns the config of the service.
func (svc *Service) Config() *config.Config {
	return svc.cfg
}

// Logger returns the logger of the service.
func (svc *Service) Logger() logger.Logger {
	return svc.logger
}

// Temporal returns the temporal client of the service.
func (svc *Service) Temporal() client.Client {
	return svc.temporalClient
}

// GRPC return the grpc server
func (svc *Service) GRPC() *grpc.Server {
	return svc.grpcServer
}

// Router returns the http router. This allows the caller to register custom routes.
func (svc *Service) Router() chi.Router {
	return svc.router
}

// initializeRouter returns a router with standard health and middleware configured.
func (svc *Service) initializeRouter(cfg *config.Config) chi.Router {
	router := chi.NewRouter()
	router.Use(middleware.Heartbeat("/healthz"))
	router.Use(ssl.NewMiddleware(config.IsDevelopment()))

	cors := cors.New(cors.Options{
		AllowedOrigins: cfg.GetStringSlice("cors.hosts"),
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		// AllowedHeaders:   []string{"Accept", "Accept-Encoding", "Accept-Language", "Authorization", "Content-Type", "sentry-trace", "X-CSRF-Token"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	router.Use(cors.Handler)
	// sample anything below
	router.Use(otelchi.Middleware(svc.serviceName))
	router.Use(middleware.RealIP)
	router.Use(middleware.RequestID)
	router.Use(userinfo.NewMiddleware())
	router.Use(loggerMiddleware.NewMiddleware(svc.logger))

	compression := middleware.NewCompressor(brotliCompressionLevel)
	compression.SetEncoder("br", func(w io.Writer, level int) io.Writer {
		return brotli.NewWriterLevel(w, level)
	})
	router.Use(compression.Handler)

	router.Use(sentryMiddleware.NewMiddleware(cfg))

	return router
}

// RegisterRoutes registers http routes with the webserver router.
func (svc *Service) RegisterRoutes(routers []RouteRegistration, middlewares ...func(http.Handler) http.Handler) {

	for _, route := range routers {
		route.RegisterRoutes(svc.router, middlewares...)
	}
}

// RegisterPublicRoutes registers http routes with the webserver router.
func (svc *Service) RegisterPublicRoutes(routers []RouteRegistration, middlewares ...func(http.Handler) http.Handler) {

	for _, route := range routers {
		route.RegisterPublicRoutes(svc.router, middlewares...)
	}
}

// RegisterWithWorker registers types with the temporal worker.
func (svc *Service) RegisterWithWorker(registrations []WorkerRegistration) {
	for _, registration := range registrations {
		registration.RegisterWithWorker(svc.temporalWorker)
	}
}

// RegisterWithCron registers crons with temporal
func (svc *Service) RegisterWithCron(registrations []CronRegister) error {
	for _, registration := range registrations {
		if err := registration.InitCron(); err != nil {
			return err
		}
	}
	return nil
}

// Start starts all internal service connections and stops all servers.
// The function blocks and should be run in a goroutine.
func (svc *Service) Start(ctx context.Context) error {
	if svc.temporalWorker != nil {
		svc.logger.InfoCtx(ctx, "Starting the temporal worker")
		if err := svc.temporalWorker.Start(); err != nil {
			return fmt.Errorf("starting temporal worker: %w", err)
		}
	}

	if svc.grpcServer != nil {
		net, err := net.Listen("tcp", svc.cfg.GetString("grpc.host"))
		if err != nil {
			return fmt.Errorf("creating listener socket: %w", err)
		}
		go func() {
			svc.logger.InfoCtx(ctx, "Starting the grpc server")
			if err := svc.grpcServer.Serve(net); err != nil {
				svc.Logger().ErrorCtx(ctx, "Error closing the grpc server", logger.ErrField(err))
			}
		}()
	}

	if svc.webserver != nil {
		svc.logger.InfoCtx(ctx, "Starting the webserver")

		if err := svc.webserver.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("starting webserver: %w", err)
		}
	}

	return nil
}

// Shutdown closes all internal service connections and stops all servers.
// It waits for all worker functions to finish.
func (svc *Service) Shutdown(ctx context.Context) error {
	var errs *multierror.Error

	if svc.temporalWorker != nil {
		svc.temporalWorker.Stop()
	}
	if svc.temporalClient != nil {
		svc.temporalClient.Close()
	}

	for _, conn := range grpcClients {
		if err := conn.Close(); err != nil {
			return err
		}
	}

	if svc.grpcServer != nil {
		svc.grpcServer.GracefulStop()
	}

	if svc.tracer != nil {
		if err := svc.tracer.Shutdown(ctx); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	if svc.webserver != nil {
		timeout, cancel := context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()

		if err := svc.webserver.Shutdown(timeout); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}
