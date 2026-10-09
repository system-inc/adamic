package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
)

// Severity is whether a rule runs, and how loudly.
type Severity int

const (
	// SeverityOff means the rule does not run for this file at all.
	SeverityOff Severity = iota

	// SeverityWarn reports without failing the run.
	SeverityWarn

	// SeverityError reports and fails the run.
	SeverityError
)

// String renders a severity the way the config spells it.
//
// Without this, fmt prints the underlying int as a control character, because Severity is an
// iota and SeverityOff is 0. A diagnostic about a config key that rendered its severity as
// '\x00' would be describing the exact problem it exists to report, illegibly.
func (s Severity) String() string {
	switch s {
	case SeverityOff:
		return "off"
	case SeverityWarn:
		return "warn"
	case SeverityError:
		return "error"
	default:
		return "unknown"
	}
}

// RuleSetting is one rule's severity and its options, if it has any.
//
// Options travel with severity rather than in a parallel map, because an override that re-states a
// rule for a subset of files usually changes both, and splitting them makes it possible to apply
// one without the other.
type RuleSetting struct {
	Severity Severity

	// Options is every element the config wrote after the severity, in order, decoded but not
	// interpreted. Nil when the rule was configured with a bare severity.
	//
	// This is ESLint's `context.options` exactly: the wire format is `[severity, ...options]`, and a
	// rule whose upstream schema has more than one element reads all of them. `eqeqeq` is
	// `["error", "always", {"null": "ignore"}]`, `consistent-this` is `["error", "self", "vm"]`.
	//
	// It used to hold `tuple[1]` alone, and then `tuple[1]` plus an `AdditionalOptions` remainder that
	// nothing outside a test read. Both were the same defect: a config entry that loads, a rule that
	// runs, and an option the author wrote that has no effect. What happens to the elements now is
	// `OptionsRegistry.Decode`'s business, and its answer is that each one is either handed to the
	// rule or refused by name. None is dropped.
	//
	// A rule that requires an option and is handed nil either guards everything or nothing, and both
	// are silent. Four rules in the gate cohere replaces were dead for months underneath exactly
	// that, so the distinction between "no options" and "options I did not read" is kept.
	Options []json.RawMessage
}

// Override is a glob-scoped block that changes rule settings for matching files.
type Override struct {
	Files []string
	Rules map[string]RuleSetting

	// Reason is why the block exists, when its author gave one, and File is the configuration that
	// wrote it. Coverage prints every reason on every run, so an override standing in for unfinished
	// work stays visible until that work lands. A scoped override needs no reason (#25benkk); one that
	// matches every file is a top-level rule and answers under `departures` instead.
	Reason string
	File   string
}

// Config is a resolved lint configuration: what runs everywhere, what never runs, and what changes
// per path.
type Config struct {
	// Rules is the base configuration, applying to every file not excluded.
	Rules map[string]RuleSetting

	// IgnorePatterns are paths no rule runs on at all.
	IgnorePatterns []string

	// Overrides apply in order, later blocks winning over earlier ones, which is how both ESLint and
	// oxlint resolve them. Order is preserved rather than sorted for exactly this reason.
	Overrides []Override

	// Plugins are the namespaces the config declares, which contribute their own rules without any
	// rules block naming them. See PluginDefaultRules.
	Plugins []string

	// Root is the directory glob patterns resolve against: the directory of the file cohere was
	// pointed at, whichever file in its `extends` chain a pattern was written in.
	Root string

	// Sources is every file the configuration was read from, the one cohere was pointed at first and
	// then each file it extends, in order. Anything keyed on the configuration's bytes (the findings
	// cache, the run cache, the scoped-run widening guard) must read all of them: a base edited alone
	// changes what runs exactly as an edit to the project's own file does.
	Sources []string

	// Departures are the rules a file in the chain sets differently from a file it extends, each with
	// the reason that file gave, keyed by rule name. Coverage reports them, so a departure is always
	// visible and never becomes a quiet allowance.
	Departures map[string]Departure

	// OffReasons are why each rule the chain turns off at top level is off, keyed by rule name: the
	// sentence a file gave under `reasons`, or its departure's reason when the off departs from an
	// inherited ruling. Coverage prints each beside the rule. An off with no entry here is an allowance.
	OffReasons map[string]OffReason

	// CohereVersion is the range of cohere releases the project's own file accepts, or nil when it
	// pins none. Only the project's own file may pin: a set is read by every project that extends it,
	// and a range there would decide for all of them which release they run.
	CohereVersion *VersionRange

	// Settings is the project's own `settings`, keyed by plugin namespace, each value as written. Only
	// the project's own file may write it, for the reason only it may pin a release: a value in a set
	// would decide for every project extending it. Which namespaces a run reads, and which keys under
	// each, is the rule packages' to say (rule.RegisterSettings), checked before any rule decodes.
	Settings map[string]json.RawMessage

	// ruleWriters is the source that turned each rule key on or set it: the file or set that wrote it at
	// top level, the one that declared the plugin for a plugin default, or the one whose override names it.
	ruleWriters map[string]string

	// house is the per-file dispatch of a zero-config configuration, nil for one that names its sets.
	house *houseSets

	// resolutionMemo is Resolve's results by match, created on first use. See resolutionMemo.
	resolutionMemo atomic.Pointer[resolutionMemo]

	// compiledGlobs is IgnorePatterns and each override's Files, compiled on first use. See Config.globs.
	compiledGlobs atomic.Pointer[configGlobs]
}

// Departure is one rule a configuration sets differently from the file it extends, and why.
type Departure struct {
	// File is the configuration that departs, and Reason is the sentence it gave under `departures`.
	File   string
	Reason string
}

// OffReason is why a rule is off, and the configuration that said so.
type OffReason struct {
	File   string
	Reason string
}

// OffReasonFor returns why the rule named ruleName is off, found by its own key or by a key that reaches
// it, the way the resolver lets a key spelled differently configure a rule.
func (c *Config) OffReasonFor(ruleName string) (OffReason, bool) {
	if c == nil {
		return OffReason{}, false
	}
	if offReason, found := lookupByReach(c.OffReasons, ruleName); found {
		return offReason, true
	}
	// Under zero config, a rule a per-file set turns on is off wherever the set does not fit, and the set
	// says why. Asked only after the chain's own reasons, so a project's own off keeps its own.
	if c.house != nil && !c.TurnsOffAtTopLevel(ruleName) {
		return lookupByReach(c.house.gatedReasons, ruleName)
	}
	return OffReason{}, false
}

