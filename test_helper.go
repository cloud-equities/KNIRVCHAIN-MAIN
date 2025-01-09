package main

import (
	"fmt"
	"net"
	"net/http/httptest"
	"os/exec"
	"strconv"
	"time"
)

// StartTestNode starts the blockchain node as a subprocess for testing
func StartTestNode(port int, minerAddress string, remoteNode string) (*exec.Cmd, error) {
	cmd := exec.Command("go", "run", "main.go", "chain",
		"--port", strconv.Itoa(port),
		"--miners_address", minerAddress,
	)
	if remoteNode != "" {
		cmd.Args = append(cmd.Args, "--remote_node", remoteNode)
	}

	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start test node: %w", err)
	}

	// Wait until the server is ready
	for i := 0; i < 10; i++ {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), time.Second)
		if err == nil {
			conn.Close()
			return cmd, nil
		}
		time.Sleep(time.Second)
	}
	return nil, fmt.Errorf("test node failed to start within timeout")
}

func TearDownTestEnv(server *httptest.Server) {
	server.Close()
}
