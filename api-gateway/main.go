package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"go.uber.org/zap"

	"techstore/internal/config"
)

type ServiceConfig struct {
	Name string
	URL  string
	Path string
}

func createProxy(targetURL, serviceName string, logger *zap.Logger, cfg *config.Config) *httputil.ReverseProxy {
	target, err := url.Parse(targetURL)
	if err != nil {
		logger.Fatal("Failed to parse target URL",
			zap.String("service", serviceName),
			zap.Error(err))
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	proxy.Director = func(req *http.Request) { // кастомизация
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host

		originalPath := req.URL.Path
		for _, service := range getServices(cfg) {
			if after, ok := strings.CutPrefix(originalPath, service.Path); ok {
				req.URL.Path = after
				if req.URL.Path == "" {
					req.URL.Path = "/"
				}
				break
			}
		}

		req.Host = target.Host

		logger.Info("Начало запроса",
			zap.String("service", serviceName),
			zap.String("method", req.Method),
			zap.String("path", originalPath))
	}

	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("Ошибка прокси",
			zap.String("service", serviceName),
			zap.String("path", r.URL.Path),
			zap.Error(err))

		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error": "Service unavailable"}`))
	}

	return proxy
}

func getServices(cfg *config.Config) []ServiceConfig {
	return []ServiceConfig{
		{
			Name: "AuthService",
			URL:  cfg.GetAuthServiceURL(),
			Path: "/auth",
		},
		{
			Name: "ProductService",
			URL:  cfg.GetProductServiceURL(),
			Path: "/products",
		},
	}
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	cfg, errParse := config.ParseConfig()
	if errParse != nil {
		logger.Error("failed to parse config: %v\n", zap.Error(errParse))
		return
	}

	services := getServices(cfg)
	for _, service := range services {
		proxy := createProxy(service.URL, service.Name, logger, cfg)
		http.Handle(service.Path+"/", proxy)
		logger.Info("Сервис зарегистрирован",
			zap.String("path", service.Path),
			zap.String("target", service.URL))
	}

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		logger.Info("Health check")
		w.Write([]byte(`{"status": "ok"}`))
	})

	logger.Info("API Gateway запущен",
		zap.Int("port", cfg.HTTPPort),
		zap.String("environment", cfg.Env))

	if err := http.ListenAndServe(fmt.Sprintf(":%d", cfg.HTTPPort), nil); err != nil {
		logger.Fatal("Ошибка запуска сервера", zap.Error(err))
	}
}
