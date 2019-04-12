package utils_test

import (
	"bookstore/utils"
	"testing"
)

func TestExtractString(t *testing.T) {

	var pattern string = ".+(tests).+"
	var str_input string = "abcdeftestsghijk"
	var got string
	var err error
	const want string = "tests"

	got, err = utils.Extract_string(pattern, str_input)

	if len(got) == 0 {
		t.Fatal("Extracted string should not be zero sized\n")
	}

	if err != nil {
		t.Fatalf("Error should be null (got '%s'\n", err)
	}

	if got != want {
		t.Fatalf("Extracted string mismatch (want '%s', got '%s'\n", want, got)
	}
}

func TestExtractBytes(t *testing.T) {
	var got []byte
	var err error
	var pattern string = ".+(tests).+"
	var bytes_input []byte = []byte("abcdeftestsghijk")
	var want []byte = []byte("tests")

	got, err = utils.Extract_bytes(pattern, bytes_input)

	if got == nil {
		t.Fatal("Extracted bytes should not be null\n")
	}

	if err != nil {
		t.Fatalf("Error should be null (got '%s'\n", err)
	}
	if string(got) != string(want) {
		t.Fatalf("Extracted string mismatch (want '%s', got '%s'\n", string(want), string(got))
	}
}