// lookupByReach finds ruleName's entry by its own key, or else by the first key, sorted, that reaches it.
func lookupByReach(reasons map[string]OffReason, ruleName string) (OffReason, bool) {
	if offReason, found := reasons[ruleName]; found {
		return offReason, true
	}
	keys := make([]string, 0, len(reasons))
	for key := range reasons {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if KeyReachesRule(key, ruleName) {
			return reasons[key], true
		}
	}
	return OffReason{}, false
}

// TurnsOffAtTopLevel reports whether the chain's top-level rules turn the rule named ruleName off, by its
// own key or, when no key names it exactly, by a key that reaches it.
func (c *Config) TurnsOffAtTopLevel(ruleName string) bool {
	if c == nil {
		return false
	}
	if setting, found := c.Rules[ruleName]; found {
		return setting.Severity == SeverityOff
	}
	for key, setting := range c.Rules {
		if setting.Severity == SeverityOff && KeyReachesRule(key, ruleName) {
			return true
		}
	}
	return false
}

// RuleKeys returns every rule name the config mentions, in the base block or any override, at any
// severity, `off` included.
//
// A key is evidence the rule exists somewhere even when cohere has not ported it: someone wrote a
// decision about it. That is what lets a suppression naming an unported rule be told apart from one
// naming a rule that exists nowhere.
func (c *Config) RuleKeys() []string {
	if c == nil {
		return nil
	}
	keys := make([]string, 0, len(c.Rules))
	for key := range c.Rules {
		keys = append(keys, key)
	}
	for _, override := range c.Overrides {
		for key := range override.Rules {
			keys = append(keys, key)
		}
	}
	return keys
}

// PluginDefaultRules are the rules a `plugins` declaration turns on without any rules block naming
// them, keyed by rule name and valued by the plugin that contributes each.
//
// # Why this is a table rather than a computation
//
// oxlint's mechanism is `warn_correctness(plugins)` in `config_builder.rs:570`: every rule whose
// category is `Correctness` AND whose plugin is declared, at severity `Warn`. Reproducing that
// directly needs each rule's category, and neither this registry nor the captured inventory carried
// one -- adding categories for 214 rules would be a fresh transcription of oxc's metadata, which is
// a larger and less trustworthy job than the one this solves.
//
// The table is not a transcription. The captured inventory's `enabledBy: pluginDefault` was
// established empirically, by planting violations and watching the real oxlint binary report rules
// named in no config -- `no-const-assign` and `react/no-children-prop` among them. That is a
// measurement of what the tool actually enables, which is what parity is about, rather than a
// second-hand copy of the metadata that produces it.
//
// # `eslint` is not a plugin, and the 16 core rules depend on that
//
// Sixteen of these have no namespace, and no `plugins` entry names `eslint`. In oxc, `ESLINT` is a
// bitflag with the value 0 (`config/plugins.rs:95`), so `plugins.contains(ESLINT)` is true whatever
// is declared, and upstream's own comment on `warn_correctness` says "there's no way to disable
// ESLint correctness rules". They are therefore contributed unconditionally, and `PluginEslint` is
// the name that records that rather than leaving sixteen entries looking unattributed.
var PluginDefaultRules = pluginDefaultRules()

// PluginEslint is the pseudo-plugin that contributes core rules. See PluginDefaultRules.
const PluginEslint = "eslint"

// pluginContributing returns which plugin a plugin-default rule arrives from.
func pluginContributing(ruleName string) string {
	namespace, _, namespaced := strings.Cut(ruleName, "/")
	if !namespaced {
		return PluginEslint
	}
	return namespace
}

// PluginDefaultSeverity is the severity a plugin declaration contributes at.
//
// `Warn`, and not by choice: `warn_correctness` inserts `AllowWarnDeny::Warn` and the inventory
// records all forty as `warn`. This is a faithful statement of what a declaration contributes and is
// deliberately not bent to encode a house preference; see below for where the preference lives.
//
// # The forty `error` lines in the config are deliberate, and removing them is a downgrade
//
// The config names all forty of these rules at `error`. They look like redundant leftovers now that
// a declaration resolves the same forty, and they are not: **the declaration contributes `warn`, so
// deleting those lines lowers all forty rather than changing nothing.**
//
// Ruled 2026-08-25 that they stay, and the reasoning is recorded here because this is where the next
// cleanup pass will look before deleting them:
//
//   - Parity is the acceptance criterion for WHICH rules run, not for how loudly. Those are separable
//     and only the first is what this migration promised.
//   - The set does not contain a rule whose violation is arguably fine, which is what `warn` would
//     assert. `no-const-assign` is a runtime TypeError, `no-this-before-super` is a crash,
//     `no-obj-calls` calls a non-function, `react/no-direct-mutation-state` silently drops a render.
//   - The tool being replaced is the floor rather than the target. cohere exists because that gate
//     was insufficient, so inheriting its severity because it is the incumbent proves too much.
//
// The choice was originally made by someone who left no note, and `CohereSettings.json` is untracked
// so there is no author or date to recover. That is a gap in the record rather than evidence the
// choice was careless: reverting an undocumented decision to a weaker one BECAUSE it is undocumented
// is how a codebase loses hard-won strictness one blameless commit at a time.
//
// `TestBothPathsAgreeOnEveryPluginDefault` reports the divergence on every run, so this cannot go
// quiet.
const PluginDefaultSeverity = SeverityWarn

// RulesFromPlugins returns the settings a `plugins` declaration contributes, given what the rules
// block already names.
//
// A rule the config names explicitly is left alone: an explicit line is a decision and a default is
// not, so the default never overrides one. That is also what oxlint does, since a rules block entry
// is applied after `warn_correctness` seeds the map.
func RulesFromPlugins(plugins []string, named map[string]RuleSetting) map[string]RuleSetting {
	declared := map[string]bool{PluginEslint: true}
	for _, plugin := range plugins {
		declared[plugin] = true
	}

	contributed := map[string]RuleSetting{}
	for ruleName, plugin := range PluginDefaultRules {
		if !declared[plugin] {
			continue
		}
		if _, explicit := named[ruleName]; explicit {
			continue
		}
		contributed[ruleName] = RuleSetting{Severity: PluginDefaultSeverity}
	}
	return contributed
}

