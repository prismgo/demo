package routedemo_test

import (
	"strings"
	"testing"

	routedemo "prismgo-demo/app/demo/route"

	"prismgo-demo/app/demo/catalog"
	demotest "prismgo-demo/app/demo/testing"
)

// expectValue executes one scenario and asserts its full observable value.
func expectValue(t *testing.T, name string, want string) {
	t.Helper()
	result, err := routedemo.Run(name)
	if err != nil {
		t.Fatalf("route demo %q error = %v, want nil", name, err)
	}
	if result.Case != name {
		t.Fatalf("route demo %q case = %q, want %q", name, result.Case, name)
	}
	if result.Value != want {
		t.Fatalf("route demo %q value = %q, want %q", name, result.Value, want)
	}
}

func TestRouteDemoArchitecture(t *testing.T) {
	expectValue(t, "architecture", "router=Router route=Route info=RouteInfo binder=Binder mount=Mount")
}

func TestRouteDemoFacade(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	expectValue(t, "facade", "status=200 body=user=7 url=/facade/users/7 singleton=true")
}

func TestRouteDemoMount(t *testing.T) {
	expectValue(t, "mount", "status=200 body=pong routes=1")
}

func TestRouteDemoIsolatedRouter(t *testing.T) {
	expectValue(t, "isolated-router", "first-one=200 first-two=404 second-two=200")
}

func TestRouteDemoHTTPMethods(t *testing.T) {
	expectValue(t, "http-methods", "get=GET post=POST put=PUT patch=PATCH delete=DELETE options=OPTIONS")
}

func TestRouteDemoMatch(t *testing.T) {
	expectValue(t, "match", "put=42 patch=42 get=404")
}

func TestRouteDemoAny(t *testing.T) {
	expectValue(t, "any", "methods=7 statuses=200,200,200,200,200,200,200")
}

func TestRouteDemoLaravelParameters(t *testing.T) {
	expectValue(t, "laravel-parameters", "id=100 slug=hello")
}

func TestRouteDemoGinParameters(t *testing.T) {
	expectValue(t, "gin-parameters", "id=7")
}

func TestRouteDemoParameterRead(t *testing.T) {
	expectValue(t, "parameter-read", "params=alpha/beta")
}

func TestRouteDemoOptionalParameters(t *testing.T) {
	expectValue(t, "optional-parameters", "with=hello without=none")
}

func TestRouteDemoWildcardParameters(t *testing.T) {
	expectValue(t, "wildcard-parameters", "path=assets/js/app.js")
}

func TestRouteDemoWhere(t *testing.T) {
	expectValue(t, "where", "valid=200 invalid=404")
}

func TestRouteDemoWhereNumber(t *testing.T) {
	expectValue(t, "where-number", "valid=200 alpha=404 negative=404")
}

func TestRouteDemoWhereAlpha(t *testing.T) {
	expectValue(t, "where-alpha", "valid=200 numeric=404")
}

func TestRouteDemoWhereAlphaNumeric(t *testing.T) {
	expectValue(t, "where-alphanumeric", "valid=200 symbol=404")
}

func TestRouteDemoWhereUUID(t *testing.T) {
	expectValue(t, "where-uuid", "valid=200 invalid=404")
}

func TestRouteDemoWhereULID(t *testing.T) {
	expectValue(t, "where-ulid", "valid=200 invalid=404")
}

func TestRouteDemoWhereIn(t *testing.T) {
	expectValue(t, "where-in", "open=200 closed=200 other=404")
}

func TestRouteDemoGroupConstraints(t *testing.T) {
	expectValue(t, "group-constraints", "users=200 orders=200 invalid=404")
}

func TestRouteDemoGlobalPattern(t *testing.T) {
	expectValue(t, "global-pattern", "valid=200 invalid=404")
}

func TestRouteDemoConstraintOverride(t *testing.T) {
	expectValue(t, "constraint-override", "pattern=200,404 local=200,404")
}

