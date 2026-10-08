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
)

// Complete original receiver and element declarations are imported, including unread members.
// The source spans are static candidate reads; these fixtures do not measure tsc reachability.
func TestCheckedViewRanked15OriginalArrays(t *testing.T) {
	declarations := os.Getenv("ADAMIC_ARRAY15_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("set ADAMIC_ARRAY15_ORIGINAL_DECLS to original15/prepare.cjs output")
	}
	data, err := os.ReadFile(checkedViewFixturePath(filepath.Join(declarations, "array15-manifest.json")))
	if err != nil {
		t.Fatal(err)
	}
	var manifest intersectionOriginalManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Commit != "050880ce59e30b356b686bd3144efe24f875ebc8" || len(manifest.Pairs) != 3 {
		t.Fatal("original provenance changed")
	}
	for _, pair := range manifest.Pairs {
		if pair.Reads != 4 || len(pair.Sites) != 4 {
			t.Fatal("candidate witnesses changed")
		}
	}
	for name, digest := range manifest.Declarations {
		data, err := os.ReadFile(checkedViewFixturePath(filepath.Join(declarations, name)))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
			t.Fatal("original declaration drift: " + name)
		}
	}
	for _, receiver := range []struct{ name, typ, kind string }{{"function", "FunctionExpression", "219"}, {"get", "GetAccessorDeclaration", "178"}, {"set", "SetAccessorDeclaration", "179"}} {
		for _, probe := range []struct{ mode, source, diagnostic string }{
			{"good", "1:7\nabsent\nabsent\n", ""},
			{"lazy", receiver.kind + "\n", ""},
			{"wrong-array", "3\n", "field read failed: (base as " + receiver.typ + ").typeParameters matches no member of NodeArray<TypeParameterDeclaration> | undefined; expected NodeArray<TypeParameterDeclaration> | undefined, found string"},
			{"wrong-pos", "bad\n", "field read failed: parameters[0]!.pos is not a number; expected number, found string"},
			{"missing-pos", "undefined\n", "field read failed: parameters[0]!.pos is not initialized; expected number, found missing"},
		} {
			t.Run(receiver.name+"-"+probe.mode, func(t *testing.T) {
				name := receiver.name + "-" + probe.mode
				input, err := os.ReadFile(checkedViewFixturePath(filepath.Join(repository, "stage3/interface-downcasts/lane2/original15", name+".a")))
				if err != nil {
					t.Fatal(err)
				}
				bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
				file := checkedViewFixtureCopyPath(filepath.Join(t.TempDir(), name+".a"), filepath.Join(repository, "stage3/interface-downcasts/lane2/original15", name+".a"))
				if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
					t.Fatal(err)
				}
				if difference := disagreement(run{stdout: []byte(probe.source)}, onNode(t, file)); difference != "" {
					t.Fatal("Node: " + difference)
				}
				program, err := lowered(t, file)
				if err != nil {
					t.Fatal(err)
				}
				for _, typ := range []string{receiver.typ, "TypeParameterDeclaration"} {
					if probe.mode == "lazy" && typ == "TypeParameterDeclaration" {
						continue
					}
					complete := false
					for _, c := range program.ViewContracts {
						if c.Name != typ {
							continue
						}
						fields := []string{}
						for _, f := range c.Fields {
							fields = append(fields, f.Name)
						}
						slices.Sort(fields)
						complete = complete || slices.Equal(fields, manifest.Fields[typ])
					}
					if !complete {
						t.Fatal("original field set reduced: " + typ)
					}
				}
				want := run{stdout: []byte(probe.source)}
				if probe.diagnostic != "" {
					want = run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.diagnostic + "\n")}
				}
				actual, binary := nativelyUncached(t, program)
				for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
					if difference := disagreement(want, got); difference != "" {
						t.Errorf("%s; stderr %q; stdout %q; exit %d", difference, got.stderr, got.stdout, got.exitCode)
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
}

// Original declarations are external generated inputs. Without those inputs, preserve their
// recorded observations; the explicit original-declaration run measures them again.
func ranked15ArrayCounts(t *testing.T) []string {
	t.Helper()
	const prefix = "stage3/interface-downcasts/lane2/original15/"
	declarations := os.Getenv("ADAMIC_ARRAY15_ORIGINAL_DECLS")
	rows := []string{}
	if declarations == "" {
		data, err := os.ReadFile(checkedViewFixturePath(filepath.Join(repository, "internal/oracle/counts.md")))
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "| "+prefix) {
				rows = append(rows, line)
			}
		}
		t.Log("original15 counts retained; set ADAMIC_ARRAY15_ORIGINAL_DECLS to remeasure")
		return rows
	}
	for _, receiver := range []string{"function", "get", "set"} {
		for _, mode := range []string{"good", "lazy", "wrong-array", "wrong-pos", "missing-pos"} {
			name := receiver + "-" + mode + ".a"
			input, err := os.ReadFile(checkedViewFixturePath(filepath.Join(repository, prefix, name)))
			if err != nil {
				t.Fatal(err)
			}
			bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
			file := checkedViewFixtureCopyPath(filepath.Join(t.TempDir(), name), filepath.Join(repository, prefix, name))
			if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
				t.Fatal(err)
			}
			absoluteRepository, err := filepath.Abs(checkedViewFixturePath(repository))
			if err != nil {
				t.Fatal(err)
			}
			relative, err := filepath.Rel(absoluteRepository, file)
			if err != nil {
				t.Fatal(err)
			}
			row := counted(t, checkedViewFixturePath(relative), false, nil, false, false)
			rows = append(rows, strings.Replace(row, relative, checkedViewFixturePath(prefix+name), 1))
		}
	}
	return rows
}

// Not parallel: the explicit refresh writes the shared counts table.
func TestCheckedViewRanked15ArrayCounts(t *testing.T) {
	if os.Getenv("ADAMIC_ARRAY15_ORIGINAL_DECLS") == "" {
		t.Skip("original15 declaration inputs required for measured counts")
	}
	rows := ranked15ArrayCounts(t)
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(checkedViewFixturePath(path))
	if err != nil {
		t.Fatal(err)
	}
	if *updateCounts {
		parts := strings.SplitN(string(data), "\n## Predicate direction counts", 2)
		lines := strings.Split(strings.TrimSuffix(parts[0], "\n"), "\n")
		for _, row := range rows {
			key := strings.Split(row, " | ")[0] + " | "
			found := false
			for i, line := range lines {
				if strings.HasPrefix(line, key) {
					lines[i] = row
					found = true
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
		return
	}
	for _, row := range rows {
		if !strings.Contains(string(data), row+"\n") {
			t.Errorf("unrecorded original15 counts: %s", row)
		}
	}
}
