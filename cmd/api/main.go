package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ShreyashSri/ChaosCI-Stats/internal/naas"
	"github.com/ShreyashSri/ChaosCI-Stats/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

func naasError(w http.ResponseWriter, msg string, code int) {
	reason := naas.GetReason()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error": %q, "naas_reason": %q}`, msg, reason)
}

func main() {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "postgres://chaos:password@localhost:5432/chaosci?sslmode=disable"
	}

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	querier := store.New(db)

	http.HandleFunc("/api/stats", func(w http.ResponseWriter, r *http.Request) {
		runs, err := querier.ListAllRuns(context.Background())
		if err != nil {
			naasError(w, "Failed to get stats", http.StatusInternalServerError)
			return
		}

		uniqueUsers := make(map[string]bool)
		var totalMinutes float64
		successCount := 0

		for _, run := range runs {
			parts := strings.Split(run.ID, "-")
			if len(parts) >= 2 {
				uniqueUsers[parts[0]] = true
			} else {
				parts = strings.Split(run.Repo, "/")
				if len(parts) > 0 {
					uniqueUsers[parts[0]] = true
				}
			}

			if run.CreatedAt.Valid {
				endTime := time.Now()
				if run.FinishedAt.Valid {
					endTime = run.FinishedAt.Time
				}
				totalMinutes += endTime.Sub(run.CreatedAt.Time).Minutes()
			}

			if run.Status == "success" || run.Status == "completed" {
				successCount++
			}
		}

		successRate := 0
		if len(runs) > 0 {
			successRate = int(float64(successCount) / float64(len(runs)) * 100)
		}

		stats := map[string]interface{}{
			"total_runs":   len(runs),
			"unique_users": len(uniqueUsers),
			"ci_minutes":   int(totalMinutes),
			"success_rate": successRate,
			"runs":         runs,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	http.HandleFunc("/api/runs", func(w http.ResponseWriter, r *http.Request) {
		runs, err := querier.ListAllRuns(context.Background())
		if err != nil {
			naasError(w, "Failed to fetch runs", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(runs)
	})

	http.HandleFunc("/api/runs/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/api/runs/"):]
		if path == "" {
			naasError(w, "Missing run ID", http.StatusBadRequest)
			return
		}

		if strings.HasSuffix(path, "/events") {
			runID := strings.TrimSuffix(path, "/events")
			handleSSE(w, r, querier, runID)
			return
		}

		runID := path
		run, err := querier.GetRun(context.Background(), runID)
		if err != nil {
			if err == sql.ErrNoRows {
				naasError(w, "Not found", http.StatusNotFound)
			} else {
				naasError(w, "Internal server error", http.StatusInternalServerError)
			}
			return
		}

		experiments, err := querier.GetExperimentsForRun(context.Background(), sql.NullString{String: runID, Valid: true})
		if err != nil {
			log.Printf("Failed to fetch experiments: %v", err)
		}

		events, err := querier.GetEventsForRun(context.Background(), sql.NullString{String: runID, Valid: true})
		if err != nil {
			log.Printf("Failed to fetch events: %v", err)
		}

		response := map[string]interface{}{
			"run":         run,
			"experiments": experiments,
			"events":      events,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	log.Println("API server listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func handleSSE(w http.ResponseWriter, r *http.Request, querier *store.Queries, runID string) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	lastEventID := int64(0)
	ctx := r.Context()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	experiments, _ := querier.GetExperimentsForRun(ctx, sql.NullString{String: runID, Valid: true})
	expMap := make(map[int64]string)
	for _, exp := range experiments {
		expMap[exp.ID] = exp.Name
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			events, err := querier.GetEventsForRun(ctx, sql.NullString{String: runID, Valid: true})
			if err == nil {
				for _, event := range events {
					if event.ID > lastEventID {
						expName := "Unknown"
						if event.ExperimentID.Valid {
							if name, ok := expMap[event.ExperimentID.Int64]; ok {
								expName = name
							}
						}

						uiEvent := map[string]interface{}{
							"id":              fmt.Sprintf("%d", event.ID),
							"run_id":          event.RunID.String,
							"experiment_name": expName,
							"status":          event.Level.String,
							"message":         event.Message.String,
							"timestamp":       event.Ts.Time.Format(time.RFC3339),
						}

						data, _ := json.Marshal(uiEvent)
						fmt.Fprintf(w, "data: %s\n\n", data)
						lastEventID = event.ID
						flusher.Flush()
					}
				}
			}
		}
	}
}
