package formatoptions

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path string, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}

// houseBlock is the house format as nexus 09b88e6 states it.
const houseBlock = `{"tabWidth": 4, "useTabs": false, "semi": true, "singleQuote": true, "printWidth": 120}`

// project writes a repository whose settings extend a Nexus tier carrying block, the shape every one of
// ours has: the project's own file states rules and no format.
func project(t *testing.T, root string, block string) {
	t.Helper()
	writeFile(t, filepath.Join(root, NexusTierFileName), `{"format": `+block+`}`)
	writeFile(t, filepath.Join(root, SettingsFileName), `{"extends": "./`+NexusTierFileName+`", "rules": {}}`)
}

// TestResolveFindsTheNearestSettingsAbove pins inheritance: a nested library with no settings of its
// own formats with its containing repository's options, as libraries/structure does inside ahra.
func TestResolveFindsTheNearestSettingsAbove(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	project(t, root, `{"tabWidth": 4, "printWidth": 120, "singleQuote": true, "bracketSameLine": true}`)
	writeFile(t, filepath.Join(root, "libraries", "structure", "package.json"), `{"name": "structure"}`)

	resolution, err := Resolve(filepath.Join(root, "libraries", "structure", "source"))
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Source != filepath.Join(root, SettingsFileName) {
		t.Fatalf("resolved from %q, want the root %s", resolution.Source, SettingsFileName)
	}
	if resolution.Options.TabWidth != 4 || resolution.Options.PrintWidth != 120 || !resolution.Options.SingleQuote || !resolution.Options.BracketSameLine {
		t.Fatalf("options %+v did not take the Nexus tier's format block", resolution.Options)
	}
}

// TestTheHouseBlockResolvesThroughTheTiers is every repository's real chain: the project extends a
// Structure (or Base) tier, which extends the Nexus tier, and only the Nexus tier holds the block.
func TestTheHouseBlockResolvesThroughTheTiers(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "nexus", NexusTierFileName), `{"rules": {}, "format": `+houseBlock+`}`)
	writeFile(t, filepath.Join(root, "structure", "StructureCohereSettings.json"), `{"extends": "../nexus/`+NexusTierFileName+`", "rules": {}}`)
	writeFile(t, filepath.Join(root, SettingsFileName), `{"extends": "./structure/StructureCohereSettings.json", "rules": {}}`)

	resolution, err := Resolve(filepath.Join(root, "source"))
	if err != nil {
		t.Fatal(err)
	}
	want := PrettierDefaults()
	want.TabWidth = 4
	want.SingleQuote = true
	want.PrintWidth = 120
	if resolution.Options != want {
		t.Fatalf("options %+v, want the house block %+v", resolution.Options, want)
	}
	if resolution.Source != filepath.Join(root, SettingsFileName) {
		t.Fatalf("resolved from %q, want the project's %s", resolution.Source, SettingsFileName)
	}
}

// TestOurTiersRefuseAFormatBlockOutsideTheNexusTier: in a chain that extends a system-inc set,
// formatting is unified, so a format key anywhere but the Nexus tier is a second place the house format
// could drift. It is refused, naming the file to change, whether it restates the house block or differs,
// in the project's own file or in a base of it. The Nexus tier itself is the set cohere carries, which
// always has its block.
func TestOurTiersRefuseAFormatBlockOutsideTheNexusTier(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name    string
		files   map[string]string
		refused string
	}{
		{"the project restating the house block", map[string]string{SettingsFileName: `{"extends": "cohere:system-inc/structure", "format": ` + houseBlock + `}`}, SettingsFileName},
		{"the project's own empty block", map[string]string{SettingsFileName: `{"extends": "cohere:system-inc/base", "format": {}}`}, SettingsFileName},
		{"a base between the project and the set", map[string]string{SettingsFileName: `{"extends": "./shared/Shared.json"}`, "shared/Shared.json": `{"extends": "cohere:system-inc/structure", "format": {"printWidth": 100}}`}, filepath.Join("shared", "Shared.json")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for file, contents := range testCase.files {
				writeFile(t, filepath.Join(root, file), contents)
			}
			resolution, err := Resolve(root)
			if err == nil {
				t.Fatalf("a format block in %s resolved to %+v", testCase.refused, resolution.Options)
			}
			if !strings.Contains(err.Error(), filepath.Join(root, testCase.refused)) || !strings.Contains(err.Error(), "only the Nexus tier") {
				t.Fatalf("refusal %v does not name %s as the file holding a format block outside the Nexus tier", err, testCase.refused)
			}
		})
	}

	root := t.TempDir()
	writeFile(t, filepath.Join(root, SettingsFileName), `{"extends": "cohere:system-inc/structure"}`)
	resolution, err := Resolve(root)
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Options.PrintWidth != 120 || resolution.Options.TabWidth != 4 || !resolution.Options.SingleQuote {
		t.Fatalf("our tier without a project block resolved to %+v, want the house block", resolution.Options)
	}
}

