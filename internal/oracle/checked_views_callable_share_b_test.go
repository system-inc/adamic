package oracle

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

type callableShareBFamily struct {
	Rank       int
	Field      string
	Read       string
	Directory  string
	ResultType string
	Expected   string
	Arity      string
	Variants   []string
	NodeErrors []string
}

func callableShareBFamilies(t *testing.T) []callableShareBFamily {
	t.Helper()
	filename := "families.json"
	if selected := os.Getenv("ADAMIC_CALLABLE_SHARE_B_FAMILIES"); selected != "" {
		filename = filepath.Base(selected)
	}
	data, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-b", filename))
	if err != nil {
		t.Fatal(err)
	}
	var families []callableShareBFamily
	if err := json.Unmarshal(data, &families); err != nil {
		t.Fatal(err)
	}
	return families
}

func callableShareBContract(t *testing.T, program *ir.Program, field string) ir.ViewContractID {
	t.Helper()
	for _, contract := range program.ViewContracts {
		for _, member := range contract.Fields {
			if member.Name == field && member.Contract != 0 && program.ViewContracts[member.Contract-1].Kind == ir.ViewCallable {
				return member.Contract
			}
		}
	}
	t.Fatal("missing original callable descriptor: " + field)
	return 0
}

func TestCheckedViewCallableShareBFamilies(t *testing.T) {
	for _, family := range callableShareBFamilies(t) {
		for _, variant := range family.Variants {
			t.Run(family.Directory+"/"+variant, func(t *testing.T) {
				program, path := interfaceFixture(t, "lane5/share-b/"+family.Directory+"/"+variant)
				truth := onNode(t, path)
				nodeError := variant == "wrong-value"
				for _, expectedError := range family.NodeErrors {
					nodeError = nodeError || variant == expectedError
				}
				if nodeError {
					if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError:") {
						t.Fatalf("Node: %#v", truth)
					}
				} else if truth.exitCode != 0 {
					t.Fatalf("Node: %#v", truth)
				}
				contract := program.ViewContracts[callableShareBContract(t, program, family.Field)-1]
				if os.Getenv("ADAMIC_CALLABLE_SHARE_B_MUTANT") == "arity" {
					callableShareBArityMutant(t, program, family)
				}
				if contract.Name != family.Expected {
					t.Fatalf("original signature: got %q want %q", contract.Name, family.Expected)
				}
				sanitized, binary := nativelyUncached(t, program)
				for backend, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
					if variant == "good" {
						if difference := disagreement(truth, got); difference != "" {
							t.Fatal(difference)
						}
						continue
					}
					found := "number"
					switch variant {
					case "wrong-arity":
						found = "function with arity " + family.Arity
					case "wrong-result":
						found = "function with incompatible result representation"
					case "wrong-members":
						found = "function with incompatible parameter representations"
					}
					expected := "adamic: panic: field read failed: " + family.Read + " expected " + family.Expected + ", found " + found + "\n"
					if variant == "wrong-value" && backend < 2 {
						expected = "adamic: panic: field read failed: " + family.Read + " is not a " + family.Expected + "; expected " + family.Expected + ", found number\n"
					}
					if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != expected {
						t.Fatalf("backend %d: exit %d stdout %q stderr %q want %q", backend, got.exitCode, got.stdout, got.stderr, expected)
					}
				}
				if variant == "good" {
					if report := leaksUncached(t, program, binary); report != "" {
						t.Fatal(report)
					}
				}
			})
		}
	}
}

// Each mutant falsifies only this read's arity obligation in lowered IR. The
// original source and immutable producer signature stay unchanged. A successful
// wrong execution must fail the original exit-70/message pin, with valid C.
func TestCheckedViewCallableShareBArityMutants(t *testing.T) {
	for _, family := range callableShareBFamilies(t) {
		t.Run(family.Directory, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/share-b/"+family.Directory+"/wrong-arity")
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node: %#v", truth)
			}
			callableShareBArityMutant(t, program, family)
			for backend, got := range []run{releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatalf("mutant must execute valid code, backend %d: %s", backend, difference)
				}
				t.Logf("arity mutant caught by exit-70/message pin: backend %d ran exit %d stdout %q", backend, got.exitCode, got.stdout)
			}
		})
	}
}