// Load reads a configuration file from disk, following its `extends` chain.
//
// A configuration that cannot be read is an error rather than an empty configuration. An empty config lints
// everything with nothing configured, which is indistinguishable from a clean run and is precisely
// the failure this project keeps finding: three configurations in the gate cohere replaces ran
// successfully having loaded zero plugins. The same holds for every file the chain names: a base
// that is missing, unreadable, or part of a cycle refuses the whole load.
//
// # How a chain merges
//
// House rulings live once, in a base each project extends (nexus, then structure or Base, then the
// project), so a ruling is made in one place and cannot drift between copies (#rkm5a31). Layers apply
// from the outermost base to the file cohere was pointed at:
//
//   - A rule entry replaces the inherited one. A bare severity keeps the inherited options, which is
//     what ESLint does, so an author reading `"rule": "off"` or `"rule": "error"` gets the reading they
//     already know.
//   - `plugins` are a union, and plugin defaults are computed once, over the merged rules.
//   - `ignorePatterns` and `overrides` concatenate, the base's first, so a later block still wins.
//     Every pattern resolves against Root: a house pattern is a shape like `**/*.test.ts`, not a path.
//   - A base may carry `format`: its reader (internal/format/formatoptions) follows the chain, applying
//     each file's block over the one it extends. A base may not carry `settings`: they are read from
//     the project's own file only, as ESLint reads them from the project's config, so a value there
//     would be ignored silently.
//   - A rule set differently from the file it extends must be named under `departures` with a reason,
//     and a `departures` entry that departs from nothing is refused, so the list cannot rot.
//   - A rule a file turns off at top level says why under `reasons`, unless the off departs from an
//     inherited ruling, where the departure's reason already does. A `reasons` entry for a rule the file
//     does not turn off at top level is refused, so that list cannot rot either. The reason follows the
//     rule up the chain until a later file turns it back on.
//
// Load matches inherited rulings by exact key. LoadFor, which the command uses, also knows which rules
// are registered, and that is what lets it tell a respelling from a twin; see sameRuling.
func Load(path string) (*Config, error) {
	return LoadFor(path, nil)
}

// LoadFor is Load knowing the registered rule names, so a key spelled differently from the one it
// inherits is matched to it by the rules the two actually reach rather than by the shape of the names.
func LoadFor(path string, registeredNames []string) (*Config, error) {
	root, err := filepath.Abs(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("resolving the config directory for %s: %w", path, err)
	}

	layers, err := readConfigLayers(path, nil)
	if err != nil {
		return nil, err
	}
	return loadLayers(layers, root, registeredNames)
}

