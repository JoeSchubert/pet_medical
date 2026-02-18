package config

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Private proxy CIDRs: when TrustPrivateProxies is true, requests from these are treated as from a trusted proxy.
var privateProxyNetworks = []*net.IPNet{
	mustCIDR("127.0.0.0/8"),
	mustCIDR("10.0.0.0/8"),
	mustCIDR("172.16.0.0/12"),
	mustCIDR("192.168.0.0/16"),
	mustCIDR("::1/128"),
	mustCIDR("fc00::/7"),
}

func mustCIDR(s string) *net.IPNet {
	_, n, err := net.ParseCIDR(s)
	if err != nil {
		panic("config: invalid CIDR " + s)
	}
	return n
}

type Config struct {
	ServerPort         int
	DBURL              string
	JWTSecret          string
	JWTAccessTTL       int    // minutes
	JWTRefreshTTL      int    // days
	CORSOrigins        string // comma-separated
	EnableDebugLogging bool
	SystemLanguage     string // e.g. "en" — server-side i18n for logs
	// Defaults for new users and when user settings are empty (env: DEFAULT_WEIGHT_UNIT, DEFAULT_CURRENCY, DEFAULT_LANGUAGE)
	DefaultWeightUnit string // "lbs" or "kg"
	DefaultCurrency   string // e.g. "USD", "EUR"
	DefaultLanguage   string // e.g. "en", "es"
	// Google OAuth (e.g. for oauth2-proxy or app-side OAuth). Read from env.
	GoogleClientID     string // OAuth2 client ID
	GoogleClientSecret string // OAuth2 client secret
	GoogleRedirectURI  string // OAuth2 redirect URI (e.g. https://app.example.com/oauth2/callback)
	// Trusted proxies: comma-separated IPs or CIDRs (e.g. "10.0.0.1,172.16.0.0/12"). When request comes from one of these, X-Forwarded-* and forwarded auth headers are trusted.
	// When TrustPrivateProxies is true (default), requests from loopback and private IP ranges are also trusted, so TRUSTED_PROXIES can be left unset when behind a proxy on the same host or in a private network.
	TrustedProxies       []*net.IPNet
	TrustPrivateProxies  bool
	// Header names for proxy auth (oauth2-proxy sets these). Defaults: X-Forwarded-Email, X-Forwarded-User
	ForwardedEmailHeader string
	ForwardedUserHeader  string
	// SameSiteCookie: use SameSiteNoneMode only when needed (e.g. cross-site). Default Lax. Set via SAME_SITE_COOKIE if required.
	SameSiteCookie http.SameSite
	// Development: when true, relaxes security (no Secure cookies, no HSTS, no startup warnings for default JWT/CORS). Default false.
	Development bool
}

func Load() *Config {
	accessTTL := 30
	if v := os.Getenv("JWT_ACCESS_TTL_MIN"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			accessTTL = n
		}
	}
	refreshTTL := 7
	if v := os.Getenv("JWT_REFRESH_TTL_DAYS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			refreshTTL = n
		}
	}
	port := 8080
	if v := os.Getenv("PORT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			port = n
		}
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:5432/pet_medical?sslmode=disable"
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production-min-32-chars"
	}
	cors := strings.TrimSpace(os.Getenv("CORS_ORIGINS"))
	// When empty, CORS middleware will allow only the request's effective origin (same-origin when behind a proxy). Set CORS_ORIGINS=* or explicit origins if you need cross-origin.
	debugLog := strings.TrimSpace(strings.ToLower(os.Getenv("ENABLE_DEBUG_LOGGING")))
	enableDebug := debugLog == "1" || debugLog == "true" || debugLog == "yes"
	systemLang := strings.TrimSpace(os.Getenv("SYSTEM_LANGUAGE"))
	if systemLang == "" {
		systemLang = "en"
	}
	weightUnit := strings.TrimSpace(strings.ToLower(os.Getenv("DEFAULT_WEIGHT_UNIT")))
	if weightUnit != "lbs" && weightUnit != "kg" {
		weightUnit = "lbs"
	}
	currency := strings.TrimSpace(strings.ToUpper(os.Getenv("DEFAULT_CURRENCY")))
	if currency == "" {
		currency = "USD"
	}
	defaultLang := strings.TrimSpace(strings.ToLower(os.Getenv("DEFAULT_LANGUAGE")))
	if defaultLang == "" {
		defaultLang = "en"
	}
	googleClientID := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_ID"))
	googleClientSecret := strings.TrimSpace(os.Getenv("GOOGLE_CLIENT_SECRET"))
	googleRedirectURI := strings.TrimSpace(os.Getenv("GOOGLE_REDIRECT_URI"))
	forwardedEmail := strings.TrimSpace(os.Getenv("FORWARDED_EMAIL_HEADER"))
	if forwardedEmail == "" {
		forwardedEmail = "X-Forwarded-Email"
	}
	forwardedUser := strings.TrimSpace(os.Getenv("FORWARDED_USER_HEADER"))
	if forwardedUser == "" {
		forwardedUser = "X-Forwarded-User"
	}
	trustedProxies := parseTrustedProxies(os.Getenv("TRUSTED_PROXIES"))
	trustPrivate := parseBoolEnv("TRUST_PRIVATE_PROXIES", true)
	development := parseBoolEnv("DEVELOPMENT", false)
	sameSite := parseSameSiteEnv(os.Getenv("SAME_SITE_COOKIE"))
	return &Config{
		ServerPort:           port,
		DBURL:                dbURL,
		JWTSecret:            jwtSecret,
		JWTAccessTTL:         accessTTL,
		JWTRefreshTTL:        refreshTTL,
		CORSOrigins:          cors,
		EnableDebugLogging:   enableDebug,
		SystemLanguage:       systemLang,
		DefaultWeightUnit:    weightUnit,
		DefaultCurrency:      currency,
		DefaultLanguage:      defaultLang,
		GoogleClientID:       googleClientID,
		GoogleClientSecret:   googleClientSecret,
		GoogleRedirectURI:    googleRedirectURI,
		TrustedProxies:       trustedProxies,
		TrustPrivateProxies:  trustPrivate,
		ForwardedEmailHeader: forwardedEmail,
		ForwardedUserHeader:  forwardedUser,
		SameSiteCookie:       sameSite,
		Development:          development,
	}
}

