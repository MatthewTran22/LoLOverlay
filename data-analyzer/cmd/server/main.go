package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"data-analyzer/internal/db"

	"github.com/joho/godotenv"
)

var database *db.DB

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	ctx := context.Background()

	// Connect to database
	var err error
	database, err = db.New(ctx)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// API routes
	http.HandleFunc("/api/stats", handleStats)
	http.HandleFunc("/api/matches", handleMatches)
	http.HandleFunc("/api/match/", handleMatchDetail)
	http.HandleFunc("/api/champions", handleChampions)

	// Static files
	http.Handle("/", http.FileServer(http.Dir("web")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on http://localhost:%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	matchCount, err := database.GetMatchCount(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	participantCount, err := database.GetParticipantCount(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"matches":      matchCount,
		"participants": participantCount,
	})
}

func handleMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	matches, err := database.GetRecentMatches(ctx, 50)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(matches)
}

func handleChampions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := database.GetChampionStats(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func handleMatchDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract match ID from URL path: /api/match/{matchId}
	matchID := strings.TrimPrefix(r.URL.Path, "/api/match/")
	if matchID == "" {
		http.Error(w, "Match ID required", http.StatusBadRequest)
		return
	}

	match, err := database.GetMatchDetail(ctx, matchID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(match)
}
