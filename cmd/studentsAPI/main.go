package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pixel2packet/studentAPI/internal/config"
	"github.com/pixel2packet/studentAPI/internal/http/handlers/student"
	"github.com/pixel2packet/studentAPI/internal/storage/sqlite"
)

func main() {
	// load config
	cfg := config.MustLoad()
	// fmt.Println("loading config:", cfg.StoragePath)
	/* The MustLoad() function does:-----------------
	       Reads CONFIG_PATH from env or command-line
	       Validates the file exists
	       Loads values into a Config struct
	   	Crashes early if anything goes wrong
	     	Returns a usable pointer to the config */

	// database setup
	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("storage intialized", slog.String("env", cfg.Env), slog.String("version", "1.0.0"))

	/* Think of *sql.DB as:
	A smart manager that:
		Keeps a pool of open connections to the DB (to reuse them)
		Sends SQL queries
		Reads query results (e.g., rows from a SELECT)
		Handles timeouts, retries, and concurrency for you
	*/

	// setup router
	router := http.NewServeMux()

	router.HandleFunc("POST /api/students", student.New(storage))
	router.HandleFunc("GET /api/students/{id}", student.GetById(storage))
	router.HandleFunc("GET /api/students/", student.GetList(storage))

	// setup server
	/*
	1. Starts server in background using goroutine
	2. Waits for shutdown signal (Ctrl+C)
	3. When received → logs "shutting down"
	4. Creates 5s timeout context
	5. Calls server.Shutdown(ctx) to gracefully close
	6. Logs error if shutdown fails
	7. Else logs "successfully stopped"
	*/
	server := http.Server{
		Addr:    cfg.Addr,
		Handler: router,
	}

	slog.Info("server started", slog.String("addr", cfg.Addr))

	// fmt.Printf("server started on: %s", cfg.HTTPServer.Addr)

	done := make(chan os.Signal, 1)

	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGALRM)
	/* Tells Go to send OS signals into your done channel
	os.Interrupt = when you press Ctrl+C
	syscall.SIGINT = same as os.Interrupt (just OS-level)
	SIGALRM is usually used for timers/alarms on Unix systems (you may not need this)
	*/


	go func() {
		err := server.ListenAndServe()
		if err != nil {
			log.Fatalf("failed to start server %s", err)
		}
	}()

	<-done

	slog.Info("shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))
	}

	// or

	// if err != nil {
	// 	slog.Error("failed to shutdown server", slog.String("error", err.Error()))
	// }

	// server.Shutdown()

	slog.Info("server shutdown successfully")

}
