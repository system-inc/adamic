package oracle

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedViewCallableFactory(t *testing.T) {
	for _, pair := range []struct {
		rank                        int
		field, conversion, expected string
	}{
		{3, "createIdentifier", "9\n110\n110\n", "{ (text: string): Identifier; (text: string, originalKeywordKind?: number | undefined, hasExtendedUnicodeEscape?: boolean | undefined): Identifier; }"},
		{9, "createStringLiteral", "3\n110\n110\n", "{ (text: string, isSingleQuote?: boolean | undefined): StringLiteral; (text: string, isSingleQuote?: boolean | undefined, hasExtendedUnicodeEscape?: boolean | undefined): StringLiteral; }"},
		{24, "createUniqueName", "11\n100\n103\n", "{ (text: string, flags?: number | undefined): Identifier; (text: string, flags?: number | undefined, prefix?: string | GeneratedNamePart | undefined, suffix?: string | undefined): Identifier; }"},
	} {
		t.Run(pair.field, func(t *testing.T) {
			variants := []string{"good", "conversion", "wrong-overload"}
			if pair.rank == 3 {
				variants = append(variants, "wrong-result")
			}
			for _, variant := range variants {
				t.Run(variant, func(t *testing.T) {
					program, path := interfaceFixture(t, fmt.Sprintf("lane5/share-factory/rank-%d/%s", pair.rank, variant))
					truth := onNode(t, path)
					want := "7\n"
					if variant == "conversion" {
						want = pair.conversion
					}
					if variant == "wrong-result" {
						want = "read\n"
					}
					if truth.exitCode != 0 || string(truth.stdout) != want || len(truth.stderr) != 0 {
						t.Fatalf("source Node: %#v, want %q", truth, want)
					}
					sanitized, binary := nativelyUncached(t, program)
					if variant == "wrong-result" {
						t.Logf("sanitized result exit=%d stdout=%q stderr=%q", sanitized.exitCode, sanitized.stdout, sanitized.stderr)
					}
					for index, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
						if variant == "good" || variant == "conversion" {
							if difference := disagreement(truth, got); difference != "" {
								t.Fatalf("backend%d: %s", index, difference)
							}
						} else {
							t.Logf("backend%d negative exit=%d stdout=%q stderr=%q", index, got.exitCode, got.stdout, got.stderr)
							ordinal := 2
							if variant == "wrong-result" {
								ordinal = 1
							}
							message := fmt.Sprintf("adamic: panic: field read failed: factory.%s expected %s, found function with incompatible overload signature %d\n", pair.field, pair.expected, ordinal)
							if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != message {
								t.Fatalf("negative did not stop: %#v", got)
							}
						}
					}
					if variant == "good" || variant == "conversion" {
						if report := leaksUncached(t, program, binary); report != "" {
							t.Fatal(report)
						}
						if variant == "conversion" {
							callableFactoryUndefinedMutant(t, program, truth)
						}
						return
					}
					if variant == "wrong-result" {
						return
					}
					changed := 0
					for index := range program.ViewContracts {
						contract := &program.ViewContracts[index]
						if contract.Kind == ir.ViewCallable && len(contract.Members) == 2 {
							contract.Members = contract.Members[:1]
							changed++
						}
					}
					if changed != 1 {
						t.Fatalf("last-overload mutant changed %d sets", changed)
					}
					mutantSanitized, mutantBinary := nativelyUncached(t, program)
					for index, got := range []run{releasedUncached(t, program), mutantSanitized, onJavaScriptBackend(t, program)} {
						if difference := disagreement(truth, got); difference != "" {
							t.Fatalf("mutant%d must execute cleanly: %s", index, difference)
						}
						t.Logf("backend%d last-overload omission caught: mutant exit 0 stdout %q, negative requires exit 70", index, got.stdout)
					}
					if report := leaksUncached(t, program, mutantBinary); report != "" {
						t.Fatal(report)
					}
				})
			}
		})
	}
}

