package common

import (
	"testing"
)

func TestParseKey(t *testing.T) {
	type TestCase struct {
		input      string
		expected   []string
		shouldFail bool
	}

	testCases := []TestCase{
		{input: "", expected: nil, shouldFail: true},
		{input: "@asdf", expected: nil, shouldFail: true},
		{input: "a@b@c", expected: nil, shouldFail: true},
		{input: "key@db", expected: []string{"key", "db"}, shouldFail: false},
		{input: "key", expected: []string{"key", DefaultDBName}, shouldFail: false},
	}

	var errMsg string
	exec := func(input string) (string, string) {
		errMsg = ""
		defer func() {
			if err := recover(); err != nil {
				errMsg = err.(string)
			}
		}()

		return ParseKey(input)
	}

	for _, testCase := range testCases {
		key, db := exec(testCase.input)

		if testCase.shouldFail {
			if errMsg == "" {
				t.Errorf("Expected input %q to fail, but it didn't", testCase.input)
			}
			continue
		}

		if !testCase.shouldFail && errMsg != "" {
			t.Errorf("Input %q failed with %q", testCase.input, errMsg)
			continue
		}

		expectedKey, expectedDB := testCase.expected[0], testCase.expected[1]
		if key != expectedKey || db != expectedDB {
			t.Errorf("Expected (%q, %q), got (%q, %q)", expectedKey, expectedDB, key, db)
		}
	}
}
