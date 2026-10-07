package wave104

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func loadOracle(t *testing.T) string {
	t.Helper()
	root := filepath.Join(repository(t), "cohere")
	local, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	virtual := filepath.Join(root, "adamic_wave104_load.go")
	sourcePath := filepath.Join(root, "internal/lint/rules/tailwind/design_system.go")
	data, err := os.ReadFile(sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	source := string(data)
	// Observe dependency output without replacing the actual Go load or wrapper.
	anchor := "result := loadDesignSystemThrough(program, fileSystem)"
	if strings.Count(source, anchor) != 1 {
		t.Fatal("Go observer anchor changed")
	}
	source = strings.Replace(source, anchor, anchor+"\n wave104ObservedLoaded = result", 1)
	anchor = "func loadDesignSystemThrough(program rule.Program, fileSystem *rule.RecordingFS) DesignSystemResult {"
	if strings.Count(source, anchor) != 1 {
		t.Fatal("Go load anchor changed")
	}
	source = strings.Replace(source, anchor, anchor+"\n wave104LoadTrace=append(wave104LoadTrace,\"load\")", 1)
	observed := filepath.Join(t.TempDir(), "design_system.go")
	if err = os.WriteFile(observed, []byte(source), 0644); err != nil {
		t.Fatal(err)
	}
	replacements := map[string]string{virtual: filepath.Join(local, "load_oracle.go.txt"), sourcePath: observed, filepath.Join(root, "internal/lint/rules/tailwind/adamic_wave104_load_exports.go"): filepath.Join(local, "load_exports.go.txt")}
	encoded, err := json.Marshal(map[string]any{"Replace": replacements})
	if err != nil {
		t.Fatal(err)
	}
	overlay := filepath.Join(t.TempDir(), "overlay.json")
	if err = os.WriteFile(overlay, encoded, 0644); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "oracle")
	run(t, root, "go", "build", "-overlay="+overlay, "-o", binary, virtual)
	return binary
}
func loadCases(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile("cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []Case
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	consumers := map[string]bool{}
	for _, row := range rows {
		consumers[row.Rule] = true
	}
	if len(consumers) != 6 {
		t.Fatal("six consumers required")
	}
	packageRoot := "/tmp/wave104-tailwind/node_modules/tailwindcss"
	if _, err = os.Stat(filepath.Join(packageRoot, "index.css")); err != nil {
		t.Fatalf("install pinned tailwindcss@4.1.18 for controls: %v", err)
	}
	for name := range consumers {
		for mode := 0; mode < 4; mode++ {
			project := t.TempDir()
			style := filepath.Join(project, "app")
			if err = os.MkdirAll(style, 0755); err != nil {
				t.Fatal(err)
			}
			css := "@import 'tailwindcss';\n@theme { --color-olive: #aaee77; }\n@utility special { color: red; }"
			if mode == 2 {
				css = "@import './missing.css';"
			}
			if mode != 3 {
				if err = os.WriteFile(filepath.Join(style, "globals.css"), []byte(css), 0644); err != nil {
					t.Fatal(err)
				}
			}
			if mode != 1 {
				modules := filepath.Join(project, "node_modules")
				if err = os.MkdirAll(modules, 0755); err != nil {
					t.Fatal(err)
				}
				if err = os.Symlink(packageRoot, filepath.Join(modules, "tailwindcss")); err != nil {
					t.Fatal(err)
				}
			}
			rows = append(rows, Case{Rule: name, Config: filepath.Join(project, "tsconfig.json"), Cwd: "unused", Present: true, Origin: "real Go design system success/missing package/missing import/missing entry control"})
		}
	}
	rows = append(rows, Case{Cwd: "", Origin: "empty project root"})
	path := filepath.Join(t.TempDir(), "cases.json")
	data, err = json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
	t.Logf("%d real-program fixture cases and 25 live filesystem controls; six consumers", len(rows)-25)
	return path
}
func loadFacts(t *testing.T) (string, []byte) {
	t.Helper()
	facts := filepath.Join(t.TempDir(), "facts.json")
	want := run(t, "", loadOracle(t), loadCases(t), facts)
	data, err := os.ReadFile(facts)
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ System, Table, Error, Reads, LoadedReads string }
	if err = json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	success, failed, nonempty := 0, 0, 0
	for _, row := range rows {
		if row.System == "1" && row.Table == "1" && row.Error == "0" {
			success++
		}
		if strings.HasPrefix(row.Error, "1") {
			failed++
		}
		if row.Reads != "[]" {
			nonempty++
		}
		if row.LoadedReads != "null" {
			t.Fatal("unexpected Go dependency read field")
		}
	}
	if success < 6 || failed < 6 || nonempty < 6 {
		t.Fatalf("incomplete coverage success=%d failed=%d nonempty=%d", success, failed, nonempty)
	}
	t.Logf("real Go success=%d failed=%d nonempty read snapshots=%d; %d exact bytes", success, failed, nonempty, len(want))
	return facts, want
}

// Not parallel: native sanitizers and actual Go CSS loads share bounded compiler resources.
func TestLoadDesignSystemForProgramMatchesGo(t *testing.T) {
	facts, want := loadFacts(t)
	entry, js, binary := buildHelper(t, "load_design_system_for_program.a", "load_main.a", "", "")
	for side, got := range observations(t, entry, js, binary, facts) {
		if !bytes.Equal(got, want) {
			t.Fatalf("%s %s", side, firstDifference(got, want))
		}
	}
	t.Logf("%d lines byte identical on source Node, emitted JavaScript and ASan/UBSan native", bytes.Count(want, []byte("\n")))
}

// Not parallel: each semantic mutant must compile and complete without sanitizer diagnostics.
func TestLoadDesignSystemForProgramMutants(t *testing.T) {
	facts, want := loadFacts(t)
	for _, change := range []struct{ name, from, to string }{
		{"discard recording snapshot", "reads: filesystem.reads()", "reads: result.reads"},
		{"snapshot before load", "const result = load(program, filesystem);", "const snapshot = filesystem.reads();\n const result = load(program, filesystem);"},
	} {
		t.Run(change.name, func(t *testing.T) {
			if change.name == "snapshot before load" {
				// One explicit source edit affects the snapshot expression and its timing.
				original, err := os.ReadFile("load_design_system_for_program.a")
				if err != nil {
					t.Fatal(err)
				}
				from := string(original)
				to := strings.Replace(from, change.from, change.to, 1)
				to = strings.Replace(to, "reads: filesystem.reads()", "reads: snapshot", 1)
				change.from, change.to = from, to
			}
			entry, js, binary := buildHelper(t, "load_design_system_for_program.a", "load_main.a", change.from, change.to)
			for side, got := range observations(t, entry, js, binary, facts) {
				if bytes.Equal(got, want) {
					t.Fatal("mutant survived", side)
				}
				t.Logf("compiled, finished cleanly, only Go comparison caught %s on %s: %s", change.name, side, firstDifference(got, want))
			}
		})
	}
}
