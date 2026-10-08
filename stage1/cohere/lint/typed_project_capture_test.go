package lint

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Overlays only: the pinned cohere sources themselves remain unchanged. Cases
// keep their original asserted findings/options plus the exact fixture project.
func typedProjectCaptureOverlay(root, directory string) (map[string]string, error) {
	replacements := map[string]string{}
	for _, edit := range []struct{ relative, anchor, replacement string }{
		{"internal/lint/testing/docs_capture.go", "options    any", "options    any\n project map[string]string"},
		{"internal/docsdata/capture/capture.go", "type Record struct {", "type Record struct {\n Project map[string]string `json:\"project,omitempty\"`"},
	} {
		original := filepath.Join(root, edit.relative)
		data, err := os.ReadFile(original)
		if err != nil {
			return nil, err
		}
		if strings.Count(string(data), edit.anchor) != 1 {
			return nil, fmt.Errorf("typed project capture anchor changed: %s", edit.relative)
		}
		text := strings.Replace(string(data), edit.anchor, edit.replacement, 1)
		if strings.HasSuffix(edit.relative, "docs_capture.go") {
			anchor := "Rule:        result.capture.rule,"
			if strings.Count(text, anchor) != 1 {
				return nil, fmt.Errorf("typed project record anchor changed")
			}
			text = strings.Replace(text, anchor, anchor+"\n Project: result.capture.project,", 1)
		}
		side := filepath.Join(directory, "project-"+filepath.Base(original))
		if err := os.WriteFile(side, []byte(text), 0644); err != nil {
			return nil, err
		}
		replacements[original] = side
	}
	virtual := filepath.Join(root, "internal/lint/testing/adamic_typed_project_capture.go")
	side := filepath.Join(directory, "typed_project_capture.go")
	if err := os.WriteFile(side, []byte(typedProjectCaptureSource), 0644); err != nil {
		return nil, err
	}
	replacements[virtual] = side
	return replacements, nil
}

const typedProjectCaptureSource = `package rule_testing

import (
 "io/fs"
 "os"
 "path/filepath"
 "testing"
 "github.com/system-inc/cohere/internal/docsdata/capture"
)

func RecordTypedProject(t *testing.T, result Result, directory string, fixtureFiles map[string]string, onDisk bool, verbatim bool) {
 t.Helper()
 if capture.Directory() == "" || result.capture == nil { return }
 files := map[string]string{}
 if !onDisk {
  // buildProgramInto constructs exactly these bytes in the memory filesystem.
  for name, contents := range fixtureFiles {
   if !verbatim { contents = FixtureText(contents) }
   // MemoryFS keys are directory + "/" + name even for slash-prefixed names.
   // Preserve the normalized path inside that fixture root, as the compiler does.
   relative, err := filepath.Rel(directory, directory+"/"+filepath.ToSlash(name))
   if err != nil { t.Fatal(err) }
   files[filepath.ToSlash(relative)] = contents
  }
  files["tsconfig.json"] = defaultTsConfig
  result.capture.project = files
  return
 }
 err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
  if err != nil { return err }
  if entry.Type() & os.ModeSymlink != 0 { return &fs.PathError{Op:"capture typed project symlink",Path:path,Err:fs.ErrInvalid} }
  if entry.IsDir() { return nil }
  bytes, err := os.ReadFile(path); if err != nil { return err }
  relative, err := filepath.Rel(directory,path); if err != nil { return err }
  files[filepath.ToSlash(relative)] = string(bytes); return nil
 })
 if err != nil { t.Fatal(err) }
 result.capture.project = files
}
`

// A captured fixture owns its tsconfig; strict-only fallback is for generated
// rows and witnesses that have no project. Marker search ends at the filesystem
// root and never treats an unrelated repository tsconfig as fixture metadata.
func typedConfigForRow(t *testing.T, source string) string {
	t.Helper()
	for directory := filepath.Dir(source); ; directory = filepath.Dir(directory) {
		marker := filepath.Join(directory, ".adamic-typed-project")
		data, err := os.ReadFile(marker)
		if err == nil {
			if string(data) != "tsconfig.json\n" {
				t.Fatalf("invalid typed project marker: %s", marker)
			}
			config := filepath.Join(directory, "tsconfig.json")
			if _, err := os.Stat(config); err != nil {
				t.Fatal(err)
			}
			return config
		}
		if !os.IsNotExist(err) {
			t.Fatal(err)
		}
		if parent := filepath.Dir(directory); parent == directory {
			break
		}
	}
	config := filepath.Join(t.TempDir(), "tsconfig.json")
	options, err := json.Marshal(map[string]any{"compilerOptions": map[string]bool{"strict": true}, "files": []string{source}})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, options, 0644); err != nil {
		t.Fatal(err)
	}
	return config
}

