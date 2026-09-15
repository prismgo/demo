package timerdemo

import (
	"fmt"
	"time"
)

// monthlyOnScenario registers a monthly day constraint and proves it waits for
// its hit day while rejecting out-of-range days at registration time.
func monthlyOnScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("statement-close").MonthlyOn(15, "08:30")
	counts := recorder.run(80 * time.Millisecond)
	invalid := expectPanic(
		func() { newTask().MonthlyOn(0, "08:30") },
		func() { newTask().MonthlyOn(32, "08:30") },
	)
	return fmt.Sprintf("day=15 at=08:30 immediate=%t invalid=%s", counts["statement-close"] > 0, invalid), nil
}

// twiceMonthlyScenario registers two monthly days and proves it waits for them.
func twiceMonthlyScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("invoice-remind").TwiceMonthly(1, 16, "10:00")
	counts := recorder.run(80 * time.Millisecond)
	invalid := expectPanic(
		func() { newTask().TwiceMonthly(0, 16) },
		func() { newTask().TwiceMonthly(1, 32) },
	)
	return fmt.Sprintf("days=1,16 at=10:00 immediate=%t invalid=%s", counts["invoice-remind"] > 0, invalid), nil
}

// lastDayOfMonthScenario registers a month-end task and proves it waits.
func lastDayOfMonthScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("statement-close").LastDayOfMonth("23:55")
	counts := recorder.run(80 * time.Millisecond)
	invalid := expectPanic(func() { newTask().LastDayOfMonth("25:00") })
	return fmt.Sprintf("at=23:55 immediate=%t invalid=%s", counts["statement-close"] > 0, invalid), nil
}

// daysOfMonthScenario registers multiple monthly days with a hit time and proves it waits.
func daysOfMonthScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("invoice-remind").DaysOfMonth(1, 15, 28).At("09:00")
	counts := recorder.run(80 * time.Millisecond)
	invalid := expectPanic(
		func() { newTask().DaysOfMonth() },
		func() { newTask().DaysOfMonth(0) },
	)
	return fmt.Sprintf("days=3 at=09:00 immediate=%t invalid=%s", counts["invoice-remind"] > 0, invalid), nil
}

// quarterlyScenario proves the first-day-of-quarter helper waits for its hit day.
func quarterlyScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("finance-close").Quarterly()
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("day=1 immediate=%t", counts["finance-close"] > 0), nil
}

// quarterlyOnScenario registers a quarter-relative day and proves it waits.
func quarterlyOnScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("finance-close").QuarterlyOn(45, "10:00")
	counts := recorder.run(80 * time.Millisecond)
	invalid := expectPanic(func() { newTask().QuarterlyOn(93) })
	return fmt.Sprintf("day=45 at=10:00 immediate=%t invalid=%s", counts["finance-close"] > 0, invalid), nil
}

// yearlyScenario proves the January-first helper waits for its hit day.
func yearlyScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("stats-reset").Yearly()
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("month=1 day=1 immediate=%t", counts["stats-reset"] > 0), nil
}

// yearlyOnScenario registers explicit and last-day yearly constraints and proves they wait.
func yearlyOnScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("stats-reset").YearlyOn(12, 31, "23:59")
	recorder.task("stats-last-day").YearlyOn(12, "last", "23:59")
	counts := recorder.run(80 * time.Millisecond)
	invalid := expectPanic(
		func() { newTask().YearlyOn(13, 1) },
		func() { newTask().YearlyOn(12, 32) },
	)
	lastToken := expectOK(func() { newTask().YearlyOn(12, "last", "23:59") })
	return fmt.Sprintf("month=12 day=31 at=23:59 last-token=%s immediate=%t invalid=%s",
		lastToken, anyFired(counts), invalid), nil
}
