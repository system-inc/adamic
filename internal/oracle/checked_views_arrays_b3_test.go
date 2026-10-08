package oracle

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

type arraysB3Probe struct {
	JavaScriptDiagnostic string `json:"javascriptDiagnostic"`
	Callable             bool   `json:"callable"`
	CodeRequired         bool   `json:"codeRequired"`
	Refusal              string `json:"refusal"`
	Nested               bool   `json:"nested"`
	originalArrayProbe
	Pair          string `json:"pair"`
	Mode          string `json:"mode"`
	StringElement bool   `json:"stringElement"`
	Leaf          string `json:"leaf"`
}

func arraysB3Inputs(t *testing.T) (string, intersectionOriginalManifest, []arraysB3Probe) {
	t.Helper()
	declarations := os.Getenv("ADAMIC_ARRAYB3_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("originalB3/prepare.cjs declaration inputs required")
	}
	var manifest intersectionOriginalManifest
	data, err := os.ReadFile(filepath.Join(declarations, "arrayB3-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Commit != "050880ce59e30b356b686bd3144efe24f875ebc8" || len(manifest.Pairs) != 20 || len(manifest.Declarations) != 79 {
		t.Fatal("original provenance drift")
	}
	for _, pair := range manifest.Pairs {
		if (pair.Reads != 1 && pair.Reads != 2) || len(pair.Sites) != pair.Reads {
			t.Fatal("original ranked witness drift")
		}
	}
	for name, digest := range manifest.Declarations {
		data, err := os.ReadFile(filepath.Join(declarations, name))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
			t.Fatal("original declaration drift: " + name)
		}
	}
	var probes []arraysB3Probe
	data, err = os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane2/originalB3/probes.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &probes); err != nil {
		t.Fatal(err)
	}
	seenNames := map[string]bool{}
	seenModes := map[string]map[string]bool{}
	for _, probe := range probes {
		if seenNames[probe.Name] {
			t.Fatal("duplicate originalB3 fixture: " + probe.Name)
		}
		seenNames[probe.Name] = true
		if seenModes[probe.Pair] == nil {
			seenModes[probe.Pair] = map[string]bool{}
		}
		seenModes[probe.Pair][probe.Mode] = true
	}
	for _, pair := range manifest.Pairs {
		key := pair.Type + "." + pair.Field
		for _, mode := range []string{"good", "bad", "lazy", "lazy-element", "alias"} {
			if !seenModes[key][mode] {
				t.Fatal("missing original pair witness: " + key + "/" + mode)
			}
		}
	}
	files, err := filepath.Glob(filepath.Join(repository, "stage3/interface-downcasts/lane2/originalB3/*.a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != len(probes) {
		t.Fatal("originalB3 fixtures are missing from the oracle manifest")
	}
	return declarations, manifest, probes
}

func arraysB3File(t *testing.T, declarations string, probe arraysB3Probe) string {
	t.Helper()
	file := originalArrayFile(t, declarations, filepath.Join(repository, "stage3/interface-downcasts/lane2/originalB3"), probe.Name)
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.ReplaceAll(string(data), "'original-tsc-private'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "private-watch.d.ts"))))
	source = strings.ReplaceAll(source, "'original-tsc-builder'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/builder.d.ts"))))
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return file
}

// Original field sets must remain complete, including unsupported unread members.
func arraysB3Complete(t *testing.T, program *ir.Program, manifest intersectionOriginalManifest, probe arraysB3Probe) {
	t.Helper()
	for _, name := range probe.Contracts {
		complete := false
		for _, contract := range program.ViewContracts {
			if contract.Name != name {
				continue
			}
			fields := []string{}
			for _, field := range contract.Fields {
				fields = append(fields, field.Name)
			}
			slices.Sort(fields)
			complete = complete || slices.Equal(fields, manifest.Fields[name])
		}
		if !complete {
			t.Fatal("original field set reduced: " + name)
		}
	}
}

