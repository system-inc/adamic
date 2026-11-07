package ir

import "testing"

func TestStdoutWriteDoesNotThrow(test *testing.T) {
	test.Parallel()
	call := NodeFSFile{Operation: "host_stdout_write", Of: Number}
	if call.MayThrow() || call.Type() != Number {
		test.Fatal("discarded blocking output must not become a catchable throw")
	}
}
