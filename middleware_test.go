//
// Copyright (c) 2021-present Ankur Srivastava and Contributors
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package fiberprometheus

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/basicauth"
	"github.com/gofiber/fiber/v3/middleware/cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
)

func TestMiddleware(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	prometheus := New("test-service")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello World")
	})
	app.Get("/error/:type", func(ctx fiber.Ctx) error {
		switch ctx.Params("type") {
		case "fiber":
			return fiber.ErrBadRequest
		default:
			return fiber.ErrInternalServerError
		}
	})
	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/error/fiber", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/error/unknown", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/metrics", nil)
	resp, _ = app.Test(req)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	got := string(body)
	want := `http_requests_total{method="GET",path="/",service="test-service",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `http_requests_total{method="GET",path="/error/fiber",service="test-service",status_code="400"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `http_requests_total{method="GET",path="/error/unknown",service="test-service",status_code="500"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `http_request_duration_seconds_count{method="GET",path="/",service="test-service",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `http_requests_in_progress_total{method="GET",service="test-service"} 0`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}
}

func TestMiddlewareWithGroup(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	prometheus := New("test-service")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	// Define Group
	public := app.Group("/public")

	// Define Group Routes
	public.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello World")
	})
	public.Get("/error/:type", func(ctx fiber.Ctx) error {
		switch ctx.Params("type") {
		case "fiber":
			return fiber.ErrBadRequest
		default:
			return fiber.ErrInternalServerError
		}
	})
	req := httptest.NewRequest("GET", "/public", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/public/error/fiber", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/public/error/unknown", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/metrics", nil)
	resp, _ = app.Test(req)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	got := string(body)
	want := `http_requests_total{method="GET",path="/public",service="test-service",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `http_requests_total{method="GET",path="/public/error/fiber",service="test-service",status_code="400"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `http_requests_total{method="GET",path="/public/error/unknown",service="test-service",status_code="500"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `http_request_duration_seconds_count{method="GET",path="/public",service="test-service",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `http_requests_in_progress_total{method="GET",service="test-service"} 0`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}
}

func TestMiddlewareWithServiceName(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	prometheus := NewWith("unique-service", "my_service_with_name", "http")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello World")
	})
	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/metrics", nil)
	resp, _ = app.Test(req)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	got := string(body)
	want := `my_service_with_name_http_requests_total{method="GET",path="/",service="unique-service",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `my_service_with_name_http_request_duration_seconds_count{method="GET",path="/",service="unique-service",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `my_service_with_name_http_requests_in_progress_total{method="GET",service="unique-service"} 0`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}
}

func TestMiddlewareWithLabels(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	constLabels := map[string]string{
		"customkey1": "customvalue1",
		"customkey2": "customvalue2",
	}
	prometheus := NewWithLabels(constLabels, "my_service", "http")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)
	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello World")
	})
	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/metrics", nil)
	resp, _ = app.Test(req)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	got := string(body)
	want := `my_service_http_requests_total{customkey1="customvalue1",customkey2="customvalue2",method="GET",path="/",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `my_service_http_request_duration_seconds_count{customkey1="customvalue1",customkey2="customvalue2",method="GET",path="/",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `my_service_http_requests_in_progress_total{customkey1="customvalue1",customkey2="customvalue2",method="GET"} 0`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}
}

func TestMiddlewareWithBasicAuth(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	// Encode password using SHA256 as required by Fiber v3
	hash := sha256.Sum256([]byte("password"))
	encodedPassword := base64.StdEncoding.EncodeToString(hash[:])

	prometheus := New("basic-auth")
	prometheus.RegisterAt(app, "/metrics", basicauth.New(basicauth.Config{
		Users: map[string]string{
			"prometheus": encodedPassword,
		},
	}))

	app.Use(prometheus.Middleware)

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/metrics", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != 401 {
		t.Fail()
	}

	req.SetBasicAuth("prometheus", "password")
	resp, _ = app.Test(req)
	if resp.StatusCode != 200 {
		t.Fail()
	}
}