func TestCheckedViewArraysB3Original(t *testing.T) {
	declarations, manifest, probes := arraysB3Inputs(t)
	for _, probe := range probes {
		t.Run(probe.Name, func(t *testing.T) {
			t.Parallel()
			file := arraysB3File(t, declarations, probe)
			node := run{stdout: []byte(probe.Source)}
			if diff := disagreement(node, onNode(t, file)); diff != "" {
				t.Fatal("Node: " + diff)
			}
			program, err := lowered(t, file)
			if probe.Refusal != "" {
				if err == nil || !strings.HasSuffix(err.Error(), probe.Refusal) {
					t.Fatalf("expected pinned refusal %q, got %v", probe.Refusal, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			arraysB3Complete(t, program, manifest, probe)
			want := node
			if probe.Diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.Diagnostic + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			for i, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				expected := want
				if i == 2 && probe.JavaScriptDiagnostic != "" {
					expected.stderr = []byte("adamic: panic: " + probe.JavaScriptDiagnostic + "\n")
				}
				if diff := disagreement(expected, got); diff != "" {
					t.Errorf("%s; stdout %q stderr %q exit %d", diff, got.stdout, got.stderr, got.exitCode)
				}
			}
			if want.exitCode == 0 {
				if report := leaksUncached(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
		})
	}
}

// Each stopping witness gets an independent read mutation. Mutants must finish
// with the expected output, so neither invalid C nor sanitizers can count.
// The nested-index default prints zero; the other mutants match source Node.
func TestCheckedViewArraysB3Mutants(t *testing.T) {
	declarations, manifest, probes := arraysB3Inputs(t)
	for _, probe := range probes {
		if probe.Diagnostic == "" || probe.Refusal != "" || probe.CodeRequired {
			continue
		}
		t.Run(probe.Name, func(t *testing.T) {
			t.Parallel()
			file := arraysB3File(t, declarations, probe)
			program, err := lowered(t, file)
			if err != nil {
				t.Fatal(err)
			}
			arraysB3Complete(t, program, manifest, probe)
			stringID := ir.ViewContractID(len(program.ViewContracts) + 1)
			program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.String, Name: "string"})
			numberID := ir.ViewContractID(len(program.ViewContracts) + 1)
			program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Number, Name: "number"})
			unionID := ir.ViewContractID(len(program.ViewContracts) + 1)
			program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewUnion, Of: ir.Union, Name: "string | number", Members: []ir.ViewContractID{stringID, numberID}})
			changed := 0
			label := "items[0]!"
			if probe.Callable {
				label += "()"
			}
			label += "." + probe.Leaf
			rewrite := func(n any) any {
				if probe.Mode == "absent" {
					if read, ok := n.(ir.Property); ok && strings.HasPrefix(read.View, "viewed") {
						read.Absent, read.Optional = true, true
						changed++
						return read
					}
					return n
				}
				if probe.Nested {
					if read, ok := n.(ir.ArrayIndex); ok && read.View == "items[0]![0]" {
						changed++
						return ir.MaybeOf{Value: ir.NumberConstant{Value: 0}, Of: ir.MaybeNumber}
					}
					return n
				}
				if probe.StringElement {
					if read, ok := n.(ir.ArrayJoin); ok {
						read.Element, read.Stringify = ir.Union, false
						read.ViewRead.Element, read.ViewRead.ViewContract = ir.Union, unionID
						changed++
						return read
					}
				} else if conversion, ok := n.(ir.BooleanToString); ok {
					if read, ok := conversion.Value.(ir.Property); ok && read.View == label {
						read.Of, read.ViewContract, read.ViewType, read.ViewAllowed = ir.String, stringID, "string", nil
						changed++
						return read
					}
				} else if conversion, ok := n.(ir.NumberToString); ok {
					if read, ok := conversion.Value.(ir.Property); ok && read.View == label {
						read.Of, read.ViewContract, read.ViewType, read.ViewAllowed = ir.String, stringID, "string", nil
						changed++
						return read
					}
				} else if read, ok := n.(ir.Property); ok && read.View == label && read.Of == ir.String {
					read.Of, read.ViewContract, read.ViewType, read.ViewAllowed = ir.Number, numberID, "number", nil
					changed++
					return ir.NumberToString{Value: read}
				}
				return n
			}
			program.Main = rewriteUnionTargetStatements(program.Main, rewrite)
			for i := range program.Functions {
				program.Functions[i].Body = rewriteUnionTargetStatements(program.Functions[i].Body, rewrite)
			}
			if changed != 1 {
				t.Fatalf("mutant changed %d reads, expected one", changed)
			}
			want := run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.Diagnostic + "\n")}
			node := run{stdout: []byte(probe.Source)}
			if probe.Nested && probe.Mode == "bad" {
				node.stdout = []byte("0\n")
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if disagreement(want, got) == "" {
					t.Fatal("mutant escaped stopping oracle")
				}
				if diff := disagreement(node, got); diff != "" {
					t.Errorf("mutant execution differs: %s; stdout %q stderr %q exit %d", diff, got.stdout, got.stderr, got.exitCode)
				}
			}
			if report := leaksUncached(t, program, binary); report != "" {
				t.Fatal(report)
			}
			t.Log("independent mutant caught by pinned stopping oracle in sanitized native, release native and JavaScript")
		})
	}
}

func arraysB3Counts(t *testing.T) []string {
	t.Helper()
	const prefix = "stage3/interface-downcasts/lane2/originalB3/"
	if os.Getenv("ADAMIC_ARRAYB3_ORIGINAL_DECLS") == "" {
		data, err := os.ReadFile(filepath.Join(repository, "internal/oracle/counts.md"))
		if err != nil {
			t.Fatal(err)
		}
		rows := []string{}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "| "+prefix) {
				rows = append(rows, line)
			}
		}
		t.Log("originalB3 counts retained without remeasurement; original declaration inputs required")
		return rows
	}
	declarations, _, probes := arraysB3Inputs(t)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	rows := []string{}
	for _, probe := range probes {
		if probe.Refusal != "" {
			continue
		}
		file := arraysB3File(t, declarations, probe)
		relative, err := filepath.Rel(root, file)
		if err != nil {
			t.Fatal(err)
		}
		row := counted(t, relative, false, nil, false, false)
		rows = append(rows, strings.Replace(row, relative, prefix+probe.Name+".a", 1))
	}
	return rows
}

// Not parallel: refreshing writes only this worker's measured rows.
func TestCheckedViewArraysB3Counts(t *testing.T) {
	if os.Getenv("ADAMIC_ARRAYB3_ORIGINAL_DECLS") == "" {
		t.Skip("original declaration inputs required for measured counts")
	}
	rows := arraysB3Counts(t)
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !*updateCounts {
		for _, row := range rows {
			if !strings.Contains(string(data), row+"\n") {
				t.Errorf("unrecorded originalB3 counts: %s", row)
			}
		}
		return
	}
	parts := strings.SplitN(string(data), "\n## Predicate direction counts", 2)
	lines := strings.Split(strings.TrimSuffix(parts[0], "\n"), "\n")
	for _, row := range rows {
		key := strings.Split(row, " | ")[0] + " | "
		found := false
		for i, line := range lines {
			if strings.HasPrefix(line, key) {
				lines[i], found = row, true
				break
			}
		}
		if !found {
			lines = append(lines, row)
		}
	}
	updated := strings.Join(lines, "\n") + "\n"
	if len(parts) == 2 {
		updated += "\n## Predicate direction counts" + parts[1]
	}
	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		t.Fatal(err)
	}
}
