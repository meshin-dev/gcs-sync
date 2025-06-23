package watcher

import (
	"context"
	"gcs_sync/internal/config"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"sync"
)

// StartAll initializes and manages watchers for all enabled synchronization rules.
// It sets up watchers to start when the application begins and ensures they stop
// gracefully when the application shuts down.
//
// Parameters:
//   - lc: An fx.Lifecycle instance used to register start and stop hooks for the watchers.
//   - cfg: A pointer to the config.Config struct containing synchronization rules and settings.
//   - log: A pointer to a logrus.Logger for logging errors and other information.
//
// This function doesn't return any value, but it sets up the necessary hooks for
// starting and stopping the watchers as part of the application's lifecycle.
func StartAll(lc fx.Lifecycle, cfg *config.Config, log *logrus.Logger) {
	stop := make(chan struct{})
	var wg sync.WaitGroup

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			for _, r := range cfg.Sync {
				if !r.Enabled {
					continue
				}
				runner, err := newRuleRunner(r)
				if err != nil {
					return err
				}
				wg.Add(1)
				go func(rr *ruleRunner) {
					defer wg.Done()
					if err := rr.run(stop); err != nil {
						log.WithError(err).Error("watcher stopped with error")
					}
				}(runner)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			close(stop)
			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-done:
				return nil
			}
		},
	})
}
