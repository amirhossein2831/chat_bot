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

// Two clients rather than one: the worker client needs the extractor, and the
// extractor needs a queue to enqueue with. Sharing one client would make that
// a cycle. River supports an insert-only client for exactly this.
type InsertClient struct{ *river.Client[pgx.Tx] }

type WorkerClient struct{ *river.Client[pgx.Tx] }

// migrated is a token that orders migrations before either client is built.
type migrated struct{}

var Module = fx.Module("jobs",
	fx.Provide(
		runMigrations,
		NewInsertClient,
		NewWorkerClient,
		NewQueue,
	),
)

// runMigrations lets River own its own tables. They would otherwise be copied
// from upstream into our migration history and drift on every upgrade.
func runMigrations(pool *pgxpool.Pool) (migrated, error) {
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		return migrated{}, err
	}
	_, err = migrator.Migrate(context.Background(), rivermigrate.DirectionUp, nil)
	return migrated{}, err
}

func NewInsertClient(pool *pgxpool.Pool, _ migrated) (InsertClient, error) {
	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
	if err != nil {
		return InsertClient{}, err
	}
	return InsertClient{client}, nil
}

func NewWorkerClient(
	lc fx.Lifecycle,
	pool *pgxpool.Pool,
	extractor Extractor,
	log *zap.Logger,
	_ migrated,
) (WorkerClient, error) {
	workers := river.NewWorkers()
	if err := river.AddWorkerSafely(workers, NewExtractWorker(extractor, log)); err != nil {
		return WorkerClient{}, err
	}

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues:  map[string]river.QueueConfig{river.QueueDefault: {MaxWorkers: queueWorkers}},
		Workers: workers,
	})
	if err != nil {
		return WorkerClient{}, err
	}

	// River keeps whatever context Start is given for the notifier's whole
	// life. Handing it fx's start context means the notifier dies when that
	// context expires seconds after boot, then retries in a hot loop forever.
	runCtx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error { return client.Start(runCtx) },
		OnStop: func(ctx context.Context) error {
			cancel()
			return client.Stop(ctx)
		},
	})

	return WorkerClient{client}, nil
}
