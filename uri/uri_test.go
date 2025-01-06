package uri

import "testing"

func TestParseURI(t *testing.T) {
	testCases := []struct {
		uri           string
		expectedRes   string
		expectedID    string
		expectedError bool
	}{
		{
			uri:           "knirv://test.nrn",
			expectedRes:   "nrn",
			expectedID:    "test",
			expectedError: false,
		},
		{
			uri:           "knirv://test.keychain",
			expectedRes:   "keychain",
			expectedID:    "test",
			expectedError: false,
		},
		{
			uri:           "invalidURI",
			expectedError: true,
		},
		{
			uri:           "knirv://invalid",
			expectedError: true,
		},
		{
			uri:           "knirv://invalid.invalid.nrn",
			expectedError: true,
		},
	}
	for _, tc := range testCases {
		res, id, err := ParseURI(tc.uri)
		if tc.expectedError && err == nil {
			t.Fatalf("expected an error for: %s, but did not get one", tc.uri)
		}
		if !tc.expectedError && err != nil {
			t.Fatalf("did not expect error but got one for: %s, %v", tc.uri, err)
		}
		if res != tc.expectedRes {
			t.Fatalf("resource does not match for uri: %s, expected %s, got %s", tc.uri, tc.expectedRes, res)
		}
		if id != tc.expectedID {
			t.Fatalf("id does not match for uri: %s, expected %s, got %s", tc.uri, tc.expectedID, id)
		}
	}
}
