package configuration

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

/*
 * Zero config is the house stack (Kirk, 2026-10-04, #bfxz13m).
 *
 * A project with no CohereSettings.json, or one whose chain names no `cohere:` set, gets every set we
 * pick, each where the program shows it fits: cohere:typescript on every file, cohere:react on a file
 * that imports react or holds JSX, cohere:next on a file that imports next or is one of Next's own
 * files, and cohere:tailwind on every file when the project's stylesheet is Tailwind's. The evidence is
 * the program, never package.json: a dependency the code never imports applies nothing.
 *
 * Each file resolves through an ordinary configuration, one per combination of the per-file sets, so
 * every rule of a chain (departures, overrides, options, plugin defaults) means in zero config exactly
 * what it means in an `extends`. A rule a set the file does not get would have turned on reads as off
 * for that file, with the set's reason, so coverage says why it did not run rather than that nobody
 * configured it.
 */

// The house sets zero config applies, by name.
const (
	TypeScriptSetName = SetPrefix + "typescript"
	ReactSetName      = SetPrefix + "react"
	NextSetName       = SetPrefix + "next"
	TailwindSetName   = SetPrefix + "tailwind"
)

// HouseDetection is what the program showed about which house sets fit where, gathered by the command
// (internal/lint/housesets) and handed to LoadHouse.
type HouseDetection struct {
	// ReactFiles and NextFiles are the absolute file names each per-file set applies to.
	ReactFiles map[string]bool
	NextFiles  map[string]bool

	// TailwindEntryPoint is the stylesheet that makes the project Tailwind's, or empty when there is
	// none, and then TailwindSkipped says why, in a clause.
	TailwindEntryPoint string
	TailwindSkipped    string
}

// AppliedSet is one house set and the evidence it was applied on, for the run's first line.
type AppliedSet struct {
	Name     string
	Evidence string
	// Applied is false for a set zero config looked for and found no evidence of, which is named rather
	// than left out, so "no file imports react" reads as a finding about the project.
	Applied bool
}

// houseVariant is which per-file sets a file gets.
type houseVariant struct {
	react bool
	next  bool
}

// houseSets is a zero-config configuration's per-file dispatch.
type houseSets struct {
	detection HouseDetection
	variants  map[houseVariant]*Config
	applied   []AppliedSet
	// gatedReasons is why each rule a per-file set turns on is off where that set does not apply.
	gatedReasons map[string]OffReason
}

// IsZeroConfig reports whether the configuration at path takes the house stack: there is no file
// there, or its chain names no set cohere carries.
//
// A chain that names any `cohere:` set chose its sets, and is loaded as written.
func IsZeroConfig(path string) (bool, error) {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return true, nil
	}
	sources, err := SourcesOf(path)
	if err != nil {
		return false, err
	}
	for _, source := range sources {
		if IsSet(source) {
			return false, nil
		}
	}
	return true, nil
}

// HouseSources is every source a zero-config configuration at path reads: the house sets, then the
// project's own file and what it extends when there is one. Anything keyed on a configuration's bytes
// reads these, so an edit to the project's file still changes the key.
func HouseSources(path string) ([]string, error) {
	sources := []string{}
	if _, err := os.Stat(path); err == nil {
		own, err := SourcesOf(path)
		if err != nil {
			return nil, err
		}
		sources = append(sources, own...)
	}
	return append(sources, TailwindSetName, NextSetName, ReactSetName, TypeScriptSetName), nil
}

// OwnSettingsOf returns the `settings` the project's own file at path writes, which zero config's
// detection reads before LoadHouse runs, or nil when there is no file or it writes none.
func OwnSettingsOf(path string) (map[string]json.RawMessage, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, nil
	}
	layers, err := readConfigLayers(path, nil)
	if err != nil {
		return nil, err
	}
	return layers[len(layers)-1].raw.Settings, nil
}

// LoadHouse loads the zero-config configuration for the project whose settings would live at path,
// applying each house set where detection shows it fits, with the project's own file, if any, on top.
//
// The returned configuration is the whole house stack: its rules are every rule any set asks for,
// which is what parity, coverage and selector validation read. Resolve answers per file from the
// variant that file gets.
func LoadHouse(path string, registeredNames []string, detection HouseDetection) (*Config, error) {
	root, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("resolving the config directory for %s: %w", path, err)
	}
	var ownLayers []configLayer
	if _, err := os.Stat(path); err == nil {
		ownLayers, err = readConfigLayers(path, nil)
		if err != nil {
			return nil, err
		}
	}

	tailwind := detection.TailwindEntryPoint != ""
	whole, err := loadHouseVariant(ownLayers, root, registeredNames, houseVariant{react: true, next: true}, true)
	if err != nil {
		return nil, err
	}
	house := &houseSets{
		detection:    detection,
		variants:     map[houseVariant]*Config{},
		gatedReasons: map[string]OffReason{},
	}
	for _, variant := range []houseVariant{{}, {react: true}, {next: true}, {react: true, next: true}} {
		loaded, err := loadHouseVariant(ownLayers, root, registeredNames, variant, tailwind)
		if err != nil {
			return nil, err
		}
		// The house ignores what it ignores for every file, whichever set wrote the pattern.
		loaded.IgnorePatterns = whole.IgnorePatterns
		house.variants[variant] = loaded
	}

	house.applied = []AppliedSet{
		{Name: TypeScriptSetName, Evidence: "every file", Applied: true},
		perFileSet(ReactSetName, len(detection.ReactFiles), "import react or contain JSX"),
		perFileSet(NextSetName, len(detection.NextFiles), "import next or are Next's own files"),
	}
	if tailwind {
		house.applied = append(house.applied, AppliedSet{Name: TailwindSetName, Applied: true,
			Evidence: "every file, since " + relativeTo(root, detection.TailwindEntryPoint) + " is a Tailwind stylesheet"})
	} else {
		house.applied = append(house.applied, AppliedSet{Name: TailwindSetName, Evidence: detection.TailwindSkipped})
	}

	// Every rule the whole stack would run, padded into each variant that lacks it as an off, so a file
	// a set does not fit reads that set's rules as off for a reason rather than as unconfigured.
	setReasons := map[string]string{
		ReactSetName:    "cohere:react applies only to files that import react or contain JSX",
		NextSetName:     "cohere:next applies only to files that import next or are Next's own files",
		TailwindSetName: "cohere:tailwind is not applied: " + detection.TailwindSkipped,
	}
	wholeKeys := whole.RuleKeys()
	for _, variant := range house.variants {
		have := map[string]bool{}
		for _, key := range variant.RuleKeys() {
			have[key] = true
		}
		for _, key := range wholeKeys {
			if have[key] {
				continue
			}
			variant.Rules[key] = RuleSetting{Severity: SeverityOff}
			if _, reasoned := house.gatedReasons[key]; !reasoned {
				house.gatedReasons[key] = OffReason{File: whole.ruleWriters[key], Reason: setReasons[whole.ruleWriters[key]]}
			}
		}
	}
	whole.house = house
	return whole, nil
}

