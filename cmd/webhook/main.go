package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ShreyashSri/ChaosCI-Stats/internal/naas"
	"github.com/ShreyashSri/ChaosCI-Stats/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

var dbSecret = os.Getenv("WEBHOOK_SECRET")

type GitHubPayload struct {
	Action      string `json:"action"`
	PullRequest struct {
		Number int `json:"number"`
		Head   struct {
			Sha  string `json:"sha"`
			Repo struct {
				FullName string `json:"full_name"`
			} `json:"repo"`
		} `json:"head"`
	} `json:"pull_request"`
}

func naasError(w http.ResponseWriter, msg string, code int) {
	reason := naas.GetReason()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error": %q, "naas_reason": %q}`, msg, reason)
}

func main() {
	if dbSecret == "" {
		dbSecret = "dev-secret"
	}

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		dbUrl = "postgres://chaos:password@localhost:5432/chaosci?sslmode=disable"
	}

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		log.Fatalf("failed to connect to db: %v", err)
	}

	querier := store.New(db)

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			naasError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			naasError(w, "Error reading body", http.StatusInternalServerError)
			return
		}

		signature := r.Header.Get("X-Hub-Signature-256")
		if !verifyHMAC(body, signature) {
			naasError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var payload GitHubPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			naasError(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.Action != "opened" && payload.Action != "synchronize" {
			w.WriteHeader(http.StatusOK)
			return
		}

		username := strings.Split(payload.PullRequest.Head.Repo.FullName, "/")[0]
		shaShort := payload.PullRequest.Head.Sha
		if len(shaShort) > 7 {
			shaShort = shaShort[:7]
		}
		runID := fmt.Sprintf("%s-%s", username, shaShort)

		_, err = querier.CreateRun(context.Background(), store.CreateRunParams{
			ID:        runID,
			Repo:      payload.PullRequest.Head.Repo.FullName,
			PrNumber:  int32(payload.PullRequest.Number),
			CommitSha: payload.PullRequest.Head.Sha,
			Engine:    "chaosmesh",
			Status:    "pending",
		})

		if err != nil {
			log.Printf("Failed to create run: %v", err)
			naasError(w, "Failed to create run", http.StatusInternalServerError)
			return
		}

		log.Printf("Created run %s for PR %d", runID, payload.PullRequest.Number)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"run_id": "%s"}`, runID)
	})

	log.Println("Webhook server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func verifyHMAC(body []byte, signature string) bool {
	if signature == "" {
		// Allow local webhook testing without an ngrok/GitHub signature.
		if dbSecret == "dev-secret" {
			return true
		}
		return false
	}

	mac := hmac.New(sha256.New, []byte(dbSecret))
	mac.Write(body)
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	if len(signature) < 7 || signature[:7] != "sha256=" {
		return false
	}

	return hmac.Equal([]byte(signature[7:]), []byte(expectedMAC))
}
