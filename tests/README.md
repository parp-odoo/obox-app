# Tests

All test code is kept here, separate from the application packages:

```text
tests/
  run.go                         Overlay test runner
  testutil/                      Shared assertions and mock devices
  fixtures/frontend/index.html   Embedded frontend fixture
  _packages/
    app_test.go                  Root main package tests
    internal/
      config/
      escpos/
      logger/
      printer/
      server/
      util/
```

## Running tests

Use the project's Go toolchain (CI uses Go 1.26), CGO, and native libusb development
dependencies. On Linux, install `libusb-1.0-0-dev`. From the repository root:

```bash
# All packages
go run ./tests/run.go

# All packages with Go test flags
go run ./tests/run.go -v -count=1 ./...

# Only the printer package
go run ./tests/run.go -v ./internal/printer

# Selected tests in one package
go run ./tests/run.go -run '^TestEncodePrinterID$' ./internal/printer

# Coverage or race detection
go run ./tests/run.go -coverprofile=coverage.out ./...
go run ./tests/run.go -race ./...
```

The runner forwards arguments to `go test`. Without arguments it selects `./...`.
When passing flags, include a package selector; flags alone use Go's usual default
of testing the root package. An explicit selector runs only the requested packages.
Relative selectors are resolved from the repository root, even when the runner is
invoked from another directory.

Packages run serially by default (`-p=1`) because several suites use mock printers
on TCP port 9100. Avoid running multiple copies of these suites simultaneously.
The runner uses the same Go installation that compiled it, generates a temporary
overlay, removes it afterward, and reports a failing test as a failed command.

## Private package access

The overlay makes each file in `_packages/` appear in its original production
package directory. This preserves access to unexported functions, methods, and
fields, including tests of the root `main` package. Production source directories
are never populated with copies of test files.

Go ignores directories starting with `_` during normal package discovery, so the
physical suites are not compiled as separate packages. Plain `go test ./...` skips
these suites; use the runner, as CI does.

## Frontend fixture

During tests, the runner makes `tests/fixtures/frontend/index.html` appear at
`frontend/dist/index.html`. The existing embed directive in `main.go` works even
when the frontend bundle is absent, and remains compatible with coverage.
This compiles the real entry point without requiring frontend assets, generated
Wails bindings, or npm packages. If a frontend bundle already exists, its other
assets are still embedded, with its index replaced by the fixture during tests.
Backend tests do not launch or validate the frontend.

Production builds continue to use the original source and frontend assets. No
fixture is copied into `frontend/dist`. If the embed path changes, update the
runner's corresponding virtual asset path.

## Adding tests

Mirror the production package path under `tests/_packages/`, keep its package
declaration, and name the file `*_test.go`. For example:

```text
internal/printer/lan.go  -> tests/_packages/internal/printer/lan_test.go
app.go                  -> tests/_packages/app_test.go
```

Use `package printer` for printer tests and `package main` for root tests. The
runner automatically discovers new files, including platform-specific test files;
no manifest needs updating. Import helpers from `epos-proxy/tests/testutil`.

Overlays affect compilation only. Runtime fixture paths remain relative to the
original package directory, and editors may require additional configuration to
understand the tests' virtual package locations.
