package settingsschema

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/format/formatoptions"
	"github.com/system-inc/cohere/internal/lint/configuration"
	"github.com/system-inc/cohere/internal/lint/registry"
)

// moduleRoot is the directory the generated paths are relative to: go test runs in the package's own
// directory, two below it.
func moduleRoot(t *testing.T) string {
	t.Helper()
	return filepath.Join("..", "..")
}

func built(t *testing.T) map[string][]byte {
	t.Helper()
	files, err := Build(registry.Names())
	if err != nil {
		t.Fatal(err)
	}
	return files
}

// TestTheCommittedFilesAreCurrent: the schemas and the reference in schema/ are what Build produces now.
// A rule renamed, a key added to the loader, a format option added: each changes Build's output, and the
// committed files must follow.
func TestTheCommittedFilesAreCurrent(t *testing.T) {
	t.Parallel()
	for path, contents := range built(t) {
		committed, err := os.ReadFile(filepath.Join(moduleRoot(t), path))
		if err != nil {
			t.Fatalf("%s: %v; run go run ./internal/settingsschema/tools/generate", path, err)
		}
		if !bytes.Equal(committed, contents) {
			t.Errorf("%s is stale; run go run ./internal/settingsschema/tools/generate", path)
		}
	}
}

// writeSettings writes files into a fresh directory and returns the path of its CohereSettings.json.
func writeSettings(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	for name, contents := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(root, "CohereSettings.json")
}

// keyFixtures is a settings file using each top-level key, valid as the loader reads it. departures needs
// a base the rule departs from, and extends one to extend.
var keyFixtures = map[string]map[string]string{
	"$schema":        {"CohereSettings.json": `{"$schema": "./schema/CohereSettings.schema.json"}`},
	"extends":        {"Base.json": `{"rules": {}}`, "CohereSettings.json": `{"extends": ["cohere:typescript", "./Base.json"]}`},
	"rules":          {"CohereSettings.json": `{"rules": {"no-debugger": "error", "eqeqeq": ["error", "always"]}}`},
	"departures":     {"Base.json": `{"rules": {"no-debugger": "warn"}}`, "CohereSettings.json": `{"extends": "./Base.json", "rules": {"no-debugger": "error"}, "departures": {"no-debugger": "A debugger statement never ships here."}}`},
	"reasons":        {"CohereSettings.json": `{"rules": {"no-continue": "off"}, "reasons": {"no-continue": "Style, with no bug class behind it."}}`},
	"overrides":      {"CohereSettings.json": `{"overrides": [{"files": ["**/*.test.ts"], "rules": {"no-debugger": "off"}, "reason": "fixtures"}]}`},
	"ignorePatterns": {"CohereSettings.json": `{"ignorePatterns": ["dist/**"]}`},
	"cohere":         {"CohereSettings.json": `{"cohere": "^1.0.0"}`},
	"plugins":        {"CohereSettings.json": `{"plugins": ["react"]}`},
	"jsPlugins":      {"CohereSettings.json": `{"jsPlugins": ["./rules.js"]}`},
	"settings":       {"CohereSettings.json": `{"settings": {"better-tailwindcss": {"callees": ["cn"]}}}`},
	"output":         {"CohereSettings.json": `{"output": {"phases": true}}`},
	FormatKey:        {"CohereSettings.json": `{"format": {"tabWidth": 4}}`},
}