// TestZeroConfigIsTheHouseFormat: no settings anywhere, settings that extend nothing, and a chain of the
// project's own files that never says how to format all format the house way, with the house ignore
// list declared, as cohere:typescript carries them (#bfxz13m).
func TestZeroConfigIsTheHouseFormat(t *testing.T) {
	t.Parallel()
	house, err := houseResolution()
	if err != nil {
		t.Fatal(err)
	}
	if house.Options.TabWidth != 4 || house.Options.PrintWidth != 120 || !house.Options.SingleQuote || !house.HouseIgnoreDeclared {
		t.Fatalf("the house resolution %+v is not cohere:typescript's block", house)
	}

	for name, files := range map[string]map[string]string{
		"settings extending nothing":         {SettingsFileName: `{"rules": {}}`},
		"a chain of the project's own files": {SettingsFileName: `{"extends": "./base/Base.json"}`, "base/Base.json": `{"rules": {}}`},
		"a cohere set that is not ours":      {SettingsFileName: `{"extends": "cohere:react"}`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for file, contents := range files {
				writeFile(t, filepath.Join(root, file), contents)
			}
			resolution, err := Resolve(root)
			if err != nil {
				t.Fatal(err)
			}
			if resolution.Options != house.Options || strings.Join(resolution.HouseIgnore, ",") != strings.Join(house.HouseIgnore, ",") || !resolution.HouseIgnoreDeclared {
				t.Fatalf("resolved %+v, want the house format and list", resolution)
			}
			if resolution.Source != filepath.Join(root, SettingsFileName) {
				t.Fatalf("resolved from %q, want the project's settings", resolution.Source)
			}
		})
	}
}

// TestAnOutsidersOwnBlockIsAppliedOverPrettiersDefaults: outside our tiers the project's own block
// decides, over Prettier's defaults rather than over the house, so `{}` is exactly Prettier's defaults.
// The most derived file with a block wins, and its `ignore` replaces the house list only when it has one.
func TestAnOutsidersOwnBlockIsAppliedOverPrettiersDefaults(t *testing.T) {
	t.Parallel()
	house, err := houseResolution()
	if err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		name   string
		files  map[string]string
		want   func(options *Options)
		ignore string
	}{
		{"an empty block", map[string]string{SettingsFileName: `{"extends": "cohere:typescript", "format": {}}`}, func(*Options) {}, strings.Join(house.HouseIgnore, ",")},
		{"a block of its own", map[string]string{SettingsFileName: `{"format": {"printWidth": 100, "ignore": ["generated/"]}}`}, func(options *Options) { options.PrintWidth = 100 }, "generated/"},
		{"a base's block", map[string]string{SettingsFileName: `{"extends": "./base/Base.json"}`, "base/Base.json": `{"format": {"semi": false}}`}, func(options *Options) { options.Semi = false }, strings.Join(house.HouseIgnore, ",")},
		{"the project's block over its base's", map[string]string{SettingsFileName: `{"extends": "./base/Base.json", "format": {"tabWidth": 8}}`, "base/Base.json": `{"format": {"semi": false}}`}, func(options *Options) { options.TabWidth = 8 }, strings.Join(house.HouseIgnore, ",")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for file, contents := range testCase.files {
				writeFile(t, filepath.Join(root, file), contents)
			}
			resolution, err := Resolve(root)
			if err != nil {
				t.Fatal(err)
			}
			want := PrettierDefaults()
			testCase.want(&want)
			if resolution.Options != want {
				t.Fatalf("options %+v, want %+v", resolution.Options, want)
			}
			if strings.Join(resolution.HouseIgnore, ",") != testCase.ignore || !resolution.HouseIgnoreDeclared {
				t.Fatalf("ignore %v, want %s", resolution.HouseIgnore, testCase.ignore)
			}
		})
	}

	t.Run("an unknown option in an outsider's block", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		settings := filepath.Join(root, SettingsFileName)
		writeFile(t, settings, `{"format": {"quoteProps": "consistent"}}`)
		if _, err := Resolve(root); err == nil || !strings.Contains(err.Error(), settings) {
			t.Fatalf("refusal %v does not name %s", err, settings)
		}
	})
}

