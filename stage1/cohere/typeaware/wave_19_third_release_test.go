package typeaware

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestWave19ThirdReleasedHandles(t *testing.T) {
	repository, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	directory := os.Getenv("ADAMIC_WAVE19_THIRD_RELEASE_ARTIFACTS")
	if directory == "" {
		directory = t.TempDir()
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		t.Fatal(err)
	}
	h := &harness{t: t, repository: repository, directory: directory}
	overlay := h.overlay("stream-registration", "bridge/tsgo/checker/facts.go", "switch mode {", "switch mode {\ncase \"declaration-ancestry\": return p.declarationAncestry(out,c,node,question)\ncase \"wave19-type-signatures\": return p.wave19TypeSignatures(out,c,node,question)\ncase \"wave19-generic-call\": return p.wave19GenericCall(out,c,node,question)\ncase \"wave19-type-members\": return p.wave19TypeMembers(out,c,node,question)\ncase \"wave19-heritage-members\": return p.wave19HeritageMembers(out,c,node,question)")
	stage0 := filepath.Join(directory, "adamic")
	h.must("stage0", exec.Command("go", "build", "-o", stage0, "./cmd/adamic"))
	archive := h.archive("stream", overlay, false)
	sanitized := h.archive("stream-asan", overlay, true)
	config := h.write("tsconfig.json", `{"compilerOptions":{"strict":true,"target":"ES2022"},"files":["probe.a"]}`)
	probe := h.write("probe.a", "class C{m(){}} work();")
	entry := h.write("released.a", `import {panic,programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';const args=programArguments();const path=args[1]??panic('path');const program=tsgoProgram(args[0]??panic('config'),[path]);const start=Number.parseInt(args[5]??'14',10);const end=Number.parseInt(args[4]??'21',10);const kind=args[3]??'CallExpression';tsgoInspect(program,path,start,end,kind,'raw-shape');tsgoRelease(program);console.log(tsgoInspect(program,path,start,end,kind,args[2]??'wave19-generic-call'));`)
	normal := h.build(stage0, "released", entry, archive, false)
	asan := h.build(stage0, "released-asan", entry, sanitized, true)
	questions := []struct{ question, kind, end, start string }{{"wave19-type-signatures\n1", "CallExpression", "21", "14"}, {"wave19-generic-call", "CallExpression", "21", "14"}, {"wave19-type-members\n1\nthen", "CallExpression", "21", "14"}, {"wave19-heritage-members", "MethodDeclaration", "13", "8"}}
	for _, binary := range []string{normal, asan} {
		for _, question := range questions {
			got := h.run("released-run", exec.Command(binary, config, probe, question.question, question.kind, question.end, question.start))
			if got.err == nil || !strings.Contains(string(got.stderr), "invalid or released checker handle") {
				t.Fatal("released handle accepted")
			}
			if bytes.Contains(got.stderr, []byte("ERROR: AddressSanitizer")) || bytes.Contains(got.stderr, []byte("runtime error:")) || bytes.Contains(got.stderr, []byte("LeakSanitizer")) {
				t.Fatal("sanitizer failed")
			}
		}
	}
	t.Log("all four questions refuse released handles in normal and ASan/UBSan/LSan builds")
	registryOverlay := h.overlay("released-registry", "bridge/tsgo/archive/main.go", "delete(programs.live, uint64(handle))", "// Mutant: retain the released handle.")
	var combined, registry map[string]map[string]string
	data, err := os.ReadFile(overlay)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &combined); err != nil {
		t.Fatal(err)
	}
	data, err = os.ReadFile(registryOverlay)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &registry); err != nil {
		t.Fatal(err)
	}
	for original, replacement := range registry["Replace"] {
		combined["Replace"][original] = replacement
	}
	data, err = json.Marshal(combined)
	if err != nil {
		t.Fatal(err)
	}
	combinedOverlay := h.write("combined-release.json", string(data))
	mutantArchive := h.archive("released-mutant-checker", combinedOverlay, false)
	mutant := h.build(stage0, "released-mutant", entry, mutantArchive, false)
	for _, question := range questions {
		got := h.must("released-mutant-run", exec.Command(mutant, config, probe, question.question, question.kind, question.end, question.start))
		if len(got.stderr) != 0 {
			t.Fatal("registry mutant failed outside the refusal check")
		}
	}
	t.Log("retained-registry mutant exits 0 with empty stderr; all four required-panic checks kill it")
}
