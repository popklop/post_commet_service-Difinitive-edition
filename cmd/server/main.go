package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"ozon-testovoe/config"
	graph "ozon-testovoe/graph/resolver"
	"ozon-testovoe/internal/repository"
	"ozon-testovoe/internal/repository/memory"
	"ozon-testovoe/internal/repository/postgres"
	"ozon-testovoe/internal/server"
	"ozon-testovoe/internal/service"
)

func main() {
	cfg := config.Load()

	var postRepo repository.PostRepos
	var commentRepo repository.CommentRepos
	var db *sql.DB

	switch cfg.StorageType {
	case "postgres":
		var err error
		db, err = sql.Open("postgres", cfg.PostgresDSN)
		if err != nil {
			log.Fatalf("failed to connect to postgres: %v", err)
		}
		if err := db.Ping(); err != nil {
			log.Fatalf("failed to ping postgres: %v", err)
		}
		log.Println("connected to PostgreSQL")
		postRepo = postgres.NewPostRepos(db)
		commentRepo = postgres.NewCommentRepos(db)
	default:
		log.Println("using in-memory storage")
		postRepo = memory.NewPostRepository()
		commentRepo = memory.NewCommentRepository()
	}

	postService := service.NewPostService(postRepo)
	commentService := service.NewCommentService(postRepo, commentRepo)

	resolver := &graph.Resolver{
		PostService:    postService,
		CommentService: commentService,
	}

	httpHandler := server.NewServer(commentService, resolver)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	httpServer := &http.Server{
		Addr:    ":" + port,
		Handler: httpHandler,
	}

	go func() {
		log.Printf("Server started at http://127.0.0.1:%s", port)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	if db != nil {
		if err := db.Close(); err != nil {
			log.Printf("error closing database connection: %v", err)
		} else {
			log.Println("database connection closed")
		}
	}
}
