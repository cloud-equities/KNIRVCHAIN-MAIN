package install

import (
	"os"
	"path/filepath"
	"testing"
	"utils"
)

func TestConfigureChainPath(t *testing.T) {
	// Create a mock input reader to simulate user input
	stdin := os.Stdin
	defer func() { os.Stdin = stdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r
	w.Write([]byte("\n"))
	w.Close()

	path, err := ConfigureChainPath()
	if err != nil {
		t.Fatalf("failed to ask for chain path: %v", err)
	}

	expectedPath := filepath.Join(utils.UserHomeDir(), ".knirv")

	if path != expectedPath {
		t.Errorf("got path %s, want %s", path, expectedPath)
	}
}

func TestDeployChain(t *testing.T) {
	testPath := "test_data_dir/test_chain"

	err := os.MkdirAll(testPath, os.ModeDir|0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	ownerAddress := "some_owner_address"
	port := "5000"
	err = DeployChain(testPath, ownerAddress, port)
	if err != nil {
		t.Fatalf("failed to create new chain: %v", err)
	}

	// add some checks here
}

func TestVerifyChainDeployment(t *testing.T) {
	testPath := "test_data_dir/test_chain"
	err := os.MkdirAll(testPath, os.ModeDir|0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	err = VerifyChainDeployment(testPath)
	if err != nil {
		t.Fatalf("failed to create new chain: %v", err)
	}
}

func TestInstallProcess(t *testing.T) {
	testPath := "test_data_dir/test_install_process/test_knirv"

	err := os.MkdirAll(testPath, os.ModeDir|0755)
	if err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	stdin := os.Stdin
	defer func() { os.Stdin = stdin }()
	r, w, _ := os.Pipe()
	os.Stdin = r

	// Write the input string followed by a newline character
	_, err = w.Write([]byte(testPath + "\n"))
	if err != nil {
		t.Fatalf("Failed to write to mock stdin: %v", err)
	}
	w.Close()

	err = InstallProcess()
	if err != nil {
		t.Fatalf("install process failed: %v", err)
	}

	// add some checks to see if database files exist
}
