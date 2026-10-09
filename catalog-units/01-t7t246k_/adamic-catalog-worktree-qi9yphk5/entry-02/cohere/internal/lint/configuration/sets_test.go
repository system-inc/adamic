package configuration

import (
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// The rule sets cohere carries (#njhfftt): a configuration names one as `cohere:<name>`, and `extends`
// takes a list, so a project composes sets that each sit on the same base.

// withSetFiles replaces the embedded sets with files for one test.
func withSetFiles(t *testing.T, files map[string]string) {
	t.Helper()
	replacement := fstest.MapFS{}
	for name, contents := range files {
		replacement["sets/"+name+".json"] = &fstest.MapFile{Data: []byte(contents)}
	}
	original := setFiles
	setFiles = replacement
	t.Cleanup(func() { setFiles = original })
}

// rootSets are the sets that extend nothing: cohere:adamic, the soundness rules, and cohere:style, the
// taste rules, which cohere:typescript composes (#drbrp8c). Every other set sits on cohere:typescript.
var rootSets = []string{SetPrefix + "adamic", SetPrefix + "style"}

func TestEverySetCohereCarriesLoads(t *testing.T) {
	t.Parallel()
	names := SetNames()
	if len(names) < 3 {
		t.Fatalf("cohere carries %d sets (%v); the embedded tree did not load", len(names), names)
	}
	for _, name := range names {
		directory := writeConfigs(t, map[string]string{"CohereSettings.json": `{"extends": "` + name + `"}`})
		loaded := loadOrFail(t, filepath.Join(directory, "CohereSettings.json"))
		if len(loaded.Rules) == 0 {
			t.Errorf("%s loaded with no rules", name)
		}
		chain := loaded.Sources[1:]
		if slices.Contains(rootSets, name) {
			if !reflect.DeepEqual(chain, []string{name}) {
				t.Errorf("%s is a root set and extends nothing, but its sources are %v", name, loaded.Sources)
			}
			continue
		}
		if !slices.Contains(chain, SetPrefix+"typescript") {
			t.Errorf("%s does not sit on cohere:typescript: its sources are %v", name, loaded.Sources)
		}
		for _, root := range rootSets {
			if !slices.Contains(chain, root) {
				t.Errorf("%s does not reach %s: its sources are %v", name, root, loaded.Sources)
			}
		}
	}
}

// cohere:style and the cohere:system-inc/ tiers each resolve to their own set (#qnjvzh3). The lookup
// is exact after the cohere: prefix, so no set name reaches another that it begins like.
func TestCohereStyleAndTheSystemIncTiersResolveToThemselves(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"cohere:style", "cohere:system-inc/structure", "cohere:system-inc/base"} {
		directory := writeConfigs(t, map[string]string{"CohereSettings.json": `{"extends": "` + name + `"}`})
		loaded := loadOrFail(t, filepath.Join(directory, "CohereSettings.json"))
		if loaded.Sources[1] != name {
			t.Errorf("%s resolved to %s first: its sources are %v", name, loaded.Sources[1], loaded.Sources)
		}
	}
	// cohere:style extends nothing, and each tier reaches it only through cohere:typescript.
	directory := writeConfigs(t, map[string]string{"CohereSettings.json": `{"extends": "cohere:style"}`})
	if sources := loadOrFail(t, filepath.Join(directory, "CohereSettings.json")).Sources; len(sources) != 2 {
		t.Errorf("cohere:style should load alone, and its sources are %v", sources)
	}
	// A name that doesn't exist is refused rather than matched to a set it begins like.
	if _, err := SourceContents("cohere:system"); err == nil {
		t.Error("cohere:system, which is no set, resolved to something")
	}
}

// Not parallel: it swaps the package's setFiles for its own sets through withSetFiles, which every test
// that reads a set would see.
func TestASetIsReadFromTheBinaryAndNamedAsASet(t *testing.T) {
	withSetFiles(t, map[string]string{
		"typescript": `{"rules": {"no-var": "error"}}`,
		"react":      `{"extends": "cohere:typescript", "rules": {"react/no-danger": "error"}}`,
	})
	directory := writeConfigs(t, map[string]string{"CohereSettings.json": `{"extends": "cohere:react"}`})
	loaded := loadOrFail(t, filepath.Join(directory, "CohereSettings.json"))

	for _, name := range []string{"no-var", "react/no-danger"} {
		if loaded.Rules[name].Severity != SeverityError {
			t.Errorf("%s is %v; the set chain must reach it", name, loaded.Rules[name].Severity)
		}
	}
	want := []string{filepath.Join(directory, "CohereSettings.json"), "cohere:react", "cohere:typescript"}
	if !reflect.DeepEqual(loaded.Sources, want) {
		t.Errorf("sources are %v, want %v", loaded.Sources, want)
	}
	if onDisk := SourcesOnDisk(loaded.Sources); !reflect.DeepEqual(onDisk, want[:1]) {
		t.Errorf("sources on disk are %v, want only the project's file", onDisk)
	}
}

