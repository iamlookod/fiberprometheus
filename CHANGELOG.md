## ChangeLog

---

## [2025-02-16] - v3.1.0

### Added

- **SetSkipPaths()** method to exclude specific paths from metrics collection
  - Useful for health checks, readiness probes, and liveness endpoints
  - Supports multiple paths configuration
- **SetIgnoreStatusCodes()** method to filter specific HTTP status codes from metrics
  - Reduces metrics cardinality by ignoring common status codes (e.g., 404, 401)
  - Helps reduce noise in monitoring dashboards

### Improved

- Achieved **100% test coverage** with comprehensive test suite
- Fixed Fiber v3 API compatibility issues in tests
  - Updated `app.Test()` method calls to match new Fiber v3 signature
  - Fixed BasicAuth to use SHA256-encoded passwords
- Added 4 new test cases for filtering functionality
- Enhanced documentation with advanced usage examples
- Added API reference documentation in README
- Added coverage badge showing 100% test coverage

### Tests

- TestSetSkipPaths: Validates path filtering functionality
- TestSetIgnoreStatusCodes: Validates status code filtering
- TestSetSkipPathsMultipleCalls: Tests map initialization on multiple calls
- TestSetIgnoreStatusCodesMultipleCalls: Tests map initialization on multiple calls
- TestRegistryNotGatherer: Tests fallback to DefaultGatherer
- TestMiddlewareSkipsMetricsEndpoint: Validates metrics endpoint self-exclusion

### Documentation

- Added advanced usage examples in README
- Added comprehensive API reference
- Added features section highlighting capabilities
- Updated badges to include coverage

## [2024-04-21] - v3.0.0

### Refactor to fiberv3

- Updated dependencies to use fiber v3
- Updated tests to use v3 as well

## [2021-03-29] - v2.1.2
### Bug Fix:
- Fixes #39, thanks @sunhailin-Leo

## [2021-02-08] - v2.1.1
### Enhancements:
- Fix the LICENSE headers and introduce MIT License

## [2021-01-18] - v2.1.0
### Enhancements:
- New method `NewWithLabels` now accepts a `map[string]string` so that users can create custom labels easily.
- Bumped gofiber to v2.3.3

## [2020-11-27] - v2.0.1
### Enhancements:
- Bug Fix: RequestInFlight won't decrease if ctx.Next() return error
- Bumped gofiber to v2.2.1
- Use go 1.15

## [2020-09-15] - v2.0.0
### Enhancements:
- Support gofiber-v2
- New import path would be github.com/ansrivas/fiberprometheus/v2


## [2020-07-08] - 0.3.2
### Enhancements:
- Upgrade gofiber to 1.14.4

## [2020-07-08] - 0.3.0
### Enhancements:
- Support a new method to provide a namespace and a subsystem for the service