// loadLayers merges a chain already read, outermost base first, the project's own file last.
//
// # Reasons are required inside our tiers, and of our own sets
//
// A chain that reaches a `cohere:system-inc/*` set is one of ours (InOurTiers), and every off and every
// departure in it says why, or the load is refused. Outside our tiers a project goes its own way
// (#bfxz13m): its own file may turn a rule off or change its options with no reason, and coverage
// counts each unreasoned off by rule so the choice stays visible. The sets cohere carries are ours
// wherever they are read, so an off or a departure written in one is held to a reason in every chain.
func loadLayers(layers []configLayer, root string, registeredNames []string) (*Config, error) {
	layerPaths := make([]string, 0, len(layers))
	for _, layer := range layers {
		layerPaths = append(layerPaths, layer.path)
	}
	inOurTiers := InOurTiers(layerPaths)

	loaded := &Config{
		Rules:      map[string]RuleSetting{},
		Root:       root,
		Departures: map[string]Departure{},
		OffReasons: map[string]OffReason{},
	}
	declaredPlugins := map[string]bool{}
	pluginDeclaredBy := map[string]string{}
	reach := newRuleReach(registeredNames)
	ancestors := ancestorsByLayer(layers)
	// Which layer last wrote each top-level rule, so a rule two unrelated sets both configure is caught.
	writtenBy := map[string]string{}

	for index, layer := range layers {
		isBase := index < len(layers)-1
		if isBase && layer.present["cohere"] {
			return nil, fmt.Errorf("lint config %s pins \"cohere\", and it is extended by %s: only the project's "+
				"own file may pin a release, since every project extending %s would inherit the pin. Move it "+
				"to the project's own file", layer.path, layers[len(layers)-1].path, layer.path)
		}
		if isBase && layer.present["settings"] {
			return nil, fmt.Errorf("lint config %s declares \"settings\", and it is extended by %s: "+
				"settings are read from the project's own file only, so the value would be ignored "+
				"silently. Keep \"settings\" in the project's own file",
				layer.path, layers[len(layers)-1].path)
		}
		if isBase && layer.present["output"] {
			return nil, fmt.Errorf("lint config %s declares \"output\", and it is extended by %s: "+
				"the run reads \"output\" from the project's own file only, so the value would be ignored "+
				"silently. Keep \"output\" in the project's own file",
				layer.path, layers[len(layers)-1].path)
		}

		// What the layers below wrote, frozen before this one writes anything. Comparing against the
		// live map would let two spellings in this same file read as one inheriting from the other, in
		// whichever order Go's map iteration happened to yield them.
		fromBases := maps.Clone(loaded.Rules)
		departed := map[string]bool{}
		holdsToReasons := inOurTiers || IsSet(layer.path)
		if err := checkReasons(layer); err != nil {
			return nil, err
		}
		for name, value := range layer.raw.Rules {
			setting, err := parseRuleSetting(value)
			if err != nil {
				return nil, fmt.Errorf("rule %q in %s: %w", name, layer.path, err)
			}

			inheritedName, inherited, isInherited, replacesInherited := inheritedRuleSetting(fromBases, name, reach)
			inheritedOffReason, inheritedIsReasoned := loaded.OffReasons[inheritedName]
			// Layers that do not extend one another are composed side by side, `cohere:react` beside
			// `cohere:next`. Each rule belongs to one of them: if both wrote it, which one wins would be
			// decided by the order of a list rather than by anyone's ruling, and neither file says so.
			if isInherited && !ancestors[layer.path][writtenBy[inheritedName]] {
				return nil, fmt.Errorf("rule %q is configured by both %s and %s, and neither extends the other: "+
					"a rule belongs to one set, so configure it in one of them or in a file that extends both",
					name, writtenBy[inheritedName], layer.path)
			}
			if isInherited {
				if setting.Options == nil {
					setting.Options = inherited.Options
				}
				if !sameRuleSetting(setting, inherited) {
					// Go whitespace: a reason in cohere's own lint config, which no JavaScript tool reads.
					reason := strings.TrimSpace(layer.raw.Departures[name])
					switch {
					case reason != "":
						departed[name] = true
						loaded.Departures[name] = Departure{File: layer.path, Reason: reason}
					case holdsToReasons:
						return nil, fmt.Errorf("lint config %s sets %q differently from the file it extends "+
							"and gives no reason: name it under \"departures\" with why this project "+
							"differs, or remove the line so the house ruling applies",
							layer.path, name)
					}
					// Outside our tiers a departure with no reason is the project's own choice. It is not
					// recorded as one, since a departure prints with its reason, and an off among them still
					// counts in coverage as an off with no reason.
				}
				// The inherited key goes only when the new one reaches every rule it did: a respelling.
				// A twin keeps its own key, or the rule only the inherited key reached would be left
				// unconfigured and silently stop running.
				if replacesInherited {
					delete(loaded.Rules, inheritedName)
					delete(loaded.OffReasons, inheritedName)
				}
			}
			loaded.Rules[name] = setting
			writtenBy[name] = layer.path

			switch {
			case setting.Severity != SeverityOff:
				delete(loaded.OffReasons, name)
			// Go whitespace: a reason in cohere's own lint config, which no JavaScript tool reads.
			case strings.TrimSpace(layer.raw.Reasons[name]) != "":
				loaded.OffReasons[name] = OffReason{File: layer.path, Reason: strings.TrimSpace(layer.raw.Reasons[name])} // Go whitespace: a reason in cohere's own lint config, which no JavaScript tool reads.
			case departed[name]:
				loaded.OffReasons[name] = OffReason{File: layer.path, Reason: loaded.Departures[name].Reason}
			case isInherited && inheritedIsReasoned:
				// Restating an inherited off keeps the reason the file that ruled it gave.
				loaded.OffReasons[name] = inheritedOffReason
			default:
				delete(loaded.OffReasons, name)
			}
		}

		loaded.IgnorePatterns = append(loaded.IgnorePatterns, layer.raw.IgnorePatterns...)
		for _, plugin := range layer.raw.Plugins {
			if !declaredPlugins[plugin] {
				declaredPlugins[plugin] = true
				pluginDeclaredBy[plugin] = layer.path
				loaded.Plugins = append(loaded.Plugins, plugin)
			}
		}

		for overrideIndex, rawOverride := range layer.raw.Overrides {
			override := Override{
				Files:  rawOverride.Files,
				Rules:  map[string]RuleSetting{},
				Reason: strings.TrimSpace(rawOverride.Reason), // Go whitespace: a reason in cohere's own lint config, which no JavaScript tool reads.
				File:   layer.path,
			}
			coversEverything := coversEveryFileOfItsKind(rawOverride.Files)
			for name, value := range rawOverride.Rules {
				setting, err := parseRuleSetting(value)
				if err != nil {
					return nil, fmt.Errorf("override %d, rule %q in %s: %w", overrideIndex, name, layer.path, err)
				}
				override.Rules[name] = setting

				// An override that matches every file of its kind is a top-level rule written in another
				// place, so it departs from an inherited ruling exactly as a top-level entry would. api's
				// `["**/*.ts", "**/*.tsx"]` block turned off thirteen rulings the Nexus tier holds at error,
				// and with only top-level entries checked it loaded with no reason and printed nothing
				// (#25benkk). A scoped override, `**/generated/**` or `**/*.test.ts`, stays a project's own
				// business and needs no reason.
				if !coversEverything {
					continue
				}
				_, inherited, isInherited, _ := inheritedRuleSetting(fromBases, name, reach)
				if !isInherited {
					continue
				}
				compared := setting
				if compared.Options == nil {
					compared.Options = inherited.Options
				}
				if sameRuleSetting(compared, inherited) {
					continue
				}
				// Go whitespace: a reason in cohere's own lint config, which no JavaScript tool reads.
				reason := strings.TrimSpace(layer.raw.Departures[name])
				if reason == "" && !holdsToReasons {
					continue
				}
				departed[name] = true
				if reason == "" {
					return nil, fmt.Errorf("lint config %s overrides %q for every file (%s) differently from the "+
						"file it extends and gives no reason: an override that matches every file is a "+
						"top-level rule, so name it under \"departures\" with why this project differs",
						layer.path, name, strings.Join(rawOverride.Files, ", "))
				}
				loaded.Departures[name] = Departure{File: layer.path, Reason: reason}
			}
			loaded.Overrides = append(loaded.Overrides, override)
		}

		for name := range layer.raw.Departures {
			if !departed[name] {
				return nil, fmt.Errorf("lint config %s names %q under \"departures\", but it sets that "+
					"rule the same as the file it extends, or not at all: remove the entry", layer.path, name)
			}
		}
	}

	if err := refuseUnreasonedOffs(loaded, writtenBy, inOurTiers); err != nil {
		return nil, err
	}

	// Applied after every layer's rules, and reading them: an explicit line is a decision and a
	// default is not, so `RulesFromPlugins` skips any rule already named. Seeding before would let a
	// default overwrite a deliberate `off`.
	loaded.ruleWriters = maps.Clone(writtenBy)
	for name, setting := range RulesFromPlugins(loaded.Plugins, loaded.Rules) {
		loaded.Rules[name] = setting
		loaded.ruleWriters[name] = pluginDeclaredBy[pluginContributing(name)]
	}
	for _, override := range loaded.Overrides {
		for name := range override.Rules {
			if _, written := loaded.ruleWriters[name]; !written {
				loaded.ruleWriters[name] = override.File
			}
		}
	}

	if own := layers[len(layers)-1]; own.present["settings"] {
		loaded.Settings = own.raw.Settings
	}

	if own := layers[len(layers)-1]; own.present["cohere"] {
		versionRange, err := ParseVersionRange(own.raw.Cohere)
		if err != nil {
			return nil, fmt.Errorf("lint config %s: %w", own.path, err)
		}
		if err := checkCoherePin(own.path, versionRange); err != nil {
			return nil, err
		}
		loaded.CohereVersion = &versionRange
	}

	for index := len(layers) - 1; index >= 0; index-- {
		loaded.Sources = append(loaded.Sources, layers[index].path)
	}
	return loaded, nil
}

