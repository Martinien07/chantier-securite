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

var (
	flagListenAddr = flag.String("listen", ":8000", "adresse d'écoute du serveur")
	flagDSN        = flag.String("dsn", "", "DSN MySQL (ex: user:pass@tcp(localhost:3306)/make_hse?parseTime=true)")
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	flag.Parse()

	// Priorité : flag -dsn > variable d'environnement MYSQL_DSN
	dsn := *flagDSN
	if dsn == "" {
		dsn = os.Getenv("MYSQL_DSN")
	}
	if dsn == "" {
		return fmt.Errorf(
			"DSN MySQL manquant.\n" +
				"Utilise l'une de ces méthodes :\n" +
				"  1. Variable d'env : export MYSQL_DSN=\"user:pass@tcp(localhost:3306)/make_hse?parseTime=true\"\n" +
				"  2. Flag           : ./safesite -dsn \"user:pass@tcp(localhost:3306)/make_hse?parseTime=true\"",
		)
	}

	database, err := db.Open(dsn)
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
