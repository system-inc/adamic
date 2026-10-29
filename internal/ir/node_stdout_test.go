package ir

import "testing"

func TestStdoutWriteDoesNotThrow(t *testing.T) {
	t.Parallel()
	call := NodeFSFile{Operation: "host_stdout_write", Of: Number}
	if call.MayThrow() || call.Type() != Number {
		t.Fatal("discarded blocking output must not become a catchable throw")
	}
}
