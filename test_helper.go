package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

// TestServer represents a test server used for integration testing
type TestServer struct {
	URL         string
	Server      *httptest.Server
	TempDir     string
	Cmd         *exec.Cmd
	CleanupFunc func()
}

// StartTestServer starts a test server for integration testing
func StartTestServer(handler http.Handler) (*TestServer, error) {
	server := httptest.NewServer(handler)

	tempDir, err := os.MkdirTemp("", "knirv-test")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	testServer := &TestServer{
		URL:     server.URL,
		Server:  server,
		TempDir: tempDir,
		CleanupFunc: func() {
			server.Close()
			os.RemoveAll(tempDir)
		},
	}
	return testServer, nil
}

// StartTestNode starts a KNIRV node process for integration testing
func StartTestNode(port int, minerAddress string, remoteNode string) (*TestServer, error) {
	tempDir, err := os.MkdirTemp("", "knirv-test")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}

	cmd := exec.Command(
		"go", "run", "../knirv/main.go",
		"chain",
		"--port", fmt.Sprintf("%d", port),
		"--miners_address", minerAddress,
		"--remote_node", remoteNode,
	)

	cmd.Dir = filepath.Dir(filepath.Join("..", "knirv", "main.go"))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}
	go func() {
		_, err := io.Copy(os.Stdout, stdout)
		if err != nil {
			fmt.Println("failed to copy stdout:", err)
		}
	}()

	go func() {
		_, err := io.Copy(os.Stderr, stderr)
		if err != nil {
			fmt.Println("failed to copy stderr:", err)
		}
	}()

	testServer := &TestServer{
		TempDir: tempDir,
		Cmd:     cmd,
		CleanupFunc: func() {
			cmd.Process.Kill()
			cmd.Wait()
			os.RemoveAll(tempDir)
		},
	}

	time.Sleep(time.Second) // Give the server some time to start

	return testServer, nil
}

func (ts *TestServer) Cleanup() {
	if ts.CleanupFunc != nil {
		ts.CleanupFunc()
	}
}

func (ts *TestServer) KillProcess() error {
	if ts.Cmd == nil || ts.Cmd.Process == nil {
		return nil // Nothing to kill
	}

	if runtime.GOOS == "windows" {
		if err := ts.Cmd.Process.Kill(); err != nil {
			return fmt.Errorf("error killing the server: %w", err)
		}
	} else {
		if err := ts.Cmd.Process.Signal(syscall.SIGKILL); err != nil {
			return fmt.Errorf("error killing the server: %w", err)
		}

	}

	err := ts.Cmd.Wait()
	if err != nil && runtime.GOOS != "windows" {
		return fmt.Errorf("error waiting for server to terminate: %w", err)
	}

	return nil
}
