package timerdemo

import (
	"fmt"
	"time"
)

// everyScenario proves that a fixed interval task runs immediately and then repeats.
func everyScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("health-check").Every(20 * time.Millisecond)
	counts := recorder.run(120 * time.Millisecond)
	runs := counts["health-check"]
	return fmt.Sprintf("immediate=%t repeated=%t", runs >= 1, runs >= 2), nil
}

// secondFrequenciesScenario registers every second-level helper and proves one fires.
func secondFrequenciesScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("every-second").EverySecond()
	recorder.task("every-two-seconds").EveryTwoSeconds()
	recorder.task("every-five-seconds").EveryFiveSeconds()
	recorder.task("every-ten-seconds").EveryTenSeconds()
	recorder.task("every-fifteen-seconds").EveryFifteenSeconds()
	recorder.task("every-twenty-seconds").EveryTwentySeconds()
	recorder.task("every-thirty-seconds").EveryThirtySeconds()
	counts := recorder.run(1050 * time.Millisecond)
	return fmt.Sprintf("methods=7 every-second-fired=%t", counts["every-second"] > 0), nil
}

// minuteFrequenciesScenario separates the minute helpers that run on start from those that wait.
func minuteFrequenciesScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("every-minute").EveryMinute()
	recorder.task("every-two-minutes").EveryTwoMinutes()
	recorder.task("every-three-minutes").EveryThreeMinutes()
	recorder.task("every-four-minutes").EveryFourMinutes()
	recorder.task("every-five-minutes").EveryFiveMinutes()
	recorder.task("every-ten-minutes").EveryTenMinutes()
	recorder.task("every-fifteen-minutes").EveryFifteenMinutes()
	recorder.task("every-thirty-minutes").EveryThirtyMinutes()
	counts := recorder.run(80 * time.Millisecond)

	immediate := 0
	for _, name := range []string{
		"every-minute", "every-two-minutes", "every-five-minutes", "every-ten-minutes",
		"every-fifteen-minutes", "every-thirty-minutes",
	} {
		if counts[name] > 0 {
			immediate++
		}
	}
	deferred := 0
	for _, name := range []string{"every-three-minutes", "every-four-minutes"} {
		if counts[name] == 0 {
			deferred++
		}
	}
	return fmt.Sprintf("immediate=%d deferred=%d", immediate, deferred), nil
}

// hourlyScenario proves the top-of-hour helper runs on start.
func hourlyScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("stats-archive").Hourly()
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("immediate=%t", anyFired(counts)), nil
}

// hourlyAtScenario registers minute offsets and proves they wait for the next hit.
func hourlyAtScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("int-offset").HourlyAt(15)
	recorder.task("string-offset").HourlyAt("0,15,30,45")
	recorder.task("slice-offset").HourlyAt([]int{5, 35})
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("offsets=3 immediate=%t", anyFired(counts)), nil
}

// hourStepsScenario registers every stepped-hour helper and proves they wait.
func hourStepsScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("odd-hour").EveryOddHour(20)
	recorder.task("two-hours").EveryTwoHours(10)
	recorder.task("three-hours").EveryThreeHours(30)
	recorder.task("four-hours").EveryFourHours(45)
	recorder.task("six-hours").EverySixHours(5)
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("methods=5 immediate=%t", anyFired(counts)), nil
}

// dailyScenario proves the midnight helper runs on start.
func dailyScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("billing-close").Daily()
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("immediate=%t", anyFired(counts)), nil
}

// dailyAtScenario registers a daily time and proves it waits for that hit time.
func dailyAtScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("billing-close").DailyAt("18:30")
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("at=18:30 immediate=%t", anyFired(counts)), nil
}

// twiceDailyScenario registers two daily hours and proves it waits.
func twiceDailyScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("summary-send").TwiceDaily(9, 18)
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("hours=2 immediate=%t", anyFired(counts)), nil
}

// twiceDailyAtScenario registers two hours with a minute offset and proves it waits.
func twiceDailyAtScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("summary-send").TwiceDailyAt(9, 18, 15)
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("hours=2 offset=15 immediate=%t", anyFired(counts)), nil
}

// atScenario composes a weekday constraint with a hit time and proves it waits.
func atScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("billing-close").Weekdays().At("18:30")
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("at=18:30 weekdays=5 immediate=%t", anyFired(counts)), nil
}

// weeklyScenario proves the default weekly helper waits for Sunday midnight.
func weeklyScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("weekly-close").Weekly()
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("weekday=sunday immediate=%t", anyFired(counts)), nil
}

// weeklyOnScenario registers weekly day and time constraints and proves they wait.
func weeklyOnScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("monday-report").WeeklyOn(time.Monday, "09:30")
	recorder.task("three-day-report").WeeklyOn("mon,wed,fri", "10:00")
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("weekday=monday at=09:30 immediate=%t", anyFired(counts)), nil
}

// weekdayGroupsScenario registers the weekday and weekend groups and proves they wait.
func weekdayGroupsScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("weekdays").Weekdays()
	recorder.task("weekends").Weekends()
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("weekdays=5 weekends=2 immediate=%t", anyFired(counts)), nil
}

// namedWeekdaysScenario registers every named weekday helper and proves they wait.
func namedWeekdaysScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("monday").Mondays()
	recorder.task("tuesday").Tuesdays()
	recorder.task("wednesday").Wednesdays()
	recorder.task("thursday").Thursdays()
	recorder.task("friday").Fridays()
	recorder.task("saturday").Saturdays()
	recorder.task("sunday").Sundays()
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("named=7 immediate=%t", anyFired(counts)), nil
}

// daysScenario registers custom weekday sets and proves they wait.
func daysScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("int-days").Days([]int{1, 3, 5}).At("09:00")
	recorder.task("string-days").Days("mon-wed,fri").At("09:00")
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("ints=3 strings=3 immediate=%t", anyFired(counts)), nil
}

// monthlyScenario proves the first-of-month helper waits for its hit day.
func monthlyScenario() (string, error) {
	recorder := newRecorder()
	recorder.task("statement-close").Monthly()
	counts := recorder.run(80 * time.Millisecond)
	return fmt.Sprintf("day=1 immediate=%t", anyFired(counts)), nil
}