func TestAListReadsASharedBaseOnceAsTheOutermostLayer(t *testing.T) {
	t.Parallel()
	directory := writeConfigs(t, map[string]string{
		"base.json":           `{"rules": {"no-var": "error"}}`,
		"react.json":          `{"extends": "./base.json", "rules": {"react/no-danger": "error"}}`,
		"next.json":           `{"extends": "./base.json", "rules": {"next/no-img-element": "error"}}`,
		"CohereSettings.json": `{"extends": ["./react.json", "./next.json"]}`,
	})
	loaded := loadOrFail(t, filepath.Join(directory, "CohereSettings.json"))

	for _, name := range []string{"no-var", "react/no-danger", "next/no-img-element"} {
		if loaded.Rules[name].Severity != SeverityError {
			t.Errorf("%s is %v; every listed chain must reach the project", name, loaded.Rules[name].Severity)
		}
	}
	// Outermost last: the project, then the list in reverse of how it applies, then the shared base,
	// once. A base read twice would apply again after react and silently undo anything react set.
	want := []string{"CohereSettings.json", "next.json", "react.json", "base.json"}
	if len(loaded.Sources) != len(want) {
		t.Fatalf("sources are %v, want %v", loaded.Sources, want)
	}
	for index, source := range loaded.Sources {
		if filepath.Base(source) != want[index] {
			t.Fatalf("sources are %v, want %v", loaded.Sources, want)
		}
	}
}

func TestARuleTwoUnrelatedLayersConfigureIsRefused(t *testing.T) {
	t.Parallel()
	// Baseline first: the same composition, each rule in one set, loads. Without it the refusal
	// below could be coming from anything about the shape.
	disjoint := writeConfigs(t, map[string]string{
		"base.json":           `{"rules": {"no-var": "error"}}`,
		"react.json":          `{"extends": "./base.json", "rules": {"eqeqeq": "error"}}`,
		"next.json":           `{"extends": "./base.json", "rules": {"no-eval": "error"}}`,
		"CohereSettings.json": `{"extends": ["./react.json", "./next.json"]}`,
	})
	loadOrFail(t, filepath.Join(disjoint, "CohereSettings.json"))

	// Both orders, so the refusal is not an accident of which set the list names first.
	for _, order := range []string{`["./react.json", "./next.json"]`, `["./next.json", "./react.json"]`} {
		directory := writeConfigs(t, map[string]string{
			"base.json":           `{"rules": {"no-var": "error"}}`,
			"react.json":          `{"extends": "./base.json", "rules": {"eqeqeq": "error"}}`,
			"next.json":           `{"extends": "./base.json", "rules": {"eqeqeq": "warn"}}`,
			"CohereSettings.json": `{"extends": ` + order + `}`,
		})
		refusedWith(t, filepath.Join(directory, "CohereSettings.json"), `"eqeqeq" is configured by both`)
	}

	// A file that extends both may still set the rule: it extends each of them, so it departs as any
	// project does, with a reason.
	departing := writeConfigs(t, map[string]string{
		"base.json":  `{"rules": {"no-var": "error"}}`,
		"react.json": `{"extends": "./base.json", "rules": {"eqeqeq": "error"}}`,
		"next.json":  `{"extends": "./base.json", "rules": {"no-eval": "error"}}`,
		"CohereSettings.json": `{"extends": ["./react.json", "./next.json"], "rules": {"eqeqeq": "off"},
			"departures": {"eqeqeq": "this project compares loosely on purpose"}}`,
	})
	loaded := loadOrFail(t, filepath.Join(departing, "CohereSettings.json"))
	if loaded.Rules["eqeqeq"].Severity != SeverityOff {
		t.Errorf("eqeqeq is %v; the project extending both sets may set it", loaded.Rules["eqeqeq"].Severity)
	}
}

func TestAnUnknownSetIsRefusedNamingTheSetsThereAre(t *testing.T) {
	t.Parallel()
	directory := writeConfigs(t, map[string]string{"CohereSettings.json": `{"extends": "cohere:typescirpt"}`})
	refusedWith(t, filepath.Join(directory, "CohereSettings.json"), "the sets are cohere:")
	_, err := Load(filepath.Join(directory, "CohereSettings.json"))
	if err == nil || !strings.Contains(err.Error(), "cohere:typescript") {
		t.Errorf("the refusal must name the set that was meant: %v", err)
	}
}

// Not parallel: it swaps the package's setFiles for its own sets through withSetFiles, which every test
// that reads a set would see.
func TestASetMayExtendOnlySets(t *testing.T) {
	withSetFiles(t, map[string]string{
		"typescript": `{"rules": {"no-var": "error"}}`,
		"react":      `{"extends": "./typescript.json", "rules": {"eqeqeq": "error"}}`,
	})
	directory := writeConfigs(t, map[string]string{"CohereSettings.json": `{"extends": "cohere:react"}`})
	refusedWith(t, filepath.Join(directory, "CohereSettings.json"), "a set may extend only other sets")
}

func TestExtendsTakesOneSourceOrAListOfThem(t *testing.T) {
	t.Parallel()
	for _, extends := range []string{`""`, `[]`} {
		directory := writeConfigs(t, map[string]string{"CohereSettings.json": `{"extends": ` + extends + `, "rules": {"no-var": "error"}}`})
		loaded := loadOrFail(t, filepath.Join(directory, "CohereSettings.json"))
		if len(loaded.Sources) != 1 {
			t.Errorf("extends %s read %v; it names nothing", extends, loaded.Sources)
		}
	}
	for extends, want := range map[string]string{
		`[""]`:     "lists an empty entry",
		`3`:        "names a path or a rule set, or a list of them",
		`{"a": 1}`: "names a path or a rule set, or a list of them",
		`["a", 2]`: "names a path or a rule set, or a list of them",
	} {
		directory := writeConfigs(t, map[string]string{"CohereSettings.json": `{"extends": ` + extends + `}`})
		refusedWith(t, filepath.Join(directory, "CohereSettings.json"), want)
	}
}
