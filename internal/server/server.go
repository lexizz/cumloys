package server

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/lexizz/cumloys/internal/config"
	"github.com/lexizz/cumloys/internal/pkg/logger"
)

type Server struct {
	ctx        context.Context
	httpServer *http.Server
	log        logger.Logger
}

func New(ctx context.Context, cfg *config.Config, handler http.Handler, logger logger.Logger) *Server {
	newDNS, errParse := parseURL(cfg.HTTP.Address, logger)
	if errParse != nil {
		return nil
	}

	return &Server{
		ctx: ctx,
		httpServer: &http.Server{
			Addr:              newDNS,
			Handler:           handler,
			ReadTimeout:       cfg.HTTP.ReadTimeout,
			ReadHeaderTimeout: cfg.HTTP.ReadTimeout,
		},
		log: logger,
	}
}

func (srv *Server) Run() error {
	return srv.httpServer.ListenAndServe()
}

func (srv *Server) Stop(ctx context.Context) error {
	return srv.httpServer.Shutdown(ctx)
}

//gocyclo:ignore
func parseURL(dns string, logger logger.Logger) (string, error) {
	urlData, errURLData := getDataURL(dns, logger)
	if errURLData != nil {
		return "", errURLData
	}

	switch {
	case urlData.Host != "":
		dns = urlData.Host
		if !strings.Contains(urlData.Host, ":") {
			dns = urlData.Host + ":8080"
		}
	case strings.Contains(urlData.Scheme, "localhost"):
		dns = urlData.Scheme + ":8080"
		if urlData.Opaque != "" {
			dns = urlData.Scheme + ":" + urlData.Opaque
		}
	case strings.Contains(urlData.Scheme, "http"):
		dns = urlData.Scheme + ":8080"
		if urlData.Opaque != "" {
			dns = urlData.Scheme + ":" + urlData.Opaque
		}
	case urlData.Scheme == "" && urlData.Host == "" && urlData.Path != "":
		dns = urlData.Path + ":8080"
	}

	logger.Info("=== Url was installed:", dns)

	return dns, nil
}

func getDataURL(dns string, logger logger.Logger) (*url.URL, error) {
	if !strings.Contains(dns, "http") {
		dns = "http://" + dns
	}

	urlData, errParse := url.Parse(dns)
	if errParse != nil {
		logger.Error("---> ERROR: failed parse dns:", errParse)
		return nil, errParse
	}

	if urlData == nil {
		logger.Error("---> ERROR: url data not found")
		return nil, errors.New("url data not found")
	}

	return urlData, nil
}