// Not parallel: this scoped updater appends only this share's measured rows.
func TestCheckedViewCallableShareBCounts(t *testing.T) {
	var rows []string
	for _, family := range callableShareBFamilies(t) {
		for _, variant := range family.Variants {
			path := filepath.Join("stage3/interface-downcasts/lane5/share-b", family.Directory, variant+".a")
			rows = append(rows, counted(t, path, false, nil, false, false))
		}
	}
	// Receiver-gap fixtures cannot produce counts until their native conversion errors are fixed.
	for _, filename := range []string{"batch-03-diagnostic-probes.json", "batch-05-diagnostic-probes.json"} {
		contents, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-b", filename))
		if err != nil {
			t.Fatal(err)
		}
		var probes []struct{ Filename string }
		if err := json.Unmarshal(contents, &probes); err != nil {
			t.Fatal(err)
		}
		for _, probe := range probes {
			rows = append(rows, counted(t, filepath.Join("stage3/interface-downcasts/lane5/share-b", probe.Filename), false, nil, false, false))
		}
	}
	rows = append(rows, counted(t, "stage3/interface-downcasts/lane5/share-b/rank-169/node-byte-view.a", false, nil, false, false))
	path := filepath.Join(repository, "internal/oracle/counts.md")
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if *updateCounts {
		text := string(contents)
		for _, row := range rows {
			key := strings.Split(row, " | ")[0] + " | "
			if !strings.Contains(text, key) {
				text += row + "\n"
			}
		}
		if err := os.WriteFile(path, []byte(text), 0644); err != nil {
			t.Fatal(err)
		}
		return
	}
	for _, row := range rows {
		if !strings.Contains(string(contents), row+"\n") {
			t.Errorf("unrecorded share b counts: %s", row)
		}
	}
}

func TestCheckedViewCallableShareBBlockers(t *testing.T) {
	for _, probe := range []struct{ path, output, refusal string }{
		{"stage3/interface-downcasts/lane5/gaps/debug-assert-contract.a", "function\n", "adamic/no-type-predicate"},
		{"stage3/interface-downcasts/lane5/share-b/blocked-assignment.a", "7\n", "checked view read of field createAssignment with unsupported callable contract"},
		{"stage3/interface-downcasts/lane5/share-b/blocked-push.a", "2\n", "checked view read of field push with unsupported callable contract"},
		{"stage3/interface-downcasts/lane5/share-b/blocked-join.a", "first|second\n", "native-array-receiver"},
		{"stage3/interface-downcasts/lane5/share-b/rank-205/good.a", "3\n", "optional chain longer than one step"},
		{"stage3/interface-downcasts/lane5/share-b/rank-448/good.a", "3\n3\n3\n3\n3\n3\n3\n3\n", "truncated-callable-name"},
	} {
		t.Run(filepath.Base(probe.path), func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, probe.path))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != probe.output {
				t.Fatalf("Node: %#v", truth)
			}
			program, err := lowered(t, path)
			if probe.refusal == "truncated-callable-name" {
				if err != nil {
					t.Fatal(err)
				}
				contract := program.ViewContracts[callableShareBContract(t, program, "updateArrowFunction")-1]
				expected := "(node: ArrowFunction, modifiers: readonly Modifier[] | undefined, typeParameters: readonly TypeParameterDeclaration[] | undefined, parameters: ..., type: TypeNode | undefined, equalsGreaterThanToken: EqualsGreaterThanToken, body: ConciseBody) => ArrowFunction"
				if contract.Name != expected {
					t.Fatalf("signature diagnostic frontier: %q", contract.Name)
				}
				sanitized, binary := nativelyUncached(t, program)
				for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
					if difference := disagreement(truth, got); difference != "" {
						t.Fatal(difference)
					}
				}
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
				t.Logf("observed signature diagnostic truncation: %s", contract.Name)
				return
			}
			if probe.refusal == "native-array-receiver" {
				if err != nil {
					t.Fatal(err)
				}
				buildErr := native.Build(native.C(program), filepath.Join(t.TempDir(), "join"), native.Options{Sanitize: true})
				if buildErr == nil || !strings.Contains(buildErr.Error(), "incompatible pointer types passing 'adamic_array *'") || !strings.Contains(buildErr.Error(), "adamic_object *") {
					t.Fatalf("array receiver frontier: %v", buildErr)
				}
				got := onJavaScriptBackend(t, program)
				expected := "adamic: panic: field read failed: result.join expected (separator?: string | undefined) => string, found function with unknown signature\n"
				if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != expected {
					t.Fatalf("intrinsic signature frontier: %#v", got)
				}
				t.Logf("JavaScript array receiver observation: exit %d stdout %q stderr %q", got.exitCode, got.stdout, got.stderr)
				t.Logf("native array receiver observation: %v", buildErr)
				return
			}
			if err != nil && !strings.Contains(err.Error(), probe.refusal) {
				t.Fatalf("wrong refusal: %v", err)
			}
			if err == nil {
				t.Fatal("candidate now lowers; requires full certification")
			}
			t.Logf("observed original callable blocker: %v", err)
		})
	}
}

