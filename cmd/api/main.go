package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"

	"github.com/Babatunde50/distributask/internal/database"
	"github.com/Babatunde50/distributask/internal/version"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Llongfile)

	err := run(logger)
	if err != nil {
		trace := debug.Stack()
		logger.Fatalf("%s\n%s", err, trace)
	}
}

type config struct {
	baseURL  string
	httpPort int
	db       struct {
		dsn         string
		automigrate bool
	}
	jwt struct {
		secretKey string
	}
}

type application struct {
	config config
	db     *database.DB
	logger *log.Logger
	wg     sync.WaitGroup
}

func run(logger *log.Logger) error {
	var cfg config

	flag.StringVar(&cfg.baseURL, "base-url", "http://localhost:4444", "base URL for the application")
	flag.IntVar(&cfg.httpPort, "http-port", 4444, "port to listen on for HTTP requests")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres://distributask:pa55word@localhost/distributask?sslmode=disable", "postgreSQL DSN")
	flag.BoolVar(&cfg.db.automigrate, "db-automigrate", true, "run migrations on startup")
	flag.StringVar(&cfg.jwt.secretKey, "jwt-secret-key", "xb37u2w4i57oooowambofjbhfbkemrj7", "secret key for JWT authentication")

	// postgres://distributask:pa55word@localhost/distributask?sslmode=disable

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	db, err := database.New(cfg.db.dsn, cfg.db.automigrate)
	if err != nil {
		return err
	}
	defer db.Close()

	app := &application{
		config: cfg,
		db:     db,
		logger: logger,
	}

	return app.serveHTTP()
}
