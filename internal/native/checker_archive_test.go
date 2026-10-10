package native

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	bridge "github.com/system-inc/adamic/bridge/tsgo"
	"github.com/system-inc/adamic/internal/buildcache"
)

// The checker archive these tests link is the tree's own, go build -buildmode=c-archive ./bridge/tsgo/archive, a
// product keyed on every file the build compiles. It never comes from outside the tree: an archive built from another
// commit, handed in by path, judged a change to the bridge's C interface against the old interface, a false red where
// the change added a function and a false green where it moved a field (#nee3cfe).
var checkerArchiveArguments = []string{"-buildmode=c-archive"}

// checkerArchiveEnvironment carries the ABI header's digest into the archive's build. The archive reaches the header
// through archive/tsgo.h's #include "../tsgo.h", outside its package, which neither go list's files nor Go's own build
// cache follow: a header changed alone would move no key and relink the object compiled from the old one. CGO_CFLAGS is
// in both the product's key and Go's action ID, so the digest moves both, and only the cgo packages compile again.
func checkerArchiveEnvironment(header []byte) []string {
	flags := os.Getenv("CGO_CFLAGS")
	if flags == "" {
		flags = "-O2 -g"
	}
	return []string{fmt.Sprintf("CGO_CFLAGS=%s -DADAMIC_TSGO_HEADER_SHA256=%x", flags, sha256.Sum256(header))}
}

// checkerArchive is the one recipe: TestProduct_CheckerArchive builds it, and every test that links the checker
// fetches it.
func checkerArchive(t testing.TB) string {
	t.Helper()
	return buildcache.GoBuild(t, "tsgo.a", "./bridge/tsgo/archive", checkerArchiveArguments, checkerArchiveEnvironment(bridge.Header)...)
}

func TestProduct_CheckerArchive(t *testing.T) {
	t.Parallel()
	checkerArchive(t)
}

// A change to the bridge's ABI header alone gives the archive a new key, so the tree that changed it links an archive
// built from it.
func TestCheckerArchiveFollowsTheBridgeHeader(t *testing.T) {
	t.Parallel()
	repository, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	key := func(header []byte) string {
		inputs, err := buildcache.GoInputs("tsgo.a", "./bridge/tsgo/archive", checkerArchiveArguments, checkerArchiveEnvironment(header))
		if err != nil {
			t.Fatal(err)
		}
		key, err := buildcache.Key(repository, inputs)
		if err != nil {
			t.Fatal(err)
		}
		return key
	}
	changed := append(append([]byte(nil), bridge.Header...), "int tsgo_another(tsgo_handle handle);\n"...)
	if key(bridge.Header) == key(changed) {
		t.Fatal("the checker archive's key didn't move when bridge/tsgo/tsgo.h did: a stale archive would be linked")
	}
}

// bridgeFunction is one function the ABI header declares, on one line, as tsgo.h writes them.
var bridgeFunction = regexp.MustCompile(`(?m)^(int|void) (tsgo_[a-z_]+)\((.*)\);$`)

// An archive whose interface isn't the tree's fails the build loudly, naming what it lacks. Two stand-in archives,
// each derived from the tree's header: one defining every function the header declares links, and one built before
// tsgo_inspect existed doesn't, and clang's error names tsgo_inspect.
func TestCheckerArchiveOfAnotherInterfaceFailsLoudly(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	if err := os.WriteFile(filepath.Join(directory, "tsgo.h"), bridge.Header, 0o644); err != nil {
		t.Fatal(err)
	}
	declared := bridgeFunction.FindAllSubmatch(bridge.Header, -1)
	if len(declared) < 2 {
		t.Fatalf("bridge/tsgo/tsgo.h declares %d functions on lines of their own", len(declared))
	}
	standIn := func(name, omitted string) string {
		var source strings.Builder
		source.WriteString("#include \"tsgo.h\"\n")
		found := false
		for _, function := range declared {
			if string(function[2]) == omitted {
				found = true
				continue
			}
			body := "{ return TSGO_OK; }"
			if string(function[1]) == "void" {
				body = "{}"
			}
			fmt.Fprintf(&source, "%s %s(%s) %s\n", function[1], function[2], function[3], body)
		}
		if omitted != "" && !found {
			t.Fatalf("bridge/tsgo/tsgo.h no longer declares %s", omitted)
		}
		path := filepath.Join(directory, name+".c")
		if err := os.WriteFile(path, []byte(source.String()), 0o644); err != nil {
			t.Fatal(err)
		}
		object := filepath.Join(directory, name+".o")
		if output, err := exec.Command("clang", "-std=c11", "-c", "-I", directory, path, "-o", object).CombinedOutput(); err != nil {
			t.Fatalf("the stand-in archive %s: %v\n%s", name, err, output)
		}
		return object
	}
	source := "#include \"tsgo_runtime.h\"\nint main(void) { return 0; }\n"
	if err := BuildTSGo(source, filepath.Join(directory, "whole"), standIn("whole", ""), Options{}); err != nil {
		t.Fatalf("an archive with the tree's whole interface didn't link: %v", err)
	}
	err := BuildTSGo(source, filepath.Join(directory, "older"), standIn("older", "tsgo_inspect"), Options{})
	if err == nil {
		t.Fatal("an archive without tsgo_inspect linked: a mismatched archive would pass silently")
	}
	if !strings.Contains(err.Error(), "tsgo_inspect") {
		t.Fatalf("an archive without tsgo_inspect failed without naming it: %v", err)
	}
}