func callableShareBArityMutant(t *testing.T, program *ir.Program, family callableShareBFamily) {
	t.Helper()
	id := callableShareBContract(t, program, family.Field)
	if family.Arity == "0" {
		program.ViewContracts[id-1].Parameters = nil
	} else {
		program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Number, Name: "number"})
		program.ViewContracts[id-1].Parameters = []ir.ViewContractID{ir.ViewContractID(len(program.ViewContracts))}
	}
}

// Not parallel: ordered admission evidence is collected for this share's ledger.
func TestCheckedViewCallableShareBAdmissionProbes(t *testing.T) {
	var data []byte
	for _, filename := range []string{"batch-02-probes.json", "batch-03-debug-probes.json", "batch-03-signature-probes.json", "batch-04-tracing-probes.json", "batch-04-fs-probes.json", "batch-05-cast-probes.json", "batch-05-signature-probes.json"} {
		contents, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-b", filename))
		if err != nil {
			t.Fatal(err)
		}
		if len(data) == 0 {
			data = contents
		} else {
			data = append(append(data[:len(data)-2], ','), contents[1:]...)
		}
	}
	var probes []struct {
		Rank     int
		Filename string
		Refusal  string
		Output   string
	}
	if err := json.Unmarshal(data, &probes); err != nil {
		t.Fatal(err)
	}
	for _, probe := range probes {
		t.Run(filepath.Dir(probe.Filename), func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-b", probe.Filename))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if probe.Output != "" {
				if truth.exitCode != 0 || string(truth.stdout) != probe.Output {
					t.Fatalf("Node: %#v", truth)
				}
			} else if truth.exitCode != 70 || !strings.Contains(string(truth.stderr), "TypeError:") {
				t.Fatalf("Node: %#v", truth)
			}
			_, err = lowered(t, path)
			if probe.Refusal != "" && (err == nil || !strings.Contains(err.Error(), probe.Refusal)) {
				t.Fatalf("expected refusal %q, got %v", probe.Refusal, err)
			}
			t.Logf("rank %d admission observation: %v", probe.Rank, err)
		})
	}
}

