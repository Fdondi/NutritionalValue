package main

import (
	"context"
	"flag"
	"log"

	"github.com/franckalain/nutritionalvalue/internal/config"
	"github.com/franckalain/nutritionalvalue/internal/database"
	"github.com/franckalain/nutritionalvalue/internal/ml"
	"github.com/franckalain/nutritionalvalue/internal/server"
)

func main() {
	configPath := flag.String("config", config.GetConfigPath(), "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database
	db, err := database.NewSQLiteDB(cfg.Database.Path)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize ML service
	model, err := ml.NewModel(cfg.ML.Type)
	if err != nil {
		log.Fatal("Failed to create ML model:", err)
	}

	if err := model.Load(context.Background()); err != nil {
		log.Fatal("Failed to load ML model:", err)
	}

	// Initialize and start server
	srv := server.New(db, model, true)
	if err := srv.Start(cfg.Server.Port, cfg.Server.StaticDir); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
