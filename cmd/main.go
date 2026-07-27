package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"player-node/internal/cache"
	"player-node/internal/config"
	"player-node/internal/db/database"
	"player-node/internal/handlers"
	"player-node/internal/middleware"
	"player-node/internal/services"
)

// version ถูกฝังตอน build โดย GitHub Actions: -ldflags="-X main.version=v1.x.x"
var version = "dev"

func main() {
	config.Load()
	log.Printf("🚀 Starting Player Node %s", version)

	// log ออก stdout อย่างเดียว ให้ systemd/journald เก็บและหมุนให้
	// (journalctl -u player-node -f) — ของเดิมเขียนลงไฟล์อย่างเดียวผ่าน
	// logger.Init ซึ่ง log.SetOutput ทับ stdout ทำให้ journal เห็นแค่บรรทัด
	// ก่อนหน้านั้น และไฟล์ที่หมุนไว้ก็ไม่มีใครลบ

	// ── Redis lookup cache (optional — ไม่มี REDIS_URL = ปิด) ────
	cache.Init(config.AppConfig.RedisURL)

	// ── MongoDB ───────────────────────────────────────────────
	if err := database.Connect(); err != nil {
		log.Printf("❌ Failed to connect to MongoDB: %v", err)
		time.Sleep(5 * time.Second) // กัน systemd restart-loop รัวๆ
		os.Exit(1)
	}
	defer database.Disconnect()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// ── Settings/domains/spaces sync: ทุก 1 นาที ──────────────
	go services.StartSettingSyncScheduler(ctx)

	// ── Templates ─────────────────────────────────────────────
	if err := handlers.InitTemplates(); err != nil {
		log.Printf("❌ Failed to load templates: %v", err)
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}

	// ── Handlers ──────────────────────────────────────────────
	h := handlers.NewHandler(handlers.Handler{})

	// ── HTTP server ───────────────────────────────────────────
	mux := http.NewServeMux()

	mux.HandleFunc("/embed/", h.Embed)
	mux.HandleFunc("/favicon.ico", handlers.Favicon)

	mux.HandleFunc("/health", handlers.Health(version))

	mux.HandleFunc("/", h.Home)

	server := &http.Server{
		Addr:    ":" + config.AppConfig.Port,
		Handler: middleware.CORS(mux),
	}

	go func() {
		log.Printf("🌐 Server started at http://localhost:%s", config.AppConfig.Port)
		log.Printf("📍 Endpoints:")
		log.Printf("   GET /embed/{fileSlug}          - Video player embed page")
		log.Printf("   GET /health                    - Health check")
		log.Printf("   GET /{slug}.{ext}              - Proxy stream files")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("🛑 Shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("⚠️ Server shutdown failed: %v", err)
	} else {
		log.Println("✅ Server gracefully stopped")
	}
}
