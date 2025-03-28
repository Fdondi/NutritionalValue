package ml

import (
	"context"
	"flag"
	"fmt"

	"github.com/franckalain/nutritionalvalue/internal/models"
)

// Model represents a machine learning model that can process images
type Model interface {
	// Load initializes the model with its configuration
	Load(ctx context.Context) error
	// ProcessImage takes an image and returns nutritional information
	ProcessImage(ctx context.Context, imageData []byte) (*models.NutritionalInfo, error)
}

// ModelFactory creates a new model instance based on configuration
type ModelFactory interface {
	// CreateModel creates a new model instance
	CreateModel() (Model, error)
}

// NewModel creates a new model instance based on the model type
func NewModel(modelType string) (Model, error) {
	var factory ModelFactory
	var configPath string

	// Parse command line flags
	flag.StringVar(&configPath, "config-"+modelType, "", "path to "+modelType+" model configuration file")
	flag.Parse()

	switch modelType {
	case "google":
		config := GoogleConfig{
			BaseConfig: BaseConfig{
				ConfigPath: configPath,
			},
		}
		if err := config.Load(); err != nil {
			return nil, fmt.Errorf("failed to load Google config: %w", err)
		}
		factory = NewGoogleModelFactory(config)
	case "local":
		config := LocalConfig{
			BaseConfig: BaseConfig{
				ConfigPath: configPath,
			},
		}
		if err := config.Load(); err != nil {
			return nil, fmt.Errorf("failed to load local config: %w", err)
		}
		factory = NewLocalModelFactory(config)
	default:
		return nil, fmt.Errorf("unsupported model type: %s", modelType)
	}
	return factory.CreateModel()
}
