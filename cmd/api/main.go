package main

import (
	"context"
	"database/sql"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"greenlight.codewithyash.dev/internal/data"
	"greenlight.codewithyash.dev/internal/mailer"
)

const version = "1.0.0"

type config struct {
	port int
	env string
	db 	struct {
		dsn				string
		maxOpenConns	int
		maxIdleConns	int
		maxIdleTime		time.Duration
	}
	limiter struct {
		rps				float64
		burst 			int
		enabled 		bool
	}
	smtp struct {
		host 			string
		port 			int
		username		string
		password		string
		sender			string
	}
}

type application struct {
	config 	config
	logger 	*slog.Logger
	models  data.Models
	mailer	mailer.Mailer
	wg 		sync.WaitGroup
}


func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	var cfg config

	err := godotenv.Load(".env")
	if err != nil {
		logger.Error("error loading .env file", "error", err)
		os.Exit(1)
	}

	err = cfg.getConfiguration()
	if err != nil {
		logger.Error("error loading configuration", "error", err)
		os.Exit(1)
	}



	db, err := openDB(cfg)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	logger.Info("database connection pool established")

	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}

	err = app.serve()
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
}




func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(cfg.db.maxIdleTime)
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}