func TestRouteDemoName(t *testing.T) {
	expectValue(t, "route-name", "name=users.show entries=1")
}

func TestRouteDemoURL(t *testing.T) {
	expectValue(t, "url", "url=/users/100")
}

func TestRouteDemoURLEscaping(t *testing.T) {
	expectValue(t, "url-escaping", "url=/search/hello%20world%2Fx")
}

func TestRouteDemoURLMissingParameter(t *testing.T) {
	expectValue(t, "url-missing-parameter", "missing=true")
}

func TestRouteDemoGroupNamePrefix(t *testing.T) {
	expectValue(t, "group-name-prefix", "url=/admin/users/100")
}

func TestRouteDemoDuplicateName(t *testing.T) {
	expectValue(t, "duplicate-route-name", "detail=/users/100 old-missing=true")
}

func TestRouteDemoGroupPrefix(t *testing.T) {
	expectValue(t, "group-prefix", "index=200 show=200")
}

func TestRouteDemoGroupChain(t *testing.T) {
	expectValue(t, "group-chain", "name=api.profile value=yes header=yes")
}

func TestRouteDemoNestedGroups(t *testing.T) {
	expectValue(t, "nested-groups", "url=/api/admin/users name=api.admin.users.index")
}

func TestRouteDemoGroupMiddleware(t *testing.T) {
	expectValue(t, "group-middleware", "group=2 total=2")
}

func TestRouteDemoRouteMiddleware(t *testing.T) {
	expectValue(t, "route-middleware", "with=middleware>action bare=bare")
}

func TestRouteDemoNamedMiddleware(t *testing.T) {
	expectValue(t, "named-middleware", "profile=on public=off names=auth")
}

func TestRouteDemoMiddlewareFunctionName(t *testing.T) {
	expectValue(t, "middleware-function-name", "name=middlewareNameGuard guarded=on open=off")
}

func TestRouteDemoRouteWithoutMiddleware(t *testing.T) {
	expectValue(t, "route-without-middleware", "both=group:on,local:on group-only=group:on,local:off")
}

func TestRouteDemoRegistrarWithoutMiddleware(t *testing.T) {
	expectValue(t, "registrar-without-middleware", "full=auth:on,audit:on callback=auth:on,audit:off")
}

func TestRouteDemoBind(t *testing.T) {
	expectValue(t, "bind", "bound=user:42")
}

func TestRouteDemoModel(t *testing.T) {
	expectValue(t, "model", "model=42")
}

func TestRouteDemoBindingContext(t *testing.T) {
	expectValue(t, "binding-context", "value=7|acct-7|true")
}

func TestRouteDemoMissingHandler(t *testing.T) {
	expectValue(t, "missing-handler", "status=404 body=missing:404")
}

func TestRouteDemoDefaultMissing(t *testing.T) {
	expectValue(t, "default-missing", "status=404 body=")
}

func TestRouteDemoControllerAction(t *testing.T) {
	expectValue(t, "controller-action", "index=index show=show:7")
}

func TestRouteDemoControllerValidation(t *testing.T) {
	expectValue(t, "controller-validation", `missing="route: controller action Missing not found" signature="route: controller action Bad must have signature func(*gin.Context), got func(*gin.Context) int" unconfigured="route: controller is not configured"`)
}

func TestRouteDemoAPIResource(t *testing.T) {
	expectValue(t, "api-resource", "routes=5 GET /photos=photos.index;POST /photos=photos.store;DELETE /photos/{photo}=photos.destroy;GET /photos/{photo}=photos.show;PUT,PATCH /photos/{photo}=photos.update")
}

func TestRouteDemoResource(t *testing.T) {
	expectValue(t, "resource", "routes=7 GET /photos=photos.index;POST /photos=photos.store;GET /photos/create=photos.create;DELETE /photos/{photo}=photos.destroy;GET /photos/{photo}=photos.show;PUT,PATCH /photos/{photo}=photos.update;GET /photos/{photo}/edit=photos.edit")
}

