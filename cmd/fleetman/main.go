package main

import (
	"database/sql"
	"log"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"

	"github.com/akalpaki/alchemy-test/internal/spacecraft"
)

func main() {
	cfg := loadConfig()
	db := setupDatabase(cfg.ConnStr)
	logger := setupLogger(cfg.LogLevel, cfg.LogFile)
	repo := spacecraft.NewRepository(db)
	h := spacecraft.Routes(repo, logger)

	srv := http.Server{
		Handler: h,
	}

	log.Println("launcing server...")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("main: starting server: %s", err.Error())
	}
}

func setupDatabase(connStr string) *sql.DB {
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		log.Fatalf("main: opening database: %s", err.Error())
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("main: pinging database: %s", err.Error())
	}

	tableSpacecrafts := `
	CREATE TABLE IF NOT EXISTS spacecrafts (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		class VARCHAR(255) NOT NULL,
		status VARCHAR(255) NOT NULL,
		image VARCHAR(255) NOT NULL,
		crew INT NOT NULL,
		value INT NOT NULL
	);
	`
	tableArmaments := `
	CREATE TABLE IF NOT EXISTS armaments (
		id INT AUTO_INCREMENT PRIMARY KEY,
		craft_id INT,
		title TEXT NOT NULL,
		quanity INT NOT NULL,
		CONSTRAINT craft_fk FOREIGN KEY (craft_id)
			REFERENCES spacecrafts(id)
			ON DELETE CASCADE
	);
	`
	if _, err := db.Exec(tableSpacecrafts); err != nil {
		log.Fatalf("main: creating tables: %s", err.Error())
	}

	if _, err := db.Exec(tableArmaments); err != nil {
		log.Fatalf("main: creating tables: %s", err.Error())
	}
	return db
}

func setupLogger(logLevel slog.Level, logFile string) *slog.Logger {
	var h slog.Handler
	if logFile != os.Stdout.Name() {
		f, err := os.Open(logFile)
		if err != nil {
			panic("invalid log output file given!")
		}
		h = slog.NewJSONHandler(f, &slog.HandlerOptions{AddSource: false, Level: logLevel})
	} else {
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{AddSource: true, Level: logLevel})
	}
	return slog.New(h)
}
