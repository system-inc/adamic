package main

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"strings"
	"testing"
)

// straceTableSum is testdata/strace-6.8-x86_64.txt's sha256: strace 6.8's x86_64 call table, read from its source at
// tag v6.8. Loom's ebpftrace/strace-6.8-x86_64.txt holds the same bytes under the same sum and expands %file from it.
const straceTableSum = "b6f73c983f46fe06594d5aa9558bbd9311881df47cc4bfc22f35c83d5762d46d"

// Every call Settle parses is one straceArguments asks for, %file as strace 6.8 expands it on x86_64. Once strace is
// gone the list is what Loom's BPF tracer records, so a call Settle reads and nobody asks for is a read no trace holds
// (#9tr6c6j). A name strace's x86_64 table doesn't hold (stat64, lstat64, fstatat64) no x86_64 trace holds either.
func TestSettleParsesOnlyAskedCalls(t *testing.T) {
	t.Parallel()
	table, err := os.ReadFile("testdata/strace-6.8-x86_64.txt")
	if err != nil {
		t.Fatal(err)
	}
	if sum := sha256.Sum256(table); hex.EncodeToString(sum[:]) != straceTableSum {
		t.Fatalf("testdata/strace-6.8-x86_64.txt's sha256 is %x, pinned %s: loom's copy changes with it", sum, straceTableSum)
	}
	known, file := map[string]bool{}, map[string]bool{}
	for _, line := range strings.Split(string(table), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || strings.HasPrefix(line, "#") {
			continue
		}
		known[fields[1]] = true
		file[fields[1]] = len(fields) > 2 && fields[2] == "file"
	}
	list, found := "", false
	for _, argument := range straceArguments {
		if value, ok := strings.CutPrefix(argument, "trace="); ok {
			list, found = value, true
		}
	}
	if !found {
		t.Fatalf("straceArguments %q ask for no trace= list", straceArguments)
	}
	asked := map[string]bool{}
	for _, item := range strings.Split(list, ",") {
		if item == "%file" {
			for name, holds := range file {
				asked[name] = asked[name] || holds
			}
			continue
		}
		asked[item] = true
	}
	// Settle's pinned calls, which its own test holds its code to (internal/buildcache's TestSettleParsesExactlyThePinnedCalls).
	pinned, err := os.ReadFile("../../testdata/settle-calls.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(pinned), "\n") {
		name := strings.TrimSpace(line)
		if name == "" || strings.HasPrefix(name, "#") || !known[name] {
			continue
		}
		if !asked[name] {
			t.Errorf("Settle parses %s, and trace=%s doesn't ask strace for it: a read in it is in no trace", name, list)
		}
	}
}