func TestRouteDemoResourceCreate(t *testing.T) {
	expectValue(t, "resource-create", "routes=1 create=200 list=404")
}

func TestRouteDemoResourceEdit(t *testing.T) {
	expectValue(t, "resource-edit", "routes=1 edit=200 list=404")
}

func TestRouteDemoResourceOnly(t *testing.T) {
	expectValue(t, "resource-only", "routes=2 index=200 show=200 store=404")
}

func TestRouteDemoResourceExcept(t *testing.T) {
	expectValue(t, "resource-except", "routes=4 destroy=404")
}

func TestRouteDemoResourceNames(t *testing.T) {
	expectValue(t, "resource-names", "index=/photos show=/photos/9")
}

func TestRouteDemoResourceParameters(t *testing.T) {
	expectValue(t, "resource-parameters", "uri=/photos/{photo_id} url=/photos/9")
}

func TestRouteDemoAPIResources(t *testing.T) {
	expectValue(t, "api-resources", "routes=10 photos=/photos/8 posts=/posts/9")
}

func TestRouteDemoNestedResource(t *testing.T) {
	expectValue(t, "nested-resource", "routes=5 index=/users/photos show=/users/photos/9")
}

func TestRouteDemoResourceControllerContract(t *testing.T) {
	expectValue(t, "resource-controller-contract", "interface=ResourceController methods=Destroy,Index,Show,Store,Update")
}

func TestRouteDemoCreateControllerContract(t *testing.T) {
	expectValue(t, "create-controller-contract", "interface=CreateController methods=Create")
}

func TestRouteDemoEditControllerContract(t *testing.T) {
	expectValue(t, "edit-controller-contract", "interface=EditController methods=Edit")
}

func TestRouteDemoRedirect(t *testing.T) {
	expectValue(t, "redirect", "default=302:/new temporary=307:/target")
}

func TestRouteDemoPermanentRedirect(t *testing.T) {
	expectValue(t, "permanent-redirect", "status=301 location=/current")
}

func TestRouteDemoStatic(t *testing.T) {
	expectValue(t, "static", "file=200:static-body missing=404")
}

func TestRouteDemoFallback(t *testing.T) {
	expectValue(t, "fallback", "known=200:ok unknown=404:fallback")
}

func TestRouteDemoDomain(t *testing.T) {
	expectValue(t, "domain", "match=200 mismatch=404 port=200")
}

func TestRouteDemoDomainPlaceholder(t *testing.T) {
	expectValue(t, "domain-placeholder", "tenant=200 nested=404 root=404")
}

func TestRouteDemoRateLimiter(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	expectValue(t, "rate-limiter", "max=5 every=1m0s key=user:42 status=200")
}

func TestRouteDemoLimit(t *testing.T) {
	expectValue(t, "limit", "max=3 every=1m0s key=tenant:9")
}

func TestRouteDemoThrottleRoute(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	expectValue(t, "throttle-route", "statuses=200,200,429")
}

func TestRouteDemoThrottleGroup(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	expectValue(t, "throttle-group", "a1=200 a2=200 b1=429")
}

func TestRouteDemoThrottleUnknown(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	expectValue(t, "throttle-unknown", "status=200 body=open")
}

func TestRouteDemoThrottleOverLimit(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	expectValue(t, "throttle-over-limit", "first=200 second=429 remaining=0 retry-positive=true")
}

func TestRouteDemoCurrentRoute(t *testing.T) {
	expectValue(t, "current-route", "info=users.show|/users/{id}|/users/:id")
}

func TestRouteDemoRouteInfo(t *testing.T) {
	expectValue(t, "route-info", "interface=RouteInfo fields=Methods:[]string,URI:string,GinPath:string,Name:string,Domain:string,Handler:string,Middleware:[]string,SourcePath:string")
}

func TestRouteDemoList(t *testing.T) {
	expectValue(t, "list", "routes=2 GET /users=users.index;GET /users/{id}=users.show")
}