// TestTheSchemaHasExactlyTheKeysTheLoaderAccepts is the test the schema exists for, in both directions,
// against the loader's behavior rather than its tables. Every key the Nexus schema names loads; a key it
// lacks is refused; the project schema names the same keys. Its `format` is optional and decides outside
// our tiers, and in a chain extending a cohere:system-inc set the formatter refuses it, which a schema
// cannot see, so the description says so.
func TestTheSchemaHasExactlyTheKeysTheLoaderAccepts(t *testing.T) {
	t.Parallel()
	files := built(t)
	nexusKeys, err := SchemaKeys(files[NexusSchemaPath])
	if err != nil {
		t.Fatal(err)
	}
	projectKeys, err := SchemaKeys(files[ProjectSchemaPath])
	if err != nil {
		t.Fatal(err)
	}

	for _, key := range nexusKeys {
		fixture, present := keyFixtures[key]
		if !present {
			t.Errorf("the schema names %q and the test has no settings file using it; add one to keyFixtures", key)
			continue
		}
		if _, err := configuration.Load(writeSettings(t, fixture)); err != nil {
			t.Errorf("the schema names %q and the loader refuses a file using it: %v", key, err)
		}
	}

	// The reverse: a key no schema names is refused, and every key the loader's own tables accept is in
	// the Nexus schema.
	for _, unknown := range []string{"env", "globals", "excludedFiles", "formatting"} {
		if _, err := configuration.Load(writeSettings(t, map[string]string{"CohereSettings.json": `{"` + unknown + `": {}}`})); err == nil {
			t.Errorf("the loader accepts %q, which no schema names", unknown)
		}
	}
	var loaderKeys []string
	for _, key := range configuration.TopLevelKeys() {
		loaderKeys = append(loaderKeys, key.Name)
	}
	if strings.Join(loaderKeys, ",") != strings.Join(nexusKeys, ",") {
		t.Errorf("loader keys %v, Nexus schema keys %v", loaderKeys, nexusKeys)
	}

	if strings.Join(projectKeys, ",") != strings.Join(nexusKeys, ",") {
		t.Errorf("project schema keys %v, want the Nexus schema's: %v", projectKeys, nexusKeys)
	}
	outsider := filepath.Dir(writeSettings(t, map[string]string{"CohereSettings.json": `{"format": {"tabWidth": 2}}`}))
	if resolution, err := formatoptions.Resolve(outsider); err != nil || resolution.Options.TabWidth != 2 {
		t.Errorf("an outsider's format block did not decide (%+v, %v), so the project schema is wrong to name format", resolution.Options, err)
	}
	ours := filepath.Dir(writeSettings(t, map[string]string{"CohereSettings.json": `{"extends": "cohere:system-inc/structure", "format": {"tabWidth": 2}}`}))
	if _, err := formatoptions.Resolve(ours); err == nil {
		t.Error("a format block in a chain extending cohere:system-inc/structure resolved, so the description is wrong to say it is refused")
	}
}

// TestTheLoaderTablesAgreeWithItsTypes: parsedTopLevelKeys and overrideKeys are the loader's refusal
// lists, and rawConfig and rawOverride are what it decodes; the schema reads the types, so the lists must
// name the same keys.
func TestTheLoaderTablesAgreeWithItsTypes(t *testing.T) {
	t.Parallel()
	var decoded []string
	for _, key := range configuration.TopLevelKeys() {
		if key.Field != nil {
			decoded = append(decoded, key.Name)
		}
	}
	sort.Strings(decoded)
	if strings.Join(decoded, ",") != strings.Join(configuration.ParsedTopLevelKeyNames(), ",") {
		t.Errorf("rawConfig decodes %v, parsedTopLevelKeys names %v", decoded, configuration.ParsedTopLevelKeyNames())
	}
	var overrideFields []string
	for _, field := range configuration.OverrideFields() {
		overrideFields = append(overrideFields, field.Name)
	}
	sort.Strings(overrideFields)
	if strings.Join(overrideFields, ",") != strings.Join(configuration.OverrideKeyNames(), ",") {
		t.Errorf("rawOverride decodes %v, overrideKeys names %v", overrideFields, configuration.OverrideKeyNames())
	}
}

