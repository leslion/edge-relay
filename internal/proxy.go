package internal

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

func NewProxyHandler(cfg *Config) (http.Handler, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.ResponseHeaderTimeout = 60 * time.Second
	transport.ExpectContinueTimeout = 1 * time.Second
	transport.IdleConnTimeout = 90 * time.Second
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 100

	if transport.TLSClientConfig == nil {
		transport.TLSClientConfig = &tls.Config{}
	} else {
		transport.TLSClientConfig = transport.TLSClientConfig.Clone()
	}
	transport.TLSClientConfig.InsecureSkipVerify = cfg.InsecureSkipVerify

	if cfg.UpstreamProxyBase != "" {
		proxyURL, err := buildProxyURL(cfg.UpstreamProxyBase, cfg.ProxyUsername, cfg.ProxyPassword)
		if err != nil {
			return nil, fmt.Errorf("Invalid proxy configuration: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	proxy := httputil.NewSingleHostReverseProxy(cfg.TargetURL)
	originalDirector := proxy.Director

	proxy.Transport = transport
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
		http.Error(w, "Bad gateway: "+err.Error(), http.StatusBadGateway)
	}

	proxy.Director = func(r *http.Request) {
		originalHost := r.Host
		originalScheme := "http"
		if r.TLS != nil {
			originalScheme = "https"
		}

		r.URL.Path = stripPrefix(r.URL.Path, "/proxy")
		originalDirector(r)

		r.Host = cfg.TargetURL.Host
		r.Header.Set("X-Forwarded-Host", originalHost)
		r.Header.Set("X-Forwarded-Proto", originalScheme)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok")
	})
	mux.Handle("/proxy/", logging(proxy))
	mux.Handle("/proxy", logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = "/"
		proxy.ServeHTTP(w, r)
	})))

	return mux, nil
}

func buildProxyURL(rawBase, username, password string) (*url.URL, error) {
	u, err := url.Parse(rawBase)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("UPSTREAM_PROXY must look like http://proxyhost:8080")
	}

	if password != "" && username == "" {
		return nil, fmt.Errorf("PROXY_PASSWORD is set but PROXY_USERNAME is empty")
	}

	if username != "" {
		u.User = url.UserPassword(username, password)
	}

	return u, nil
}

func stripPrefix(path, prefix string) string {
	path = strings.TrimPrefix(path, prefix)
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

func redactProxyURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.User != nil {
		username := u.User.Username()
		if username != "" {
			u.User = url.UserPassword(username, "*****")
		} else {
			u.User = url.User("*****")
		}
	}
	return u.String()
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s", r.Method, r.URL.String())
		next.ServeHTTP(w, r)
	})
}
