package main

import (
	"chain"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"KNIRVCHAIN/constants"
	knirvlog "KNIRVCHAIN/log"
)

func init() {
	log.SetPrefix(constants.BLOCKCHAIN_NAME + ":") // initialize the program.
}

func main() {
	chainCmdSet := flag.NewFlagSet("chain", flag.ExitOnError)
	lockboxCmdSet := flag.NewFlagSet("vault", flag.ExitOnError)
	chainPort := chainCmdSet.Uint64("port", 5000, "HTTP port to launch our blockchain server")
	chainMiner := chainCmdSet.String("miners_address", "", "Miners address to credit mining reward")
	remoteNode := chainCmdSet.String("remote_node", "", "Remote Node from where the blockchain will be synced")
	dbPath := chainCmdSet.String("database_path", filepath.Join(".", "knirv.db"), "Filepath for saving chain's database") // Use struct variable and implement using new structs during method calling in core app for specific types.

	vaultPort := lockboxCmdSet.Uint64("port", 8080, "HTTP port to launch our vault server") // you can implement this as a new structure in our object when that is relevant with specific implementations to other data structures or methods with struct access and parameters with valid interfaces that tests should verify with all data for scope and workflow parameters implementation requirements.

	blockchainNodeAddress := lockboxCmdSet.String("node_address", "http://127.0.0.1:5000", "Blockchain node address for the vault gateway") // this is still static url you might make dynamic later by a parameter for testing data flow.

	if len(os.Args) < 2 {
		fmt.Println("Error:Expected chain or vault subcommand")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "chain":
		chainCmdSet.Parse(os.Args[2:])
		if chainCmdSet.Parsed() {
			if *chainMiner == "" || chainCmdSet.NFlag() == 0 {
				fmt.Println("Usage of chain subcommand: ")
				chainCmdSet.PrintDefaults()
				os.Exit(1)
			}
			db, err := chain.NewDBClient(*dbPath)
			if err != nil {
				knirvlog.FatalError("failed to create LevelDB: ", err) // if errors use this object's methods for proper type logging during resource access issues that may occur during system runtime.
				os.Exit(1)

			}
			defer db.Close()                                                    // close open db, on exit of this section since scope is local only.
			chainAddress := "http://127.0.0.1:" + strconv.Itoa(int(*chainPort)) // method is now type safe

			if *remoteNode == "" { // create genesis block from current source parameters only
				genesisBlock := chain.Block{}
				blockchain1 := chain.NewBlockchain(&genesisBlock, chainAddress, db) // pass database information as expected. Use constructor for required and correct objects rather than using system or implicit object declarations as parameters which may have wrong parameters from testing environment as its logic or data workflow parameters were defined by core scope implementation of other resource files during code review.
				blockchain1.Peers[blockchain1.ChainAddress] = true

				bcs := chain.NewBlockchainServer(*chainPort, blockchain1, *&chainAddress) // use struct object for type safe operation
				go bcs.Start()                                                            // run in separate go routines since a synchronous implementation that includes other long tasks in methods should never exist for this program test flow or test object type verification procedures as those calls should be explicit for thread safety in an interface object workflow instead.
				go bcs.BlockchainPtr.ProofOfWorkMining(*chainMiner)                       // call only local functions and pass by interface objects not the entire struct by name.
				go bcs.BlockchainPtr.DialAndUpdatePeers()
				go bcs.BlockchainPtr.RunConsensus() // run logic
				time.Sleep(2 * time.Second)
			} else {

				blockchain1, err := chain.SyncBlockchain(*remoteNode, db, chainAddress)
				if err != nil {
					knirvlog.LogError("unable to initialize from remote blockchain ", err)
					os.Exit(1)
				}
				blockchain2, err := chain.NewBlockchainFromSync(blockchain1, chainAddress, db, chain.BlockchainOptions{
					WalletAddress: *chainMiner, // only set what parameters we require for scope requirements or testing scenarios, other method parameters for more modular data type usage, which may be implemented during runtime object loading by those test suite specific implementations or other application object instantiations or interface declarations as needed (but if type or logic parameters for those tests are incorrect in their execution methods an object's method validation fails or when that test code type does not correspond with function return that code method will cause compile issues and help surface them at that build state using our consistent process steps to perform more accurate type safety test in all our method design choice requirements during project implementation steps).

				})
				if err != nil { // make sure to correctly pass through information for test to use parameters and method type requirements as intended and are available when you expect these to occur in tests, as testing parameters.
					knirvlog.LogError("Unable to initialize new sync object:", err) // added error messaging with new logging scope for these variables during method initialization.
					os.Exit(1)                                                      // test failures will be handled by implementation types so this method can only create and use type methods. If methods for specific type struct data are to create errors, all calls can still check them, which has been tested in many other methods and structs previously, but more data will also improve that tests.
				}
				var wg sync.WaitGroup

				blockchain2.Peers[blockchain2.ChainAddress] = true                        // use method variables only
				bcs := chain.NewBlockchainServer(*chainPort, blockchain2, *&chainAddress) // correctly defined data type scope to help test framework components are now tested better
				wg.Add(4)

				go bcs.Start() // create resource using a thread. you now expect a parameter of a struct. ( use struct value where they are available rather than relying on values outside their required method calls scope ).
				go bcs.BlockchainPtr.ProofOfWorkMining(*chainMiner)
				go bcs.BlockchainPtr.DialAndUpdatePeers() // chain resource management logic operations.
				go bcs.BlockchainPtr.RunConsensus()       // Implement resource and object interaction by struct. This also promotes proper usage of methods by using proper object types that adhere to test parameters during logic test workflow by separating out resource logic during object and test workflow steps by method scope types for data operations.

				wg.Wait()
			}
		}

	case "lockbox":
		lockboxCmdSet.Parse(os.Args[2:])
		if lockboxCmdSet.Parsed() {
			if lockboxCmdSet.NFlag() == 0 {
				fmt.Println("Usage of LockBox subcommand: ") // only pass local defined string values that will only produce information if requirements were not met, based on that command. to keep console as clean and consistent as it possibly can be at runtime, instead implement error and tests where those conditions fail. to help track the workflow data chain across method signatures.
				lockboxCmdSet.PrintDefaults()
				os.Exit(1) // if type not verified for http or url path it can now exit at the correct logical place where it did not meet parameter checks of test cases.
			}

			vs := vault.CreateNewVault(uint16(*vaultPort), *blockchainNodeAddress) // all calls for type safe operation from constructor to objects parameters
			vs.Start()                                                             // Start will block thread where other tests can check code correctness
		}

	default:
		fmt.Println("Error:Expected chain or wallet subcommand") // check method input type for known types in system implementation.

		os.Exit(1) // Exit is called only at core parts of application to force thread close, and to report errors from missing core data when program logic scope implementation methods fails when the types of declared variables are unexpected at the current method.
	}

}