// TestTheOverrideSchemaHasExactlyTheKeysTheLoaderAccepts: each override key the schema names loads, and
// one it lacks is refused.
func TestTheOverrideSchemaHasExactlyTheKeysTheLoaderAccepts(t *testing.T) {
	t.Parallel()
	var schema struct {
		Defs struct {
			Override struct {
				Properties map[string]json.RawMessage `json:"properties"`
			} `json:"override"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(built(t)[ProjectSchemaPath], &schema); err != nil {
		t.Fatal(err)
	}
	fixtures := map[string]string{"files": `"files": ["**/*.ts"]`, "rules": `"rules": {"no-debugger": "off"}`, "reason": `"reason": "why"`}
	for key := range schema.Defs.Override.Properties {
		fixture, present := fixtures[key]
		if !present {
			t.Errorf("the override schema names %q and the test has no block using it", key)
			continue
		}
		block := `{` + fixture + `}`
		if key != "files" {
			block = `{"files": ["**/*.ts"], ` + fixture + `}`
		}
		if _, err := configuration.Load(writeSettings(t, map[string]string{"CohereSettings.json": `{"overrides": [` + block + `]}`})); err != nil {
			t.Errorf("the override schema names %q and the loader refuses it: %v", key, err)
		}
	}
	if len(schema.Defs.Override.Properties) != len(fixtures) {
		t.Errorf("the override schema names %d keys, the test knows %d", len(schema.Defs.Override.Properties), len(fixtures))
	}
	if _, err := configuration.Load(writeSettings(t, map[string]string{"CohereSettings.json": `{"overrides": [{"files": ["**/*.ts"], "excludedFiles": ["x.ts"]}]}`})); err == nil {
		t.Error("the loader accepts an override's excludedFiles, which the schema does not name")
	}
}

// TestTheFormatSchemaHasExactlyTheKeysTheFormatterAccepts: every format option the Nexus schema names
// resolves in a Nexus tier, and one it lacks is refused.
func TestTheFormatSchemaHasExactlyTheKeysTheFormatterAccepts(t *testing.T) {
	t.Parallel()
	var schema struct {
		Properties map[string]struct {
			Properties map[string]struct {
				Type string   `json:"type"`
				Enum []string `json:"enum"`
			} `json:"properties"`
		} `json:"properties"`
	}
	if err := json.Unmarshal(built(t)[NexusSchemaPath], &schema); err != nil {
		t.Fatal(err)
	}
	options := schema.Properties[FormatKey].Properties
	if len(options) != len(formatoptions.BlockKeys) {
		t.Errorf("the format schema names %d options, the formatter's table %d", len(options), len(formatoptions.BlockKeys))
	}
	for name, option := range options {
		value := map[string]string{"integer": `4`, "boolean": `true`, "array": `["pnpm-lock.yaml"]`}[option.Type]
		if option.Type == "string" {
			value = `"value"`
			if len(option.Enum) > 0 {
				value = `"` + option.Enum[0] + `"`
			}
		}
		root := filepath.Dir(writeSettings(t, map[string]string{
			"NexusCohereSettings.json": `{"format": {"` + name + `": ` + value + `}}`,
			"CohereSettings.json":      `{"extends": "./NexusCohereSettings.json"}`,
		}))
		if _, err := formatoptions.Resolve(root); err != nil {
			t.Errorf("the format schema names %q and the formatter refuses it: %v", name, err)
		}
	}
	root := filepath.Dir(writeSettings(t, map[string]string{
		"NexusCohereSettings.json": `{"format": {"quoteProps": "as-needed"}}`,
		"CohereSettings.json":      `{"extends": "./NexusCohereSettings.json"}`,
	}))
	if _, err := formatoptions.Resolve(root); err == nil {
		t.Error("the formatter accepts quoteProps, which the format schema does not name")
	}
}

// TestTheRulesAreTheRegistrys: the schema offers exactly the registered rule names, so a rename under
// the registry reaches editors at the next generation rather than never.
func TestTheRulesAreTheRegistrys(t *testing.T) {
	t.Parallel()
	var schema struct {
		Defs struct {
			Rules struct {
				Properties map[string]json.RawMessage `json:"properties"`
			} `json:"rules"`
		} `json:"$defs"`
	}
	if err := json.Unmarshal(built(t)[ProjectSchemaPath], &schema); err != nil {
		t.Fatal(err)
	}
	var offered []string
	for name := range schema.Defs.Rules.Properties {
		offered = append(offered, name)
	}
	sort.Strings(offered)
	registered := append([]string(nil), registry.Names()...)
	sort.Strings(registered)
	if strings.Join(offered, ",") != strings.Join(registered, ",") {
		t.Errorf("the schema offers %d rules, the registry has %d", len(offered), len(registered))
	}
}

// TestTheSeveritiesAreTheLoaders: each spelling the schema allows is one parseSeverity accepts, in any
// case, and a spelling outside them is refused.
func TestTheSeveritiesAreTheLoaders(t *testing.T) {
	t.Parallel()
	for _, spelling := range configuration.SeveritySpellings {
		for _, variant := range []string{spelling, strings.ToUpper(spelling), strings.ToUpper(spelling[:1]) + spelling[1:]} {
			if err := configuration.ParseSeverity(variant); err != nil {
				t.Errorf("the schema allows %q and the loader refuses it: %v", variant, err)
			}
		}
	}
	for _, refused := range []string{"fatal", "on", "0", "2", "errors"} {
		if err := configuration.ParseSeverity(refused); err == nil {
			t.Errorf("the loader accepts %q, which the schema does not allow", refused)
		}
	}
}

// TestEveryExampleIsJSON: the reference's examples are copied into real files, so each must parse.
func TestEveryExampleIsJSON(t *testing.T) {
	t.Parallel()
	for _, key := range configuration.TopLevelKeys() {
		example, present := examples[key.Name]
		if !present {
			t.Errorf("the reference has no example for %q", key.Name)
			continue
		}
		if !json.Valid([]byte(example)) {
			t.Errorf("the example for %q is not JSON: %s", key.Name, example)
		}
	}
	if len(examples) != len(configuration.TopLevelKeys()) {
		t.Errorf("%d examples for %d keys", len(examples), len(configuration.TopLevelKeys()))
	}
}

// TestEveryDescriptionNamesAKeyThatExists: prose for a key the loader no longer accepts is refused too.
func TestEveryDescriptionNamesAKeyThatExists(t *testing.T) {
	t.Parallel()
	accepted := map[string]bool{}
	for _, key := range configuration.TopLevelKeys() {
		accepted[key.Name] = true
	}
	for key := range topLevelDescriptions {
		if !accepted[key] {
			t.Errorf("topLevelDescriptions describes %q, which the loader does not accept", key)
		}
	}
	fields := map[string]bool{}
	for _, field := range configuration.OverrideFields() {
		fields[field.Name] = true
	}
	for key := range overrideDescriptions {
		if !fields[key] {
			t.Errorf("overrideDescriptions describes %q, which an override does not accept", key)
		}
	}
	options := map[string]bool{}
	for _, option := range formatoptions.BlockKeys {
		options[option.Name] = true
		if formatDescriptions[option.Name] == "" {
			t.Errorf("format option %q has no description", option.Name)
		}
	}
	for key := range formatDescriptions {
		if !options[key] {
			t.Errorf("formatDescriptions describes %q, which the format block does not accept", key)
		}
	}
}

// TestEveryCarriedSetFitsTheSchema: the rule sets cohere carries are settings files too. Each names only
// keys the project schema allows, except the Nexus tier's set, which holds the format block and fits the
// Nexus schema.
func TestEveryCarriedSetFitsTheSchema(t *testing.T) {
	t.Parallel()
	files := built(t)
	nexusKeys, _ := SchemaKeys(files[NexusSchemaPath])
	projectKeys, _ := SchemaKeys(files[ProjectSchemaPath])
	allowed := func(keys []string) map[string]bool {
		set := map[string]bool{}
		for _, key := range keys {
			set[key] = true
		}
		return set
	}
	for _, name := range configuration.SetNames() {
		contents, err := configuration.SourceContents(name)
		if err != nil {
			t.Fatal(err)
		}
		var keys map[string]json.RawMessage
		if err := json.Unmarshal(contents, &keys); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		schemaKeys := allowed(projectKeys)
		if name == formatoptions.NexusTierSetName {
			schemaKeys = allowed(nexusKeys)
			if _, present := keys[FormatKey]; !present {
				t.Errorf("%s is the Nexus tier and holds no format block", name)
			}
		}
		for key := range keys {
			if !schemaKeys[key] {
				t.Errorf("%s names %q, which its schema does not allow", name, key)
			}
		}
	}
}

// TestTheDocumentedSetupsWork: the reference opens with two setups, the carried set alone and a tier of
// one's own. Both must load and say how to format, or the first thing a reader copies fails.
func TestTheDocumentedSetupsWork(t *testing.T) {
	t.Parallel()
	for name, files := range map[string]map[string]string{
		"the carried set": {
			"CohereSettings.json": `{"extends": "` + formatoptions.NexusTierSetName + `", "rules": {"no-debugger": "error"}, "ignorePatterns": ["dist/**"]}`,
		},
		"a tier of one's own": {
			"CohereSettings.json":           `{"extends": "./` + formatoptions.NexusTierFileName + `", "rules": {"no-debugger": "error"}}`,
			formatoptions.NexusTierFileName: `{"format": ` + examples[FormatKey] + `}`,
		},
	} {
		path := writeSettings(t, files)
		if _, err := configuration.Load(path); err != nil {
			t.Errorf("%s: the loader refuses it: %v", name, err)
		}
		if _, err := formatoptions.Resolve(filepath.Dir(path)); err != nil {
			t.Errorf("%s: the formatter refuses it: %v", name, err)
		}
	}
}
