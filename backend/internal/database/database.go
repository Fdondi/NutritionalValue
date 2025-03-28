package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"time"

	"github.com/franckalain/nutritionalvalue/internal/models"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaFS embed.FS

// DB interface defines the methods our database should implement
type DB interface {
	SaveNutritionalInfo(ctx context.Context, info *models.NutritionalInfo) error
	GetNutritionalInfo(ctx context.Context, id string) (*models.NutritionalInfo, error)
	SaveScan(ctx context.Context, scan *models.NutritionScan) error
	UpdateScanStatus(ctx context.Context, id, status string, errMsg string) error
	GetRecentNutritionalInfo(ctx context.Context, limit int) ([]*models.NutritionalInfo, error)
	Close() error
}

// SQLiteDB implements the DB interface
type SQLiteDB struct {
	db *sql.DB
}

// NewSQLiteDB creates a new SQLite database connection
func NewSQLiteDB(dbPath string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Enable foreign keys and WAL mode for better concurrency
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, fmt.Errorf("error enabling foreign keys: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, fmt.Errorf("error enabling WAL mode: %w", err)
	}

	// Initialize database schema
	if err := initializeSchema(db); err != nil {
		return nil, fmt.Errorf("error initializing schema: %w", err)
	}

	return &SQLiteDB{db: db}, nil
}

func initializeSchema(db *sql.DB) error {
	// Read schema file
	schemaBytes, err := schemaFS.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("error reading schema file: %w", err)
	}

	// Execute schema
	if _, err := db.Exec(string(schemaBytes)); err != nil {
		return fmt.Errorf("error executing schema: %w", err)
	}

	log.Println("Database schema initialized successfully")
	return nil
}

// SaveNutritionalInfo saves nutritional information to the database
func (s *SQLiteDB) SaveNutritionalInfo(ctx context.Context, info *models.NutritionalInfo) error {
	query := `
		INSERT INTO nutritional_info (
			id, total_weight, calories, protein, carbs, fat, fiber, sugar,
			image_path, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			total_weight = excluded.total_weight,
			calories = excluded.calories,
			protein = excluded.protein,
			carbs = excluded.carbs,
			fat = excluded.fat,
			fiber = excluded.fiber,
			sugar = excluded.sugar,
			image_path = excluded.image_path,
			updated_at = excluded.updated_at
	`

	now := time.Now()
	if info.CreatedAt.IsZero() {
		info.CreatedAt = now
	}
	info.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, query,
		info.ID, info.TotalWeight,
		info.Calories, info.Protein, info.Carbs, info.Fat, info.Fiber,
		info.Sugar, info.ImagePath, info.CreatedAt, info.UpdatedAt,
	)
	return err
}

// GetNutritionalInfo retrieves nutritional information from the database
func (s *SQLiteDB) GetNutritionalInfo(ctx context.Context, id string) (*models.NutritionalInfo, error) {
	query := `
		SELECT id, total_weight, calories, protein, carbs, fat, fiber, sugar,
			image_path, created_at, updated_at
		FROM nutritional_info WHERE id = ?
	`

	info := &models.NutritionalInfo{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&info.ID, &info.TotalWeight,
		&info.Calories, &info.Protein, &info.Carbs, &info.Fat, &info.Fiber,
		&info.Sugar, &info.ImagePath, &info.CreatedAt, &info.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return info, nil
}

// SaveScan saves a nutrition scan to the database
func (s *SQLiteDB) SaveScan(ctx context.Context, scan *models.NutritionScan) error {
	query := `
		INSERT OR REPLACE INTO nutrition_scans (
			id, image_data, status, error, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	if scan.CreatedAt.IsZero() {
		scan.CreatedAt = now
	}
	scan.UpdatedAt = now

	_, err := s.db.ExecContext(ctx, query,
		scan.ID, scan.ImageData, scan.Status, scan.Error,
		scan.CreatedAt, scan.UpdatedAt,
	)
	return err
}

// UpdateScanStatus updates the status of a scan
func (s *SQLiteDB) UpdateScanStatus(ctx context.Context, id, status string, errMsg string) error {
	query := `
		UPDATE nutrition_scans
		SET status = ?, error = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query, status, errMsg, time.Now(), id)
	return err
}

// Close closes the database connection
func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

// GetRecentNutritionalInfo retrieves the most recent nutritional info entries
func (s *SQLiteDB) GetRecentNutritionalInfo(ctx context.Context, limit int) ([]*models.NutritionalInfo, error) {
	query := `
		SELECT id, total_weight, calories, protein, carbs, fat, fiber, sugar, image_path, created_at, updated_at
		FROM nutritional_info
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.NutritionalInfo
	for rows.Next() {
		var info models.NutritionalInfo
		var createdAt, updatedAt string

		err := rows.Scan(
			&info.ID, &info.TotalWeight, &info.Calories, &info.Protein,
			&info.Carbs, &info.Fat, &info.Fiber, &info.Sugar,
			&info.ImagePath, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse time strings
		info.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		info.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		results = append(results, &info)
	}

	return results, nil
}
