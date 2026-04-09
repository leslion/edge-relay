package internal

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"golang.org/x/term"
)

type Config struct {
	ListenAddress      string
	TargetURL          *url.URL
	UpstreamProxyBase  string
	ProxyUsername      string
	ProxyPassword      string
	InsecureSkipVerify bool
	CustomCACertFile   string
}

func LoadConfig() (*Config, error) {
	if err := loadEnv(); err != nil {
		return nil, err
	}

	targetBase, err := requireEnv("TARGET_BASE")
	if err != nil {
		return nil, err
	}

	listenAddress := getenv("LISTEN_ADDRESS", "127.0.0.1:8787")
	upstreamProxyBase := strings.TrimSpace(os.Getenv("UPSTREAM_PROXY"))
	proxyUsername := os.Getenv("PROXY_USERNAME")
	proxyPassword := os.Getenv("PROXY_PASSWORD")
	promptForProxyPassword := strings.EqualFold(getenv("PROMPT_FOR_PROXY_PASSWORD", "false"), "true")
	insecureSkipVerify := strings.EqualFold(getenv("INSECURE_SKIP_VERIFY", "false"), "true")
	customCACertFile := strings.TrimSpace(os.Getenv("CUSTOM_CA_CERT_FILE"))

	if promptForProxyPassword && proxyPassword == "" {
		if proxyUsername == "" {
			return nil, fmt.Errorf("PROMPT_FOR_PROXY_PASSWORD=true requires PROXY_USERNAME to be set")
		}

		pw, err := promptPassword("Enter proxy password: ")
		if err != nil {
			return nil, fmt.Errorf("Failed reading proxy password: %w", err)
		}
		proxyPassword = pw
	}

	targetURL, err := url.Parse(targetBase)
	if err != nil || targetURL.Scheme == "" || targetURL.Host == "" {
		return nil, fmt.Errorf("Invalid TARGET_BASE: %q", targetBase)
	}

	return &Config{
		ListenAddress:      listenAddress,
		TargetURL:          targetURL,
		UpstreamProxyBase:  upstreamProxyBase,
		ProxyUsername:      proxyUsername,
		ProxyPassword:      proxyPassword,
		InsecureSkipVerify: insecureSkipVerify,
		CustomCACertFile:   customCACertFile,
	}, nil
}

func loadEnv() error {
	envFile := strings.TrimSpace(os.Getenv("ENV_FILE"))

	if envFile != "" {
		if err := godotenv.Load(envFile); err != nil {
			return fmt.Errorf("Failed to load ENV_FILE=%q: %w", envFile, err)
		}
		log.Printf("Loaded environment from %s", envFile)
		return nil
	}

	_, statErr := os.Stat(".env")
	if statErr == nil {
		if err := godotenv.Load(".env"); err != nil {
			return fmt.Errorf("Failed to load .env: %w", err)
		}
		log.Printf("Loaded environment from .env")
		return nil
	}

	if errors.Is(statErr, os.ErrNotExist) {
		log.Printf("No .env file found. Continuing with process environment.")
		return nil
	}

	return fmt.Errorf("Failed to inspect .env: %w", statErr)
}

func promptPassword(prompt string) (string, error) {
	_, _ = fmt.Fprint(os.Stderr, prompt)
	pw, err := term.ReadPassword(int(os.Stdin.Fd()))
	_, _ = fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", err
	}
	return string(pw), nil
}

func requireEnv(key string) (string, error) {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return "", fmt.Errorf("Missing required env var %s", key)
	}
	return val, nil
}

func getenv(key, fallback string) string {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	return val
}
