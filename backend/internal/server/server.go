package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/franckalain/nutritionalvalue/internal/database"
	"github.com/franckalain/nutritionalvalue/internal/ml"
	"github.com/franckalain/nutritionalvalue/internal/models"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // In production, this should be more restrictive
	},
}

type Server struct {
	db            database.DB
	model         ml.Model
	clients       sync.Map
	tempImageData sync.Map // Temporary storage for image data
	debug         bool
}

func New(db database.DB, model ml.Model, debug bool) *Server {
	if debug {
		log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
		log.Println("Debug logging enabled")
	}
	return &Server{
		db:    db,
		model: model,
		debug: debug,
	}
}

func (s *Server) Start(port, staticDir string) error {
	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Setup HTTP routes
	http.HandleFunc("/ws", s.handleWebSocket)
	http.HandleFunc("/health", s.handleHealth)

	// Serve static files
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/", fs)

	// Start server
	go func() {
		log.Printf("Starting server on port %s\n", port)
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatal("ListenAndServe:", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down server...")
	return nil
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	// Store client connection
	clientID := uuid.New().String()
	s.clients.Store(clientID, conn)
	defer s.clients.Delete(clientID)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}

		// Parse message
		var msg map[string]any
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Error parsing message:", err)
			continue
		}

		s.handleWebSocketMessage(conn, msg)
	}
}

func (s *Server) handleWebSocketMessage(conn *websocket.Conn, message map[string]any) {
	messageType, ok := message["type"].(string)
	if !ok {
		s.sendError(conn, "Invalid message format")
		return
	}

	data, _ := message["data"].(map[string]any)

	switch messageType {
	case "scan":
		s.handleScan(conn, data)
	case "confirm_scan":
		s.handleConfirmScan(conn, data)
	case "get_history":
		s.handleGetHistory(conn)
	default:
		s.sendError(conn, "Unknown message type")
	}
}

func (s *Server) handleScan(conn *websocket.Conn, data map[string]any) {
	// Validate input data
	imageStr, ok := data["image"].(string)
	if !ok {
		s.sendError(conn, "Invalid image data")
		return
	}

	totalWeight, ok := data["totalWeight"].(float64)
	if !ok {
		s.sendError(conn, "Invalid weight value")
		return
	}

	// Decode base64 image
	imageData, err := base64.StdEncoding.DecodeString(imageStr)
	if err != nil {
		log.Printf("Error decoding image: %v", err)
		s.sendError(conn, "Invalid image format")
		return
	}

	// Process image
	nutritionInfo, err := s.model.ProcessImage(context.Background(), imageData)
	if err != nil {
		log.Printf("Error processing image: %v", err)
		s.sendError(conn, "Failed to process image")
		return
	}

	log.Printf("Successfully processed image! Nutritional values - Calories: %.1f, Protein: %.1fg, Carbs: %.1fg, Fat: %.1fg",
		nutritionInfo.Calories, nutritionInfo.Protein, nutritionInfo.Carbs, nutritionInfo.Fat)

	// Set the total weight from user input
	nutritionInfo.TotalWeight = totalWeight
	nutritionInfo.ID = uuid.New().String()
	nutritionInfo.CreatedAt = time.Now()
	nutritionInfo.UpdatedAt = time.Now()

	// Store the image data in the server's memory temporarily
	// We'll use a map with the nutrition info ID as the key
	s.tempImageData.Store(nutritionInfo.ID, imageData)

	// Send results back to client for confirmation
	s.sendMessage(conn, "scan_result", nutritionInfo)
}

func (s *Server) handleGetHistory(conn *websocket.Conn) {
	// Get recent nutritional info from database
	ctx := context.Background()
	nutritionInfos, err := s.db.GetRecentNutritionalInfo(ctx, 20) // Get last 20 entries
	if err != nil {
		log.Printf("Error retrieving history: %v", err)
		s.sendError(conn, "Failed to retrieve history")
		return
	}

	// Calculate session and weekly totals
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, startOfWeek.Location())

	var dayTotal, weekTotal struct {
		Calories float64
		Protein  float64
		Carbs    float64
		Fat      float64
	}

	// Calculate totals
	for _, info := range nutritionInfos {
		// Add to weekly total if within this week
		if info.CreatedAt.After(startOfWeek) {
			weekTotal.Calories += info.Calories
			weekTotal.Protein += info.Protein
			weekTotal.Carbs += info.Carbs
			weekTotal.Fat += info.Fat

			// Add to day total if within today
			if info.CreatedAt.After(startOfDay) {
				dayTotal.Calories += info.Calories
				dayTotal.Protein += info.Protein
				dayTotal.Carbs += info.Carbs
				dayTotal.Fat += info.Fat
			}
		}
	}

	// Prepare response
	response := map[string]interface{}{
		"items": nutritionInfos,
		"day_total": map[string]float64{
			"calories": dayTotal.Calories,
			"protein":  dayTotal.Protein,
			"carbs":    dayTotal.Carbs,
			"fat":      dayTotal.Fat,
		},
		"week_total": map[string]float64{
			"calories": weekTotal.Calories,
			"protein":  weekTotal.Protein,
			"carbs":    weekTotal.Carbs,
			"fat":      weekTotal.Fat,
		},
	}

	s.sendMessage(conn, "history", response)
}

