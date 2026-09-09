package jobs

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

const queueWorkers = 4

var Module = fx.Module("jobs",
	fx.Provide(NewClient),
)

// NewClient runs River's own migrations at boot. River owns its tables, so
// letting it manage them keeps them out of the application's migration
// history, where they would only ever be copied from upstream anyway.
func NewClient(lc fx.Lifecycle, pool *pgxpool.Pool, extractor Extractor, log *zap.Logger) (*river.Client[pgx.Tx], error) {
	driver := riverpgxv5.New(pool)

	migrator, err := rivermigrate.New(driver, nil)
	if err != nil {
		return nil, err
	}
	if _, err := migrator.Migrate(context.Background(), rivermigrate.DirectionUp, nil); err != nil {
		return nil, err
	}

	workers := river.NewWorkers()
	if err := river.AddWorkerSafely(workers, NewExtractWorker(extractor, log)); err != nil {
		return nil, err
	}

	client, err := river.NewClient(driver, &river.Config{
		Queues:  map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: queueWorkers}},
		Workers: workers,
	})
	if err != nil {
		return nil, err
	}

	lc.Append(fx.Hook{
		OnStart: client.Start,
		OnStop:  client.Stop,
	})

	return client, nil
}
