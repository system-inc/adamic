//go:build darwin

// Package apple holds Adamic's programs that call Apple's frameworks to account. Node can't run
// AppKit, so these answer to a witness instead (docs/apple.md): each testdata/<name>.a has a
// testdata/<name>.m that makes the same calls in Objective-C and prints the same lines, compiled by
// Apple's own toolchain, independent of everything Adamic's compiler does.
package apple

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// served is what the test's server serves, by path, for the programs that fetch: a stack, the
// supplements app's list (with a field its type doesn't name), and the list drifted from its type.
var served = map[string]string{
	"/stack.json":       `[{"identifier":"magnesium","name":"Magnesium","milligrams":200,"timing":"Evening"},{"identifier":"vitamin-d","name":"Vitamin D","milligrams":0.05,"timing":"Morning"}]`,
	"/supplements.json": `[{"name":"Creatine","dose":"5 g","symbol":"bolt.fill","timing":"Morning","brand":"dropped"},{"name":"Vitamin D","dose":"2000 IU","symbol":"sun.max.fill","timing":"Morning"},{"name":"Magnesium","dose":"400 mg","symbol":"moon.fill","timing":"Evening"}]`,
	"/drifted.json":     `[{"name":"Creatine","dose":"5 g","symbol":"bolt.fill","timing":"Noon"}]`,
}

// TestProgramsAgreeWithTheirWitnesses builds each program and its witness under the address and
// undefined-behavior sanitizers, and holds the program to the witness: the same stdout, byte for
// byte, and exit 0 for both. Then the memory both runtimes share, from a counted build: Adamic's heap
// must free every value it allocated, an action's closure included once Apple lets go of it, and the
// conversions must let go of every Objective-C reference they made.
//
// macOS's leaks tool isn't one of the checks: inside an AppKit process it reports nothing for an
// object leaked on purpose (an NSString retained twice and dropped, in Objective-C as in Adamic),
// though it finds the same leak in a program that only uses Foundation. A check that can't fail
// proves nothing, so the counts hold the line instead.
func TestProgramsAgreeWithTheirWitnesses(t *testing.T) {
	t.Parallel()
	// Every program and witness is given the address of a server of the test's own as its first
	// argument; those that don't fetch ignore it.
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		body, found := served[request.URL.Path]
		if !found {
			http.NotFound(writer, request)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(body))
	}))
	t.Cleanup(server.Close)
	address := server.URL
	programs, err := filepath.Glob(filepath.Join("testdata", "*.a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(programs) == 0 {
		t.Fatal("no programs in testdata: the test would pass without checking anything")
	}
	for _, program := range programs {
		name := strings.TrimSuffix(filepath.Base(program), ".a")
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// The witness is Objective-C, or Swift where the API is Swift's only (SwiftUI).
			witnessSource := strings.TrimSuffix(program, ".a") + ".m"
			if _, err := os.Stat(witnessSource); err != nil {
				witnessSource = strings.TrimSuffix(program, ".a") + ".swift"
			}
			if _, err := os.Stat(witnessSource); err != nil {
				t.Fatalf("%s has no witness: write %s.m or .swift, the same calls in Objective-C or Swift", program, strings.TrimSuffix(program, ".a"))
			}
			directory := t.TempDir()

			witness := filepath.Join(directory, "witness")
			if strings.HasSuffix(witnessSource, ".swift") {
				compile(t, "xcrun", "swiftc", "-O", "-sanitize=address", witnessSource, "-o", witness)
			} else {
				compile(t, "clang", "-fobjc-arc", "-fsanitize=address,undefined", "-fno-sanitize-recover=all", "-g", "-Wall", "-Werror", witnessSource, "-framework", "AppKit", "-o", witness)
			}
			expected := run(t, witness, address)

			sanitized := filepath.Join(directory, "sanitized")
			build(t, program, sanitized, native.Options{Sanitize: true})
			if observed := run(t, sanitized, address); !bytes.Equal(observed, expected) {
				t.Fatalf("%s disagrees with its witness\nwitness:\n%s\nadamic:\n%s", program, expected, observed)
			}

			counted := filepath.Join(directory, "counted")
			build(t, program, counted, native.Options{Count: true})
			allocations, frees, owed := counts(t, counted, address)
			if allocations != frees {
				t.Errorf("%s allocated %d values and freed %d", program, allocations, frees)
			}
			if owed != 0 {
				t.Errorf("%s's conversions owe %d Objective-C references they never let go of", program, owed)
			}
		})
	}
}

func build(t *testing.T, path string, output string, options native.Options) {
	t.Helper()
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	if !native.UsesApple(lowered) {
		t.Fatalf("%s makes no Apple call: it's in the wrong place", path)
	}
	if err := native.BuildApple(native.C(lowered), output, options); err != nil {
		t.Fatal(err)
	}
}

func compile(t *testing.T, command string, arguments ...string) {
	t.Helper()
	if combined, err := exec.Command(command, arguments...).CombinedOutput(); err != nil {
		t.Fatalf("%s: %v\n%s", command, err, combined)
	}
}

// deadline is how long a program may run: one waiting for a wake that never comes (an await Apple's
// hooks never resume) fails here, rather than holding the test until go test's own timeout.
const deadline = time.Minute

// run runs a binary and returns its stdout, failing on any exit but 0 or anything on stderr.
func run(t *testing.T, binary string, arguments ...string) []byte {
	t.Helper()
	var stdout, stderr bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()
	command := exec.CommandContext(ctx, binary, arguments...)
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil || stderr.Len() > 0 {
		t.Fatalf("%s: %v\nstderr:\n%s", filepath.Base(binary), err, stderr.Bytes())
	}
	return stdout.Bytes()
}

var (
	countsLine = regexp.MustCompile(`adamic: counts: allocations (\d+) frees (\d+)`)
	owedLine   = regexp.MustCompile(`adamic: apple: owed (-?\d+)`)
)

// counts runs a counted build and reads what it allocated and freed, and what Objective-C references
// its conversions still owe.
func counts(t *testing.T, binary string, arguments ...string) (int, int, int) {
	t.Helper()
	var stderr bytes.Buffer
	ctx, cancel := context.WithTimeout(context.Background(), deadline)
	defer cancel()
	command := exec.CommandContext(ctx, binary, arguments...)
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		t.Fatalf("%s: %v\n%s", filepath.Base(binary), err, stderr.Bytes())
	}
	match := countsLine.FindSubmatch(stderr.Bytes())
	if match == nil {
		t.Fatalf("a counted build wrote no counts:\n%s", stderr.Bytes())
	}
	owedMatch := owedLine.FindSubmatch(stderr.Bytes())
	if owedMatch == nil {
		t.Fatalf("a counted build wrote nothing about what its conversions owe:\n%s", stderr.Bytes())
	}
	var allocations, frees, owed int
	fmt.Sscan(string(match[1]), &allocations)
	fmt.Sscan(string(match[2]), &frees)
	fmt.Sscan(string(owedMatch[1]), &owed)
	return allocations, frees, owed
}

// leakedType is a leaked block in leaks' tree: <NSArray 0x7965044840>, or <malloc in f 0x...>.
