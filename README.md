# fiberprometheus

Prometheus middleware for gofiber.

Forked from [ansrivas](https://github.com/ansrivas/fiberprometheus) <br>

Please feel free to open any issue with possible enhancements or bugs

**Note: Requires Go 1.22 and above**

![Release](https://img.shields.io/github/release/iamlookod/fiberprometheus.svg)
![Test](https://github.com/iamlookod/fiberprometheus/workflows/Test/badge.svg)
![Security](https://github.com/iamlookod/fiberprometheus/workflows/Security/badge.svg)
![Linter](https://github.com/iamlookod/fiberprometheus/workflows/Linter/badge.svg)
![Coverage](https://img.shields.io/badge/coverage-100%25-brightgreen)

Following metrics are available by default:

```
http_requests_total
http_request_duration_seconds
http_requests_in_progress_total
http_cache_results
```

### Install v3

```
go get -u github.com/gofiber/fiber/v3
go get -u github.com/iamlookod/fiberprometheus/v3
```

### Example using v3

#### Basic Usage

```go
package main

import (
	"github.com/iamlookod/fiberprometheus/v3"
	"github.com/gofiber/fiber/v3"
)

func main() {
  app := fiber.New()

  // This here will appear as a label, one can also use
  // fiberprometheus.NewWith(servicename, namespace, subsystem )
  // or
  // labels := map[string]string{"custom_label1":"custom_value1", "custom_label2":"custom_value2"}
  // fiberprometheus.NewWithLabels(labels, namespace, subsystem )
  prom := fiberprometheus.New("my-service-name")
  prom.RegisterAt(app, "/metrics")
  app.Use(prom.Middleware)

  app.Get("/", func(c fiber.Ctx) error {
    return c.SendString("Hello World")
  })

  app.Post("/some", func(c fiber.Ctx) error {
    return c.SendString("Welcome!")
  })

  app.Listen(":3000")
}
```

#### Advanced Usage with Skip Paths and Ignore Status Codes

```go
package main

import (
	"github.com/iamlookod/fiberprometheus/v3"
	"github.com/gofiber/fiber/v3"
)

func main() {
  app := fiber.New()

  prom := fiberprometheus.New("my-service-name")

  // Skip specific paths from being tracked in metrics
  // Useful for health checks, readiness probes, etc.
  prom.SetSkipPaths([]string{"/health", "/readiness", "/livez"})

  // Ignore specific HTTP status codes from being recorded in metrics
  // Useful to reduce noise from common status codes like 404
  prom.SetIgnoreStatusCodes([]int{404, 401})

  prom.RegisterAt(app, "/metrics")
  app.Use(prom.Middleware)

  app.Get("/", func(c fiber.Ctx) error {
    return c.SendString("Hello World")
  })

  app.Get("/health", func(c fiber.Ctx) error {
    return c.SendString("OK") // This won't be tracked in metrics
  })

  app.Listen(":3000")
}
```

### Result

- Hit the default url at http://localhost:3000
- Navigate to http://localhost:3000/metrics

### Features

- ✅ **100% Test Coverage** - Comprehensive test suite covering all functionality
- 📊 **Multiple Metrics** - Tracks requests, duration, in-flight requests, and cache hits/misses
- 🔧 **Configurable** - Skip specific paths and status codes
- 🏷️ **Custom Labels** - Add custom labels to your metrics
- 🎯 **Flexible Registration** - Use custom registries and namespaces
- 🔐 **Authentication Support** - Works with Fiber middleware like BasicAuth
- ⚡ **High Performance** - Minimal overhead with efficient metric collection

### Grafana Board

- https://grafana.com/grafana/dashboards/14331
