# ePOS Proxy

A lightweight desktop app that exposes your USB Epson printer as a local HTTP endpoint, making it compatible with Odoo Point of Sale.

## Installation

### Download

Download the latest release for your platform from the [Releases](../../releases) page.

## Usage

- Connect your printer via USB to your computer
- Launch ePOS Proxy - the application will detect your printer automatically

![screenshot](readme/printer.png)

- Copy the printer address shown in the UI
- Configure Odoo Point of Sale to use this address 

![screenshot](readme/config.png)

-  Paste the printer IP in the "Epson Printer IP Address" field.
-  Check Use Local Network access option
-  Save and open your POS session


### Linux

Before running the application, you may need to make it executable and install a required dependency.
```bash
sudo apt update
sudo apt install libwebkit2gtk-4.1-0
```

## Development tests

All Go test suites, helpers, fixtures, and the runner live under
[`tests/`](tests/README.md). From the repository root, run:

```bash
# All packages
go run ./tests/run.go

# One package or selected tests
go run ./tests/run.go -v -run TestEncodePrinterID ./internal/printer

# Coverage
go run ./tests/run.go -coverprofile=coverage.out ./...
```

The runner uses a temporary Go overlay to preserve access to private package
members and supplies a frontend fixture, so backend tests do not require an npm
or Wails build. Packages run serially because several suites share TCP port 9100.
Use this command instead of plain `go test ./...`, which skips the relocated
tests. CI uses the same runner. See the [test guide](tests/README.md) for
prerequisites, package selection, and adding tests.

