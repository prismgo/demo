// Package schedule registers the demo application's scheduled tasks.
package schedule

import (
	"context"

	"github.com/prismgo/framework/timer"
)

// Register adds the demo's scheduled tasks to s.
func Register(s *timer.Schedule) {
	s.Call(func(ctx context.Context) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		return nil
	}).
		DailyAt("02:00").
		Name("app:daily-maintenance").
		Description("Run the starter application's daily maintenance task.")
}