// Not parallel: ordered original-receiver gap observations share their ledger.
func TestCheckedViewCallableShareBCollectionReceivers(t *testing.T) {
	for _, group := range []struct{ Filename, NativeType string }{{"batch-03-map-probes.json", "map"}, {"batch-03-array-probes.json", "array"}, {"batch-04-map-probes.json", "map"}, {"batch-05-map-probes.json", "map"}, {"batch-05-array-probes.json", "array"}} {
		data, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-b", group.Filename))
		if err != nil {
			t.Fatal(err)
		}
		var probes []struct {
			Rank                                       int
			Field, Filename, Output, Refusal, Expected string
		}
		if err := json.Unmarshal(data, &probes); err != nil {
			t.Fatal(err)
		}
		for _, probe := range probes {
			t.Run(filepath.Dir(probe.Filename), func(t *testing.T) {
				path, err := filepath.Abs(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-b", probe.Filename))
				if err != nil {
					t.Fatal(err)
				}
				truth := onNode(t, path)
				if truth.exitCode != 0 || string(truth.stdout) != probe.Output {
					t.Fatalf("Node: %#v", truth)
				}
				program, err := lowered(t, path)
				if err != nil {
					if probe.Refusal == "" {
						t.Fatalf("unexpected collection lowering refusal: %v", err)
					}
					if !strings.Contains(err.Error(), probe.Refusal) {
						t.Fatal(err)
					}
					t.Logf("rank %d lowering observation: %v", probe.Rank, err)
					return
				}
				if probe.Refusal != "" {
					t.Fatal("receiver now lowers; requires full certification")
				}
				buildErr := native.Build(native.C(program), filepath.Join(t.TempDir(), "map"), native.Options{Sanitize: true})
				if buildErr == nil || !strings.Contains(buildErr.Error(), "incompatible pointer types passing 'adamic_"+group.NativeType+" *'") || !strings.Contains(buildErr.Error(), "adamic_object *") {
					t.Fatalf("Map receiver frontier: %v", buildErr)
				}
				got := onJavaScriptBackend(t, program)
				if got.exitCode != 70 || len(got.stdout) != 0 || !strings.Contains(string(got.stderr), "function with unknown signature") {
					t.Fatalf("JavaScript intrinsic frontier: %#v", got)
				}
				if probe.Expected != "" && string(got.stderr) != probe.Expected {
					t.Fatalf("expected %q, got %q", probe.Expected, got.stderr)
				}
				t.Logf("rank %d native observation: %v", probe.Rank, buildErr)
				t.Logf("rank %d JavaScript observation: %q", probe.Rank, got.stderr)
			})
		}
	}
}

func TestCheckedViewCallableShareBByteViewModule(t *testing.T) {
	t.Parallel()
	program, path := interfaceFixture(t, "lane5/share-b/rank-169/node-byte-view")
	truth := onNode(t, path)
	sanitized, binary := nativelyUncached(t, program)
	for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatal(difference)
		}
	}
	if report := leaksUncached(t, program, binary); report != "" {
		t.Fatal(report)
	}
}

func TestCheckedViewCallableShareBDiagnosticNames(t *testing.T) {
	contents, err := os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane5/share-b/batch-05-diagnostic-probes.json"))
	if err != nil {
		t.Fatal(err)
	}
	var probes []struct{ Field, Directory, Expected, ObservedExpected string }
	if err := json.Unmarshal(contents, &probes); err != nil {
		t.Fatal(err)
	}
	for _, probe := range probes {
		t.Run(probe.Directory, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane5/share-b/"+probe.Directory+"/good")
			truth := onNode(t, path)
			if truth.exitCode != 0 {
				t.Fatalf("Node: %#v", truth)
			}
			contract := program.ViewContracts[callableShareBContract(t, program, probe.Field)-1]
			if contract.Name != probe.ObservedExpected || contract.Name == probe.Expected {
				t.Fatalf("callable diagnostic: %q", contract.Name)
			}
			sanitized, binary := nativelyUncached(t, program)
			for _, got := range []run{releasedUncached(t, program), sanitized, onJavaScriptBackend(t, program)} {
				if difference := disagreement(truth, got); difference != "" {
					t.Fatal(difference)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Logf("original type: %s; observed type: %s", probe.Expected, contract.Name)
		})
	}
}