// Dropping only the resolved signature's undefined packing must change valid
// native output. The producer metadata and overload-set checks remain intact.
func callableFactoryUndefinedMutant(t *testing.T, program *ir.Program, truth run) {
	t.Helper()
	changed := 0
	rewrite := func(value any) any {
		if call, ok := value.(ir.CallClosure); ok {
			for index, argument := range call.Arguments {
				if maybe, ok := argument.(ir.MaybeOf); ok && maybe.Value == nil {
					call.Arguments[index] = ir.Undefined{}
					changed++
				}
			}
			return call
		}
		return value
	}
	program.Main = rewriteUnionTargetStatements(program.Main, rewrite)
	for index := range program.Functions {
		program.Functions[index].Body = rewriteUnionTargetStatements(program.Functions[index].Body, rewrite)
	}
	if changed == 0 {
		t.Fatal("undefined conversion mutant did not reach a call")
	}
	sanitized, binary := nativelyUncached(t, program)
	for index, got := range []run{releasedUncached(t, program), sanitized} {
		if got.exitCode != 0 || len(got.stderr) != 0 || string(got.stdout) == string(truth.stdout) {
			t.Fatalf("conversion mutant%d did not fail by clean output disagreement: %#v", index, got)
		}
		t.Logf("native%d resolved undefined conversion mutant caught: stdout %q, Node requires %q", index, got.stdout, truth.stdout)
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

// Not parallel: -update-counts appends this unit's measured fixture rows.
func TestCheckedViewCallableFactoryCounts(t *testing.T) {
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	fixtures, err := filepath.Glob(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-factory/rank-*/*.a"))
	if err != nil || len(fixtures) != 10 {
		t.Fatalf("factory fixture inventory: %d, %v", len(fixtures), err)
	}
	for _, fixture := range fixtures {
		relative, err := filepath.Rel(repository, fixture)
		if err != nil {
			t.Fatal(err)
		}
		row := counted(t, checkedViewFixturePath(relative), false, nil, false, false)
		if strings.Contains(text, row+"\n") {
			continue
		}
		key := strings.Split(row, " | ")[0] + " | "
		if strings.Contains(text, key) {
			t.Fatalf("existing factory count changed: %s", row)
		}
		if !*updateCounts {
			t.Errorf("unrecorded factory count: %s", row)
			continue
		}
		text += row + "\n"
	}
	if *updateCounts {
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCheckedViewCallableFactoryDeclarations(t *testing.T) {
	root := filepath.Join(repository, "stage3/interface-downcasts/lane5/share-factory")
	data, err := os.ReadFile(filepath.Join(root, "certificates.json"))
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Pairs, Reads int
		Members      []struct {
			Rank, Reads  int
			Field, Read  string
			Declarations []string
		}
	}
	if err := json.Unmarshal(data, &evidence); err != nil {
		t.Fatal(err)
	}
	reads := 0
	for _, member := range evidence.Members {
		reads += member.Reads
		original, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-a", fmt.Sprintf("gap-%d/good.a", member.Rank)))
		if err != nil {
			t.Fatal(err)
		}
		variants := []string{"good", "conversion", "wrong-overload"}
		if member.Rank == 3 {
			variants = append(variants, "wrong-result")
		}
		for _, variant := range variants {
			fixture, err := os.ReadFile(filepath.Join(root, fmt.Sprintf("rank-%d/%s.a", member.Rank, variant)))
			if err != nil {
				t.Fatal(err)
			}
			if variant == "good" && string(fixture) != string(original) {
				t.Fatalf("rank %d original control changed", member.Rank)
			}
			for _, declaration := range member.Declarations {
				if !strings.Contains(string(fixture), declaration) {
					t.Fatalf("rank %d %s lost original declaration %q", member.Rank, variant, declaration)
				}
			}
			if !strings.Contains(string(fixture), member.Read) {
				t.Fatalf("rank %d %s lost original read", member.Rank, variant)
			}
		}
	}
	if evidence.Pairs != 3 || len(evidence.Members) != 3 || evidence.Reads != 368 || reads != 368 {
		t.Fatalf("certificate inventory changed: %#v", evidence)
	}
}
