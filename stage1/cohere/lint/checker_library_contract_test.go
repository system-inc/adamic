package lint

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This probe isolates library-member-provenance from the upstream helper's
// separate compiler-options read. It keeps the upstream finding as its oracle.
func TestCheckerLibraryMemberReadContract(t *testing.T) {
	t.Parallel()
	directory := mutant(t, "if(!isSymbolFromDefaultLibrary(checker, member)) { return false; }", "", "rules/nexus-consistency-no-iso-string-date-cut/rule.a")
	descriptor := filepath.Join(directory, "rules/nexus-consistency-no-iso-string-date-cut/rule.json")
	data, err := os.ReadFile(descriptor)
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatal(err)
	}
	writeReads := func(reads []string) {
		settings["programReads"] = reads
		encoded, err := json.Marshal(settings)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(descriptor, encoded, 0644); err != nil {
			t.Fatal(err)
		}
	}
	writeReads([]string{"ReadsDefaultLibrary"})
	source := ownedWitnesses(t, directory, "nexus-consistency-no-iso-string-date-cut")[0]
	config := typedConfigForRow(t, source)
	path := manifest(t, []string{"program " + config, source + "\tnexus/consistency-no-iso-string-date-cut"})
	prefix := filepath.Join(t.TempDir(), "library-transcript")
	want := execute(t, "", goOracle(t), "--manifest", path).output
	native := execute(t, "", buildPort(t, directory, true), "--manifest", path, "--record", prefix).output
	if diff := difference(native, want); diff != "" {
		t.Fatal(diff)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	replay := func(module string) []byte {
		return execute(t, "", "node", "--disable-warning=ExperimentalWarning", runner, module, "--manifest", path, "--replay", prefix).output
	}
	for runtime, module := range map[string]string{"Node": filepath.Join(directory, "main.ts"), "emitted JavaScript": emittedJavaScript(t, directory)} {
		if diff := difference(replay(module), want); diff != "" {
			t.Fatalf("%s: %s", runtime, diff)
		}
	}

	if !bytes.Contains(want, []byte("isoStringCutToDate")) {
		t.Fatalf("library witness did not exercise the question: %s", want)
	}
	t.Log("library member facts replay identically with ReadsDefaultLibrary alone")
	writeReads([]string{})
	prepareRegistry(t, directory)
	refused := replay(filepath.Join(directory, "main.ts"))
	if !bytes.Contains(refused, []byte("undeclared program read ReadsDefaultLibrary")) || bytes.Equal(refused, want) {
		t.Fatalf("missing declaration survived: %s", refused)
	}
	t.Log("removing the default-library declaration refuses visibly")

	// A legacy full payload also carries library flags. An OtherFiles declaration
	// cannot authorize that separate read. Early refusal records no checker entry.
	writeReads([]string{"ReadsOtherFiles"})
	helper := filepath.Join(directory, "checker_library_member.a")
	code, err := os.ReadFile(helper)
	if err != nil {
		t.Fatal(err)
	}
	before := "new FileQuestion('ReadsDefaultLibrary', `library-member-provenance\\n"
	after := "new FileQuestion('ReadsOtherFiles', `symbol-provenance\\n"
	if strings.Count(string(code), before) != 1 {
		t.Fatal("full declaration probe anchor changed")
	}
	if err := os.WriteFile(helper, []byte(strings.Replace(string(code), before, after, 1)), 0644); err != nil {
		t.Fatal(err)
	}
	prepareRegistry(t, directory)
	prefix = filepath.Join(t.TempDir(), "full-declaration-transcript")
	full := execute(t, "", buildPort(t, directory, true), "--manifest", path, "--record", prefix).output
	if !bytes.Contains(full, []byte("undeclared program read ReadsDefaultLibrary")) {
		t.Fatalf("full payload bypassed its library read: %s", full)
	}
	for runtime, module := range map[string]string{"Node": filepath.Join(directory, "main.ts"), "emitted JavaScript": emittedJavaScript(t, directory)} {
		if diff := difference(replay(module), full); diff != "" {
			t.Fatalf("%s full declaration refusal: %s", runtime, diff)
		}
	}
	// Remove only this read guard. Empty replay now gives a missing-entry refusal,
	// so comparison catches the mutant without a native bridge panic.
	checkerPath := filepath.Join(directory, "checker.a")
	code, err = os.ReadFile(checkerPath)
	if err != nil {
		t.Fatal(err)
	}
	before = "if(libraryRead && !this.reads.includes('ReadsDefaultLibrary'))"
	if strings.Count(string(code), before) != 1 {
		t.Fatal("library guard mutant anchor changed")
	}
	if err := os.WriteFile(checkerPath, []byte(strings.Replace(string(code), before, "if(false && !this.reads.includes('ReadsDefaultLibrary'))", 1)), 0644); err != nil {
		t.Fatal(err)
	}
	for runtime, module := range map[string]string{"Node": filepath.Join(directory, "main.ts"), "emitted JavaScript": emittedJavaScript(t, directory)} {
		bad := replay(module)
		if bytes.Equal(bad, full) || !bytes.Contains(bad, []byte("missing transcript entry")) {
			t.Fatalf("%s library guard mutant survived: %s", runtime, bad)
		}
	}
	bad := execute(t, "", buildPort(t, directory, true), "--manifest", path, "--replay", prefix).output
	if bytes.Equal(bad, full) || !bytes.Contains(bad, []byte("missing transcript entry")) {
		t.Fatalf("native library guard mutant survived: %s", bad)
	}
	t.Log("full declaration library guard mutant caught on native and both replays")
}
