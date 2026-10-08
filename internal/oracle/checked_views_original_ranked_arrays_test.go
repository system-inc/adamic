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

type originalArrayProbe struct {
	Name       string   `json:"name"`
	Source     string   `json:"source"`
	Diagnostic string   `json:"diagnostic"`
	Contracts  []string `json:"contracts"`
}

func originalArrayInputs(t *testing.T, group string) (string, string, []originalArrayProbe) {
	t.Helper()
	declarations := os.Getenv("ADAMIC_ARRAY" + group + "_ORIGINAL_DECLS")
	directory := filepath.Join(repository, "stage3/interface-downcasts/lane2/original"+group)
	data, err := os.ReadFile(filepath.Join(directory, "probes.json"))
	if err != nil {
		t.Fatal(err)
	}
	var probes []originalArrayProbe
	if err := json.Unmarshal(data, &probes); err != nil {
		t.Fatal(err)
	}
	return declarations, directory, probes
}

func originalArrayFile(t *testing.T, declarations, directory, name string) string {
	t.Helper()
	input, err := os.ReadFile(filepath.Join(directory, name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
	file := filepath.Join(t.TempDir(), name+".a")
	if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
		t.Fatal(err)
	}
	return file
}

// Original declarations include every unread member. Candidate spans do not measure runtime reachability.
func originalRankedArrayOracle(t *testing.T, group string, pairCount, reads int) {
	t.Helper()
	declarations, directory, probes := originalArrayInputs(t, group)
	if declarations == "" {
		t.Skip("original" + group + "/prepare.cjs declaration inputs required")
	}
	data, err := os.ReadFile(filepath.Join(declarations, "array"+group+"-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest intersectionOriginalManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Commit != "050880ce59e30b356b686bd3144efe24f875ebc8" || len(manifest.Pairs) != pairCount {
		t.Fatal("original provenance changed")
	}
	for _, pair := range manifest.Pairs {
		if pair.Reads != reads || len(pair.Sites) != reads {
			t.Fatal("candidate witnesses changed")
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
	for _, probe := range probes {
		t.Run(probe.Name, func(t *testing.T) {
			file := originalArrayFile(t, declarations, directory, probe.Name)
			if difference := disagreement(run{stdout: []byte(probe.Source)}, onNode(t, file)); difference != "" {
				t.Fatal("Node: " + difference)
			}
			program, err := lowered(t, file)
			if err != nil {
				t.Fatal(err)
			}
			for _, typ := range probe.Contracts {
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
			want := run{stdout: []byte(probe.Source)}
			if probe.Diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + probe.Diagnostic + "\n")}
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

func originalRankedArrayCounts(t *testing.T, group string) []string {
	t.Helper()
	declarations, directory, probes := originalArrayInputs(t, group)
	prefix := "stage3/interface-downcasts/lane2/original" + group + "/"
	rows := []string{}
	if declarations == "" {
		data, err := os.ReadFile(filepath.Join(repository, "internal/oracle/counts.md"))
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(line, "| "+prefix) {
				rows = append(rows, line)
			}
		}
		t.Log("original" + group + " counts retained without remeasurement; declaration inputs required")
		return rows
	}
	absoluteRepository, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	for _, probe := range probes {
		file := originalArrayFile(t, declarations, directory, probe.Name)
		relative, err := filepath.Rel(absoluteRepository, file)
		if err != nil {
			t.Fatal(err)
		}
		row := counted(t, relative, false, nil, false, false)
		rows = append(rows, strings.Replace(row, relative, prefix+probe.Name+".a", 1))
	}
	return rows
}

// Not parallel: a requested refresh writes only this group's rows in the shared table.
func originalRankedArrayCountTest(t *testing.T, group string) {
	t.Helper()
	if os.Getenv("ADAMIC_ARRAY"+group+"_ORIGINAL_DECLS") == "" {
		t.Skip("original declaration inputs required for measured counts")
	}
	rows := originalRankedArrayCounts(t, group)
	path := filepath.Join(repository, "internal/oracle/counts.md")
	data, err := os.ReadFile(path)
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
			t.Errorf("unrecorded original array counts: %s", row)
		}
	}
}

func TestCheckedViewRanked16OriginalArrays(t *testing.T) {
	originalRankedArrayOracle(t, "16", 7, 4)
}

// Not parallel: the explicit refresh writes this group's rows in the shared counts file.
func TestCheckedViewRanked16ArrayCounts(t *testing.T) {
	originalRankedArrayCountTest(t, "16")
}
