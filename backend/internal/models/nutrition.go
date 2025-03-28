package models

import (
	"time"
)

// NutritionalInfo represents the nutritional information extracted from a label
type NutritionalInfo struct {
	ID          string  `json:"id"`
	TotalWeight float64 `json:"total_weight"` // in grams

	// Macronutrients (per 100g)
	Calories float64 `json:"calories"` // kcal
	Protein  float64 `json:"protein"`  // grams
	Carbs    float64 `json:"carbs"`    // grams
	Fat      float64 `json:"fat"`      // grams
	Fiber    float64 `json:"fiber"`    // grams
	Sugar    float64 `json:"sugar"`    // grams

	// Additional information
	ImagePath string    `json:"image_path"` // path to the stored image
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NutritionScan represents a scanning session
type NutritionScan struct {
	ID        string           `json:"id"`
	ImageData []byte           `json:"image_data"` // Base64 encoded image
	Status    string           `json:"status"`     // "pending", "processing", "completed", "failed"
	Result    *NutritionalInfo `json:"result,omitempty"`
	Error     string           `json:"error,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}