func perFileSet(name string, files int, evidence string) AppliedSet {
	if files == 0 {
		return AppliedSet{Name: name, Evidence: "no file " + singularVerb(1, evidence)}
	}
	noun := "files that"
	if files == 1 {
		noun = "file that"
	}
	return AppliedSet{Name: name, Applied: true, Evidence: fmt.Sprintf("%d %s %s", files, noun, singularVerb(files, evidence))}
}

// singularVerb turns "import react or contain JSX" into "imports react or contains JSX" for one file.
func singularVerb(files int, evidence string) string {
	if files != 1 {
		return evidence
	}
	return strings.NewReplacer("import ", "imports ", "contain ", "contains ", "are ", "is one of ").Replace(evidence)
}

func relativeTo(root string, path string) string {
	if relative, err := filepath.Rel(root, path); err == nil && !strings.HasPrefix(relative, "..") {
		return filepath.ToSlash(relative)
	}
	return path
}

// loadHouseVariant loads cohere:typescript, the per-file sets variant names, cohere:tailwind when
// tailwind is true, and the project's own layers on top, each of which extends every set so a rule it
// writes replaces the set's ruling rather than colliding with it.
func loadHouseVariant(ownLayers []configLayer, root string, registeredNames []string, variant houseVariant, tailwind bool) (*Config, error) {
	sets := []string{TypeScriptSetName}
	if variant.react {
		sets = append(sets, ReactSetName)
	}
	if variant.next {
		sets = append(sets, NextSetName)
	}
	if tailwind {
		sets = append(sets, TailwindSetName)
	}
	var layers []configLayer
	read := map[string]bool{}
	for _, set := range sets {
		setLayers, err := readConfigLayers(set, nil)
		if err != nil {
			return nil, err
		}
		for _, setLayer := range setLayers {
			if !read[setLayer.path] {
				read[setLayer.path] = true
				layers = append(layers, setLayer)
			}
		}
	}
	for _, ownLayer := range ownLayers {
		ownLayer.extended = append(append([]string(nil), ownLayer.extended...), sets...)
		layers = append(layers, ownLayer)
	}
	return loadLayers(layers, root, registeredNames)
}

// Applied is the house sets a zero-config run applied, and the ones it looked for and did not, with
// the evidence for each. Nil for a configuration that names its sets.
func (c *Config) Applied() []AppliedSet {
	if c == nil || c.house == nil {
		return nil
	}
	return c.house.applied
}

// SetsLine is the run's first line under zero config: each house set and why it applies or does not.
// Empty for a configuration that names its own sets, since its file already says.
func (c *Config) SetsLine() string {
	applied := c.Applied()
	if len(applied) == 0 {
		return ""
	}
	terms := make([]string, 0, len(applied))
	for _, set := range applied {
		if set.Applied {
			terms = append(terms, set.Name+" on "+set.Evidence)
			continue
		}
		terms = append(terms, set.Name+" not applied, "+set.Evidence)
	}
	return "sets: " + strings.Join(terms, "; ")
}

// DetectionFingerprint is a digest of which files each per-file set applied to and the Tailwind entry,
// for a cache keyed on what decides which rules run: a file can change sets without its own bytes
// changing, when another file starts importing next. Empty for a configuration that names its sets.
func (c *Config) DetectionFingerprint() string {
	if c == nil || c.house == nil {
		return ""
	}
	digest := sha256.New()
	for _, files := range []map[string]bool{c.house.detection.ReactFiles, c.house.detection.NextFiles} {
		names := make([]string, 0, len(files))
		for name := range files {
			names = append(names, name)
		}
		sort.Strings(names)
		fmt.Fprintf(digest, "%d\n%s\n", len(names), strings.Join(names, "\n"))
	}
	fmt.Fprintf(digest, "tailwind %s\n", c.house.detection.TailwindEntryPoint)
	return fmt.Sprintf("%x", digest.Sum(nil))
}

// variantFor is the configuration a file resolves through under zero config.
func (house *houseSets) variantFor(path string) *Config {
	return house.variants[houseVariant{react: house.detection.ReactFiles[path], next: house.detection.NextFiles[path]}]
}
