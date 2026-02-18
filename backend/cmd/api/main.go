package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pet-medical/api/internal/auth"
	"github.com/pet-medical/api/internal/config"
	"github.com/pet-medical/api/internal/db"
	"github.com/pet-medical/api/internal/debuglog"
	"github.com/pet-medical/api/internal/handlers"
	"github.com/pet-medical/api/internal/i18n"
	"github.com/pet-medical/api/internal/middleware"
)

//go:embed static/*
var staticFS embed.FS

func main() {
	cfg := config.Load()
	i18n.Init(cfg.SystemLanguage)
	debuglog.SetEnabled(cfg.EnableDebugLogging)
	debuglog.Debugf("config loaded: port=%d debug=%v system_language=%s", cfg.ServerPort, cfg.EnableDebugLogging, cfg.SystemLanguage)
	gormDB, err := db.NewGORMWithRetry(cfg.DBURL)
	if err != nil {
		log.Fatalf("gorm: %v", err)
	}
	sqlDB, _ := gormDB.DB()
	defer func() { _ = sqlDB.Close() }()

	if err := db.AutoMigrateAll(gormDB); err != nil {
		log.Fatalf("automigrate: %v", err)
	}
	if err := db.SeedDefaultAdmin(gormDB, cfg); err != nil {
		log.Fatalf("seed: %v", err)
	}
	if err := db.SeedDefaultDropdownOptions(gormDB); err != nil {
		log.Fatalf("seed defaults: %v", err)
	}

	jwt := auth.NewJWT(cfg.JWTSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	refreshStore := auth.NewRefreshStore(gormDB)

	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}

	authHandler := &handlers.AuthHandler{
		DB:                gormDB,
		JWT:               jwt,
		RefreshStore:      refreshStore,
		Config:            cfg,
		DefaultWeightUnit: cfg.DefaultWeightUnit,
		DefaultCurrency:   cfg.DefaultCurrency,
		DefaultLanguage:   cfg.DefaultLanguage,
		SameSiteCookie:    int(cfg.SameSiteCookie),
	}
	petsHandler := &handlers.PetsHandler{DB: gormDB, UploadDir: uploadDir}
	vaccHandler := &handlers.VaccinationsHandler{DB: gormDB}
	weightsHandler := &handlers.WeightsHandler{DB: gormDB}
	docsHandler := &handlers.DocumentsHandler{DB: gormDB, UploadDir: uploadDir, MaxDocumentBytes: cfg.MaxUploadDocumentBytes}
	photosHandler := &handlers.PhotosHandler{DB: gormDB, UploadDir: uploadDir, MaxPhotoBytes: cfg.MaxUploadPhotoBytes}
	usersHandler := &handlers.UsersHandler{
		DB:                gormDB,
		DefaultWeightUnit: cfg.DefaultWeightUnit,
		DefaultCurrency:   cfg.DefaultCurrency,
		DefaultLanguage:   cfg.DefaultLanguage,
	}
	settingsHandler := &handlers.SettingsHandler{
		DB:                gormDB,
		DefaultWeightUnit: cfg.DefaultWeightUnit,
		DefaultCurrency:   cfg.DefaultCurrency,
		DefaultLanguage:   cfg.DefaultLanguage,
	}
	customOptsHandler := &handlers.CustomOptionsHandler{GORM: gormDB}
	defaultOptsHandler := &handlers.DefaultOptionsHandler{GORM: gormDB}

	router := mux.NewRouter()
	healthOK := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	}
	router.HandleFunc("/health", healthOK).Methods(http.MethodGet)
	router.HandleFunc("/api/health", healthOK).Methods(http.MethodGet)

	// Public auth
	router.HandleFunc("/api/auth/login", authHandler.Login).Methods(http.MethodPost)
	router.HandleFunc("/api/auth/refresh", authHandler.Refresh).Methods(http.MethodPost)
	router.HandleFunc("/api/auth/logout", authHandler.Logout).Methods(http.MethodPost)

	// Protected API
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.AuthRequired(jwt))
	api.HandleFunc("/auth/me", authHandler.Me).Methods(http.MethodGet)
	api.HandleFunc("/auth/change-password", authHandler.ChangePassword).Methods(http.MethodPost, http.MethodPut)
	api.HandleFunc("/settings", settingsHandler.GetMine).Methods(http.MethodGet)
	api.HandleFunc("/settings", settingsHandler.UpdateMine).Methods(http.MethodPut, http.MethodPatch)
	api.HandleFunc("/custom-options", customOptsHandler.Get).Methods(http.MethodGet)
	api.HandleFunc("/custom-options", customOptsHandler.Add).Methods(http.MethodPost)
	api.HandleFunc("/pets", petsHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/pets", petsHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/pets/{id}", petsHandler.Get).Methods(http.MethodGet)
	api.HandleFunc("/pets/{id}", petsHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/pets/{id}", petsHandler.Delete).Methods(http.MethodDelete)
	api.HandleFunc("/pets/{petId}/vaccinations", vaccHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/pets/{petId}/vaccinations", vaccHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/pets/{petId}/vaccinations/{id}", vaccHandler.Get).Methods(http.MethodGet)
	api.HandleFunc("/pets/{petId}/vaccinations/{id}", vaccHandler.Update).Methods(http.MethodPut)
	api.HandleFunc("/pets/{petId}/vaccinations/{id}", vaccHandler.Delete).Methods(http.MethodDelete)
	api.HandleFunc("/pets/{petId}/weights", weightsHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/pets/{petId}/weights", weightsHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/pets/{petId}/weights/{id}", weightsHandler.Delete).Methods(http.MethodDelete)
	api.HandleFunc("/pets/{petId}/documents", docsHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/pets/{petId}/documents", docsHandler.Create).Methods(http.MethodPost)
	api.HandleFunc("/pets/{petId}/documents/{id}", docsHandler.Get).Methods(http.MethodGet)
	api.HandleFunc("/pets/{petId}/documents/{id}", docsHandler.Update).Methods(http.MethodPut, http.MethodPatch)
	api.HandleFunc("/pets/{petId}/documents/{id}", docsHandler.Delete).Methods(http.MethodDelete)
	api.HandleFunc("/pets/{petId}/photos", photosHandler.List).Methods(http.MethodGet)
	api.HandleFunc("/pets/{petId}/photos", photosHandler.Upload).Methods(http.MethodPost)
	api.HandleFunc("/pets/{petId}/photos/{id}/avatar", photosHandler.SetAvatar).Methods(http.MethodPut, http.MethodPatch)
	api.HandleFunc("/pets/{petId}/photos/{id}", photosHandler.Delete).Methods(http.MethodDelete)

	// Serve uploaded files (photos, documents) — under API so auth applies
	api.PathPrefix("/uploads/").Handler(handlers.ServeUploads(uploadDir, "/api/uploads"))

	// Admin-only: default dropdown options (species, breeds, vaccinations)
	api.Handle("/admin/default-options", middleware.AdminRequired(http.HandlerFunc(defaultOptsHandler.List))).Methods(http.MethodGet)
	api.Handle("/admin/default-options", middleware.AdminRequired(http.HandlerFunc(defaultOptsHandler.Create))).Methods(http.MethodPost)
	api.Handle("/admin/default-options/{id}", middleware.AdminRequired(http.HandlerFunc(defaultOptsHandler.Update))).Methods(http.MethodPut, http.MethodPatch)
	api.Handle("/admin/default-options/{id}", middleware.AdminRequired(http.HandlerFunc(defaultOptsHandler.Delete))).Methods(http.MethodDelete)

	// Admin-only: user management
	api.Handle("/users", middleware.AdminRequired(http.HandlerFunc(usersHandler.List))).Methods(http.MethodGet)
	api.Handle("/users", middleware.AdminRequired(http.HandlerFunc(usersHandler.Create))).Methods(http.MethodPost)
	api.Handle("/users/{id}/role", middleware.AdminRequired(http.HandlerFunc(usersHandler.UpdateRole))).Methods(http.MethodPut, http.MethodPatch)
	api.Handle("/users/{id}/settings", middleware.AdminRequired(http.HandlerFunc(settingsHandler.GetForUser))).Methods(http.MethodGet)
	api.Handle("/users/{id}/settings", middleware.AdminRequired(http.HandlerFunc(settingsHandler.UpdateForUser))).Methods(http.MethodPut, http.MethodPatch)

	// SPA static files (index.html for non-API routes)
	sub, _ := fs.Sub(staticFS, "static")
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if !strings.HasPrefix(path, "api/") {
			if _, err := sub.Open(path); err == nil {
				http.FileServer(http.FS(sub)).ServeHTTP(w, r)
				return
			}
			// SPA fallback: serve index.html
			if data, err := fs.ReadFile(sub, "index.html"); err == nil {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write(data)
				return
			}
		}
		http.NotFound(w, r)
	})

	// Security: in production (!Development), warn about default/permissive settings and enable HSTS when using Secure cookies
	if cfg.Development {
		log.Printf("Development mode: security relaxed (cookies not Secure, no HSTS, no strict warnings).")
	} else {
		if cfg.JWTSecret == "change-me-in-production-min-32-chars" {
			log.Printf("[SECURITY] WARNING: JWT_SECRET is not set; using default. Set a strong secret in production (e.g. openssl rand -base64 32).")
		}
		if cfg.CORSOrigins == "*" {
			log.Printf("[SECURITY] WARNING: CORS_ORIGINS is '*' (allow all origins). Omit CORS_ORIGINS for same-origin-only (recommended when frontend and API share a host).")
		}
		if cfg.ForwardedEmailHeader != "" {
			log.Printf("[SECURITY] Trusted proxy auth is enabled (proxy IPs from TRUSTED_PROXIES or private/loopback when TRUST_PRIVATE_PROXIES=true). Ensure only your proxy can reach the app to prevent header spoofing.")
		}
	}

	hstsMaxAge := 0
	if !cfg.Development {
		hstsMaxAge = 31536000 // 1 year when using HTTPS in production
	}
	handler := middleware.TrustedProxyAuth(cfg, gormDB, jwt, refreshStore, cfg.DefaultWeightUnit, cfg.DefaultCurrency, cfg.DefaultLanguage)(router)
	throttled := middleware.ThrottleByPath(cfg, cfg.RateLimitAuthLoginPerMin, cfg.RateLimitAuthOtherPerMin, cfg.RateLimitAPIPerMin)(handler)
	chain := middleware.Logging(middleware.SecurityHeaders(hstsMaxAge, cfg)(middleware.CORS(cfg.CORSOrigins, cfg)(throttled)))

	addr := ":" + strconv.Itoa(cfg.ServerPort)
	log.Printf("Listening on %s", addr)
	debuglog.Debugf("upload dir: %s", uploadDir)
	if err := http.ListenAndServe(addr, chain); err != nil {
		log.Fatalf("server: %v", err)
	}
}