// A reconstructed project must preserve the upstream run's findings count even
// if all three runtimes would otherwise agree on a damaged, smaller project.
func checkCapturedFindingCount(source string, output []byte) error {
	data, err := os.ReadFile(source + ".capture-findings.json")
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var expected []json.RawMessage
	if err := json.Unmarshal(data, &expected); err != nil {
		return err
	}
	actual := bytes.Count(output, []byte("\nrange "))
	if bytes.HasPrefix(output, []byte("range ")) {
		actual++
	}
	if actual != len(expected) {
		return fmt.Errorf("captured project changed upstream findings for %s: got %d, captured %d", source, actual, len(expected))
	}
	return nil
}

func TestTypedProjectInputsAndMutants(t *testing.T) {
	t.Parallel()
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	source := ownedWitnesses(t, directory, "nexus-correctness-no-mock-on-module-namespace")[0]
	config := typedConfigForRow(t, source)
	path := manifest(t, []string{"program " + config, source + "\tnexus/correctness-no-mock-on-module-namespace"})
	oracle := goOracle(t)
	// The witness has plain, asserted and non-null namespace targets; Go reports
	// one finding for each. Keep this literal independent of the run being checked.
	if err := os.WriteFile(source+".capture-findings.json", []byte("[{},{},{}]"), 0644); err != nil {
		t.Fatal(err)
	}
	run := func() []byte { return execute(t, "", oracle, "--manifest", path).output }
	if err := checkCapturedFindingCount(source, run()); err != nil {
		t.Fatal(err)
	}
	foreign := filepath.Join(filepath.Dir(source), "node.d.ts")
	foreignBytes, err := os.ReadFile(foreign)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(foreign); err != nil {
		t.Fatal(err)
	}
	if err := checkCapturedFindingCount(source, run()); err == nil {
		t.Fatal("missing foreign input mutant survived capture guard")
	} else {
		t.Logf("missing foreign input caught: %v", err)
	}
	if err := os.WriteFile(foreign, foreignBytes, 0644); err != nil {
		t.Fatal(err)
	}
	configBytes, err := os.ReadFile(config)
	if err != nil {
		t.Fatal(err)
	}
	var options map[string]any
	if err := json.Unmarshal(configBytes, &options); err != nil {
		t.Fatal(err)
	}
	options["compilerOptions"].(map[string]any)["module"] = "CommonJS"
	options["compilerOptions"].(map[string]any)["moduleResolution"] = "Node10"
	changed, err := json.Marshal(options)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(config, changed, 0644); err != nil {
		t.Fatal(err)
	}
	if err := checkCapturedFindingCount(source, run()); err == nil {
		t.Fatal("changed tsconfig mutant survived capture guard")
	} else {
		t.Logf("changed tsconfig caught: %v", err)
	}
	if err := os.WriteFile(config, configBytes, 0644); err != nil {
		t.Fatal(err)
	}
	if err := checkCapturedFindingCount(source, run()); err != nil {
		t.Fatal(err)
	}
	t.Log("restoring foreign inputs and the exact config restores the captured verdict")
	// Preserve the distinct control where upstream deliberately supplies no checker.
	if err := os.WriteFile(source+".capture-findings.json", []byte("[]"), 0644); err != nil {
		t.Fatal(err)
	}
	marker := source + ".capture-no-program"
	if err := os.WriteFile(marker, []byte("syntax-only\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runCaptured := func() []byte {
		rows := []string{source + "\tnexus/correctness-no-mock-on-module-namespace"}
		if !capturedWithoutProgram(t, source) {
			rows = append([]string{"program " + config}, rows...)
		}
		return execute(t, "", oracle, "--manifest", manifest(t, rows)).output
	}
	if err := checkCapturedFindingCount(source, runCaptured()); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(marker); err != nil {
		t.Fatal(err)
	}
	if err := checkCapturedFindingCount(source, runCaptured()); err == nil {
		t.Fatal("adding a checker to a syntax-only capture survived")
	} else {
		t.Logf("changed checker presence caught: %v", err)
	}
}

// A Run (rather than RunTyped) capture must retain its absent checker.
func capturedWithoutProgram(t *testing.T, source string) bool {
	t.Helper()
	data, err := os.ReadFile(source + ".capture-no-program")
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "syntax-only\n" {
		t.Fatalf("invalid capture program marker: %s", source)
	}
	return true
}
