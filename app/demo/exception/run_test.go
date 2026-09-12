package exceptiondemo_test

import (
	"strings"
	"testing"

	"prismgo-demo/app/demo/catalog"
	exceptiondemo "prismgo-demo/app/demo/exception"
	demotest "prismgo-demo/app/demo/testing"
)

func TestExceptionDemoFirstBatch(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{name: "architecture", want: "handler=*exception.Handler; should_report=true; render=500"},
		{name: "config", want: "app.debug=false; exception_handler=true"},
		{name: "debug-config", want: "configured=false; standalone=false"},
		{name: "middleware-config", want: "configured=true; enabled=500; disabled=200; direct=500"},
		{name: "default-report", want: "error_channel=true"},
		{name: "custom-reporter", want: "context=context.backgroundCtx; error=true; status=503"},
		{name: "reporter-order", want: "logged_before_reporter=true"},
		{name: "package-report", want: "logged=true; caller=true"},
		{name: "log-context", want: "log_tenant=tenant-42; response_tenant=false"},
		{name: "default-level", want: "500=error; 422=warn"},
		{name: "custom-level", want: "custom=info; fallback=error"},
		{name: "level-constants", want: "debug,info,warn,error"},
		{name: "dont-report", want: "should_report=false; reporters=0"},
		{name: "predicate-order", want: "first,second"},
		{name: "context-ignored", want: "canceled=false; deadline=false"},
		{name: "clear-ignored", want: "canceled=true"},
		{name: "default-render", want: "status=500; type=internal_error; detail=Internal Server Error"},
		{name: "renderer-chain", want: "order=skip,handle; status=409; fallback=500"},
		{name: "problem-renderer", want: "status=409; type=order_conflict; code=1001; message=order already closed"},
		{name: "response-renderer", want: "status=503; html=true"},
		{name: "response-fallback", want: "status=500; type=internal_error"},
		{name: "panic-recovery", want: "status=500; reported=true; safe=true"},
		{name: "gin-errors", want: "status=409; joined=true"},
		{name: "status-report", want: "reported=[404 503]"},
		{name: "request-id", want: "header=demo-request-42; problem=demo-request-42; log=demo-request-42"},
		{name: "problem-response", want: "status=409; type=order_conflict; code=1001; message=order already closed"},
		{name: "problem-optional", want: "error=required; omitted=true; request_id=demo-request-42"},
		{name: "problem-fields", want: "public_status=500; plain_trace=false; debug_trace=true"},
		{name: "http-error", want: "status=404; detail=order missing"},
		{name: "public-detail", want: "client=invalid order; server=Internal Server Error; debug_internal=true"},
	}
	if len(cases) != 30 {
		t.Fatalf("exception case count = %d, want 30", len(cases))
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("LOG_CHANNEL", "error")
			t.Setenv("ERROR_LOGGER_DRIVER", "single")
			t.Setenv("ERROR_LOGGER_FILE", t.TempDir()+"/exception.log")
			_ = demotest.NewApplication(t, demotest.Options{})
			entry, ok := catalog.Find("exception", test.name)
			if !ok || entry.Status != catalog.StatusImplemented {
				t.Fatalf("exception catalog case %q = %#v, found %t; want implemented", test.name, entry, ok)
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

func TestExceptionDemoUnknownScenario(t *testing.T) {
	_, err := exceptiondemo.Run("missing")
	if err == nil || !strings.Contains(err.Error(), `unknown exception scenario "missing"`) {
		t.Fatalf("unknown exception scenario error = %v, want named unknown scenario", err)
	}
}
