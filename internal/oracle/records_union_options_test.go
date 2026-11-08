package oracle

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// tsc writes CompilerOptions through its index signature (commandLineParser's options[opt.name] = value), so any value
// of the slot's union can sit under a declared key. records_union_options.ts reads declared options as Node does; the
// lying programs put a number and a null under outFile?: string, which Node prints and Adamic stops at, exit 70, with a
// message naming the read, what it found, and the write that may have stored it.
var recordUnionLying = []struct{ name, node, found string }{
	{"lying", "3\n", "number"},
	{"lying_null", "null\n", "null"},
}

func recordDeclaredMemberPanic(found string) string {
	return "adamic: panic: field read failed: options.outFile is not a string; expected string, found " +
		found +
		", which a computed-key or index-signature write may have stored\n"
}

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/records_union_options.ts", true, false})
	for _, lying := range recordUnionLying {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/records_union_" + lying.name + ".ts", true, true})
	}
}

// recordDeclaredReadStop says how a lying program failed to stop at the declared read's checked unbox on each backend
// that didn't, sanitized, release or JavaScript, or "" when every one of them stopped there.
func recordDeclaredReadStop(test *testing.T, program *ir.Program, found string) string {
	test.Helper()
	sanitized, _ := natively(test, program)
	want := run{exitCode: 70, stderr: []byte(recordDeclaredMemberPanic(found))}
	failures := []string{}
	for _, backend := range []struct {
		name string
		got  run
	}{
		{"sanitized", sanitized},
		{"release", released(test, program)},
		{"JavaScript", onJavaScriptBackend(test, program)},
	} {
		if difference := disagreement(want, backend.got); difference != "" {
			failures = append(failures, fmt.Sprintf(
				"%s: %s: exit %d, stdout %q, stderr %q",
				backend.name,
				difference,
				backend.got.exitCode,
				backend.got.stdout,
				backend.got.stderr,
			))
		}
	}
	return strings.Join(failures, "; ")
}

func TestRecordUnionDeclaredReadStops(test *testing.T) {
	test.Parallel()
	for _, lying := range recordUnionLying {
		test.Run(lying.name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/records_union_"+lying.name+".ts"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			if truth := onNode(
				test,
				path,
			); truth.exitCode != 0 || string(truth.stdout) != lying.node || len(truth.stderr) != 0 {
				test.Fatalf("Node no longer reads the lying value: %+v", truth)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			if failure := recordDeclaredReadStop(test, program, lying.found); failure != "" {
				test.Fatal(failure)
			}
		})
	}
}

// Mutant: the declared read with its checked unbox dropped, so the slot is read at the declared type unchecked. The
// stop's pin must fail on it.
func TestRecordUnionDeclaredReadMutant(test *testing.T) {
	test.Parallel()
	for _, lying := range recordUnionLying {
		test.Run(lying.name, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/records_union_"+lying.name+".ts"),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			changed := false
			for index, function := range program.Functions {
				if function.Name != "record_declared_member" {
					continue
				}
				body := []ir.Statement{}
				for _, statement := range function.Body {
					if check, ok := statement.(ir.If); ok && len(check.Then) == 1 {
						if _, ok := check.Then[0].(ir.Panic); ok {
							changed = true
							continue
						}
					}
					body = append(body, statement)
				}
				program.Functions[index].Body = body
			}
			if !changed {
				test.Fatal("mutant removed no checked unbox")
			}
			failure := recordDeclaredReadStop(test, program, lying.found)
			if failure == "" {
				test.Fatal("the unchecked declared read still stopped on every backend")
			}
			test.Logf("unchecked declared read caught by the stop's pin: %s", failure)
		})
	}
}

// Both lying programs are ruled negative witnesses (#bq3fvdt, Adamic, Oct 27): the { [name]: 3 } probe and its null
// twin, the same lie stopped at the same checked read. Each is listed in cloud/admission-corpus/negative-witnesses.json
// by its content hash with the stop the backends print, so the admission delta accepts it. Its repair, writing the
// declared key by name with a string, agrees with Node. Mutant: an entry removed from the list, recorded in
// testdata/records-union-witness-mutant.json.
func TestRecordUnionWitnessIsRuled(test *testing.T) {
	test.Parallel()
	listed, readError := os.ReadFile(filepath.Join(repository, "cloud/admission-corpus/negative-witnesses.json"))
	if readError != nil {
		test.Fatal(readError)
	}
	var list struct {
		Witnesses map[string]struct {
			Form              string `json:"form"`
			RulingTask        string `json:"ruling_task"`
			Ruling            string `json:"ruling"`
			Declaration       string `json:"declaration"`
			DeclaredType      string `json:"declared_type"`
			RepairFrom        string `json:"repair_from"`
			RepairTo          string `json:"repair_to"`
			TypeCorrectSha256 string `json:"type_correct_sha256"`
			Exit              int    `json:"exit"`
			Stderr            string `json:"stderr"`
		} `json:"witnesses"`
	}
	if decodeError := json.Unmarshal(listed, &list); decodeError != nil {
		test.Fatal(decodeError)
	}
	contentHash := func(text string) string {
		hash := sha256.Sum256([]byte(text))
		return hex.EncodeToString(hash[:])
	}
	for _, lying := range recordUnionLying {
		test.Run(lying.name, func(test *testing.T) {
			test.Parallel()
			path := "internal/oracle/testdata/records_union_" + lying.name + ".ts"
			source, readError := os.ReadFile(filepath.Join(repository, path))
			if readError != nil {
				test.Fatal(readError)
			}
			witness, found := list.Witnesses[contentHash(string(source))]
			if !found {
				test.Fatalf("%s is not a listed negative witness", path)
			}
			if witness.Form != "checked_boundary" ||
				witness.RulingTask != "#bq3fvdt" ||
				witness.Ruling != "Adamic, Oct 27" ||
				witness.Declaration != path+":1" ||
				witness.DeclaredType != "string" ||
				witness.Exit != 70 ||
				witness.Stderr != recordDeclaredMemberPanic(lying.found) {
				test.Fatalf("the listed witness doesn't name this ruling and this stop: %+v", witness)
			}
			repaired := strings.Replace(string(source), witness.RepairFrom, witness.RepairTo, 1)
			if strings.Count(string(source), witness.RepairFrom) != 1 ||
				contentHash(repaired) != witness.TypeCorrectSha256 {
				test.Fatal("the listed repair isn't one place, or its hash isn't the listed type-correct hash")
			}
			repairedPath := filepath.Join(test.TempDir(), "records_union_"+lying.name+".ts")
			if writeError := os.WriteFile(repairedPath, []byte(repaired), 0o644); writeError != nil {
				test.Fatal(writeError)
			}
			program, loweredError := lowered(test, repairedPath)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			truth := onNode(test, repairedPath)
			sanitized, _ := natively(test, program)
			for _, backend := range []struct {
				name string
				got  run
			}{
				{"sanitized", sanitized},
				{"release", released(test, program)},
				{"JavaScript", onJavaScriptBackend(test, program)},
			} {
				if truth.exitCode != 0 || disagreement(truth, backend.got) != "" {
					test.Fatalf(
						"the repair on %s: exit %d, stdout %q, stderr %q; Node exit %d, stdout %q",
						backend.name,
						backend.got.exitCode,
						backend.got.stdout,
						backend.got.stderr,
						truth.exitCode,
						truth.stdout,
					)
				}
			}
		})
	}
}
