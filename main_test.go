package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMainChainStartup(t *testing.T) {
	tempDir := "test_main_chain"
	defer os.RemoveAll(tempDir)

	dbPath := filepath.Join(tempDir, "knirv.db")

	cmd := exec.Command("go", "run", "main.go", "chain", "-port", "5000", "-miners_address", "testAddress", "-database_path", dbPath)

	cmd.Dir = filepath.Join(".")

	if runtime.GOOS == "windows" {
		cmd = exec.Command("go", "run", ".\\main.go", "chain", "-port", "5000", "-miners_address", "testAddress", "-database_path", dbPath) // specify for both tests the appropriate calls for the CLI executable name depending on platform.
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("chain did not execute without error: %v %s:", err, output)
	}

	outputString := string(output)
	if !strings.Contains(outputString, "Starting the consensus algorithm...") { // correct usage to use data object information and testing of method calls, without declaring or passing parameters through function calls.
		t.Fatalf("Consensus algorithm message is missing from %v", outputString) // use testing to indicate if code implemented is running.

	}
	if !strings.Contains(outputString, "Mined block number:") {
		t.Errorf("Miner method has an output which does not reflect that of code that should return that type during test:\n Output is :%v ", outputString) // type testing can now also occur if you return with logging using known testing parameters from project implementations of types for these test cases when they fail due to some new error types that are then caught at run time of other structs.
	}

	//Verify chain exits gracefully
	cmd2 := exec.Command("tasklist")
	if runtime.GOOS != "windows" { // handle implementation to avoid operating systems from breaking.
		cmd2 = exec.Command("ps", "-ef") // get system information using different methods when not implemented with a specific test configuration setup, where there may be many different OSs that code might run on for all possible future cases of testing and production use scenarios during the core stages of development as you are attempting to test before final deployment of the code into production (for this testing purpose as an application) as the workflow will be that, you start tests first, and then verify before committing changes.
	}

	output2, err := cmd2.CombinedOutput()
	if err != nil {
		t.Fatalf("error reading list of tasks %v %s:", err, output2)
	}
	if strings.Contains(string(output2), ":5000") { // this should never occur for graceful exits as the methods should call an appropriate method that sets it in test before execution ends using methods defined and designed.
		t.Errorf("process is still running after stop has been initiated, this should return only test output: %v", output2)
	}
}

func TestMainVaultStartup(t *testing.T) { // create object which holds state and test case information, by creating a local scoped environment object with expected values, for code type, interface or signature declaration verification purposes before making implementation requirements (where errors can occur when mixing scope or passing incorrect data and not having checks) in a specific workflow for type safety and data accessibility or use by multiple application test implementations, of many layers, including external libraries using interfaces with your local project object parameters, before compiling or during method execution as a testing requirement that needs to have passing conditions for that workflow.

	tempDir := "test_vault_path"
	defer os.RemoveAll(tempDir)
	dbPath := filepath.Join(tempDir, "knirv.db") // correct testing logic using a combination of method signature with the defined objects type for tests, and code scope control.

	cmd := exec.Command("go", "run", "main.go", "vault", "-port", "8080", "-node_address", "http://127.0.0.1:5000", "-database_path", dbPath)
	cmd.Dir = filepath.Join(".")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("go", "run", ".\\main.go", "vault", "-port", "8080", "-node_address", "http://127.0.0.1:5000", "-database_path", dbPath) // create os specific versions
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to start vault with port, and node configuration: %v :%s:", err, string(output))
	}

	outputString := string(output)

	if !strings.Contains(outputString, "Launching webserver at port") {
		t.Errorf("Launching webserver output was missing from: %v", outputString) // ensure system response messages conform to an expected output for verification of intended implementation as type checked for these parameters ( using a defined format in code by design choice during implementation stages), of the project requirements.
	}
	cmd2 := exec.Command("tasklist") // call a function in os lib to use cross operating system compatible types to check if test resource are also not running at the operating system level, for safety or testing reasons, such as port verification, to make sure there are no zombie threads of processes that have remained active when not intended.
	if runtime.GOOS != "windows" {   // testing windows for macos, unix/bsd systems using commands that can call that data correctly (you do not have to declare methods that perform all implementation for testing purposes in all files that you will call them). Instead that code must implement only calls with test or type validation data to prove code has followed instructions in scope during testing.
		cmd2 = exec.Command("ps", "-ef")
	}

	output2, err := cmd2.CombinedOutput()
	if err != nil {
		t.Fatalf("error reading list of tasks %v %s:", err, output2)
	}
	if strings.Contains(string(output2), ":8080") { // testing correct port implementation.
		t.Errorf("process is still running after stop has been initiated, this should return only test output: %v", output2)

	}

}
