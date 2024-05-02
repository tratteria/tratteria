package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"go.uber.org/zap"
)

func NewReverseProxy(targetURL *url.URL, logger *zap.Logger) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		logger.Info("Proxying request.", zap.String("target", targetURL.Host), zap.String("path", req.URL.Path))
	}

	return proxy
}