// parseBoolEnv returns true if the env var is set to 1, true, or yes (case-insensitive). Otherwise returns defaultVal.
func parseBoolEnv(key string, defaultVal bool) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if v == "" {
		return defaultVal
	}
	return v == "1" || v == "true" || v == "yes"
}

// parseSameSiteEnv parses SAME_SITE_COOKIE: "none" -> SameSiteNoneMode, "lax" -> LaxMode, "strict" -> StrictMode. Default Lax.
func parseSameSiteEnv(v string) http.SameSite {
	switch strings.TrimSpace(strings.ToLower(v)) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	case "lax", "":
		return http.SameSiteLaxMode
	default:
		return http.SameSiteLaxMode
	}
}

// parseTrustedProxies parses a comma-separated list of IPs or CIDRs (e.g. "10.0.0.1,172.16.0.0/12").
func parseTrustedProxies(s string) []*net.IPNet {
	var out []*net.IPNet
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if strings.Contains(part, "/") {
			_, n, err := net.ParseCIDR(part)
			if err != nil {
				continue
			}
			out = append(out, n)
		} else {
			ip := net.ParseIP(part)
			if ip == nil {
				continue
			}
			mask := net.CIDRMask(len(ip)*8, len(ip)*8)
			out = append(out, &net.IPNet{IP: ip, Mask: mask})
		}
	}
	return out
}

// IsTrustedProxy returns true if the given remote address is from a trusted proxy (explicit TRUSTED_PROXIES or, when TrustPrivateProxies is true, from loopback/private IP).
func (c *Config) IsTrustedProxy(remoteAddr string) bool {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	if c.TrustPrivateProxies && isPrivateOrLoopback(ip) {
		return true
	}
	for _, n := range c.TrustedProxies {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

func isPrivateOrLoopback(ip net.IP) bool {
	for _, n := range privateProxyNetworks {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// IsRequestHTTPS returns true if the request is considered HTTPS: direct TLS or X-Forwarded-Proto: https from a trusted proxy.
// Used to set cookie Secure flag without requiring SECURE_COOKIES env.
func (c *Config) IsRequestHTTPS(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if c.IsTrustedProxy(r.RemoteAddr) && strings.TrimSpace(strings.ToLower(r.Header.Get("X-Forwarded-Proto"))) == "https" {
		return true
	}
	return false
}

// RequestOrigin returns the effective origin of this request (scheme://host), using X-Forwarded-Proto and X-Forwarded-Host when from a trusted proxy.
// Used for CORS when CORS_ORIGINS is unset (allow same-origin only).
func (c *Config) RequestOrigin(r *http.Request) string {
	scheme := "http"
	if c.IsRequestHTTPS(r) {
		scheme = "https"
	}
	host := r.Host
	if c.IsTrustedProxy(r.RemoteAddr) {
		if h := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); h != "" {
			// Can be "host1, host2" from multiple proxies; use the first (client-facing).
			if i := strings.Index(h, ","); i >= 0 {
				host = strings.TrimSpace(h[:i])
			} else {
				host = h
			}
		}
	}
	return scheme + "://" + host
}
