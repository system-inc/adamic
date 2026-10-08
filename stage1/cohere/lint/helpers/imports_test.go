package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/system-inc/adamic/internal/coherepin"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Retained captures query unchanged Go helpers on every consumer input and factory control.
func TestImportsAgreementAndMutants(t *testing.T) {
	root, _ := filepath.Abs("../../../..")
	cohere := filepath.Join(root, "cohere")
	pin, err := coherepin.Pinned(root)
	if err != nil {
		t.Fatal(err)
	}
	if err = coherepin.Check(root, pin); err != nil {
		t.Fatal(err)
	}
	var provenance struct{ GoPin string }
	data, err := os.ReadFile("imports/testdata/provenance.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &provenance); err != nil {
		t.Fatal(err)
	}
	if err = coherepin.Check(root, provenance.GoPin); err != nil {
		t.Fatal(err)
	}
	capturePin, err := json.Marshal(map[string]string{"pin": pin})
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct{ name, file, from, to string }{
		{"bindings", "bindings_of.a", "namespace: clause.namedBindings", "namespace: binding.name"},
		{"filename", "normalized_file_name.a", "file.fileName.split('\\\\').join('/')", "file.fileName.replace('\\\\', '/')"},
		{"imported", "imported_name_of.a", "if(specifier.propertyName >= 0)", "if(false)"},
		{"call", "call_expression_source.a", "node.arguments.length === 1", "node.arguments.length >= 1"},
		{"segment", "has_path_segment.a", "byte === 92", "false"},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir, _ := filepath.Abs("imports/testdata/" + c.name)
			scratch := t.TempDir()
			if err := os.WriteFile(filepath.Join(scratch, "capture-pin.json"), capturePin, 0644); err != nil {
				t.Fatal(err)
			}
			virtual := filepath.Join(cohere, "adamic_imports_oracle.go")
			replacements := map[string]string{virtual: filepath.Join(dir, "oracle.go")}
			if c.name == "bindings" {
				replacements[filepath.Join(cohere, "internal/lint/ecmascript/react/adamic_imports.go")] = filepath.Join(dir, "react_export.go")
			}
			if c.name == "filename" {
				replacements[filepath.Join(cohere, "TypeScript/tsc/internal/ast/adamic_imports.go")] = filepath.Join(dir, "filename_ast.go")
				replacements[filepath.Join(cohere, "TypeScript-shim/ast/adamic_imports.go")] = filepath.Join(dir, "filename_shim.go")
			}
			overlay, _ := json.Marshal(map[string]any{"Replace": replacements})
			overlayPath := filepath.Join(scratch, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
				t.Fatal(err)
			}
			oracle := filepath.Join(scratch, "oracle")
			run(t, cohere, "go", "build", "-overlay="+overlayPath, "-o", oracle, virtual)
			cases := filepath.Join(scratch, "cases.json")
			var want []byte
			switch c.name {
			case "bindings":
				t.Log(string(run(t, "", oracle, root, scratch, "github.com/system-inc/cohere/internal/lint/ecmascript/imports.BindingsOf")))
			case "filename":
				t.Log(string(run(t, "", oracle, root, scratch)))
			case "imported":
				t.Log(string(run(t, "", oracle, root, scratch, "imported")))
			case "segment":
				want = run(t, "", oracle, filepath.Join(dir, "sources.jsonl.gz"), cases)
			case "call":
				// The inherited capture includes sibling JSX consumers; retain the inputs,
				// then compare only this helper's observations.
				data, err := os.ReadFile(filepath.Join(dir, "sources.jsonl.gz"))
				if err != nil {
					t.Fatal(err)
				}
				_ = data
				inputs := importsCapturedInputs(t, filepath.Join(dir, "sources.jsonl.gz"))
				raw, _ := json.Marshal(inputs)
				config := filepath.Join(scratch, "sources.json")
				if err = os.WriteFile(config, raw, 0644); err != nil {
					t.Fatal(err)
				}
				raw = run(t, "", oracle, config)
				var corpus map[string]json.RawMessage
				if err = json.Unmarshal(raw, &corpus); err != nil {
					t.Fatal(err)
				}
				var full string
				if err = json.Unmarshal(corpus["Want"], &full); err != nil {
					t.Fatal(err)
				}
				for _, line := range strings.Split(full, "\n") {
					if strings.HasPrefix(line, "call:") {
						want = append(want, []byte(line+"\n")...)
					}
				}
				delete(corpus, "Want")
				raw, _ = json.Marshal(corpus)
				if err = os.WriteFile(cases, raw, 0644); err != nil {
					t.Fatal(err)
				}
				t.Logf("%d captured inputs", len(inputs))
			}
			if want == nil {
				var err error
				want, err = os.ReadFile(filepath.Join(scratch, "want.txt"))
				if err != nil {
					t.Fatal(err)
				}
			}
			entry := filepath.Join(dir, "main.a")
			check := func(entry string, mutant bool) {
				program, err := load.Load([]string{entry})
				if err != nil {
					t.Fatal(err)
				}
				ir, err := lower.Lower(context.Background(), program)
				if err != nil {
					t.Fatal(err)
				}
				binary := filepath.Join(t.TempDir(), "imports")
				if err = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); err != nil {
					t.Fatal(err)
				}
				js := binary + ".mjs"
				if err = os.WriteFile(js, []byte(javascript.JavaScript(ir)), 0644); err != nil {
					t.Fatal(err)
				}
				runner := filepath.Join(root, "oracle/node.mjs")
				for _, side := range []struct {
					name string
					cmd  []string
				}{{"Node", []string{"node", "--disable-warning=ExperimentalWarning", runner, entry, cases}}, {"JavaScript", []string{"node", "--disable-warning=ExperimentalWarning", runner, js, cases}}, {"native", []string{binary, cases}}} {
					got := run(t, "", side.cmd[0], side.cmd[1:]...)
					if mutant {
						if bytes.Equal(got, want) {
							t.Fatal(side.name + " semantic mutant survived")
						}
						t.Log(side.name + " mutant caught by output disagreement")
					} else {
						compare(t, got, want)
					}
				}
			}
			check(entry, false)
			t.Logf("Go agreement: %d output lines, %d bytes on Node, JavaScript, sanitized native", bytes.Count(want, []byte{'\n'}), len(want))
			variant := t.TempDir()
			// Copy the production graph, never a second helper implementation.
			production, _ := filepath.Abs("imports")
			entries, err := os.ReadDir(production)
			if err != nil {
				t.Fatal(err)
			}
			for _, f := range entries {
				if !strings.HasSuffix(f.Name(), ".a") {
					continue
				}
				data, err := os.ReadFile(filepath.Join(production, f.Name()))
				if err != nil {
					t.Fatal(err)
				}
				if f.Name() == c.file {
					if strings.Count(string(data), c.from) != 1 {
						t.Fatal("mutation anchor drift")
					}
					data = []byte(strings.Replace(string(data), c.from, c.to, 1))
				}
				if err = os.WriteFile(filepath.Join(variant, f.Name()), data, 0644); err != nil {
					t.Fatal(err)
				}
			}
			driver, err := os.ReadFile(entry)
			if err != nil {
				t.Fatal(err)
			}
			driver = []byte(strings.ReplaceAll(strings.ReplaceAll(string(driver), "'../../../options_json.ts'", "'"+filepath.Join(root, "stage1/cohere/lint/helpers/options_json.ts")+"'"), "'../../", "'./"))
			mutatedEntry := filepath.Join(variant, "main.a")
			if err = os.WriteFile(mutatedEntry, driver, 0644); err != nil {
				t.Fatal(err)
			}
			check(mutatedEntry, true)
		})
	}
}

func TestImportsLiveSourceVisitorsAndSpecifier(t *testing.T) {
	for _, name := range []string{"visitors", "specifier"} {
		t.Run(name, func(t *testing.T) {
			script, _ := filepath.Abs("imports/testdata/" + name + "/validate.py")
			t.Log(string(run(t, "", "python3", script)))
		})
	}
}