func (s *Server) handleConfirmScan(conn *websocket.Conn, data map[string]any) {
	// Log the received data for debugging
	log.Printf("Received confirm_scan data: %+v", data)

	// Extract the nutrition info ID with better error handling
	var nutritionInfoID string
	if id, ok := data["id"]; ok {
		if strID, ok := id.(string); ok {
			nutritionInfoID = strID
		} else {
			log.Printf("ID is not a string: %v (type: %T)", id, id)
			s.sendError(conn, "Invalid nutrition info ID format")
			return
		}
	} else {
		log.Printf("No ID field in data: %+v", data)
		s.sendError(conn, "Missing nutrition info ID")
		return
	}

	log.Printf("Looking for image data with ID: %s", nutritionInfoID)

	// Retrieve the image data from temporary storage
	imageDataAny, ok := s.tempImageData.Load(nutritionInfoID)
	if !ok {
		// Dump all keys in tempImageData for debugging
		var keys []string
		s.tempImageData.Range(func(key, value interface{}) bool {
			keys = append(keys, fmt.Sprintf("%v", key))
			return true
		})
		log.Printf("Image data not found for ID: %s. Available keys: %v", nutritionInfoID, keys)
		s.sendError(conn, "Image data not found")
		return
	}

	// Type assertion with safety check
	imageData, ok := imageDataAny.([]byte)
	if !ok {
		log.Printf("Stored data is not []byte: %T", imageDataAny)
		s.sendError(conn, "Invalid stored image data")
		return
	}

	// Clean up the temporary storage
	s.tempImageData.Delete(nutritionInfoID)

	// Safely extract numeric values with defaults
	var totalWeight, calories, protein, carbs, fat, fiber, sugar float64

	if tw, ok := data["total_weight"].(float64); ok {
		totalWeight = tw
	} else {
		log.Printf("Invalid total_weight: %v", data["total_weight"])
		totalWeight = 100 // Default
	}

	if cal, ok := data["calories"].(float64); ok {
		calories = cal
	} else {
		log.Printf("Invalid calories: %v", data["calories"])
	}

	if p, ok := data["protein"].(float64); ok {
		protein = p
	} else {
		log.Printf("Invalid protein: %v", data["protein"])
	}

	if c, ok := data["carbs"].(float64); ok {
		carbs = c
	} else {
		log.Printf("Invalid carbs: %v", data["carbs"])
	}

	if f, ok := data["fat"].(float64); ok {
		fat = f
	} else {
		log.Printf("Invalid fat: %v", data["fat"])
	}

	if fb, ok := data["fiber"].(float64); ok {
		fiber = fb
	} else {
		log.Printf("Invalid fiber: %v", data["fiber"])
	}

	if s, ok := data["sugar"].(float64); ok {
		sugar = s
	} else {
		log.Printf("Invalid sugar: %v", data["sugar"])
	}

	// Convert the data back to NutritionalInfo
	nutritionInfo := &models.NutritionalInfo{
		ID:          nutritionInfoID,
		TotalWeight: totalWeight,
		Calories:    calories,
		Protein:     protein,
		Carbs:       carbs,
		Fat:         fat,
		Fiber:       fiber,
		Sugar:       sugar,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Save the nutritional info to the database
	if err := s.db.SaveNutritionalInfo(context.Background(), nutritionInfo); err != nil {
		log.Printf("Error saving nutritional info: %v", err)
		s.sendError(conn, "Failed to save results")
		return
	}

	// Create and save the scan record
	scan := &models.NutritionScan{
		ID:        uuid.New().String(),
		ImageData: imageData,
		Status:    "completed",
		Result:    nutritionInfo,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.db.SaveScan(context.Background(), scan); err != nil {
		log.Printf("Error saving scan: %v", err)
		s.sendError(conn, "Failed to save scan")
		return
	}

	log.Printf("Successfully saved nutritional info and scan")
	s.sendMessage(conn, "scan_saved", nil)
}

func (s *Server) sendMessage(conn *websocket.Conn, messageType string, data any) {
	msg := map[string]any{
		"type": messageType,
		"data": data,
	}

	log.Printf("Sending message to client - Type: %s, Data: %+v", messageType, data)
	if err := conn.WriteJSON(msg); err != nil {
		log.Println("Error sending message:", err)
	}
	log.Printf("Message sent successfully")
}

func (s *Server) sendError(conn *websocket.Conn, message string) {
	msg := map[string]any{
		"type":    "error",
		"message": message,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Println("Error sending error message:", err)
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
