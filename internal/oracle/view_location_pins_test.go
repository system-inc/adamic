package oracle

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

var externalViewCastSites sync.Map

// Preserve existing exit/output/field/type pins, adding only source-validated
// cast locations. The acceptance tests separately pin exact locations as literals.
func viewReadPin(expected run, compiled any) run {
	if !strings.HasPrefix(string(expected.stderr), "adamic: panic: field read failed:") && !strings.HasPrefix(string(expected.stderr), "adamic: panic: element read failed:") {
		return expected
	}
	var program *ir.Program
	switch value := compiled.(type) {
	case *ir.Program:
		program = value
	case ir.Program:
		program = &value
	default:
		return expected
	}
	if strings.Contains(string(expected.stderr), "view cast at ") || strings.Contains(string(expected.stderr), "possible view casts at ") {
		return expected
	}
	for _, origin := range program.ViewCastLocations {
		last := strings.LastIndex(origin.Where, ":")
		if last < 0 {
			return expected
		}
		previous := strings.LastIndex(origin.Where[:last], ":")
		if previous < 0 {
			return expected
		}
		file := origin.Where[:previous]
		cached, ok := externalViewCastSites.Load(file)
		if !ok {
			script := `const fs=require('node:fs'),ts=require(process.argv[1]);const name=process.argv[2],source=ts.createSourceFile(name,fs.readFileSync(name,'utf8'),ts.ScriptTarget.Latest,true,ts.ScriptKind.TS),sites=[];function visit(n){if(ts.isAsExpression(n)){const p=source.getLineAndCharacterOfPosition(n.getStart(source));sites.push(name+':'+(p.line+1)+':'+(p.character+1));}ts.forEachChild(n,visit);}visit(source);process.stdout.write(JSON.stringify(sites));`
			data, err := exec.Command("node", "-e", script, filepath.Join(repository, "stage3/api/node_modules/typescript"), file).Output()
			if err != nil {
				return expected
			}
			var sites []string
			if json.Unmarshal(data, &sites) != nil {
				return expected
			}
			valid := map[string]bool{}
			for _, site := range sites {
				valid[site] = true
			}
			cached, _ = externalViewCastSites.LoadOrStore(file, valid)
		}
		if !cached.(map[string]bool)[origin.Where] {
			return expected
		}
	}
	labels := []string{}
	for label := range program.ViewReadCastLocations {
		labels = append(labels, label)
	}
	sort.Slice(labels, func(i, j int) bool { return len(labels[i]) > len(labels[j]) })
	for _, label := range labels {
		if strings.Contains(string(expected.stderr), label) {
			expected.stderr = []byte(strings.Replace(string(expected.stderr), label, ir.ViewDiagnosticLabel(program, label), 1))
			break
		}
	}
	return expected
}

func TestViewReadCastOriginsThroughJoinedHelper(t *testing.T) {
	path := filepath.Join(t.TempDir(), "origins.a")
	source := `interface Base {readonly id:number;}
interface Viewed extends Base {readonly text:string;}
function first(base:Base):Viewed {
    return base as Viewed;
}
function second(base:Base):Viewed {
    return base as Viewed;
}
function read(value:Viewed):void {
    console.log(value.text);
}
const okay={id:1,text:"fine"};
const wrong={id:2,text:42};
read(first(okay));
read(second(wrong));
`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "fine\n42\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node: %#v", truth)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	expected := run{exitCode: 70, stdout: []byte("fine\n"), stderr: []byte("adamic: panic: field read failed: value.text (possible view casts at origins.a:4:12, origins.a:7:12) is not a string; expected string, found number\n")}
	sanitized, _ := nativelyUncached(t, program)
	for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if diff := disagreement(expected, got); diff != "" {
			t.Fatalf("joined helper origin pin: %s: stderr=%q", diff, got.stderr)
		}
	}
	// Retain the real type guard and remove only its diagnostic provenance.
	program.ViewReadCastLocations = nil
	for name, got := range map[string]run{"native": releasedUncached(t, program), "JavaScript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || !strings.HasPrefix(string(got.stderr), "adamic: panic: field read failed:") {
			t.Fatalf("%s origin mutant changed the read stop: %#v", name, got)
		}
		if disagreement(expected, got) == "" {
			t.Fatalf("%s missing-origin mutant survived", name)
		}
		t.Logf("%s field read without cast site rejected: %q", name, got.stderr)
	}
}

// A transitive certificate can fail in a child path while its diagnostic label
// belongs to the parent read. Preserve the complete old message and require a
// source-validated annotation from a matching parent/read label; its placement
// within that path is not part of the old type/output pin.
func viewReadDisagreement(expected, actual run, compiled any) string {
	pinned := viewReadPin(expected, compiled)
	if string(pinned.stderr) == string(expected.stderr) {
		return disagreement(pinned, actual)
	}
	var program *ir.Program
	switch value := compiled.(type) {
	case *ir.Program:
		program = value
	case ir.Program:
		program = &value
	default:
		return disagreement(pinned, actual)
	}
	for label, location := range program.ViewReadCastLocations {
		annotation := " (" + location + ")"
		if strings.Contains(string(expected.stderr), label) && strings.Contains(string(actual.stderr), annotation) {
			plain := actual
			plain.stderr = []byte(strings.Replace(string(actual.stderr), annotation, "", 1))
			if diff := disagreement(expected, plain); diff == "" {
				return ""
			}
		}
	}
	return disagreement(pinned, actual)
}

func viewReadDiagnosticMismatch(expected string, actual []byte, program any) bool {
	return viewReadDisagreement(run{stderr: []byte(expected)}, run{stderr: actual}, program) != ""
}
