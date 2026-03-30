package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artshop/backend/internal/config"
	"github.com/artshop/backend/internal/database"
	"github.com/artshop/backend/internal/handlers"
	"github.com/artshop/backend/internal/middleware"
	"github.com/artshop/backend/internal/repository"
	"github.com/artshop/backend/internal/services"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// -------------------------------------------------------------------------
	// 1. Load configuration
	// -------------------------------------------------------------------------
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}
	slog.Info("configuration loaded", "env", cfg.Env, "port", cfg.Port)

	// -------------------------------------------------------------------------
	// 2. Connect to database
	// -------------------------------------------------------------------------
	db, err := database.InitDB(cfg)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	// -------------------------------------------------------------------------
	// 3. Create repositories
	// -------------------------------------------------------------------------
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)
	orderRepo := repository.NewOrderRepository(db)
	cartRepo := repository.NewCartRepository(db)
	reviewRepo := repository.NewReviewRepository(db)

	// -------------------------------------------------------------------------
	// 4. Create services
	// -------------------------------------------------------------------------
	authService := services.NewAuthService(userRepo, cfg)
	productService := services.NewProductService(productRepo, cfg)
	orderService := services.NewOrderService(orderRepo, cartRepo, productRepo, cfg)
	aiService := services.NewAIService(cfg)
	adminService := services.NewAdminService(userRepo, productRepo, orderRepo, db)

	// -------------------------------------------------------------------------
	// 5. Create handlers
	// -------------------------------------------------------------------------
	authHandler := handlers.NewAuthHandler(authService)
	productHandler := handlers.NewProductHandler(productService)
	orderHandler := handlers.NewOrderHandler(orderService)
	cartHandler := handlers.NewCartHandler(cartRepo)
	reviewHandler := handlers.NewReviewHandler(reviewRepo)
	wishlistHandler := handlers.NewWishlistHandler(cartRepo)
	adminHandler := handlers.NewAdminHandler(adminService)
	aiHandler := handlers.NewAIHandler(aiService, db)
	uploadHandler := handlers.NewUploadHandler(cfg)
	notificationHandler := handlers.NewNotificationHandler(db)

	// -------------------------------------------------------------------------
	// 6. Setup chi router with global middleware
	// -------------------------------------------------------------------------
	r := chi.NewRouter()

	// Rate limiter (with background cleanup goroutine).
	rateLimiter := middleware.NewRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst)
	defer rateLimiter.Stop()

	// Global middleware.
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Timeout(60 * time.Second))

	// CORS.
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting.
	r.Use(rateLimiter.Middleware)

	// -------------------------------------------------------------------------
	// 7. Mount all routes
	// -------------------------------------------------------------------------

	// Health check.
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok"}`)
	})

	r.Route("/api", func(r chi.Router) {
		// =====================================================================
		// Public routes (no authentication required)
		// =====================================================================

		// Auth.
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.RefreshToken)

			// Auth-required auth routes.
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAuth(cfg.JWTSecret))
				r.Post("/register-seller", authHandler.RegisterSeller)
				r.Get("/me", authHandler.Me)
			})
		})

		// Products (public listing).
		r.Route("/products", func(r chi.Router) {
			r.Get("/", productHandler.List)
			r.Get("/featured", productHandler.Featured)
			r.Get("/search", productHandler.Search)
			r.Get("/{id}", productHandler.GetByID)

			// Product reviews (public read).
			r.Get("/{id}/reviews", reviewHandler.GetByProduct)

			// Auth-required review creation.
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAuth(cfg.JWTSecret))
				r.Post("/{id}/reviews", reviewHandler.Create)
			})

			// Seller/admin product management.
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAuth(cfg.JWTSecret))
				r.Use(middleware.RequireRole("seller", "admin"))
				r.Post("/", productHandler.Create)
				r.Put("/{id}", productHandler.Update)
				r.Delete("/{id}", productHandler.Delete)
			})
		})

		// Categories (public).
		r.Get("/categories", productHandler.Categories)

		// =====================================================================
		// Authenticated routes
		// =====================================================================
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(cfg.JWTSecret))

			// Cart.
			r.Route("/cart", func(r chi.Router) {
				r.Get("/", cartHandler.GetCart)
				r.Post("/", cartHandler.AddItem)
				r.Put("/{productId}", cartHandler.UpdateQuantity)
				r.Delete("/{productId}", cartHandler.RemoveItem)
				r.Delete("/", cartHandler.Clear)
			})

			// Wishlist.
			r.Route("/wishlist", func(r chi.Router) {
				r.Get("/", wishlistHandler.GetWishlist)
				r.Post("/", wishlistHandler.AddToWishlist)
				r.Delete("/{productId}", wishlistHandler.RemoveFromWishlist)
			})

			// Orders.
			r.Route("/orders", func(r chi.Router) {
				r.Post("/", orderHandler.Create)
				r.Get("/", orderHandler.List)
				r.Get("/{id}", orderHandler.GetByID)
				r.Put("/{id}/status", orderHandler.UpdateStatus)
			})

			// Notifications.
			r.Route("/notifications", func(r chi.Router) {
				r.Get("/", notificationHandler.List)
				r.Get("/unread-count", notificationHandler.UnreadCount)
				r.Put("/read-all", notificationHandler.MarkAllAsRead)
				r.Put("/{id}/read", notificationHandler.MarkAsRead)
			})

			// Reviews (delete).
			r.Delete("/reviews/{id}", reviewHandler.Delete)

			// Upload.
			r.Post("/upload", uploadHandler.Upload)

			// AI recommendations (any authenticated user).
			r.Get("/ai/recommendations", aiHandler.Recommendations)
		})

		// =====================================================================
		// Seller routes
		// =====================================================================
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(cfg.JWTSecret))
			r.Use(middleware.RequireRole("seller", "admin"))

			r.Get("/seller/orders", orderHandler.SellerOrders)

			// AI content generation.
			r.Post("/ai/generate-description", aiHandler.GenerateDescription)
			r.Post("/ai/generate-tags", aiHandler.GenerateTags)
		})

		// =====================================================================
		// Admin routes
		// =====================================================================
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.RequireAuth(cfg.JWTSecret))
			r.Use(middleware.RequireRole("admin"))

			r.Get("/dashboard", adminHandler.Dashboard)
			r.Get("/users", adminHandler.ListUsers)
			r.Put("/users/{id}/role", adminHandler.UpdateUserRole)
			r.Put("/users/{id}/toggle-active", adminHandler.ToggleUserActive)
			r.Get("/orders", adminHandler.ListOrders)
			r.Put("/products/{id}/toggle-featured", adminHandler.ToggleProductFeatured)
			r.Get("/revenue", adminHandler.Revenue)
		})
	})

	// -------------------------------------------------------------------------
	// 8. Start HTTP server with graceful shutdown
	// -------------------------------------------------------------------------
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to listen for shutdown signals.
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine.
	go func() {
		slog.Info("starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal.
	sig := <-shutdown
	slog.Info("shutdown signal received", "signal", sig)

	// Graceful shutdown with a 30-second deadline.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped gracefully")
}