// TestAChainRefusesWhatOneFileWouldBeRefusedFor: a bad key in the Nexus tier is refused naming the
// tier, which is the file to change, and a missing base is refused.
func TestAChainRefusesWhatOneFileWouldBeRefusedFor(t *testing.T) {
	t.Parallel()
	t.Run("an unknown option in the Nexus tier", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		nexusTier := filepath.Join(root, "tiers", NexusTierFileName)
		writeFile(t, nexusTier, `{"format": {"quoteProps": "consistent"}}`)
		writeFile(t, filepath.Join(root, SettingsFileName), `{"extends": "./tiers/`+NexusTierFileName+`"}`)
		_, err := Resolve(root)
		if err == nil || !strings.Contains(err.Error(), nexusTier) {
			t.Fatalf("refusal %v does not name the Nexus tier %s", err, nexusTier)
		}
	})

	t.Run("a base that is missing", func(t *testing.T) {
		t.Parallel()
		root := t.TempDir()
		writeFile(t, filepath.Join(root, SettingsFileName), `{"extends": "./tiers/`+NexusTierFileName+`"}`)
		if resolution, err := Resolve(root); err == nil {
			t.Fatalf("a chain naming a missing base resolved to %+v", resolution.Options)
		}
	})
}

// TestALeftoverIsRefusedEvenWhenItAgrees: old Prettier config is refused whatever it says. Agreeing
// with the Nexus tier today is no reason to keep a second statement of the options that nothing reads,
// since nothing would notice the day it stopped agreeing. ahra's move was the last one the tolerance
// served (#dv5ng7g).
func TestALeftoverIsRefusedEvenWhenItAgrees(t *testing.T) {
	t.Parallel()
	for name, leftover := range map[string]struct{ file, contents string }{
		"package.json, the house options":        {"package.json", `{"name": "x", "prettier": {"tabWidth": 4, "singleQuote": true, "printWidth": 120}}`},
		"package.json, with the Tailwind plugin": {"package.json", `{"name": "x", "prettier": {"plugins": ["prettier-plugin-tailwindcss"], "tabWidth": 4, "singleQuote": true, "printWidth": 120, "tailwindFunctions": ["mergeClassNames"]}}`},
		".prettierrc, the house options":         {".prettierrc", `{"tabWidth": 4, "singleQuote": true, "printWidth": 120}`},
		"package.json, an empty prettier key":    {"package.json", `{"name": "x", "prettier": {}}`},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			project(t, root, `{"tabWidth": 4, "singleQuote": true, "printWidth": 120}`)
			writeFile(t, filepath.Join(root, leftover.file), leftover.contents)
			resolution, err := Resolve(root)
			if !errors.Is(err, ErrPrettierConfigRemains) {
				t.Fatalf("a leftover %s resolved to %+v, %v; want it refused", leftover.file, resolution.Options, err)
			}
			if !strings.Contains(err.Error(), filepath.Join(root, leftover.file)) || !strings.Contains(err.Error(), "delete it") {
				t.Fatalf("refusal %v does not name %s and say to delete it", err, leftover.file)
			}
		})
	}
}

// TestALeftoverBelowTheSettingsIsRefused: the walk up refuses old config in every directory it passes,
// not only beside the settings, so a package in a workspace cannot keep its own.
func TestALeftoverBelowTheSettingsIsRefused(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	project(t, root, houseBlock)
	writeFile(t, filepath.Join(root, "packages", "inner", "package.json"), `{"name": "inner", "prettier": {"printWidth": 120}}`)
	if _, err := Resolve(filepath.Join(root, "packages", "inner", "source")); !errors.Is(err, ErrPrettierConfigRemains) {
		t.Fatalf("a leftover below the settings was not refused: %v", err)
	}
}