func TestRouteDemoListCommand(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	expectValue(t, "list-command", "routes=1 name=api.users.show")
}

func TestRouteDemoHandlerOrder(t *testing.T) {
	expectValue(t, "handler-order", "order=group>route>action")
}

func TestRouteDemoResolve(t *testing.T) {
	demotest.NewApplication(t, demotest.Options{})
	expectValue(t, "resolve", "singleton=true routes=1")
}

func TestRouteDemoReset(t *testing.T) {
	expectValue(t, "reset", "before=1 after=0 name-missing=true pattern=200")
}

func TestRouteDemoClone(t *testing.T) {
	expectValue(t, "clone", "original=1 cloned=2 isolated=true url=/cloned")
}

func TestRouteDemoAdd(t *testing.T) {
	expectValue(t, "add", "methods=GET,POST get=200 post=200")
}

func TestRouteDemoRouterGroup(t *testing.T) {
	expectValue(t, "router-group", "routes=2 inside=200 outside=200")
}

func TestRouteDemoRouteScopeBindings(t *testing.T) {
	expectValue(t, "route-scope-bindings", "chain=true routes=1")
}

func TestRouteDemoRegistrarScopeBindings(t *testing.T) {
	expectValue(t, "registrar-scope-bindings", "chain=true routes=1")
}

func TestRouteDemoRegistrarOverrides(t *testing.T) {
	expectValue(t, "registrar-overrides", "domain=inner.test action=inner outer=404")
}

func TestRouteDemoFacadeContract(t *testing.T) {
	expectValue(t, "facade-contract", "facade=Resolve,Bind,Model,Pattern,Mount,List,URL,Get,Post,Put,Patch,Delete,Options,Match,Any,Redirect,PermanentRedirect,Static,Fallback,Prefix,Name,Domain,Middleware,WithoutMiddleware,Controller,Group,Resource,ApiResource,ApiResources")
}

func TestRouteDemoProviderRegistration(t *testing.T) {
	expectValue(t, "provider-registration", "name=route bound=true lazy=true router=true")
}

func TestRouteDemoProviderPreservesRouter(t *testing.T) {
	expectValue(t, "provider-preserves-router", "preserved=true routes=1")
}

func TestRouteDemoProviderSingleton(t *testing.T) {
	expectValue(t, "provider-singleton", "singleton=true resolved=true")
}

func TestRouteDemoBestPractices(t *testing.T) {
	expectValue(t, "best-practices", "url=/api/v1/prismgos/7 routes=1 name=api.prismgos.show")
}

func TestRouteDemoRejectsUnknownScenario(t *testing.T) {
	_, err := routedemo.Run("unknown")
	if err == nil || !strings.Contains(err.Error(), `unknown scenario "unknown"`) {
		t.Fatalf("Run(unknown) error = %v, want unknown scenario error", err)
	}
}

func TestRouteDemoCatalogCoverage(t *testing.T) {
	entries := catalog.Filter("route", "", catalog.StatusImplemented)
	if len(entries) != 87 {
		t.Fatalf("implemented route entries = %d, want 87", len(entries))
	}
	if item, ok := catalog.Find("route", "architecture"); !ok || item.Status != catalog.StatusImplemented {
		t.Fatalf("route architecture entry = %#v, %v; want implemented", item, ok)
	}
	if item, ok := catalog.Find("route", "domain-placeholder"); !ok || item.Status != catalog.StatusImplemented {
		t.Fatalf("route domain-placeholder entry = %#v, %v; want implemented", item, ok)
	}
	if item, ok := catalog.Find("route", "best-practices"); !ok || item.Status != catalog.StatusImplemented {
		t.Fatalf("route best-practices entry = %#v, %v; want implemented", item, ok)
	}
	if item, ok := catalog.Find("route", "rate-limiter"); !ok || item.Status != catalog.StatusImplemented {
		t.Fatalf("route rate-limiter entry = %#v, %v; want implemented", item, ok)
	}
}
