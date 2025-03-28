package ml

import (
	"context"
	"fmt"
	"os"

	"github.com/franckalain/nutritionalvalue/internal/models"
)

// LocalConfig holds configuration for the local model
type LocalConfig struct {
	BaseConfig
	ModelPath    string `json:"model_path"`
	GPUEnabled   bool   `json:"gpu_enabled"`
	GPUDeviceID  int    `json:"gpu_device_id"`
	MaxBatchSize int    `json:"max_batch_size"`
}

// Load loads the local configuration
func (c *LocalConfig) Load() error {
	if err := c.LoadConfig(c.ConfigPath, "local", c); err != nil {
		return err
	}

	// Fall back to environment variables if not set
	if c.ModelPath == "" {
		c.ModelPath = os.Getenv("LOCAL_MODEL_PATH")
	}
	if c.GPUEnabled == false {
		c.GPUEnabled = os.Getenv("LOCAL_GPU_ENABLED") == "true"
	}
	// TODO: Add environment variables for GPUDeviceID and MaxBatchSize

	return nil
}

// LocalModel implements the Model interface for local ML models
type LocalModel struct {
	config LocalConfig
}

// LocalModelFactory implements ModelFactory for local models
type LocalModelFactory struct {
	config LocalConfig
}

// NewLocalModelFactory creates a new local model factory
func NewLocalModelFactory(config LocalConfig) *LocalModelFactory {
	return &LocalModelFactory{config: config}
}

// CreateModel creates a new local model instance
func (f *LocalModelFactory) CreateModel() (Model, error) {
	return &LocalModel{
		config: f.config,
	}, nil
}

// Load initializes the local model
func (m *LocalModel) Load(ctx context.Context) error {
	// TODO: Implement actual model loading
	return nil
}

// ProcessImage processes an image using the local model
func (m *LocalModel) ProcessImage(ctx context.Context, imageData []byte) (*models.NutritionalInfo, error) {
	// TODO: Implement actual image processing
	return nil, fmt.Errorf("unimplemented: local model processing not yet implemented")
}
