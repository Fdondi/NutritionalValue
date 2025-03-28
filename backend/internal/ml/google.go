package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"cloud.google.com/go/vertexai/genai"
	"github.com/franckalain/nutritionalvalue/internal/models"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

// GoogleConfig holds configuration for the Google model
type GoogleConfig struct {
	BaseConfig
	ProjectID       string `json:"project_id"`
	Location        string `json:"location"`
	CredentialsFile string `json:"credentials_file"`
}

// Load loads the Google configuration
func (c *GoogleConfig) Load() error {
	if err := c.LoadConfig(c.ConfigPath, "google", c); err != nil {
		return err
	}

	// Fall back to environment variables if not set
	if c.ProjectID == "" {
		c.ProjectID = os.Getenv("GOOGLE_PROJECT_ID")
	}
	if c.Location == "" {
		c.Location = os.Getenv("GOOGLE_LOCATION")
	}
	if c.CredentialsFile == "" {
		c.CredentialsFile = os.Getenv("GOOGLE_CREDENTIALS_FILE")
	}

	return nil
}

// GoogleModel implements the Model interface for Google's Vertex AI
type GoogleModel struct {
	config GoogleConfig
	client *genai.Client
	model  *genai.GenerativeModel
}

// GoogleModelFactory implements ModelFactory for Google models
type GoogleModelFactory struct {
	config GoogleConfig
}

// NewGoogleModelFactory creates a new Google model factory
func NewGoogleModelFactory(config GoogleConfig) *GoogleModelFactory {
	return &GoogleModelFactory{config: config}
}

// CreateModel creates a new Google model instance
func (f *GoogleModelFactory) CreateModel() (Model, error) {
	return &GoogleModel{
		config: f.config,
	}, nil
}

// Load initializes the Google model
func (m *GoogleModel) Load(ctx context.Context) error {
	opts := []option.ClientOption{}

	if m.config.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(m.config.CredentialsFile))
	}

	client, err := genai.NewClient(ctx, m.config.ProjectID, m.config.Location, opts...)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	m.client = client
	m.model = client.GenerativeModel("gemini-pro-vision")
	return nil
}

// ProcessImage processes an image using Google's Vertex AI
func (m *GoogleModel) ProcessImage(ctx context.Context, imageData []byte) (*models.NutritionalInfo, error) {
	if m.model == nil {
		return nil, fmt.Errorf("model not loaded")
	}

	// Create a prompt for the model
	prompt := `Analyze this nutritional label image and extract the values per 100g in a structured format:
- Calories
- Protein
- Carbohydrates
- Fat
- Fiber
- Sugar

Format the response as a JSON object with exactly one of "error" or "success" populated. 
Not all values can be zero. If most values are zero, raise an error explaining what went wrong.
{
	"error": {
		"error_reason": "string",
		"suggestion_for_better_results": "string"
	},
	"success": {
		"calories": number,
		"protein": number,
		"carbs": number,
		"fat": number,
		"fiber": number,
		"sugar": number,
	}
}`
	// Create the image part for the model
	img := genai.ImageData("image/jpeg", imageData)

	// Parse the content
	fmt.Println("Calling the model")
	resp, err := m.model.GenerateContent(ctx, genai.Text(prompt), img)
	if err != nil {
		return nil, fmt.Errorf("failed to call ai: %w", err)
	}

	// Parse the response
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("no response generated")
	}

	// Extract the JSON response
	candidate := resp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		return nil, fmt.Errorf("no content in response")
	}

	// Get the text content and parse it as JSON
	textContent := fmt.Sprintf("%v", candidate.Content.Parts[0])

	// Response should be multi-linee string
	// First line should be ```json; assert that it is then discard it
	textContent = strings.TrimPrefix(textContent, " ```json\n")
	textContent = strings.TrimSuffix(textContent, "\n```")

	// Parse the JSON response
	var output struct {
		Error struct {
			ErrorReason string `json:"error_reason"`
			Suggestion  string `json:"suggestion_for_better_results"`
		} `json:"error"`
		Success struct {
			Calories float64 `json:"calories"`
			Protein  float64 `json:"protein"`
			Carbs    float64 `json:"carbs"`
			Fat      float64 `json:"fat"`
			Fiber    float64 `json:"fiber"`
			Sugar    float64 `json:"sugar"`
		} `json:"success"`
	}

	// First unmarshal into a map to check for missing fields
	var rawMap map[string]interface{}
	if err := json.Unmarshal([]byte(textContent), &rawMap); err != nil {
		return nil, fmt.Errorf("failed to parse model response: %w while parsing %s", err, textContent)
	}

	// Check if success object exists and has all required fields
	successObj, ok := rawMap["success"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("missing or invalid success object in response")
	}

	requiredFields := []string{"calories", "protein", "carbs", "fat"}
	for _, field := range requiredFields {
		if _, exists := successObj[field]; !exists {
			return nil, fmt.Errorf("missing required field '%s' in response", field)
		}
	}

	// Now unmarshal into our struct
	if err := json.Unmarshal([]byte(textContent), &output); err != nil {
		return nil, fmt.Errorf("failed to parse model response: %w", err)
	}

	if output.Error.ErrorReason != "" {
		return nil, fmt.Errorf("error: %s; suggestion: %s", output.Error.ErrorReason, output.Error.Suggestion)
	}

	// Create and return the nutritional info
	return &models.NutritionalInfo{
		ID:       uuid.New().String(),
		Calories: output.Success.Calories,
		Protein:  output.Success.Protein,
		Carbs:    output.Success.Carbs,
		Fat:      output.Success.Fat,
		Fiber:    output.Success.Fiber,
		Sugar:    output.Success.Sugar,
	}, nil
}