// SourcesOf returns every file the configuration at path reads, itself first and then each file it
// extends, in order: what Config.Sources holds, for a caller that keys on the configuration's bytes
// before or without building a Config.
//
// The run cache, the findings cache and the scoped-run widening guard each keyed on the project's
// own file alone. With `extends`, an edit to a base changes what runs exactly as an edit to the
// project's file does, and a key blind to the base replays the old verdict.
func SourcesOf(path string) ([]string, error) {
	layers, err := readConfigLayers(path, nil)
	if err != nil {
		return nil, err
	}
	sources := make([]string, 0, len(layers))
	for index := len(layers) - 1; index >= 0; index-- {
		sources = append(sources, layers[index].path)
	}
	return sources, nil
}

// IgnorePatternsOf returns the ignorePatterns of the configuration at path, the base's first, exactly as
// Load concatenates them, without decoding a single rule.
//
// The format walk shares this list with lint: a path the project excludes from checking is excluded
// from formatting too, from one list. It reads the list here rather than through Load because the walk
// has no opinion on rules, and a rule key Load would refuse must not stop the formatter from knowing
// which files it may touch.
func IgnorePatternsOf(path string) ([]string, error) {
	layers, err := readConfigLayers(path, nil)
	if err != nil {
		return nil, err
	}
	var patterns []string
	for _, layer := range layers {
		patterns = append(patterns, layer.raw.IgnorePatterns...)
	}
	return patterns, nil
}

// configLayer is one file in an `extends` chain: what it says, which top-level keys it wrote, and the
// layers it names in `extends`, by path.
type configLayer struct {
	path     string
	raw      rawConfig
	present  map[string]bool
	extended []string
}

// readConfigLayers reads path and everything it extends, outermost base first.
//
// chain is the files already being read, so a file that extends itself, directly or through others,
// is refused by naming the loop rather than overflowing the stack.
//
// path is a file or an embedded set (`cohere:typescript`). A set is its own name, never a path, and it
// may extend only other sets: it has no directory for a relative path to resolve against.
func readConfigLayers(path string, chain []string) ([]configLayer, error) {
	absolute := path
	if !IsSet(path) {
		resolved, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("resolving lint config %s: %w", path, err)
		}
		absolute = resolved
	}
	for _, seen := range chain {
		if seen == absolute {
			return nil, fmt.Errorf("lint config %s extends itself: %s",
				absolute, strings.Join(append(append([]string(nil), chain...), absolute), " -> "))
		}
	}

	contents, err := SourceContents(absolute)
	if err != nil {
		return nil, fmt.Errorf("reading lint config %s: %w", absolute, err)
	}
	if err := checkTopLevelKeys(contents, absolute); err != nil {
		return nil, err
	}
	if err := checkOverrideKeys(contents, absolute); err != nil {
		return nil, err
	}

	var raw rawConfig
	if err := json.Unmarshal(contents, &raw); err != nil {
		return nil, fmt.Errorf("parsing lint config %s: %w", absolute, err)
	}
	var keyed map[string]json.RawMessage
	if err := json.Unmarshal(contents, &keyed); err != nil {
		return nil, fmt.Errorf("parsing lint config %s: %w", absolute, err)
	}
	present := make(map[string]bool, len(keyed))
	for key := range keyed {
		present[key] = true
	}

	layer := configLayer{path: absolute, raw: raw, present: present}

	// Each entry's chain in order, a layer already read keeping its first place: two sets that sit on
	// one base share it, and it applies once, outermost.
	var layers []configLayer
	read := map[string]bool{}
	for _, extended := range raw.Extends {
		base := extended
		if !IsSet(base) {
			if IsSet(absolute) {
				return nil, fmt.Errorf("rule set %s extends %s, which is a path: a set may extend only other sets", absolute, extended)
			}
			if !filepath.IsAbs(base) {
				base = filepath.Join(filepath.Dir(absolute), base)
			}
		}
		baseLayers, err := readConfigLayers(base, append(chain, absolute))
		if err != nil {
			return nil, fmt.Errorf("lint config %s extends %s: %w", absolute, extended, err)
		}
		for _, baseLayer := range baseLayers {
			if !read[baseLayer.path] {
				read[baseLayer.path] = true
				layers = append(layers, baseLayer)
			}
		}
		layer.extended = append(layer.extended, baseLayers[len(baseLayers)-1].path)
	}
	return append(layers, layer), nil
}

// ancestorsByLayer is every layer each layer extends, directly or through others, keyed by path.
func ancestorsByLayer(layers []configLayer) map[string]map[string]bool {
	extendedByPath := make(map[string][]string, len(layers))
	for _, layer := range layers {
		extendedByPath[layer.path] = layer.extended
	}
	ancestors := make(map[string]map[string]bool, len(layers))
	var collect func(path string, into map[string]bool)
	collect = func(path string, into map[string]bool) {
		for _, extended := range extendedByPath[path] {
			if !into[extended] {
				into[extended] = true
				collect(extended, into)
			}
		}
	}
	for _, layer := range layers {
		ancestors[layer.path] = map[string]bool{}
		collect(layer.path, ancestors[layer.path])
	}
	return ancestors
}

// inheritedRuleSetting finds the entry an earlier layer wrote for the same ruling as name, and says
// whether name replaces that entry outright.
//
// The same key is always the same ruling. A different spelling is the same ruling when the two keys
// reach a registered rule in common (see sameRuling), so a project cannot step around a house ruling
// by spelling the key differently. Without registered names only the same key matches.
func inheritedRuleSetting(rules map[string]RuleSetting, name string, reach *ruleReach) (string, RuleSetting, bool, bool) {
	if setting, found := rules[name]; found {
		return name, setting, true, true
	}
	matches := make([]string, 0, 1)
	for inheritedName := range rules {
		if reach.sameRuling(inheritedName, name) {
			matches = append(matches, inheritedName)
		}
	}
	if len(matches) == 0 {
		return "", RuleSetting{}, false, false
	}
	// Sorted so two spellings in one base resolve the same way on every run.
	sort.Strings(matches)
	inheritedName := matches[0]
	return inheritedName, rules[inheritedName], true, reach.reachesEvery(name, inheritedName)
}

