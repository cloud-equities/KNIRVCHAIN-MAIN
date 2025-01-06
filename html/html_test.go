package html

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// use for creating common setup variables for re-use by other test functions
type dataForTest struct {
	chainAddress  string
	walletAddress string
	message       string
	ownerAddress  string
}

var mytest dataForTest

func setup() { // function used to correctly get a default set of parameters without hardcoded value.
	mytest.chainAddress = "http://localhost:5000"         // a url string from test parameters of method parameters, all variables will always be of consistent type where types for parameters is defined.
	mytest.walletAddress = "someFakeAddress"              // testing with a constant.
	mytest.ownerAddress = "testOwner"                     // verifying a state where testing the owner property that it is used with consistent values.
	mytest.message = "this message should be in the dom." // text is a known string value, to compare values to after implementing.
}
func TestHomeTemplate(t *testing.T) {

	w := httptest.NewRecorder()
	setup() // get all global type struct initializations with base set of parameters
	params := DashboardParams{
		ChainAddress:  mytest.chainAddress,  // access using your struct types where possible for safety and test correctness.
		WalletAddress: mytest.walletAddress, //  use defined struct
		Owner:         mytest.ownerAddress,  //  test data can now be assigned using local vars to help show how data can also be handled by test data parameters that are consistent in design.
		Message:       mytest.message,
	}
	err := Dashboard(w, params)
	if err != nil {
		t.Fatalf("Expected Home template to render without an error, but got error: %v", err)
	}
	if w.Code != http.StatusOK {
		t.Fatalf("StatusCode is not 200: but is '%d'", w.Code)
	}

	output := w.Body.String()

	if !strings.Contains(output, params.Message) { // testing if data can also get printed.
		t.Errorf("String message output is incorrect:\n got:%s \n want: contains '%v'", output, params.Message)
	}
	if !strings.Contains(output, params.ChainAddress) {
		t.Errorf("chain Address output is incorrect:\n got:%s \n want: contains '%v'", output, params.ChainAddress) // all required method properties and requirements must all be testable for expected type and method calls in tests as is needed for all future iterations.

	}
	if !strings.Contains(output, params.WalletAddress) {
		t.Errorf("Wallet Address output is incorrect:\n got:%s \n want: contains '%v'", output, params.WalletAddress)
	}
	if !strings.Contains(output, params.Owner) {
		t.Errorf("owner Address output is incorrect:\n got:%s \n want: contains '%v'", output, params.Owner)
	}

	byte_output := []byte(output)
	json_text := extractTextBetweenTags(byte_output, "pre") // test code must implement expected responses that are created with actual expected code structures that you have created with proper use of those objects that represent type specific data sets.
	if json_text == "" {
		t.Fatalf("test requires valid <pre> to view parsed data, could not read it:\n Output: %v\n Expected Data inside <pre>", output)
	}
	var parsedJSON interface{}

	err = json.Unmarshal([]byte(json_text), &parsedJSON)
	if err != nil {
		t.Fatalf("response from HTML response output is an invalid Json Struct object:  %v : %v,  text:", string(byte_output), err) //  Use logging, with known variable formats in tests to help isolate why some implementations and variables that may have not being handled correctly may still cause a fail condition when tests do not produce output correctly during test executions.

	}
	_, ok := parsedJSON.(map[string]interface{}) // proper checking and verification to assert that you will return the appropriate method type.
	if ok == false {
		t.Fatalf("type of data passed is an incorrect struct %v:  should be an map with string values, using method call, to pass struct types to rendering engine in next section", output)
	}

	_, err = template.New("html test").Parse(output) // correct output parsing of data for correct formatting with test outputs.

	if err != nil {
		t.Fatalf("failed to create an implementation of template %v ", err)
	}
}

func TestDashboardTemplate(t *testing.T) {
	w := httptest.NewRecorder() // you should be able to follow this specific pattern throughout other files.

	setup() // set this up to have access to parameters for all html method calls when creating new structs that may depend on test conditions for test implementations of code modules and component, methods.

	params := DashboardParams{ // verify output has a format and implements parameters correctly.
		ChainAddress:  mytest.chainAddress,  // access data for method via this method signature of parameters from test object instead of calling test type directly which leads to potential logical errors.
		WalletAddress: mytest.walletAddress, // check that these are always as we expect
		Message:       mytest.message,       //check the string input value.
		Owner:         mytest.ownerAddress,  // all of the tests now should follow a consistent pattern and make use of proper variable typing instead of assuming a data types value in all locations, where this information can be better represented through methods as has been requested.
	}
	err := Dashboard(w, params)
	if err != nil {
		t.Fatalf("Expected Dashboard template to render without an error, but got error %v", err) // verification step
	}
}

func TestProfileTemplate(t *testing.T) {

	w := httptest.NewRecorder() // check for this in all new and other components, where you intend to test a component, that must function under a very controlled type validation logic as defined by our framework here, with new patterns that you must implement to make the best version of this project using testable and known method signatures
	setup()                     // new logic for reusable variables in test framework for html test components to use.
	params := ProfileParams{    // struct has defined values from type implementation, make a variable that implements that type, where these test and helper functions may need to have local variables which implement struct to hold data before passing on.
		ChainAddress:  mytest.chainAddress,
		WalletAddress: mytest.walletAddress,
		Owner:         mytest.ownerAddress,
		Message:       mytest.message,
	}
	err := Profile(w, params) // calling method with struct
	if err != nil {
		t.Fatalf("Expected Profile template to render without an error, but got error: %v", err) // should return error when tests do not conform
	}
}

func TestSettingsTemplate(t *testing.T) {
	w := httptest.NewRecorder() // add helper function in each method.
	setup()
	params := SettingsParams{ // Test cases using method signatures as structs, for data values
		ChainAddress:  mytest.chainAddress,
		WalletAddress: mytest.walletAddress,
		Owner:         mytest.ownerAddress,
		Message:       mytest.message,
	}
	err := Settings(w, params) // test that this test structure works, using different parameter for rendering implementation logic.
	if err != nil {
		t.Fatalf("Expected Settings template to render without an error, but got error %v", err) // confirm logic returns an error here so if issues ever arise with formatting these tests will be failing instead of producing an error.
	}
}

// helper method for getting value from HTML body using the HTML node's text value to retrieve a json response for string matching, since it can easily serialize an array struct to JSON when generating html content dynamically based on object properties. This type must also match the type in code to properly test your types.
func extractTextBetweenTags(html []byte, tag string) string {
	openTag := fmt.Sprintf("<%s>", tag)
	closeTag := fmt.Sprintf("</%s>", tag)

	startIndex := bytes.Index(html, []byte(openTag))
	if startIndex == -1 {
		return ""
	}

	startIndex += len(openTag)
	endIndex := bytes.Index(html[startIndex:], []byte(closeTag))

	if endIndex == -1 {
		return ""
	}

	endIndex += startIndex

	return string(html[startIndex:endIndex])
}
