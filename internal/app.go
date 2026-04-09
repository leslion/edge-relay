package internal

import (
	"log"
	"net/http"
	"time"
)

func Run() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	handler, err := NewProxyHandler(cfg)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              cfg.ListenAddress,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("Listening on http://%s", cfg.ListenAddress)
	log.Printf("Target base: %s", cfg.TargetURL.String())

	if cfg.UpstreamProxyBase != "" {
		redacted, err := buildProxyURL(cfg.UpstreamProxyBase, cfg.ProxyUsername, cfg.ProxyPassword)
		if err == nil {
			log.Printf("Upstream proxy: %s", redactProxyURL(redacted.String()))
		}
	} else {
		log.Printf("Upstream proxy: none")
	}

	return server.ListenAndServe()
}
