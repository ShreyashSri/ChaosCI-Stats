package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ShreyashSri/ChaosCI-Stats/internal/config"
	"github.com/ShreyashSri/ChaosCI-Stats/internal/engine"
	"github.com/ShreyashSri/ChaosCI-Stats/internal/engine/chaosmesh"
	"github.com/ShreyashSri/ChaosCI-Stats/internal/engine/litmus"
	"github.com/ShreyashSri/ChaosCI-Stats/internal/github"
	"github.com/ShreyashSri/ChaosCI-Stats/internal/queue"
	"github.com/ShreyashSri/ChaosCI-Stats/internal/store"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

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

	k8sClient, err := engine.NewDynamicClient()
	if err != nil {
		log.Fatalf("Failed to create k8s client: %v", err)
	}

	cmEngine := chaosmesh.NewAdapter(k8sClient)
	litmusEngine := litmus.NewAdapter(k8sClient)

	ghClient, err := github.NewClient()
	if err != nil {
		log.Printf("Warning: failed to init github client: %v", err)
	}

	runExperiment := func(ctx context.Context, activeEngine engine.ChaosEngine, run store.Run, expConfig config.Experiment, kind string) bool {
		exp, err := querier.CreateExperiment(ctx, store.CreateExperimentParams{
			RunID:  sql.NullString{String: run.ID, Valid: true},
			Name:   expConfig.Name,
			Kind:   kind,
			Type:   expConfig.Type,
			Status: "pending",
		})
		if err != nil {
			log.Printf("Failed to create experiment in DB: %v", err)
			return false
		}

		var yamlData []byte
		if ghClient != nil {
			yamlData, err = ghClient.GetFileContent(ctx, run.Repo, run.CommitSha, expConfig.File)
			if err != nil {
				log.Printf("Failed to fetch file %s: %v", expConfig.File, err)
				return false
			}
		}

		err = activeEngine.Apply(ctx, exp, yamlData)
		if err != nil {
			log.Printf("Failed to apply experiment: %v", err)
			return false
		}

		ch, err := activeEngine.Watch(ctx, exp, yamlData)
		if err != nil {
			log.Printf("Failed to watch experiment: %v", err)
			return false
		}

		success := true
		for res := range ch {
			log.Printf("Run %s | Exp %d status: %s", run.ID, exp.ID, res.Status)

			if res.Status != "success" && res.Status != "pending" && res.Status != "running" {
				success = false
			}

			now := sql.NullTime{Time: time.Now(), Valid: true}
			_, _ = querier.UpdateExperimentStatus(ctx, store.UpdateExperimentStatusParams{
				ID:         exp.ID,
				Status:     res.Status,
				StartedAt:  now,
				FinishedAt: now,
			})

			_, _ = querier.CreateEvent(ctx, store.CreateEventParams{
				RunID:        sql.NullString{String: run.ID, Valid: true},
				ExperimentID: sql.NullInt64{Int64: exp.ID, Valid: true},
				Level:        sql.NullString{String: "info", Valid: true},
				Message:      sql.NullString{String: res.Message, Valid: true},
			})
		}

		_ = activeEngine.Cleanup(ctx, exp, yamlData)
		return success
	}

	handler := func(ctx context.Context, runID string) error {
		log.Printf("Worker processing run %s", runID)

		run, err := querier.GetRun(ctx, runID)
		if err != nil {
			log.Printf("Failed to fetch run %s: %v", runID, err)
			return err
		}

		var activeEngine engine.ChaosEngine
		switch run.Engine {
		case "chaosmesh":
			activeEngine = cmEngine
		case "litmus":
			activeEngine = litmusEngine
		default:
			return fmt.Errorf("unknown engine type: %s", run.Engine)
		}

		var configYAML []byte
		if ghClient != nil {
			configYAML, err = ghClient.GetFileContent(ctx, run.Repo, run.CommitSha, "chaos.yaml")
			if err != nil {
				log.Printf("Failed to fetch chaos.yaml: %v", err)
				if run.CheckID.Valid {
					_ = ghClient.UpdateCheckRun(ctx, run.Repo, run.CheckID.Int64, "completed", "failure", "Missing chaos.yaml")
				}
				return err
			}
		}

		cfg, err := config.ParseConfig(configYAML)
		if err != nil {
			log.Printf("Failed to parse config: %v", err)
			if ghClient != nil && run.CheckID.Valid {
				_ = ghClient.UpdateCheckRun(ctx, run.Repo, run.CheckID.Int64, "completed", "failure", "Invalid chaos.yaml")
			}
			return err
		}

		hasEssentialFailure := false
		for _, e := range cfg.Essential {
			if !runExperiment(ctx, activeEngine, run, e, "essential") {
				hasEssentialFailure = true
			}
		}

		conclusion := "success"
		if hasEssentialFailure {
			conclusion = "action_required"
		}

		if ghClient != nil && run.CheckID.Valid {
			_ = ghClient.UpdateCheckRun(ctx, run.Repo, run.CheckID.Int64, "completed", conclusion, "Chaos tests completed.")
		}

		_, err = querier.UpdateRunStatus(ctx, store.UpdateRunStatusParams{
			ID:         runID,
			Status:     conclusion,
			FinishedAt: sql.NullTime{Time: time.Now(), Valid: true},
		})

		go func() {
			bgCtx := context.Background()
			for _, e := range cfg.Extended {
				runExperiment(bgCtx, activeEngine, run, e, "extended")
			}
		}()

		log.Printf("Run %s completed essential tests with conclusion %s", runID, conclusion)
		return nil
	}

	pool := queue.NewWorkerPool(querier, 3, handler)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool.Start(ctx)
	log.Println("Worker started. Listening for jobs...")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down worker...")
	cancel()
	time.Sleep(1 * time.Second)
}
