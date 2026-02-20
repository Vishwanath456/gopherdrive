package main

import (
	"context"
	"database/sql"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcserver "gopherdrive/internal/grpc"
	"gopherdrive/internal/repository"
	"gopherdrive/internal/rest"
	worker "gopherdrive/internal/work"
	pb "gopherdrive/proto"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
)

func main() {

	// ✅ Structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// ✅ Connect to MySQL
	db, err := sql.Open("mysql", "gopher:gopher123@tcp(127.0.0.1:3306)/gopherdrive")
	if err != nil {
		slog.Error("db open failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		slog.Error("db ping failed", "error", err)
		os.Exit(1)
	}

	slog.Info("database connected")

	// ✅ Repository
	repo := repository.NewMySQLRepo(db)

	// ✅ Worker Pool
	wp := worker.NewWorkerPool(5)

	// ✅ REST handler
	handler := rest.NewHandler(wp, repo)

	// ✅ Create mux ONCE
	mux := http.NewServeMux()

	// API routes

	// Health route
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {

		if err := db.Ping(); err != nil {
			http.Error(w, "database down", http.StatusInternalServerError)
			return
		}

		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		if err != nil {
			http.Error(w, "grpc down", http.StatusInternalServerError)
			return
		}
		conn.Close()

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Serve frontend

	// ✅ HTTP Server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start HTTP
	go func() {
		slog.Info("http server running", "port", 8080)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("http server error", "error", err)
		}
	}()

	// ✅ Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		slog.Error("grpc listen failed", "error", err)
		os.Exit(1)
	}

	grpcSrv := grpc.NewServer()
	pb.RegisterMetadataServiceServer(grpcSrv, grpcserver.NewServer(repo))

	go func() {
		slog.Info("grpc server running", "port", 50051)
		if err := grpcSrv.Serve(lis); err != nil {
			slog.Error("grpc server error", "error", err)
		}
	}()

	mux.HandleFunc("/files", handler.UploadHandler)
	mux.HandleFunc("/files/", handler.GetFileHandler)
	mux.HandleFunc("/api/files", handler.ListFilesHandler)

	mux.Handle("/", http.FileServer(http.Dir("./web")))

	// ✅ Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	slog.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	server.Shutdown(ctx)
	grpcSrv.GracefulStop()
	wp.Wg.Wait()

	slog.Info("server exited cleanly")
}
