package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A read or write in its temporal dead zone throws Node's ReferenceError natively, named
// ReferenceError with "Cannot access 'name' before initialization" as its message alone, so a try
// catches it and e.name and e.message read as Node's (#bxmdk73, Adamic's ruling on #d2p19w9).
// String(error) of an Error doesn't lower on main yet, so each prints the two it's made of. The throw is counted
// wherever the analyses ask what can throw (ir.Program's DeadZoneMayThrow, read by lower's throwsOut and flow's
// CanThrow), so a caller tests for it through a function that can't throw on its own (read_global.a's twice), and an
// async function's rejection lands in its awaiter's try. Each agrees with Node natively, sanitized and on JavaScript.
// Mutants: adamic_panic back in the emitter's deadZone, and every program goes red natively; the class
// prefix put back in the message, and each e.message reads doubled; DeadZoneMayThrow false for a
// read, and read_global.a's twice no longer tests the word.
func TestDeadZoneThrowsAreCaughtAsNodes(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/dead_zone_caught"
	entries, readDirectoryError := os.ReadDir(filepath.Join(repository, directory))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".a") {
			continue
		}
		programs++
		test.Run(entry.Name(), func(test *testing.T) {
			test.Parallel()
			loomViewsDAgrees(test, directory+"/"+entry.Name())
		})
	}
	if programs != 6 {
		test.Fatalf(
			"%d programs in %s, want 6: a program added or removed needs this count changed with it",
			programs,
			directory,
		)
	}
}

// A caught dead-zone throw in the middle of an expression lets go of everything the expression
// already held: built strings, a call's argument, a template's part. A panic never had to, so this is
// the one test that a throw site's cleanup (the inline release or the statement's ladder,
// internal/native/exceptions.go) frees on the way out. It agrees with Node in
// TestDeadZoneThrowsAreCaughtAsNodes; here it must also leak nothing. Mutants: a ladder rung that
// skips its release, and an inline site that skips its owned list; each leaks the strings it held.
func TestDeadZoneOwnedTemporaryLetsGo(test *testing.T) {
	test.Parallel()
	program, loweredError := lowered(test, repository+"/internal/oracle/testdata/dead_zone_caught/owned_temporary.a")
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	_, binary := natively(test, program)
	if report := leaks(test, program, binary); report != "" {
		test.Fatalf("a caught dead-zone throw leaked what its expression held: %s", report)
	}
}
