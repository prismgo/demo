package timerdemo

import (
	"fmt"
	"time"

	"github.com/prismgo/framework/config"
	"github.com/prismgo/framework/container"
)

// timezoneScenario shows that calendar scheduling reads the process local time zone.
func timezoneScenario() (string, error) {
	previous := time.Local
	time.Local = time.FixedZone("demo-zone", 8*60*60)
	defer func() { time.Local = previous }()

	offset := time.Now().In(time.Local).Format("-07:00")
	return fmt.Sprintf("location=%s offset=%s", time.Local.String(), offset), nil
}

// debugLoggingScenario reports the app.debug flag the scheduler reads for success logs.
func debugLoggingScenario() (string, error) {
	return fmt.Sprintf("app.debug=%t", config.GetBool("app.debug", false)), nil
}

// overlapCacheScenario reports the default cache store backing WithoutOverlapping locks.
func overlapCacheScenario() (string, error) {
	return fmt.Sprintf("cache.default=%s", config.GetString("cache.default", "")), nil
}

// exceptionConfigurationScenario reports whether the exception reporter dependency is bound.
func exceptionConfigurationScenario() (string, error) {
	state := "missing"
	if _, err := container.Make[any](exceptionHandlerKey); err == nil {
		state = "bound"
	}
	return "reporter=" + state, nil
}
