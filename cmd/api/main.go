package main

import (
	"flag"
	"fmt"
	"runtime/debug"
	"sync"

	"github.com/Babatunde50/distributask/internal/database"
	"github.com/Babatunde50/distributask/internal/version"
	"github.com/Babatunde50/distributask/internal/worker"
	"github.com/hibiken/asynq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	err := run()
	if err != nil {
		trace := debug.Stack()
		log.Fatal().Msgf("%s\n%s", err, trace)

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
	config          config
	db              *database.DB
	wg              sync.WaitGroup
	taskDistributor worker.TaskDistributor
}

func run() error {
	var cfg config

	flag.StringVar(&cfg.baseURL, "base-url", "http://api:4444", "base URL for the application")
	flag.IntVar(&cfg.httpPort, "http-port", 4444, "port to listen on for HTTP requests")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "distributask:pa55word@postgres/distributask?sslmode=disable", "postgreSQL DSN")
	flag.BoolVar(&cfg.db.automigrate, "db-automigrate", true, "run migrations on startup")
	flag.StringVar(&cfg.jwt.secretKey, "jwt-secret-key", "xb37u2w4i57oooowambofjbhfbkemrj7", "secret key for JWT authentication")

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

	redisConnOpt := asynq.RedisClientOpt{
		Addr: "redis:6379",
		DB:   0,
	}

	processor := worker.NewRedisTaskProcessor(redisConnOpt, db)

	go processor.Start()

	taskDistributor := worker.NewRedisTaskDistributor(redisConnOpt)

	defer db.Close()

	app := &application{
		config:          cfg,
		db:              db,
		taskDistributor: taskDistributor,
	}

	return app.serveHTTP()
}
