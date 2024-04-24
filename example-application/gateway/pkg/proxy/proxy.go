package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"go.uber.org/zap"
)

func NewReverseProxy(targetURL string, logger *zap.Logger) *httputil.ReverseProxy {
	url, err := url.Parse(targetURL)
	if err != nil {
		panic("Failed to parse gateway proxy target URL: " + targetURL)
	}

	proxy := httputil.NewSingleHostReverseProxy(url)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		logger.Info("Proxying request.", zap.String("target", url.Host), zap.String("path", req.URL.Path))
	}

	return proxy
}
