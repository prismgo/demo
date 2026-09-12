package eventdemo_test

import (
	"context"
	"strings"
	"testing"

	eventdemo "prismgo-demo/app/demo/event"
	demotest "prismgo-demo/app/demo/testing"
)

func TestEventDemoSecondBatch(t *testing.T) {
	cases := []struct{ name, want string }{
		{"async-durability", "async-goroutine=true; queued-worker-path=true; durability-depends-on-driver"},
		{"app-lifecycle", "app.booting,app.booted,app.terminating,app.terminated"},
		{"provider-lifecycle", "app.provider.registering,app.provider.registered,app.provider.booting,app.provider.booted"},
		{"server-lifecycle", "server.starting,server.started,server.stopping,server.stopped"},
		{"request-lifecycle", "request.received,request.handled,request.finished,request.received,request.failed,request.finished; status=200,500"},
		{"request-finished-ordering", "request.received,request.handled,request.finished,request.received,request.failed,request.finished"},
		{"console-lifecycle", "console.application.starting,console.command.starting,console.command.finished"},
		{"vendor-publish-event", "tag=event-demo; published=1; skipped=0"},
		{"listener-error-isolation", "handled=1; reported=1"},
		{"listener-panic-isolation", "handled=1; reported=1"},
		{"async-failure-isolation", "handled=1; reported=1"},
		{"queued-failure-handling", "handled=1; reported=1"},
		{"consistency-boundary", "dispatch-return=void"},
		{"isolated-testing", "handled=1; reported=0"},
		{"queued-sync-testing", "handled=1; event=demo.event.registered"},
		{"event-interface", "name=demo.event.contract"},
		{"listener-interface", "handled=1"},
		{"listener-func-interface", "handled=1"},
		{"dispatcher-interface", "has=true; handled=1"},
		{"subscriber-interface", "handled=1"},
		{"should-queue-interface", "should-queue=true"},
		{"async-listener-interface", "async=true; implemented=true"},
		{"queue-options-interface", "connection=sync; queue=event-demo; tries=3"},
		{"provider-register", "bound=true; dispatcher=*event.Dispatcher"},
		{"provider-boot", "queued-job-registered=true"},
		{"laravel-compatibility", "named Event, Listener, Subscriber"},
	}
	if len(cases) != 26 {
		t.Fatalf("second-batch cases = %d, want 26", len(cases))
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			demotest.NewApplication(t, demotest.Options{})
			got, err := eventdemo.Run(context.Background(), tc.name, "sync")
			if err != nil {
				t.Fatalf("Run(%q) error = %v, want nil", tc.name, err)
			}
			if got.Case != tc.name || !strings.Contains(got.Value, tc.want) {
				t.Fatalf("Run(%q) = %#v, want case %q and value containing %q", tc.name, got, tc.name, tc.want)
			}
		})
	}
}
