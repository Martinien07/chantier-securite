package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"srv.exe.dev/db"
	"srv.exe.dev/srv"
)

var flagListenAddr = flag.String("listen", ":8000", "address to listen on")

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()

	database, err := db.Open("db.sqlite3")
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer database.Close()

	if err := db.RunMigrations(database); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	server, err := srv.NewServer(database)
	if err != nil {
		return fmt.Errorf("create server: %w", err)
	}

	slog.Info("server starting", "addr", *flagListenAddr)
	return http.ListenAndServe(*flagListenAddr, server.Handler())
}
