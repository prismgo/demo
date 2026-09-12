package exceptiondemo_test

import (
	"testing"

	"prismgo-demo/app/demo/catalog"
	exceptiondemo "prismgo-demo/app/demo/exception"
	demotest "prismgo-demo/app/demo/testing"
)

func TestExceptionDemoSecondBatch(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{name: "safe-response", want: "status=500; concealed=true"},
		{name: "request-fields", want: "status=500; method=GET; path=/failure; url=/failure; query=source=demo; ip=true; duration=true"},
		{name: "diagnostic-fields", want: "request_id=diagnostic-42; errors=true; panic=demo panic; stack=true"},
		{name: "business-fields", want: "status=409; code=1001; type=order_conflict; context=map[order_id:42]; fields=map[order:closed]"},
		{name: "handler-defaults", want: "ignore=1; recovery=true; logging=true; client=true; stack=true; debug=false"},
		{name: "handler-flags", want: "report=false; client=false; recovery=false; stack=false"},
		{name: "options", want: "logging=true; debug=true"},
		{name: "with-dont-report", want: "matched=true; other=true"},
		{name: "with-level", want: "custom=info; fallback=error"},
		{name: "with-reporter", want: "calls=1"},
		{name: "with-recovery", want: "recovery=false; propagated=propagated"},
		{name: "with-logging", want: "should_report=false"},
		{name: "with-client-logging", want: "client=false; server=true"},
		{name: "with-panic-stack", want: "handler_capture=false,true; log_stack=true,true"},
		{name: "with-debug", want: "debug=true; exposed=true"},
		{name: "with-debug-resolver", want: "before=false; after=true"},
		{name: "with-context", want: "status=500; tenant=tenant-42; response_private=true"},
		{name: "with-renderer", want: "status=409; type=demo"},
		{name: "with-response-renderer", want: "status=503; body=maintenance"},
		{name: "predicate-type", want: "matched=true; other=true"},
		{name: "level-resolver-type", want: "custom=info; fallback=error"},
		{name: "reporter-type", want: "contexts=context.backgroundCtx,*gin.Context; error=true; statuses=[503 500]"},
		{name: "context-extractor-type", want: "status=500; tenant=tenant-42; response_private=true"},
		{name: "renderer-type", want: "status=409; type=demo"},
		{name: "response-renderer-type", want: "status=503; body=maintenance"},
		{name: "resolve", want: "registered=true; type=*exception.Handler"},
		{name: "facade-report", want: "reported=true"},
		{name: "facade-render", want: "status=500; type=internal_error"},
		{name: "build-register", want: "factory=true; logging=false; debug=true"},
		{name: "handler-report", want: "reported=true"},
	}
	if len(cases) != 30 {
		t.Fatalf("second exception batch cases = %d, want 30", len(cases))
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("LOG_CHANNEL", "error")
			t.Setenv("ERROR_LOGGER_DRIVER", "single")
			t.Setenv("ERROR_LOGGER_FILE", t.TempDir()+"/exception.log")
			_ = demotest.NewApplication(t, demotest.Options{})
			entry, ok := catalog.Find("exception", test.name)
			if !ok || entry.Status != catalog.StatusImplemented || entry.Test != "TestExceptionDemoSecondBatch/"+test.name {
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