// TestNothingConfiguredIsTheHouseNotPrettier: a tree with no settings anywhere above it is zero config,
// which formats the house way, not with Prettier's defaults.
func TestNothingConfiguredIsTheHouseNotPrettier(t *testing.T) {
	t.Parallel()
	resolution, err := Resolve(t.TempDir())
	if err != nil {
		t.Skipf("something above the temp directory configures formatting (%v), so this machine cannot test the empty case", err)
	}
	if resolution.Source != "" {
		t.Skipf("settings exist above the temp directory (%s), so this machine cannot test the empty case", resolution.Source)
	}
	if resolution.Options.TabWidth != 4 || resolution.Options.PrintWidth != 120 || !resolution.HouseIgnoreDeclared {
		t.Fatalf("nothing configured resolved to %+v, want the house format and list", resolution)
	}
}

// TestResolveRefusesEveryOptionNobodyChose covers each refusal, because each one replaces a format run
// with options nobody chose: Prettier's width 80 where the repository meant 120.
func TestResolveRefusesEveryOptionNobodyChose(t *testing.T) {
	t.Parallel()
	extendsNexus := `{"extends": "./` + NexusTierFileName + `"}`
	for name, testCase := range map[string]struct {
		files    map[string]string
		leftover bool
	}{
		"unknown option":             {files: map[string]string{SettingsFileName: extendsNexus, NexusTierFileName: `{"format": {"quoteProps": "consistent"}}`}},
		"plugin key in the block":    {files: map[string]string{SettingsFileName: extendsNexus, NexusTierFileName: `{"format": {"plugins": ["prettier-plugin-tailwindcss"]}}`}},
		"crlf":                       {files: map[string]string{SettingsFileName: extendsNexus, NexusTierFileName: `{"format": {"endOfLine": "crlf"}}`}},
		"package.json":               {files: map[string]string{SettingsFileName: extendsNexus, NexusTierFileName: `{"format": {"tabWidth": 4}}`, "package.json": `{"prettier": {"tabWidth": 2}}`}, leftover: true},
		".prettierrc":                {files: map[string]string{SettingsFileName: extendsNexus, NexusTierFileName: `{"format": {"tabWidth": 4}}`, ".prettierrc": `{"printWidth": 100, "tabWidth": 4}`}, leftover: true},
		"old config and no settings": {files: map[string]string{"package.json": `{"prettier": {"tabWidth": 4}}`}, leftover: true},
		"javascript config":          {files: map[string]string{SettingsFileName: extendsNexus, NexusTierFileName: `{"format": {}}`, "prettier.config.js": `module.exports = {}`}, leftover: true},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for file, contents := range testCase.files {
				writeFile(t, filepath.Join(root, file), contents)
			}
			resolution, err := Resolve(root)
			if err == nil {
				t.Fatalf("resolved %+v from %s, want a refusal", resolution.Options, resolution.Source)
			}
			if errors.Is(err, ErrPrettierConfigRemains) != testCase.leftover {
				t.Fatalf("refusal %v: leftover-config sentinel %v, want %v", err, errors.Is(err, ErrPrettierConfigRemains), testCase.leftover)
			}
		})
	}
}

// TestALeftoverIsNamedWithItsFile: a refusal a reader can act on says which file to change.
func TestALeftoverIsNamedWithItsFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	project(t, root, `{"tabWidth": 4}`)
	writeFile(t, filepath.Join(root, "package.json"), `{"name": "x", "prettier": {"tabWidth": 2}}`)
	_, err := Resolve(root)
	if err == nil || !strings.Contains(err.Error(), filepath.Join(root, "package.json")) {
		t.Fatalf("refusal %v does not name the package.json carrying the old options", err)
	}
}

// TestAPackageJSONWithoutPrettierIsNotALeftover: almost every repository has a package.json, and only
// its "prettier" key is old config.
func TestAPackageJSONWithoutPrettierIsNotALeftover(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	project(t, root, `{"printWidth": 120}`)
	writeFile(t, filepath.Join(root, "package.json"), `{"name": "x", "devDependencies": {}}`)
	resolution, err := Resolve(root)
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Options.PrintWidth != 120 {
		t.Fatalf("print width %d, want the Nexus tier's 120", resolution.Options.PrintWidth)
	}
}

