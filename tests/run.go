//go:build ignore

// Run with: go run ./tests/run.go [go test arguments]
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func main() {
	if err := run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		return errors.New("cannot locate test runner source")
	}
	root := filepath.Dir(filepath.Dir(source))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		return fmt.Errorf("cannot locate repository go.mod: %w", err)
	}

	replacements := make(map[string]string)
	testsDir := filepath.Join(root, "tests", "_packages")
	err := filepath.WalkDir(testsDir, func(source string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), "_test.go") {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("test must be a regular file: %s", source)
		}
		relative, err := filepath.Rel(testsDir, source)
		if err != nil {
			return err
		}
		destination := filepath.Join(root, relative)
		if info, err := os.Stat(filepath.Dir(destination)); err != nil || !info.IsDir() {
			return fmt.Errorf("test %s must mirror an existing package directory", relative)
		}
		if _, err := os.Lstat(destination); !os.IsNotExist(err) {
			return fmt.Errorf("virtual test path must not already exist: %s", destination)
		}
		replacements[destination] = source
		// Go's package discovery already ignores the backing _packages directory.
		return nil
	})
	if err != nil {
		return err
	}
	if len(replacements) == 0 {
		return fmt.Errorf("no *_test.go files found under tests/_packages/")
	}

	// Keep temporary files out of package discovery if TMPDIR is inside the repo.
	tempDir, err := os.MkdirTemp("", ".epos-proxy-test-overlay-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	// Supply the fixture at the existing embed path. Keeping main.go unchanged
	// also lets coverage instrumentation use its original embed directive.
	fixture := filepath.Join(root, "tests", "fixtures", "frontend", "index.html")
	if info, err := os.Stat(fixture); err != nil || !info.Mode().IsRegular() {
		return fmt.Errorf("frontend test fixture must be a regular file: %s", fixture)
	}
	replacements[filepath.Join(root, "frontend", "dist", "index.html")] = fixture

	overlay, err := json.Marshal(struct {
		Replace map[string]string
	}{Replace: replacements})
	if err != nil {
		return err
	}
	overlayPath := filepath.Join(tempDir, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0600); err != nil {
		return err
	}

	// Multiple suites use TCP port 9100, so run packages serially by default.
	args := append([]string{"test", "-overlay=" + overlayPath, "-p=1"}, os.Args[1:]...)
	if len(os.Args) == 1 {
		args = append(args, "./...")
	}
	goName := "go"
	if runtime.GOOS == "windows" {
		goName += ".exe"
	}
	command := exec.Command(filepath.Join(runtime.GOROOT(), "bin", goName), args...)
	command.Dir = root
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	return command.Run()
}
