package main

import (
	"strings"
	"testing"
)

// A timed run is kept only when it measured more than its timer can resolve, with all three fields
// the engines print.
func TestReadingRefusesWhatMeasuredNothing(test *testing.T) {
	test.Parallel()
	if seconds, checksum, err := reading(
		"0.012000000 4096 0.000000001\n",
	); err != nil || seconds != 0.012 || checksum != "4096" {
		test.Fatalf("a real reading: %v %q %v", seconds, checksum, err)
	}
	for _, output := range []string{
		"0.000000000 4096 0.000000001",
		"0.000000001 4096 0.000000001",
		"0.000000500 4096 0.000001000",
		"0.012 4096 0",
		"NaN 4096 0.000000001",
	} {
		if _, _, err := reading(output); err == nil || !strings.HasPrefix(err.Error(), "measured nothing: ") {
			test.Fatalf("%q was kept: %v", output, err)
		}
	}
	for _, output := range []string{"0.012 4096", "0.012"} {
		if _, _, err := reading(output); err == nil {
			test.Fatalf("%q, missing a field, was kept", output)
		}
	}
}