// TestTheHouseIgnoreListRidesInTheFormatBlock: the block's `ignore` is the house list, not a printing
// option. It is read beside the options, declared only when written, and a value that is not a list of
// patterns is refused naming the Nexus tier.
func TestTheHouseIgnoreListRidesInTheFormatBlock(t *testing.T) {
	t.Parallel()
	declared := t.TempDir()
	project(t, declared, `{"tabWidth": 4, "ignore": ["pnpm-lock.yaml", "*.sqlite"]}`)
	resolution, err := Resolve(declared)
	if err != nil {
		t.Fatal(err)
	}
	if !resolution.HouseIgnoreDeclared || strings.Join(resolution.HouseIgnore, ",") != "pnpm-lock.yaml,*.sqlite" || resolution.Options.TabWidth != 4 {
		t.Fatalf("resolved %+v, want the house list beside tab width 4", resolution)
	}

	// An outsider's block without the key keeps the house list, cohere:typescript's.
	undeclared := t.TempDir()
	project(t, undeclared, `{"tabWidth": 4}`)
	house, err := houseResolution()
	if err != nil {
		t.Fatal(err)
	}
	if resolution, err := Resolve(undeclared); err != nil || !resolution.HouseIgnoreDeclared || strings.Join(resolution.HouseIgnore, ",") != strings.Join(house.HouseIgnore, ",") {
		t.Fatalf("a block without the key did not keep the house list: %+v, %v", resolution, err)
	}

	malformed := t.TempDir()
	project(t, malformed, `{"tabWidth": 4, "ignore": "pnpm-lock.yaml"}`)
	if _, err := Resolve(malformed); err == nil || !strings.Contains(err.Error(), NexusTierFileName) {
		t.Fatalf("an ignore that is not a list was accepted or not named: %v", err)
	}

}

// TestIgnorePatternsResolveWithTheOptions: the walk takes the chain's ignorePatterns from the same
// resolution as its options, base first, so it reads the list lint reads.
func TestIgnorePatternsResolveWithTheOptions(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "nexus", NexusTierFileName), `{"rules": {}, "ignorePatterns": ["**/generated/**"], "format": `+houseBlock+`}`)
	writeFile(t, filepath.Join(root, SettingsFileName), `{"extends": "./nexus/`+NexusTierFileName+`", "rules": {}, "ignorePatterns": ["data/**"]}`)
	resolution, err := Resolve(root)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(resolution.IgnorePatterns, ",") != "**/generated/**,data/**" {
		t.Fatalf("ignorePatterns %v, want the base's first", resolution.IgnorePatterns)
	}
}

// TestADroppedKeyCanNeverReachPrettiersDefaults: the move's failure is a repository whose prettier key
// went and whose options never arrived, formatting at width 80 with nothing said. Settings that say
// nothing about formatting format the house way, and old config with no settings above it is refused.
func TestADroppedKeyCanNeverReachPrettiersDefaults(t *testing.T) {
	t.Parallel()
	withoutBlock := t.TempDir()
	writeFile(t, filepath.Join(withoutBlock, SettingsFileName), `{"rules": {"no-debugger": "error"}}`)
	writeFile(t, filepath.Join(withoutBlock, "package.json"), `{"name": "x"}`)
	resolution, err := Resolve(withoutBlock)
	if err != nil {
		t.Fatal(err)
	}
	if resolution.Options == PrettierDefaults() || resolution.Options.PrintWidth != 120 {
		t.Fatalf("settings that say nothing about formatting resolved to %+v, want the house format", resolution.Options)
	}

	neverMoved := t.TempDir()
	writeFile(t, filepath.Join(neverMoved, "package.json"), `{"name": "x", "prettier": {"tabWidth": 4, "printWidth": 120}}`)
	if resolution, err := Resolve(neverMoved); err == nil {
		t.Fatalf("old config with no settings resolved to %+v", resolution.Options)
	}
}

// resolveUnremembered is Resolve as it was before the Resolver: every directory walks to its settings
// and reads the chain afresh. The reference the remembered path is held to.
func resolveUnremembered(directory string) (Resolution, error) {
	for current := directory; ; current = filepath.Dir(current) {
		leftover, found, err := prettierConfigIn(current)
		if err != nil {
			return Resolution{}, err
		}
		if found {
			return Resolution{}, fmt.Errorf("%s: %w; delete it, since cohere reads format options only from a \"format\" block in %s",
				leftover, ErrPrettierConfigRemains, SettingsFileName)
		}
		path := filepath.Join(current, SettingsFileName)
		if _, err := os.Stat(path); err == nil {
			return resolveChain(path)
		}
		if parent := filepath.Dir(current); parent == current {
			return houseResolution()
		}
	}
}

