package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"chelbo/backend/internal/ai"
	"chelbo/backend/internal/auth"
	"chelbo/backend/internal/chat"
	"chelbo/backend/internal/database"
	"chelbo/backend/internal/file"
	"chelbo/backend/internal/pkg/config"
	"chelbo/backend/internal/pkg/logger"
	"chelbo/backend/internal/redis"
	"chelbo/backend/internal/user"

	"github.com/gorilla/mux"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Init(cfg.Server.Environment)
	logger.Info("🚀 Starting Chelbo Messenger server...")

	// Initialize database
	if err := database.Init(&cfg.Database); err != nil {
		logger.Errorf("Failed to initialize database: %v", err)
		log.Fatal("Failed to initialize database: ", err)
	}
	defer database.Close()

	// Initialize Redis (опционально)
	if err := redis.Init(&cfg.Redis); err != nil {
		logger.Warnf("Failed to initialize Redis: %v (continuing without Redis)", err)
	} else {
		defer redis.Close()
	}

	// Initialize WebSocket hub
	hub := chat.GetHub()

	// Initialize handlers
	authHandler := auth.NewAuthHandler(cfg, database.DB)
	userHandler := user.NewHandler(database.DB)
	chatHandler := chat.NewChatHandler(database.DB, hub)
	fileHandler := file.NewHandler(cfg)
	aiHandler := ai.NewHandler(cfg.AI.Enabled, cfg.AI.MockResponses)

	// Setup router
	router := mux.NewRouter()

	// Глобальный CORS middleware для всех роутов
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Разрешаем все источники для разработки
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Обрабатываем preflight запросы
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Health check endpoint
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}).Methods("GET")

	// Metrics endpoint
	router.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# HELP chelbo_http_requests_total Total HTTP requests\n# TYPE chelbo_http_requests_total counter\nchelbo_http_requests_total 0\n"))
	}).Methods("GET")

	// WebSocket route
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			logger.Error("WebSocket: no token provided")
			http.Error(w, "token required", http.StatusUnauthorized)
			return
		}

		claims, err := auth.ValidateJWT(token, &cfg.JWT)
		if err != nil {
			logger.Errorf("WebSocket: invalid token: %v", err)
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		logger.Infof("WebSocket: user %d connected", claims.UserID)
		hub.HandleWebSocket(w, r, claims.UserID)
	})

	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Public routes (no auth required)
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST", "OPTIONS")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST", "OPTIONS")

	// Protected routes (auth required)
	protected := api.PathPrefix("").Subrouter()
	protected.Use(auth.AuthMiddleware(&cfg.JWT))

	// Auth routes
	protected.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST", "OPTIONS")
	protected.HandleFunc("/auth/me", authHandler.GetMe).Methods("GET", "OPTIONS")

	// User routes
	protected.HandleFunc("/users/me", userHandler.UpdateProfile).Methods("PUT", "OPTIONS")
	protected.HandleFunc("/users/me/avatar", fileHandler.UploadAvatar).Methods("POST", "OPTIONS")
	protected.HandleFunc("/users/search", userHandler.SearchUsers).Methods("GET", "OPTIONS")
	protected.HandleFunc("/users/{id:[0-9]+}", userHandler.GetUserByID).Methods("GET", "OPTIONS")

	// Chat routes
	protected.HandleFunc("/chats", chatHandler.GetChats).Methods("GET", "OPTIONS")
	protected.HandleFunc("/chats/private/{id:[0-9]+}", chatHandler.CreatePrivateChat).Methods("POST", "OPTIONS")
	protected.HandleFunc("/chats/group", chatHandler.CreateGroup).Methods("POST", "OPTIONS")
	protected.HandleFunc("/chats/{id:[0-9]+}/messages", chatHandler.GetMessages).Methods("GET", "OPTIONS")
	protected.HandleFunc("/chats/{id:[0-9]+}/messages", chatHandler.SendMessage).Methods("POST", "OPTIONS")

	// Message routes
	protected.HandleFunc("/messages/{id:[0-9]+}", chatHandler.DeleteMessage).Methods("DELETE", "OPTIONS")
	protected.HandleFunc("/messages/{id:[0-9]+}/read", chatHandler.MarkAsRead).Methods("POST", "OPTIONS")
	protected.HandleFunc("/messages/{id:[0-9]+}/forward", chatHandler.ForwardMessage).Methods("POST", "OPTIONS")

	// File routes
	protected.HandleFunc("/files/upload", fileHandler.UploadFile).Methods("POST", "OPTIONS")
	router.HandleFunc("/uploads/{filename}", fileHandler.GetFile).Methods("GET", "OPTIONS")

	// AI routes
	protected.HandleFunc("/ai/ask", aiHandler.Ask).Methods("POST", "OPTIONS")
	protected.HandleFunc("/ai/translate", aiHandler.Translate).Methods("POST", "OPTIONS")
	protected.HandleFunc("/ai/summarize", aiHandler.Summarize).Methods("POST", "OPTIONS")

	// Serve static files
	router.PathPrefix("/").Handler(http.FileServer(http.Dir("./web")))

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.Infof("🌐 Server listening on %s:%s", cfg.Server.Host, cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Failed to start server: %v", err)
			log.Fatal(err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	logger.Info("✅ Server stopped gracefully")
}
