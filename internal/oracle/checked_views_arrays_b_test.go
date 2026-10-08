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

type arraysBProbe struct {
	CodeRequired bool   `json:"codeRequired"`
	Refusal      string `json:"refusal"`
	Nested       bool   `json:"nested"`
	originalArrayProbe
	Pair          string `json:"pair"`
	Mode          string `json:"mode"`
	StringElement bool   `json:"stringElement"`
	Leaf          string `json:"leaf"`
}

func arraysBInputs(t *testing.T) (string, intersectionOriginalManifest, []arraysBProbe) {
	t.Helper()
	declarations := os.Getenv("ADAMIC_ARRAYB_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("originalB/prepare.cjs declaration inputs required")
	}
	var manifest intersectionOriginalManifest
	data, err := os.ReadFile(filepath.Join(declarations, "arrayB-manifest.json"))
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
		if pair.Reads != 1 || len(pair.Sites) != 1 {
			t.Fatal("original one-read witness drift")
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
	var probes []arraysBProbe
	data, err = os.ReadFile(filepath.Join(repository, "stage3/interface-downcasts/lane2/originalB/probes.json"))
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
			t.Fatal("duplicate originalB fixture: " + probe.Name)
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
	files, err := filepath.Glob(filepath.Join(repository, "stage3/interface-downcasts/lane2/originalB/*.a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != len(probes) {
		t.Fatal("originalB fixtures are missing from the oracle manifest")
	}
	return declarations, manifest, probes
}

func arraysBFile(t *testing.T, declarations string, probe arraysBProbe) string {
	t.Helper()
	file := originalArrayFile(t, declarations, filepath.Join(repository, "stage3/interface-downcasts/lane2/originalB"), probe.Name)
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
func arraysBComplete(t *testing.T, program *ir.Program, manifest intersectionOriginalManifest, probe arraysBProbe) {
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

func TestCheckedViewArraysBOriginal(t *testing.T) {
	declarations, manifest, probes := arraysBInputs(t)
	for _, probe := range probes {
		t.Run(probe.Name, func(t *testing.T) {
			t.Parallel()
			file := arraysBFile(t, declarations, probe)
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
			arraysBComplete(t, program, manifest, probe)
			want := node
			if probe.Diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.Diagnostic + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if diff := disagreement(want, got); diff != "" {
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
func TestCheckedViewArraysBMutants(t *testing.T) {
	declarations, manifest, probes := arraysBInputs(t)
	for _, probe := range probes {
		if probe.Diagnostic == "" || probe.Refusal != "" || probe.CodeRequired {
			continue
		}
		t.Run(probe.Name, func(t *testing.T) {
			t.Parallel()
			file := arraysBFile(t, declarations, probe)
			program, err := lowered(t, file)
			if err != nil {
				t.Fatal(err)
			}
			arraysBComplete(t, program, manifest, probe)
			stringID := ir.ViewContractID(len(program.ViewContracts) + 1)
			program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.String, Name: "string"})
			numberID := ir.ViewContractID(len(program.ViewContracts) + 1)
			program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewScalar, Of: ir.Number, Name: "number"})
			unionID := ir.ViewContractID(len(program.ViewContracts) + 1)
			program.ViewContracts = append(program.ViewContracts, ir.ViewContract{Kind: ir.ViewUnion, Of: ir.Union, Name: "string | number", Members: []ir.ViewContractID{stringID, numberID}})
			changed := 0
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
				} else if conversion, ok := n.(ir.NumberToString); ok {
					if read, ok := conversion.Value.(ir.Property); ok && read.View == "items[0]!."+probe.Leaf {
						read.Of, read.ViewContract, read.ViewType, read.ViewAllowed = ir.String, stringID, "string", nil
						changed++
						return read
					}
				} else if read, ok := n.(ir.Property); ok && read.View == "items[0]!."+probe.Leaf && read.Of == ir.String {
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

func arraysBCounts(t *testing.T) []string {
	t.Helper()
	const prefix = "stage3/interface-downcasts/lane2/originalB/"
	if os.Getenv("ADAMIC_ARRAYB_ORIGINAL_DECLS") == "" {
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
		t.Log("originalB counts retained without remeasurement; original declaration inputs required")
		return rows
	}
	declarations, _, probes := arraysBInputs(t)
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	rows := []string{}
	for _, probe := range probes {
		if probe.Refusal != "" {
			continue
		}
		file := arraysBFile(t, declarations, probe)
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
func TestCheckedViewArraysBCounts(t *testing.T) {
	if os.Getenv("ADAMIC_ARRAYB_ORIGINAL_DECLS") == "" {
		t.Skip("original declaration inputs required for measured counts")
	}
	rows := arraysBCounts(t)
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !*updateCounts {
		for _, row := range rows {
			if !strings.Contains(string(data), row+"\n") {
				t.Errorf("unrecorded originalB counts: %s", row)
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
