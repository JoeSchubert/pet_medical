package config

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

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
	TrustedProxies []*net.IPNet
	// Header names for proxy auth (oauth2-proxy sets these). Defaults: X-Forwarded-Email, X-Forwarded-User
	ForwardedEmailHeader string
	ForwardedUserHeader  string
	// SecureCookies: when true, cookies use Secure flag (HTTPS only). In development mode this is forced false; otherwise from SECURE_COOKIES (default true).
	SecureCookies bool
	// SameSiteCookie: use SameSiteNoneMode behind some reverse proxies so cookies are sent. When None, Secure must be true.
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
	cors := os.Getenv("CORS_ORIGINS")
	if cors == "" {
		cors = "*"
	}
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
	development := parseBoolEnv("DEVELOPMENT", false)
	var secureCookies bool
	if development {
		secureCookies = false
	} else {
		secureCookies = parseBoolEnv("SECURE_COOKIES", true)
	}
	sameSite := parseSameSiteEnv(os.Getenv("SAME_SITE_COOKIE"))
	if sameSite == http.SameSiteNoneMode && !secureCookies {
		sameSite = http.SameSiteLaxMode // None requires Secure
	}
	return &Config{
		ServerPort:           port,
		DBURL:                dbURL,
		JWTSecret:             jwtSecret,
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
		ForwardedEmailHeader: forwardedEmail,
		ForwardedUserHeader:  forwardedUser,
		SecureCookies:        secureCookies,
		SameSiteCookie:        sameSite,
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

// IsTrustedProxy returns true if the given remote address (e.g. from req.RemoteAddr, "host:port") is one of the trusted proxy IPs/CIDRs.
func (c *Config) IsTrustedProxy(remoteAddr string) bool {
	if len(c.TrustedProxies) == 0 {
		return false
	}
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	for _, n := range c.TrustedProxies {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