func TestMiddlewareWithCustomRegistry(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	registry := prometheus.NewRegistry()

	srv := httptest.NewServer(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	t.Cleanup(srv.Close)

	promfiber := NewWithRegistry(registry, "unique-service", "my_service_with_name", "http", nil)
	app.Use(promfiber.Middleware)

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello World")
	})
	req := httptest.NewRequest("GET", "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fail()
	}
	if resp.StatusCode != 200 {
		t.Fail()
	}

	resp, err = srv.Client().Get(srv.URL)
	if err != nil {
		t.Fail()
	}
	if resp == nil {
		t.Fatal("response is nil")
	}
	if resp.StatusCode != 200 {
		t.Fail()
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	got := string(body)
	want := `my_service_with_name_http_requests_total{method="GET",path="/",service="unique-service",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `my_service_with_name_http_request_duration_seconds_count{method="GET",path="/",service="unique-service",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `my_service_with_name_http_requests_in_progress_total{method="GET",service="unique-service"} 0`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}
}

func TestCustomRegistryRegisterAt(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	registry := prometheus.NewRegistry()
	registry.Register(collectors.NewGoCollector())
	registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	fpCustom := NewWithRegistry(registry, "custom-registry", "custom_name", "http", nil)
	fpCustom.RegisterAt(app, "/metrics")

	app.Use(fpCustom.Middleware)

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, world!")
	})
	req := httptest.NewRequest("GET", "/", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(fmt.Errorf("GET / failed: %w", err))
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatal(fmt.Errorf("GET /: Status=%d", res.StatusCode))
	}

	req = httptest.NewRequest("GET", "/metrics", nil)
	resMetr, err := app.Test(req)
	if err != nil {
		t.Fatal(fmt.Errorf("GET /metrics failed: %W", err))
	}
	defer resMetr.Body.Close()
	if res.StatusCode != 200 {
		t.Fatal(fmt.Errorf("GET /metrics: Status=%d", resMetr.StatusCode))
	}
	body, err := io.ReadAll(resMetr.Body)
	if err != nil {
		t.Fatal(fmt.Errorf("GET /metrics: read body: %w", err))
	}
	got := string(body)

	want := `custom_name_http_requests_total{method="GET",path="/",service="custom-registry",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}
}

func TestWithCacheMiddleware(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	registry := prometheus.NewRegistry()
	registry.Register(collectors.NewGoCollector())
	registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	fpCustom := NewWithRegistry(registry, "custom-registry", "custom_name", "http", nil)
	fpCustom.RegisterAt(app, "/metrics")

	app.Use(fpCustom.Middleware)
	app.Use(cache.New())

	app.Get("/myPath", func(c fiber.Ctx) error {
		return c.SendString("Hello, world!")
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/myPath", nil)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(fmt.Errorf("GET / failed: %w", err))
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			t.Fatal(fmt.Errorf("GET /: Status=%d", res.StatusCode))
		}
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(fmt.Errorf("GET /metrics failed: %W", err))
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatal(fmt.Errorf("GET /metrics: Status=%d", res.StatusCode))
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(fmt.Errorf("GET /metrics: read body: %w", err))
	}
	got := string(body)
	want := `custom_name_http_requests_total{method="GET",path="/myPath",service="custom-registry",status_code="200"} 2`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `custom_name_http_cache_results{cache_result="hit",method="GET",path="/myPath",service="custom-registry",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `custom_name_http_cache_results{cache_result="miss",method="GET",path="/myPath",service="custom-registry",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}
}

func TestWithCacheMiddlewareWithCustomKey(t *testing.T) {
	t.Parallel()
	app := fiber.New()
	registry := prometheus.NewRegistry()
	registry.Register(collectors.NewGoCollector())
	registry.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	fpCustom := NewWithRegistry(registry, "custom-registry", "custom_name", "http", nil)
	fpCustom.RegisterAt(app, "/metrics")
	fpCustom.CustomCacheKey("my-custom-cache-header")

	app.Use(fpCustom.Middleware)
	app.Use(cache.New(
		cache.Config{
			CacheHeader: "my-custom-cache-header",
		},
	))

	app.Get("/myPath", func(c fiber.Ctx) error {
		return c.SendString("Hello, world!")
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/myPath", nil)
		res, err := app.Test(req)
		if err != nil {
			t.Fatal(fmt.Errorf("GET / failed: %w", err))
		}
		defer res.Body.Close()
		if res.StatusCode != 200 {
			t.Fatal(fmt.Errorf("GET /: Status=%d", res.StatusCode))
		}
	}

	req := httptest.NewRequest("GET", "/metrics", nil)
	res, err := app.Test(req)
	if err != nil {
		t.Fatal(fmt.Errorf("GET /metrics failed: %W", err))
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		t.Fatal(fmt.Errorf("GET /metrics: Status=%d", res.StatusCode))
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatal(fmt.Errorf("GET /metrics: read body: %w", err))
	}
	got := string(body)
	want := `custom_name_http_requests_total{method="GET",path="/myPath",service="custom-registry",status_code="200"} 2`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `custom_name_http_cache_results{cache_result="hit",method="GET",path="/myPath",service="custom-registry",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	want = `custom_name_http_cache_results{cache_result="miss",method="GET",path="/myPath",service="custom-registry",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}
}

func TestRegistryNotGatherer(t *testing.T) {
	t.Parallel()

	// Create a custom registerer that is NOT a gatherer
	type registererOnly struct {
		prometheus.Registerer
	}

	// Use a real registry but wrap it to hide the Gatherer interface
	realRegistry := prometheus.NewRegistry()
	registererOnlyInstance := registererOnly{Registerer: realRegistry}

	// This should trigger the fallback to DefaultGatherer
	fpCustom := NewWithRegistry(registererOnlyInstance, "test-service", "test", "http", nil)

	if fpCustom.gatherer != prometheus.DefaultGatherer {
		t.Errorf("Expected gatherer to be DefaultGatherer when registry is not a Gatherer")
	}
}

func TestMiddlewareSkipsMetricsEndpoint(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	prometheus := New("test-service")

	// Register middleware first, then the metrics endpoint
	// This ensures the middleware will be called for all routes including /metrics
	app.Use(prometheus.Middleware)
	prometheus.RegisterAt(app, "/metrics")

	app.Get("/test", func(c fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	// Make a request to /test
	req := httptest.NewRequest("GET", "/test", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Fail()
	}

	// Make multiple requests to /metrics endpoint itself to ensure middleware sees it
	for i := 0; i < 3; i++ {
		req = httptest.NewRequest("GET", "/metrics", nil)
		resp, _ = app.Test(req)
		if resp.StatusCode != 200 {
			t.Errorf("Expected 200 for /metrics, got %d", resp.StatusCode)
		}
		resp.Body.Close()
	}

	// Now fetch metrics to verify
	req = httptest.NewRequest("GET", "/metrics", nil)
	resp, _ = app.Test(req)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fail()
	}

	body, _ := io.ReadAll(resp.Body)
	got := string(body)

	// Verify that /test was tracked
	want := `http_requests_total{method="GET",path="/test",service="test-service",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("got %s; want %s", got, want)
	}

	// Verify that /metrics itself was NOT tracked (should not appear in metrics)
	notWant := `path="/metrics"`
	if strings.Contains(got, notWant) {
		t.Errorf("Metrics endpoint should not track itself, but found: %s", notWant)
	}
}

func TestSetSkipPaths(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	prometheus := New("test-service")
	prometheus.RegisterAt(app, "/metrics")

	// Set paths to skip
	prometheus.SetSkipPaths([]string{"/health", "/readiness"})

	app.Use(prometheus.Middleware)

	app.Get("/api/users", func(c fiber.Ctx) error {
		return c.SendString("Users")
	})

	app.Get("/health", func(c fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/readiness", func(c fiber.Ctx) error {
		return c.SendString("Ready")
	})

	// Make requests to all endpoints
	req := httptest.NewRequest("GET", "/api/users", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/health", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != 200 {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/readiness", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != 200 {
		t.Fail()
	}

	// Fetch metrics
	req = httptest.NewRequest("GET", "/metrics", nil)
	resp, _ = app.Test(req)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	got := string(body)

	// Verify that /api/users was tracked
	want := `http_requests_total{method="GET",path="/api/users",service="test-service",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("Expected /api/users to be tracked, but got: %s", got)
	}

	// Verify that /health was NOT tracked
	notWant := `path="/health"`
	if strings.Contains(got, notWant) {
		t.Errorf("Expected /health to be skipped, but found: %s", notWant)
	}

	// Verify that /readiness was NOT tracked
	notWant = `path="/readiness"`
	if strings.Contains(got, notWant) {
		t.Errorf("Expected /readiness to be skipped, but found: %s", notWant)
	}
}

func TestSetIgnoreStatusCodes(t *testing.T) {
	t.Parallel()
	app := fiber.New()

	prometheus := New("test-service")
	prometheus.RegisterAt(app, "/metrics")

	// Set status codes to ignore
	prometheus.SetIgnoreStatusCodes([]int{404, 401})

	app.Use(prometheus.Middleware)

	app.Get("/success", func(c fiber.Ctx) error {
		return c.SendString("Success")
	})

	app.Get("/unauthorized", func(c fiber.Ctx) error {
		return c.Status(401).SendString("Unauthorized")
	})

	app.Get("/notfound", func(c fiber.Ctx) error {
		return c.Status(404).SendString("Not Found")
	})

	app.Get("/error", func(c fiber.Ctx) error {
		return c.Status(500).SendString("Internal Server Error")
	})

	// Make requests to all endpoints
	req := httptest.NewRequest("GET", "/success", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode != 200 {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/unauthorized", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != 401 {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/notfound", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != 404 {
		t.Fail()
	}

	req = httptest.NewRequest("GET", "/error", nil)
	resp, _ = app.Test(req)
	if resp.StatusCode != 500 {
		t.Fail()
	}

	// Fetch metrics
	req = httptest.NewRequest("GET", "/metrics", nil)
	resp, _ = app.Test(req)
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	got := string(body)

	// Verify that 200 status was tracked
	want := `http_requests_total{method="GET",path="/success",service="test-service",status_code="200"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("Expected /success (200) to be tracked, but got: %s", got)
	}

	// Verify that 500 status was tracked
	want = `http_requests_total{method="GET",path="/error",service="test-service",status_code="500"} 1`
	if !strings.Contains(got, want) {
		t.Errorf("Expected /error (500) to be tracked, but got: %s", got)
	}

	// Verify that 401 status was NOT tracked
	notWant := `status_code="401"`
	if strings.Contains(got, notWant) {
		t.Errorf("Expected 401 status to be ignored, but found: %s", notWant)
	}

	// Verify that 404 status was NOT tracked
	notWant = `status_code="404"`
	if strings.Contains(got, notWant) {
		t.Errorf("Expected 404 status to be ignored, but found: %s", notWant)
	}
}

func TestSetSkipPathsMultipleCalls(t *testing.T) {
	t.Parallel()

	prometheus := New("test-service")

	// Call SetSkipPaths multiple times to test map initialization
	prometheus.SetSkipPaths([]string{"/health"})
	prometheus.SetSkipPaths([]string{"/readiness"})

	// Both paths should be in the skip map
	if !prometheus.skipPaths["/health"] {
		t.Errorf("Expected /health to be in skipPaths")
	}
	if !prometheus.skipPaths["/readiness"] {
		t.Errorf("Expected /readiness to be in skipPaths")
	}
}

func TestSetIgnoreStatusCodesMultipleCalls(t *testing.T) {
	t.Parallel()

	prometheus := New("test-service")

	// Call SetIgnoreStatusCodes multiple times to test map initialization
	prometheus.SetIgnoreStatusCodes([]int{404})
	prometheus.SetIgnoreStatusCodes([]int{401})

	// Both codes should be in the ignore map
	if !prometheus.ignoreStatusCodes[404] {
		t.Errorf("Expected 404 to be in ignoreStatusCodes")
	}
	if !prometheus.ignoreStatusCodes[401] {
		t.Errorf("Expected 401 to be in ignoreStatusCodes")
	}
}

func Benchmark_Middleware(b *testing.B) {
	app := fiber.New()

	prometheus := New("test-benchmark")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	h := app.Handler()
	ctx := &fasthttp.RequestCtx{}

	req := &fasthttp.Request{}
	req.Header.SetMethod(fiber.MethodOptions)
	req.SetRequestURI("/")
	ctx.Init(req, nil, nil)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		h(ctx)
	}
}

func Benchmark_Middleware_Parallel(b *testing.B) {
	app := fiber.New()

	prometheus := New("test-benchmark")
	prometheus.RegisterAt(app, "/metrics")
	app.Use(prometheus.Middleware)

	app.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello World")
	})

	h := app.Handler()

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		ctx := &fasthttp.RequestCtx{}
		req := &fasthttp.Request{}
		req.Header.SetMethod(fiber.MethodOptions)
		req.SetRequestURI("/metrics")
		ctx.Init(req, nil, nil)

		for pb.Next() {
			h(ctx)
		}
	})
}