// ruleReach is which registered rules each config key configures, by the resolver's own test, each
// key asked once.
//
// Matching one key to the inherited spellings of its ruling compares it with every key the layers
// below wrote, and each comparison asked the whole registry about both keys again: on ahra that was a
// fresh 464-name scan, and a fresh map, for every pair, and it was 590ms of every run, the whole of
// what the run spent outside its phases (#kgv1pry). A key reaches the same rules every time it is
// asked, so each is asked once per load. One load runs on one goroutine, so a plain map does.
type ruleReach struct {
	registeredNames []string
	byKey           map[string]map[string]bool
}

func newRuleReach(registeredNames []string) *ruleReach {
	return &ruleReach{registeredNames: registeredNames, byKey: map[string]map[string]bool{}}
}

// of is the registered rules key configures.
func (reach *ruleReach) of(key string) map[string]bool {
	if reached, found := reach.byKey[key]; found {
		return reached
	}
	reached := map[string]bool{}
	for _, registered := range reach.registeredNames {
		if KeyReachesRule(key, registered) {
			reached[registered] = true
		}
	}
	reach.byKey[key] = reached
	return reached
}

// sameRuling reports whether two differently spelled keys configure a registered rule in common.
//
// Decided by the registry rather than by the names. `nexus/x` and `x` look like one ruling and are,
// when only `x` is registered. `@typescript-eslint/no-invalid-this` and `no-invalid-this` look the same
// way and are two rules, a core rule and its typescript-eslint twin; they still share a ruling, because
// the resolver lets the qualified key configure the core rule when no key names it exactly (#hprjh4s).
func (reach *ruleReach) sameRuling(left string, right string) bool {
	leftReached := reach.of(left)
	for registered := range reach.of(right) {
		if leftReached[registered] {
			return true
		}
	}
	return false
}

// reachesEvery reports whether key reaches every registered rule inherited reaches, so that inherited
// can be removed without leaving any rule it configured unconfigured.
func (reach *ruleReach) reachesEvery(key string, inherited string) bool {
	reached := reach.of(key)
	for registered := range reach.of(inherited) {
		if !reached[registered] {
			return false
		}
	}
	return true
}

// wholeTreePattern is a glob that names no directory and no file name, only an extension or nothing
// at all: `**/*`, `**/*.ts`, `**/*.{ts,tsx}`. It matches every file of its kind wherever the file sits,
// which is decided by its shape rather than guessed from a file listing. `**/*.test.ts` names part of a
// file name and `source/**/*.ts` names a directory, so neither is one.
var wholeTreePattern = regexp.MustCompile(`^\*\*/\*(\.[A-Za-z0-9]+|\.\{[A-Za-z0-9]+(,[A-Za-z0-9]+)*\})?$`)

// coversEveryFileOfItsKind reports whether every pattern of an override is a whole-tree pattern, so
// the override applies to every file of the kinds it names.
func coversEveryFileOfItsKind(patterns []string) bool {
	if len(patterns) == 0 {
		return false
	}
	for _, pattern := range patterns {
		if !wholeTreePattern.MatchString(pattern) {
			return false
		}
	}
	return true
}

// sameRuleSetting reports whether two settings run the rule identically: one severity, and options
// equal element by element after compaction, so whitespace in the JSON is not a departure.
func sameRuleSetting(left RuleSetting, right RuleSetting) bool {
	if left.Severity != right.Severity || len(left.Options) != len(right.Options) {
		return false
	}
	for index := range left.Options {
		if compactJson(left.Options[index]) != compactJson(right.Options[index]) {
			return false
		}
	}
	return true
}

// compactJson renders a raw element without insignificant whitespace, or as written if it does not
// parse, which parseRuleSetting has already ruled out.
func compactJson(raw json.RawMessage) string {
	var buffer bytes.Buffer
	if err := json.Compact(&buffer, raw); err != nil {
		return string(raw)
	}
	return buffer.String()
}

// checkReasons refuses a `reasons` entry that explains nothing: one with no sentence, one for a rule the
// file does not turn off at top level, and one for an off the file already explains under `departures`.
//
// The second is the guard that keeps the list from rotting. A reason left behind after its rule was
// turned back on, or written for a rule another file turns off, reads as a decision while deciding
// nothing, exactly as a departure that departs from nothing would.
func checkReasons(layer configLayer) error {
	names := make([]string, 0, len(layer.raw.Reasons))
	for name := range layer.raw.Reasons {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		// Go whitespace: a reason in cohere's own lint config, which no JavaScript tool reads.
		if strings.TrimSpace(layer.raw.Reasons[name]) == "" {
			return fmt.Errorf("lint config %s names %q under \"reasons\" and gives no reason: say why the "+
				"rule is off, in a sentence", layer.path, name)
		}
		value, written := layer.raw.Rules[name]
		turnsOff := false
		if written {
			setting, err := parseRuleSetting(value)
			turnsOff = err == nil && setting.Severity == SeverityOff
		}
		if !turnsOff {
			return fmt.Errorf("lint config %s names %q under \"reasons\", but it does not turn that rule "+
				"off in its own top-level rules: a reason belongs to the file whose off it explains, so "+
				"remove the entry", layer.path, name)
		}
		if _, departs := layer.raw.Departures[name]; departs {
			return fmt.Errorf("lint config %s names %q under both \"reasons\" and \"departures\": the "+
				"departure's reason already says why this file turns it off, so remove the \"reasons\" entry",
				layer.path, name)
		}
	}
	return nil
}

