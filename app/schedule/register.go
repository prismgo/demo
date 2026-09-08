package schedule

import (
	"context"

	"github.com/prismgo/framework/timer"
)

// Register declares application scheduled tasks.
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
