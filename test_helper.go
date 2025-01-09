package main

import (
	"fmt"
	"net/http/httptest"
	"os/exec"
	"strconv"
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
	return cmd, nil
}

func TearDownTestEnv(server *httptest.Server) {
	server.Close()
}