// refuseUnreasonedOffs refuses a chain that turns a rule off at top level with no file saying why.
//
// An off with no reason is an allowance: it reads as a decision while recording none, and the next
// reader cannot tell a ruling from a rule somebody wanted quiet. Coverage printed each one while the sets
// and repositories were given their reasons; once every one carried a reason, the standard became a
// refusal so it cannot rot back (#2qq4yr7). The error names each rule and the file that turned it off, so
// the fix is one entry under that file's "reasons".
//
// Inside our tiers, and in the sets cohere carries. Outside them an off in the project's own files is
// its choice to make (#bfxz13m), and coverage counts it rather than the load refusing it.
func refuseUnreasonedOffs(loaded *Config, writtenBy map[string]string, inOurTiers bool) error {
	var unreasoned []string
	for name, setting := range loaded.Rules {
		if setting.Severity != SeverityOff {
			continue
		}
		if _, reasoned := loaded.OffReasons[name]; reasoned {
			continue
		}
		if !inOurTiers && !IsSet(writtenBy[name]) {
			continue
		}
		unreasoned = append(unreasoned, fmt.Sprintf("%q in %s", name, writtenBy[name]))
	}
	if len(unreasoned) == 0 {
		return nil
	}
	sort.Strings(unreasoned)
	return fmt.Errorf("lint config turns off %s with no reason: an off that says nothing reads as an "+
		"allowance, so name each under \"reasons\" in the file that turns it off, with why, in a sentence",
		strings.Join(unreasoned, ", "))
}

type rawConfig struct {
	Extends        extendsList                `json:"extends"`
	Departures     map[string]string          `json:"departures"`
	Reasons        map[string]string          `json:"reasons"`
	Plugins        []string                   `json:"plugins"`
	Rules          map[string]json.RawMessage `json:"rules"`
	IgnorePatterns []string                   `json:"ignorePatterns"`
	Overrides      []rawOverride              `json:"overrides"`
	Cohere         string                     `json:"cohere"`
	Settings       map[string]json.RawMessage `json:"settings"`
}

// parsedTopLevelKeys are the keys `rawConfig` decodes and the loader acts on.
var parsedTopLevelKeys = map[string]bool{
	"extends":        true,
	"departures":     true,
	"reasons":        true,
	"plugins":        true,
	"rules":          true,
	"ignorePatterns": true,
	"overrides":      true,
	"cohere":         true,
	"settings":       true,
}

// ignoredTopLevelKeys are the keys the loader deliberately does not act on, each with the reason.
//
// A reason is required, and the requirement is enforced below rather than left to convention. An
// entry with no reason says only that somebody wanted the config to load, which is the move this
// whole guard exists to stop -- the same defect one level down.
var ignoredTopLevelKeys = map[string]string{
	"$schema": "editor metadata; nothing in the linter reads it and nothing should, so it is " +
		"ignored on purpose rather than by omission",

	"jsPlugins": "paths to the JavaScript rule implementations the gate being replaced loads at " +
		"runtime. This linter compiles its rules in, so there is nothing to load and nothing to " +
		"honour. Ignored deliberately, and it stays in the config because oxlint still reads it " +
		"while both tools run side by side.",

	"format": "the formatter's options (printWidth, tabWidth and the rest), read by " +
		"internal/format/formatoptions. They moved here from package.json's prettier block when cohere's " +
		"native printers replaced Prettier, so the linter leaves them alone on purpose: they are the " +
		"formatter's, not a rule's.",

	"output": "how a run prints, read by the command (command/cohere/output_settings.go) from the file cohere " +
		"reads first. It decides nothing a rule does, so the linter leaves it alone on purpose.",
}

// checkTopLevelKeys refuses a config carrying a key the loader does not implement.
//
// # Why this refuses rather than warning
//
// `encoding/json` drops an unlisted key silently and returns no error, so a config could declare
// something and have it do nothing with every check passing. Four keys were in exactly that state:
// `plugins`, `jsPlugins` and `settings` were all being discarded, and `settings` named a Tailwind
// entry point that a hardcoded candidate list happened to probe first.
//
// None of them was found by a mechanism. `plugins` surfaced while someone investigated why 40
// rules ran on zero files; the other two surfaced only because that investigation prompted a read of
// the whole file. Nothing would have surfaced the next one.
//
// That is the same defect as a rule offered no files, sitting in this tool's own front door: the
// author believes something is configured, every check is green, and nothing runs. Refusing is the
// shape the rest of the tool already takes -- `Load` errors rather than returning an empty config for
// exactly this reason, and its comment says three configurations in the gate cohere replaces "ran
// successfully having loaded zero plugins."
//
// Warning was the alternative and it was rejected on the tool's own thesis. A warning goes into an
// output that already carries two hundred coverage notes, and a config author who edits one key does
// not re-read that output. A warning nobody reads is the silence this is meant to break, with a
// line of text in front of it. The kindness argument for warning is real, and the answer is that
// the error names the key and the file, so the fix is one edit rather than an investigation.
func checkTopLevelKeys(contents []byte, path string) error {
	var keyed map[string]json.RawMessage
	if err := json.Unmarshal(contents, &keyed); err != nil {
		// Not this function's error to report. Returning nil hands the real parse failure to the
		// caller's own Unmarshal, which produces the better message for a malformed file.
		return nil
	}

	var unimplemented []string
	for key := range keyed {
		if parsedTopLevelKeys[key] {
			continue
		}
		reason, ignored := ignoredTopLevelKeys[key]
		if !ignored {
			unimplemented = append(unimplemented, key)
			continue
		}
		if reason == "" {
			return fmt.Errorf(
				"lint config %s declares %q, which is listed as deliberately ignored with no reason "+
					"given: an entry with no reason says only that somebody wanted this to load",
				path, key)
		}
	}
	if len(unimplemented) == 0 {
		return nil
	}
	sort.Strings(unimplemented)

	return fmt.Errorf(
		"lint config %s declares %s, which this loader does not implement: the key would be "+
			"discarded silently and whatever it configures would never take effect. Either "+
			"implement it in `rawConfig` and act on it, or add it to `ignoredTopLevelKeys` "+
			"with a note saying why ignoring it is correct",
		path, strings.Join(quoteEach(unimplemented), ", "))
}

// quoteEach quotes each name so a multi-key message reads unambiguously.
func quoteEach(names []string) []string {
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, strconv.Quote(name))
	}
	return quoted
}

type rawOverride struct {
	Files  []string                   `json:"files"`
	Rules  map[string]json.RawMessage `json:"rules"`
	Reason string                     `json:"reason"`
}

