package main

import (
	"bytes"
	"os"
	"testing"
)

func TestCommandPrintsFixture(t *testing.T) {
	var output, errors bytes.Buffer
	path := "../../internal/split/testdata/decisions.a"
	if code := run([]string{path}, &output, &errors); code != 0 {
		t.Fatalf("exit %d: %s", code, errors.String())
	}
	expected, err := os.ReadFile("../../internal/split/testdata/decisions.txt")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output.Bytes(), expected) {
		t.Fatalf("output:\n%s", output.String())
	}
	if errors.Len() != 0 {
		t.Fatal(errors.String())
	}
}
func TestCommandRejectsMissingFile(t *testing.T) {
	var output, errors bytes.Buffer
	if code := run([]string{"missing.a"}, &output, &errors); code != 1 || errors.Len() == 0 || output.Len() != 0 {
		t.Fatalf("exit %d, stdout %q, stderr %q", code, output.String(), errors.String())
	}
}
