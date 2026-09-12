package exceptiondemo_test

import (
	"testing"

	"prismgo-demo/app/demo/catalog"
	exceptiondemo "prismgo-demo/app/demo/exception"
	demotest "prismgo-demo/app/demo/testing"
)

func TestExceptionDemoThirdBatch(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{name: "handler-render", want: "returned=500; response=500; type=internal_error"},
		{name: "should-report", want: "ignored=false; client=true; disabled_client=false"},
		{name: "handler-level", want: "429=info; 500=error"},
		{name: "handler-debug", want: "before=false; after=true"},
		{name: "apply-options", want: "logging=true; debug=true"},
		{name: "cli-report", want: "reports=2; component=cli; commands=exception:fail,exception:panic"},
		{name: "routine-report", want: "reports=2; component=routine; routine=exception-demo"},
		{name: "queue-report", want: "reports=1; component=queue; subsystem=worker"},
		{name: "horizon-report", want: "reports=2; component=horizon; subsystems=worker,supervisor"},
		{name: "event-report", want: "reports=2; component=event; event=exception.demo.failed"},
		{name: "non-http-fields", want: "reports=1; status=500; caller=non-http"},
		{name: "scrub-keys", want: "password=[redacted]; token=[redacted]"},
		{name: "scrub-nested", want: "authorization=[redacted]; cookie=[redacted]"},
		{name: "scrub-service-key", want: "service_key=exception.handler"},
		{name: "wrap-handler", want: "reported=true; status=500; safe=true"},
		{name: "replace-handler", want: "replaced=true; status=503; body=service unavailable; logging=false; recovery=false; debug=false"},
		{name: "laravel-mapping", want: "dont_report=2; reporters=1; status=418; level=error"},
	}
	if len(cases) != 17 {
		t.Fatalf("third exception batch cases = %d, want 17", len(cases))
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("LOG_CHANNEL", "error")
			t.Setenv("ERROR_LOGGER_DRIVER", "single")
			t.Setenv("ERROR_LOGGER_FILE", t.TempDir()+"/exception.log")
			_ = demotest.NewApplication(t, demotest.Options{})
			entry, ok := catalog.Find("exception", test.name)
			if !ok || entry.Status != catalog.StatusImplemented || entry.Test != "TestExceptionDemoThirdBatch/"+test.name {
				t.Fatalf("exception catalog case %q = %#v, found %t; want implemented entry linked to subtest", test.name, entry, ok)
			}
			result, err := exceptiondemo.Run(test.name)
			if err != nil {
				t.Fatalf("exception case %q error = %v, want nil", test.name, err)
			}
			if result.Case != test.name || result.Value != test.want {
				t.Fatalf("exception case %q result = %#v, want case %q and value %q", test.name, result, test.name, test.want)
			}
		})
	}
}
