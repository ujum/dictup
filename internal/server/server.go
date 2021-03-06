package server

import (
	"context"
	"fmt"
	"github.com/kataras/iris/v12"
	"github.com/ujum/dictap/internal/api/v1"
	"github.com/ujum/dictap/internal/config"
	"github.com/ujum/dictap/internal/service"
	"github.com/ujum/dictap/pkg/logger"
	"log"
	"net/http"
	"time"
)

// @title Swagger API
// @version 1.0
// @description This is a Dictup server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

type Server struct {
	Logger     logger.Logger
	Iris       *iris.Application
	httpServer *http.Server
	Cfg        *config.ServerConfig
}

// logWriter implement io.Writer interface to adapt
// app logger to log.Logger (http server logger)
type logWriter struct {
	logger logger.Logger
}

func (lw *logWriter) Write(p []byte) (n int, err error) {
	lw.logger.Error(string(p))
	return len(p), nil
}

func New(cfg *config.ServerConfig, appLogger logger.Logger, services *service.Services) *Server {
	irisApp := iris.New()
	irisApp.Logger().Install(appLogger)
	srv := &http.Server{
		Addr:           fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:        irisApp,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
		ErrorLog:       log.New(&logWriter{appLogger}, "", 0),
	}

	appSrv := &Server{
		Logger:     appLogger,
		Iris:       irisApp,
		httpServer: srv,
		Cfg:        cfg,
	}
	requestHandler := v1.NewHandler(appLogger, cfg, services)
	requestHandler.RegisterRoutes(irisApp)
	return appSrv
}

func (appSrv *Server) Start(ctx context.Context) error {
	// stop the server if context closed
	go func() {
		<-ctx.Done()
		if err := appSrv.Stop(); err != nil {
			appSrv.Logger.Debugf("http: web server shutdown err: %v", err)
		}
	}()

	if err := appSrv.Iris.Run(iris.Server(appSrv.httpServer)); err != nil {
		if err == http.ErrServerClosed {
			appSrv.Logger.Info("http: web server shutdown complete")
			return nil
		}
		appSrv.Logger.Errorf("http: web server closed unexpect: %v", err)
		return err
	}
	return nil
}

func (appSrv *Server) Stop() error {
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := appSrv.Iris.Shutdown(ctxShutDown); err != nil {
		return err
	}
	appSrv.Logger.Debug("http: web server closed")
	return nil
}
