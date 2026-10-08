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

type intersectionOriginalManifest struct {
	Commit       string              `json:"upstream_commit"`
	Declarations map[string]string   `json:"declarations"`
	Fields       map[string][]string `json:"fields"`
	Pairs        []struct {
		ID            int               `json:"type_id"`
		Type          string            `json:"type"`
		Field         string            `json:"field"`
		Reads         int               `json:"read_count"`
		Sites         []json.RawMessage `json:"sites"`
		PresentFields []string          `json:"present_fields"`
	} `json:"pairs"`
}

func intersectionOriginalInputs(t *testing.T) (string, intersectionOriginalManifest) {
	t.Helper()
	declarations := os.Getenv("ADAMIC_INTERSECTION_ORIGINAL_DECLS")
	if declarations == "" {
		t.Skip("set ADAMIC_INTERSECTION_ORIGINAL_DECLS to pinned declaration output")
	}
	data, err := os.ReadFile(filepath.Join(declarations, "intersection-manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var manifest intersectionOriginalManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Commit != "050880ce59e30b356b686bd3144efe24f875ebc8" || len(manifest.Declarations) != 78 || len(manifest.Pairs) != 2 || manifest.Pairs[0].ID != 10236 || manifest.Pairs[0].Reads != 1 || len(manifest.Pairs[0].Sites) != 1 {
		t.Fatal("original provenance changed")
	}
	for name, digest := range manifest.Declarations {
		data, err := os.ReadFile(filepath.Join(declarations, name))
		if err != nil {
			t.Fatal(err)
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != digest {
			t.Fatal("declaration drift: " + name)
		}
	}
	return declarations, manifest
}

func intersectionOriginalProgram(t *testing.T, declarations, name, source string) (*ir.Program, string) {
	t.Helper()
	input, err := os.ReadFile("../../stage3/interface-downcasts/lane7/original/" + name + ".a")
	if err != nil {
		t.Fatal(err)
	}
	bound := strings.Replace(string(input), "'original-tsc-types'", fmt.Sprintf("%q", filepath.ToSlash(filepath.Join(declarations, "compiler/types.d.ts"))), 1)
	file := filepath.Join(t.TempDir(), name+".a")
	if err := os.WriteFile(file, []byte(bound), 0600); err != nil {
		t.Fatal(err)
	}
	if difference := disagreement(run{stdout: []byte(source)}, onNode(t, file)); difference != "" {
		t.Fatal("Node: " + difference)
	}
	program, err := lowered(t, file)
	if err != nil {
		t.Fatal(err)
	}
	return program, file
}

func requireIntersectionOriginalFields(t *testing.T, program *ir.Program, manifest intersectionOriginalManifest) {
	t.Helper()
	for _, name := range []string{"SymbolTracker", "ModuleSpecifierResolutionHost"} {
		complete := false
		for _, c := range program.ViewContracts {
			if c.Name != name {
				continue
			}
			fields := []string{}
			for _, f := range c.Fields {
				fields = append(fields, f.Name)
			}
			slices.Sort(fields)
			complete = complete || len(fields) > 0 && slices.Equal(fields, manifest.Fields[name])
		}
		if !complete {
			t.Fatal("original field set was reduced: " + name)
		}
	}
	complete := false
	for _, c := range program.ViewContracts {
		if !c.Intersection || c.Unsupported != "" {
			continue
		}
		fields := []string{}
		for _, f := range c.Fields {
			fields = append(fields, f.Name)
		}
		slices.Sort(fields)
		complete = complete || slices.Equal(fields, manifest.Pairs[0].PresentFields)
	}
	if !complete {
		t.Fatal("original intersection obligations missing")
	}
}

func TestCheckedViewIntersectionOriginalPairs(t *testing.T) {
	declarations, manifest := intersectionOriginalInputs(t)
	for _, test := range []struct{ name, source, diagnostic string }{
		{"tracker-good", "true\n", ""}, {"tracker-helpers-good", "true\n", ""},
		{"tracker-absent", "false\n", ""}, {"tracker-undefined", "false\n", ""},
		{"tracker-wrong", "true\n", "field read failed: tracker?.moduleResolverHost.useCaseSensitiveFileNames is not a () => boolean; expected () => boolean, found string"},
		{"tracker-helpers-wrong", "true\n", "field read failed: tracker?.moduleResolverHost.useCaseSensitiveFileNames is not a () => boolean; expected () => boolean, found string"},
		{"tracker-missing", "true\n", "field read failed: tracker?.moduleResolverHost.useCaseSensitiveFileNames is not initialized; expected () => boolean, found missing"},
		{"tracker-root-wrong", "true\n", "field read failed: tracker?.moduleResolverHost matches no member of (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) | undefined; expected (ModuleSpecifierResolutionHost & { getCommonSourceDirectory(): string; }) | undefined, found string"},
	} {
		t.Run(test.name, func(t *testing.T) {
			program, _ := intersectionOriginalProgram(t, declarations, test.name, test.source)
			requireIntersectionOriginalFields(t, program, manifest)
			if kind := os.Getenv("ADAMIC_INTERSECTION_ORIGINAL_MUTANT"); kind != "" {
				intersectionOriginalMutant(t, program, kind)
			}
			want := run{stdout: []byte(test.source)}
			if test.diagnostic != "" {
				want = run{exitCode: 70, stderr: []byte("adamic: panic: " + test.diagnostic + "\n")}
			}
			actual, binary := nativelyUncached(t, program)
			if want.exitCode == 0 {
				if report := leaks(t, program, binary); report != "" {
					t.Fatal(report)
				}
			}
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s; got %#v", difference, got)
				}
			}
		})
	}
}

func intersectionOriginalMutant(t *testing.T, program *ir.Program, kind string) {
	t.Helper()
	changed := 0
	for i := range program.ViewContracts {
		c := &program.ViewContracts[i]
		if kind == "skip" && c.Intersection {
			c.Intersection = false
			changed++
		}
		for j, f := range c.Fields {
			if f.Name != "useCaseSensitiveFileNames" {
				continue
			}
			switch kind {
			case "nested":
				c.Fields = append(c.Fields[:j:j], c.Fields[j+1:]...)
				changed++
			case "shape":
				child := program.ViewContracts[f.Contract-1]
				child.Kind = ir.ViewScalar
				child.Of = ir.String
				child.Name = "string"
				c.Fields[j].Contract = ir.ViewContractID(len(program.ViewContracts) + 1)
				program.ViewContracts = append(program.ViewContracts, child)
				changed++
			case "presence":
				c.Fields[j].Optional = true
				changed++
			}
			break
		}
	}
	if kind == "outer" || kind == "absence" {
		changed = changeObjectPrimitiveRead(program, func(p ir.Property) bool { return p.Name == "moduleResolverHost" }, func(p ir.Property) ir.Property {
			if kind == "outer" {
				p.View = ""
			} else {
				p.Absent = false
			}
			return p
		})
	}
	if changed == 0 {
		t.Fatal("original mutant found no obligation")
	}
}
