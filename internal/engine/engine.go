package engine

import (
	"context"

	"github.com/ShreyashSri/ChaosCI-Stats/internal/store"
)

type Result struct {
	ExperimentID int64
	Status       string
	Message      string
}

type ChaosEngine interface {
	Apply(ctx context.Context, exp store.Experiment) error
	Watch(ctx context.Context, exp store.Experiment) (<-chan Result, error)
	Cleanup(ctx context.Context, exp store.Experiment) error
}
