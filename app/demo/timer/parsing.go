package timerdemo

import "fmt"

// timeParsingScenario accepts HH:MM and HH:MM:SS and rejects invalid clocks.
func timeParsingScenario() (string, error) {
	hhmm := expectOK(
		func() { newTask().At("18:30") },
		func() { newTask().At("0:0") },
	)
	hhmmss := expectOK(func() { newTask().At("18:30:15") })
	invalid := expectPanic(
		func() { newTask().At("25:00") },
		func() { newTask().At("12:99") },
		func() { newTask().At("bad") },
	)
	return fmt.Sprintf("hhmm=%s hhmmss=%s invalid=%s", hhmm, hhmmss, invalid), nil
}

// offsetParsingScenario accepts every documented minute offset form.
func offsetParsingScenario() (string, error) {
	forms := expectOK(
		func() { newTask().HourlyAt(15) },
		func() { newTask().HourlyAt("0,15,30,45") },
		func() { newTask().HourlyAt("10-20") },
		func() { newTask().HourlyAt([]int{5, 35}) },
		func() { newTask().HourlyAt([]string{"0", "30"}) },
	)
	invalid := expectPanic(func() { newTask().HourlyAt("60") })
	return fmt.Sprintf("forms=%s invalid=%s", forms, invalid), nil
}

// weekdayParsingScenario accepts numeric, abbreviated, and ranged weekday forms.
func weekdayParsingScenario() (string, error) {
	aliases := []string{
		"sun", "sunday", "mon", "monday", "tue", "tues", "tuesday",
		"wed", "wednesday", "thu", "thur", "thurs", "thursday",
		"fri", "friday", "sat", "saturday",
	}
	aliasProbes := make([]func(), 0, len(aliases))
	for _, alias := range aliases {
		alias := alias
		aliasProbes = append(aliasProbes, func() { newTask().Days(alias) })
	}
	ranges := expectOK(
		func() { newTask().Days("mon-wed,fri") },
		func() { newTask().WeeklyOn("mon,wed,fri", "10:00") },
	)
	slices := expectOK(
		func() { newTask().Days([]int{1, 3, 5}) },
		func() { newTask().WeeklyOn([]int{1, 3, 5}, "10:00") },
	)
	invalid := expectPanic(
		func() { newTask().Days("noday") },
		func() { newTask().Days(7) },
	)
	return fmt.Sprintf("aliases=%s ranges=%s slices=%s invalid=%s",
		expectOK(aliasProbes...), ranges, slices, invalid), nil
}