// overrideKeys are the keys an override block may carry. Anything else is refused for the reason
// checkTopLevelKeys gives, one level down: `encoding/json` drops an unknown key without a word. The
// likeliest one is ESLint's `excludedFiles`, and dropping it widens the block to files its author
// meant to leave out.
var overrideKeys = map[string]bool{"files": true, "rules": true, "reason": true}

// checkOverrideKeys refuses an override block carrying a key the loader does not read.
func checkOverrideKeys(contents []byte, path string) error {
	var shaped struct {
		Overrides []map[string]json.RawMessage `json:"overrides"`
	}
	if err := json.Unmarshal(contents, &shaped); err != nil {
		// The caller's own Unmarshal reports a malformed file with the better message.
		return nil
	}
	for index, override := range shaped.Overrides {
		var unknown []string
		for key := range override {
			if !overrideKeys[key] {
				unknown = append(unknown, key)
			}
		}
		if len(unknown) == 0 {
			continue
		}
		sort.Strings(unknown)
		return fmt.Errorf(
			"lint config %s: override %d declares %s, which this loader does not read: the key would be "+
				"discarded silently. An override takes \"files\", \"rules\" and \"reason\"",
			path, index, strings.Join(quoteEach(unknown), ", "))
	}
	return nil
}

// parseRuleSetting decodes the two shapes a rule value takes: a bare severity, or an array of the
// severity followed by every option element, which is ESLint's `[severity, ...options]`.
//
// Every element after the severity is kept. Whether a rule accepts that many is decided by
// `OptionsRegistry.Decode`, which knows each rule's arity; this layer does not, so the one thing it
// must not do is decide by truncating.
func parseRuleSetting(value json.RawMessage) (RuleSetting, error) {
	var severityName string
	if err := json.Unmarshal(value, &severityName); err == nil {
		severity, err := parseSeverity(severityName)
		return RuleSetting{Severity: severity}, err
	}

	var tuple []json.RawMessage
	if err := json.Unmarshal(value, &tuple); err != nil {
		return RuleSetting{}, fmt.Errorf("expected a severity string or a [severity, ...options] array, got %s", truncate(string(value)))
	}
	if len(tuple) == 0 {
		return RuleSetting{}, fmt.Errorf("empty rule configuration array")
	}

	if err := json.Unmarshal(tuple[0], &severityName); err != nil {
		return RuleSetting{}, fmt.Errorf("first element is not a severity string: %s", truncate(string(tuple[0])))
	}
	severity, err := parseSeverity(severityName)
	if err != nil {
		return RuleSetting{}, err
	}

	setting := RuleSetting{Severity: severity}
	if len(tuple) > 1 {
		setting.Options = tuple[1:]
	}
	return setting, nil
}

// parseSeverity accepts the spellings the config files use, and refuses the rest loudly.
//
// An unrecognized severity silently treated as "off" is the worst available outcome: a typo in a
// rule name's severity would disable that rule across the whole tree with nothing reporting it.
func parseSeverity(name string) (Severity, error) {
	switch strings.ToLower(name) {
	case "off", "allow":
		return SeverityOff, nil
	case "warn", "warning":
		return SeverityWarn, nil
	case "error", "deny":
		return SeverityError, nil
	}
	return SeverityOff, fmt.Errorf("unknown severity %q (expected off, warn, or error)", name)
}

func truncate(value string) string {
	if len(value) <= 60 {
		return value
	}
	return value[:60] + "..."
}

// pluginDefaultRules builds the table. See PluginDefaultRules for why it is a table.
//
// These 41 entries were derived from a captured inventory of what oxlint's `plugins` declarations
// turn on without any rules block naming them, proven at capture time by planting violations rather
// than by reading a config. That inventory is deleted and this table is now the only record of it,
// which is a deliberate trade: the catalog was a frozen snapshot, so the test that pinned this
// against it could only ever catch a hand-edit here, never real upstream drift.
//
// What that means for a change: adding or removing an entry is no longer checked by anything. If
// oxlint's defaults move, the way to find out is to plant a violation for the rule and see whether
// the gate reports it with no config entry naming it -- the same probe that produced this list.
func pluginDefaultRules() map[string]string {
	return map[string]string{
		"constructor-super":                   PluginEslint,
		"getter-return":                       PluginEslint,
		"no-caller":                           PluginEslint,
		"no-class-assign":                     PluginEslint,
		"no-const-assign":                     PluginEslint,
		"no-dupe-class-members":               PluginEslint,
		"no-eval":                             PluginEslint,
		"no-func-assign":                      PluginEslint,
		"no-import-assign":                    PluginEslint,
		"no-iterator":                         PluginEslint,
		"no-new-native-nonconstructor":        PluginEslint,
		"no-obj-calls":                        PluginEslint,
		"no-setter-return":                    PluginEslint,
		"no-this-before-super":                PluginEslint,
		"no-unsafe-negation":                  PluginEslint,
		"no-with":                             PluginEslint,
		"react/forward-ref-uses-ref":          "react",
		"react/jsx-no-duplicate-props":        "react",
		"react/jsx-no-undef":                  "react",
		"react/jsx-props-no-spread-multi":     "react",
		"react/no-children-prop":              "react",
		"react/no-danger-with-children":       "react",
		"react/no-did-mount-set-state":        "react",
		"react/no-did-update-set-state":       "react",
		"react/no-direct-mutation-state":      "react",
		"react/no-find-dom-node":              "react",
		"react/no-is-mounted":                 "react",
		"react/no-render-return-value":        "react",
		"react/no-string-refs":                "react",
		"react/no-this-in-sfc":                "react",
		"react/no-unescaped-entities":         "react",
		"react/no-unsafe":                     "react",
		"react/no-will-update-set-state":      "react",
		"react/void-dom-elements-no-children": "react",
		"react-hooks/void-use-memo":           "react-hooks",
		"@typescript-eslint/no-array-delete":  "@typescript-eslint",
		"@typescript-eslint/no-for-in-array":  "@typescript-eslint",
		"@typescript-eslint/no-implied-eval":  "@typescript-eslint",
		"@typescript-eslint/no-unnecessary-parameter-property-assignment": "@typescript-eslint",
		"@typescript-eslint/no-unsafe-unary-minus":                        "@typescript-eslint",
		"@typescript-eslint/no-useless-empty-export":                      "@typescript-eslint",
	}
}