// TestAResolverAnswersAsResolveDidOnOurTrees: one Resolver, shared across every directory of a real
// tree as a format run shares it, answers each directory exactly as the unremembered walk does:
// options, source, both ignore lists, and every refusal word for word (ahra's projects/www-ahra-ai,
// with its prettier key, is one). Reads files as text only.
//
// Off unless COHERE_RESOLVE_TREES lists the trees, separated as PATH is: the reference walks every
// directory afresh, and on 2026-10-04 ahra's 90,798 directories, www's 1,075 and api's 835 took 317s,
// all identical, 437 of them refusals.
func TestAResolverAnswersAsResolveDidOnOurTrees(t *testing.T) {
	t.Parallel()
	trees := filepath.SplitList(os.Getenv("COHERE_RESOLVE_TREES"))
	if len(trees) == 0 {
		t.Skip("COHERE_RESOLVE_TREES names no tree")
	}
	skipped := map[string]bool{".git": true, "node_modules": true, ".cache": true, ".next": true, "data": true, "dist": true}
	for _, root := range trees {
		t.Run(root, func(t *testing.T) {
			t.Parallel()
			resolver := NewResolver()
			directories, refusals := 0, 0
			walkError := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
				if err != nil || !entry.IsDir() {
					return nil
				}
				if skipped[entry.Name()] {
					return filepath.SkipDir
				}
				directories++
				remembered, rememberedError := resolver.Resolve(path)
				reference, referenceError := resolveUnremembered(path)
				if fmt.Sprint(rememberedError) != fmt.Sprint(referenceError) {
					t.Fatalf("%s: the Resolver answered %v, the walk %v", path, rememberedError, referenceError)
				}
				if referenceError != nil {
					refusals++
					return nil
				}
				if fmt.Sprintf("%+v", remembered) != fmt.Sprintf("%+v", reference) {
					t.Fatalf("%s: the Resolver answered %+v, the walk %+v", path, remembered, reference)
				}
				return nil
			})
			if walkError != nil {
				t.Fatal(walkError)
			}
			if directories < 50 {
				t.Fatalf("compared only %d directories under %s, so the comparison measured nothing", directories, root)
			}
			t.Logf("%d directories identical, %d of them refusals", directories, refusals)
		})
	}
}

// TestAnEditAnywhereInTheChainIsSeen: a remembered chain is used only while its files read what they
// did, so a later run sees an edit to the Nexus tier and to the project's own `extends`, even one that
// leaves the file's size unchanged.
func TestAnEditAnywhereInTheChainIsSeen(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	nexusTier := filepath.Join(root, "nexus", NexusTierFileName)
	writeFile(t, nexusTier, `{"format": {"printWidth": 120}}`)
	writeFile(t, filepath.Join(root, "other", NexusTierFileName), `{"format": {"printWidth": 90}}`)
	settings := filepath.Join(root, SettingsFileName)
	writeFile(t, settings, `{"extends": "./nexus/`+NexusTierFileName+`"}`)

	width := func() int {
		t.Helper()
		resolution, err := NewResolver().Resolve(filepath.Join(root, "source"))
		if err != nil {
			t.Fatal(err)
		}
		return resolution.Options.PrintWidth
	}
	if got := width(); got != 120 {
		t.Fatalf("print width %d, want the Nexus tier's 120", got)
	}

	writeFile(t, nexusTier, `{"format": {"printWidth": 100}}`)
	if got := width(); got != 100 {
		t.Fatalf("after the Nexus tier changed to 100 at the same size, a run read %d", got)
	}

	writeFile(t, settings, `{"extends": "./other/`+NexusTierFileName+`"}`)
	if got := width(); got != 90 {
		t.Fatalf("after the project's extends moved to another tier, a run read %d, want 90", got)
	}
}

// TestARememberedParentDoesNotAnswerForAChildWithLeftoverConfig: a Resolver that already knows a
// directory's options still walks a child that holds old Prettier config, and refuses it.
func TestARememberedParentDoesNotAnswerForAChildWithLeftoverConfig(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	project(t, root, houseBlock)
	writeFile(t, filepath.Join(root, "packages", "inner", ".prettierrc"), `{"printWidth": 120}`)

	resolver := NewResolver()
	if _, err := resolver.Resolve(filepath.Join(root, "packages")); err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.Resolve(filepath.Join(root, "packages", "inner", "source")); !errors.Is(err, ErrPrettierConfigRemains) {
		t.Fatalf("a child with a leftover resolved through its remembered parent: %v", err)
	}
}
