package nexus

import (
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/testing"
)

const correctnessNoDiscardedOutcomeFile = "/repository/modules/meta/CorrectnessNoDiscardedOutcome.ts"

const correctnessNoDiscardedOutcomeNexus = "/repository/libraries/structure/libraries/nexus/source/"

// correctnessNoDiscardedOutcomeNexusFiles model the four Nexus files that declare its outcome types,
// as they stand in ahra's Structure checkout on 2026-10-03, trimmed to the aliases and the signatures
// of the functions returning them. The aliases are copied as written, arms and parentheses included.
var correctnessNoDiscardedOutcomeNexusFiles = map[string]string{
	correctnessNoDiscardedOutcomeNexus + "structured-text/json/Json.ts": strings.Join([]string{
		"export type JsonParseOutcomeType<ValueType = unknown> =",
		"    { outcome: 'Parsed'; value: ValueType } | { outcome: 'Invalid'; message: string };",
		"export declare function parseJson<ValueType = unknown>(text: string, context: string): JsonParseOutcomeType<ValueType>;",
	}, "\n"),
	correctnessNoDiscardedOutcomeNexus + "structured-text/json/Jsonc.ts": strings.Join([]string{
		"export type JsoncParseOutcomeType<ValueType = unknown> =",
		"    { outcome: 'Parsed'; value: ValueType } | { outcome: 'Invalid'; message: string };",
		"export declare function parseJsonc<ValueType = unknown>(text: string, context: string): JsoncParseOutcomeType<ValueType>;",
	}, "\n"),
	correctnessNoDiscardedOutcomeNexus + "structured-text/json/JsonFile.ts": strings.Join([]string{
		"export type JsonFileReadOutcomeType<ValueType = unknown> =",
		"    | { outcome: 'Parsed'; value: ValueType }",
		"    // Missing, a directory, or permission denied.",
		"    | { outcome: 'Unreadable'; message: string }",
		"    // Read, but not valid JSON.",
		"    | { outcome: 'Invalid'; message: string };",
		"export declare function readJsonFile<ValueType = unknown>(path: string, context: string): JsonFileReadOutcomeType<ValueType>;",
		"export type JsonFileWriteOutcomeType =",
		"    | { outcome: 'Written' }",
		"    // A value that will not serialize, a bad path, permissions, or a full disk.",
		"    | { outcome: 'Unwritable'; message: string };",
		"export interface JsonFileWriteOptionsInterface { mode?: number }",
		"export declare function writeJsonFile(",
		"    path: string,",
		"    value: unknown,",
		"    context: string,",
		"    options?: JsonFileWriteOptionsInterface,",
		"): JsonFileWriteOutcomeType;",
	}, "\n"),
	correctnessNoDiscardedOutcomeNexus + "system/PackageJson.ts": strings.Join([]string{
		"export type PackageJsonReadOutcomeType<ValueType = unknown> =",
		"    { outcome: 'Read'; value: ValueType } | { outcome: 'Missing' } | { outcome: 'Invalid'; message: string };",
		"export declare function readPackageJson(packageJsonFile: string): PackageJsonReadOutcomeType;",
	}, "\n"),
}

// correctnessNoDiscardedOutcomePrelude imports every producer the cases use and declares the rest,
// so the checker resolves each name and no verdict rests on an unresolved identifier. lib ES2022 has
// no `console`, so it is declared here.
var correctnessNoDiscardedOutcomePrelude = strings.Join([]string{
	"import { parseJson } from '../../libraries/structure/libraries/nexus/source/structured-text/json/Json';",
	"import { parseJsonc } from '../../libraries/structure/libraries/nexus/source/structured-text/json/Jsonc';",
	"import {",
	"    readJsonFile,",
	"    writeJsonFile,",
	"    type JsonFileReadOutcomeType,",
	"    type JsonFileWriteOutcomeType,",
	"} from '../../libraries/structure/libraries/nexus/source/structured-text/json/JsonFile';",
	"import { readPackageJson } from '../../libraries/structure/libraries/nexus/source/system/PackageJson';",
	"declare const console: { warn(message: string): void };",
	"declare const statePath: string;",
	"declare const state: { lastSeen: Record<string, string> };",
	"declare const text: string;",
	"declare const flag: boolean;",
	"declare const items: string[];",
	"declare function use(value: unknown): void;",
	"",
}, "\n")

func correctnessNoDiscardedOutcomeSource(lines ...string) string {
	return correctnessNoDiscardedOutcomePrelude + strings.Join(lines, "\n") + "\n"
}

func correctnessNoDiscardedOutcomeRun(t *testing.T, sourceText string) rule_testing.Result {
	t.Helper()
	files := map[string]string{correctnessNoDiscardedOutcomeFile: sourceText}
	for fileName, fileText := range correctnessNoDiscardedOutcomeNexusFiles {
		files[fileName] = fileText
	}
	return rule_testing.RunTypedFiles(t, CorrectnessNoDiscardedOutcome, files, correctnessNoDiscardedOutcomeFile)
}

// correctnessNoDiscardedOutcomeFinding is one finding as the outcome type its message names and the
// text it points at.
type correctnessNoDiscardedOutcomeFinding struct {
	typeName string
	span     string
}

// correctnessNoDiscardedOutcomeExpect asserts the findings in source order. The harness trims the
// subject file before writing it, so spans are read from the program's own text.
func correctnessNoDiscardedOutcomeExpect(t *testing.T, result rule_testing.Result, want []correctnessNoDiscardedOutcomeFinding) {
	t.Helper()
	diagnostics := append(result.Diagnostics[:0:0], result.Diagnostics...)
	sort.SliceStable(diagnostics, func(left int, right int) bool {
		return diagnostics[left].Range.Pos() < diagnostics[right].Range.Pos()
	})
	sorted := result
	sorted.Diagnostics = diagnostics
	wantIds := make([]string, 0, len(want))
	for range want {
		wantIds = append(wantIds, correctnessNoDiscardedOutcomeId)
	}
	rule_testing.ExpectFindings(t, sorted, wantIds...)
	text := result.SourceFile.Text()
	var got []correctnessNoDiscardedOutcomeFinding
	for _, diagnostic := range diagnostics {
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
		typeName := ""
		for _, known := range correctnessNoDiscardedOutcomeDeclarations {
			if diagnostic.Message.Description == correctnessNoDiscardedOutcomeMessage(known.typeName).Description {
				typeName = known.typeName
			}
		}
		got = append(got, correctnessNoDiscardedOutcomeFinding{typeName, strings.TrimSpace(text[diagnostic.Range.Pos():diagnostic.Range.End()])})
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("finding %d is %v, want %v (all: %v)", index, got[index], want[index], got)
		}
	}
}

func correctnessNoDiscardedOutcomeWrite(span string) correctnessNoDiscardedOutcomeFinding {
	return correctnessNoDiscardedOutcomeFinding{"JsonFileWriteOutcomeType", span}
}

// correctnessNoDiscardedOutcomeRealSites are the four ahra sites the rule was built from, as they
// stand on 2026-10-03, each before and after its fix. Every one saves a ledger with a bare
// `writeJsonFile(...)` in a function returning `void`.
var correctnessNoDiscardedOutcomeRealSites = []struct {
	name    string
	before  []string
	after   []string
	finding string
}{
	{
		// `modules/meta/MetaPollingApi.ts:41`: a lost write leaves the old cursor, and the next poll
		// replays every item as new.
		name: "MetaPollingApi saveState",
		before: []string{
			"interface PollingStateInterface {",
			"    lastSeen: Record<string, string>;",
			"}",
			"export function saveState(state: PollingStateInterface): void {",
			"    writeJsonFile(statePath, state, 'meta polling state');",
			"}",
		},
		after: []string{
			"interface PollingStateInterface {",
			"    lastSeen: Record<string, string>;",
			"}",
			"export function saveState(state: PollingStateInterface): void {",
			"    const stateWrite = writeJsonFile(statePath, state, 'meta polling state');",
			"    if(stateWrite.outcome === 'Unwritable') {",
			"        console.warn(stateWrite.message);",
			"    }",
			"}",
		},
		finding: "writeJsonFile(statePath, state, 'meta polling state');",
	},
	{
		// `modules/google/youtube/YouTubeSafety.ts:57`: the ledger that stops a repeated action, so a
		// failed write is a safety limit that stopped counting. The fix stops rather than warns.
		name: "YouTubeSafety saveLedger",
		before: []string{
			"interface LedgerEntryInterface { type: string; at: number }",
			"type LedgerInterface = Record<string, LedgerEntryInterface[]>;",
			"declare const ledgerPath: string;",
			"export function saveLedger(ledger: LedgerInterface): void {",
			"    writeJsonFile(ledgerPath, ledger, 'youtube safety ledger');",
			"}",
		},
		after: []string{
			"interface LedgerEntryInterface { type: string; at: number }",
			"type LedgerInterface = Record<string, LedgerEntryInterface[]>;",
			"declare const ledgerPath: string;",
			"export function saveLedger(ledger: LedgerInterface): void {",
			"    const ledgerWrite = writeJsonFile(ledgerPath, ledger, 'youtube safety ledger');",
			"    if(ledgerWrite.outcome === 'Unwritable') {",
			"        throw new Error(ledgerWrite.message);",
			"    }",
			"}",
		},
		finding: "writeJsonFile(ledgerPath, ledger, 'youtube safety ledger');",
	},
	{
		// `modules/pensieve/PensieveWindows.ts:43`: the return type is inferred, `void` today.
		name: "PensieveWindows saveJournalState",
		before: []string{
			"interface JournalStateInterface { processedWindows: Record<string, string>; lastRun: string | null }",
			"declare const JournalStatePath: string;",
			"export function saveJournalState(state: JournalStateInterface) {",
			"    state.lastRun = new Date().toISOString();",
			"    writeJsonFile(JournalStatePath, state, 'pensieve journal state');",
			"}",
		},
		after: []string{
			"interface JournalStateInterface { processedWindows: Record<string, string>; lastRun: string | null }",
			"declare const JournalStatePath: string;",
			"export function saveJournalState(state: JournalStateInterface) {",
			"    state.lastRun = new Date().toISOString();",
			"    const journalWrite = writeJsonFile(JournalStatePath, state, 'pensieve journal state');",
			"    if(journalWrite.outcome === 'Unwritable') console.warn(journalWrite.message);",
			"}",
		},
		finding: "writeJsonFile(JournalStatePath, state, 'pensieve journal state');",
	},
	{
		// `modules/pensieve/PensieveWeeklies.ts:99`.
		name: "PensieveWeeklies recordPartialSource",
		before: []string{
			"declare function getPartialSourcesPath(weekStart: string): string;",
			"declare function loadPartialSources(weekStart: string): Record<string, string>;",
			"export function recordPartialSource(weekStart: string, partialFilename: string, dailyDigest: string): void {",
			"    const sources = loadPartialSources(weekStart);",
			"    sources[partialFilename] = dailyDigest;",
			"    writeJsonFile(getPartialSourcesPath(weekStart), sources, `pensieve partial sources ${weekStart}`);",
			"}",
		},
		after: []string{
			"declare function getPartialSourcesPath(weekStart: string): string;",
			"declare function loadPartialSources(weekStart: string): Record<string, string>;",
			"export function recordPartialSource(weekStart: string, partialFilename: string, dailyDigest: string): void {",
			"    const sources = loadPartialSources(weekStart);",
			"    sources[partialFilename] = dailyDigest;",
			"    const sourcesWrite = writeJsonFile(getPartialSourcesPath(weekStart), sources, `pensieve partial sources ${weekStart}`);",
			"    if(sourcesWrite.outcome === 'Unwritable') console.warn(sourcesWrite.message);",
			"}",
		},
		finding: "writeJsonFile(getPartialSourcesPath(weekStart), sources, `pensieve partial sources ${weekStart}`);",
	},
}

// The four real sites before their fix: each bare write is reported, and nothing else is.
func TestCorrectnessNoDiscardedOutcomeFiresOnTheRealSites(t *testing.T) {
	t.Parallel()

	for _, site := range correctnessNoDiscardedOutcomeRealSites {
		t.Run(site.name, func(t *testing.T) {
			t.Parallel()

			result := correctnessNoDiscardedOutcomeRun(t, correctnessNoDiscardedOutcomeSource(site.before...))
			correctnessNoDiscardedOutcomeExpect(t, result, []correctnessNoDiscardedOutcomeFinding{
				correctnessNoDiscardedOutcomeWrite(site.finding),
			})
		})
	}
}

// The same four after their fix: the outcome is kept and its failure arm handled.
func TestCorrectnessNoDiscardedOutcomeStaysSilentOnTheFixedSites(t *testing.T) {
	t.Parallel()

	for _, site := range correctnessNoDiscardedOutcomeRealSites {
		t.Run(site.name, func(t *testing.T) {
			t.Parallel()

			rule_testing.ExpectClean(t, correctnessNoDiscardedOutcomeRun(t, correctnessNoDiscardedOutcomeSource(site.after...)))
		})
	}
}

func TestCorrectnessNoDiscardedOutcomeFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
		want  []correctnessNoDiscardedOutcomeFinding
	}{
		{"a JSON parse dropped", []string{
			"export function check() {",
			"    parseJson(text, 'webhook payload');",
			"}",
		}, []correctnessNoDiscardedOutcomeFinding{{"JsonParseOutcomeType", "parseJson(text, 'webhook payload');"}}},
		{"a JSONC parse dropped", []string{
			"export function check() {",
			"    parseJsonc(text, 'ProjectSecrets.jsonc');",
			"}",
		}, []correctnessNoDiscardedOutcomeFinding{{"JsoncParseOutcomeType", "parseJsonc(text, 'ProjectSecrets.jsonc');"}}},
		{"a file read dropped, instantiated with a type argument", []string{
			"export function check() {",
			"    readJsonFile<{ lastSeen: Record<string, string> }>(statePath, 'meta polling state');",
			"}",
		}, []correctnessNoDiscardedOutcomeFinding{{"JsonFileReadOutcomeType", "readJsonFile<{ lastSeen: Record<string, string> }>(statePath, 'meta polling state');"}}},
		{"a package.json read dropped", []string{
			"export function check() {",
			"    readPackageJson('package.json');",
			"}",
		}, []correctnessNoDiscardedOutcomeFinding{{"PackageJsonReadOutcomeType", "readPackageJson('package.json');"}}},
		{"a write with a mode, inside parentheses", []string{
			"export function save() {",
			"    (writeJsonFile(statePath, state, 'plaid pending links', { mode: 0o600 }));",
			"}",
		}, []correctnessNoDiscardedOutcomeFinding{correctnessNoDiscardedOutcomeWrite("(writeJsonFile(statePath, state, 'plaid pending links', { mode: 0o600 }));")}},
		{"a project wrapper with an inferred return type", []string{
			"function saveState() {",
			"    return writeJsonFile(statePath, state, 'meta polling state');",
			"}",
			"export function poll() {",
			"    saveState();",
			"}",
		}, []correctnessNoDiscardedOutcomeFinding{correctnessNoDiscardedOutcomeWrite("saveState();")}},
		{"an async wrapper awaited and dropped", []string{
			"async function saveState(): Promise<JsonFileWriteOutcomeType> {",
			"    return writeJsonFile(statePath, state, 'meta polling state');",
			"}",
			"export async function poll() {",
			"    await saveState();",
			"}",
		}, []correctnessNoDiscardedOutcomeFinding{correctnessNoDiscardedOutcomeWrite("await saveState();")}},
		{"a method on a store", []string{
			"class StateStore {",
			"    save(): JsonFileWriteOutcomeType {",
			"        return writeJsonFile(statePath, state, 'meta polling state');",
			"    }",
			"}",
			"export function poll(store: StateStore) {",
			"    store.save();",
			"}",
		}, []correctnessNoDiscardedOutcomeFinding{correctnessNoDiscardedOutcomeWrite("store.save();")}},
		{"an optional call, whose type loses the alias to `| undefined`", []string{
			"interface StateStoreInterface { save(): JsonFileWriteOutcomeType }",
			"export function poll(store: StateStoreInterface | undefined) {",
			"    store?.save();",
			"}",
		}, []correctnessNoDiscardedOutcomeFinding{correctnessNoDiscardedOutcomeWrite("store?.save();")}},
		{"a project union that adds an arm beside Nexus's", []string{
			"type StateSaveOutcomeType = JsonFileWriteOutcomeType | { outcome: 'Skipped' };",
			"declare function saveIfChanged(): StateSaveOutcomeType;",
			"export function poll() {",
			"    saveIfChanged();",
			"}",
		}, []correctnessNoDiscardedOutcomeFinding{correctnessNoDiscardedOutcomeWrite("saveIfChanged();")}},
		{"one of each, in a loop and a branch", []string{
			"export function saveAll() {",
			"    for(const each of items) {",
			"        writeJsonFile(each, state, 'each');",
			"    }",
			"    if(flag) writeJsonFile(statePath, state, 'flagged');",
			"}",
		}, []correctnessNoDiscardedOutcomeFinding{
			correctnessNoDiscardedOutcomeWrite("writeJsonFile(each, state, 'each');"),
			correctnessNoDiscardedOutcomeWrite("writeJsonFile(statePath, state, 'flagged');"),
		}},
		{"a write at module level", []string{
			"writeJsonFile(statePath, state, 'meta polling state');",
		}, []correctnessNoDiscardedOutcomeFinding{correctnessNoDiscardedOutcomeWrite("writeJsonFile(statePath, state, 'meta polling state');")}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result := correctnessNoDiscardedOutcomeRun(t, correctnessNoDiscardedOutcomeSource(testCase.lines...))
			correctnessNoDiscardedOutcomeExpect(t, result, testCase.want)
		})
	}
}

func TestCorrectnessNoDiscardedOutcomeStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"ignored on purpose, with void", []string{
			"export function save() {",
			"    void writeJsonFile(statePath, state, 'scratch cache');",
			"}",
		}},
		{"returned to the caller", []string{
			"export function save(): JsonFileWriteOutcomeType {",
			"    return writeJsonFile(statePath, state, 'meta polling state');",
			"}",
		}},
		{"kept and handed on", []string{
			"export function save() {",
			"    const stateWrite = writeJsonFile(statePath, state, 'meta polling state');",
			"    use(stateWrite);",
			"}",
		}},
		{"the read the Nexus doc comment shows", []string{
			"export function load() {",
			"    const outcome = readJsonFile(statePath, 'report write');",
			"    if(outcome.outcome === 'Unreadable') return {};",
			"    return outcome.outcome === 'Parsed' ? outcome.value : {};",
			"}",
		}},
		{"an arrow returning the outcome", []string{
			"export const outcomes = items.map((each) => writeJsonFile(each, state, 'each'));",
		}},
		{"an unawaited async wrapper, a floating promise and not this rule's", []string{
			"async function saveState(): Promise<JsonFileWriteOutcomeType> {",
			"    return writeJsonFile(statePath, state, 'meta polling state');",
			"}",
			"export function poll() {",
			"    saveState().catch(use);",
			"}",
		}},
		{"a value narrowed to the success arm", []string{
			"declare function markWritten(): Extract<JsonFileWriteOutcomeType, { outcome: 'Written' }>;",
			"export function poll() {",
			"    markWritten();",
			"}",
		}},
		{"a project outcome of the same name and shape, declared here", []string{
			"type JsonParseOutcomeType = { outcome: 'Parsed'; value: unknown } | { outcome: 'Invalid'; message: string };",
			"declare function parseLocally(text: string): JsonParseOutcomeType;",
			"export function check() {",
			"    parseLocally(text);",
			"}",
		}},
		{"a call returning nothing", []string{
			"export function warn() {",
			"    console.warn('nothing to save');",
			"}",
		}},
		{"a read of the outcome field as a statement", []string{
			"export function check(outcome: JsonFileReadOutcomeType) {",
			"    outcome.outcome;",
			"}",
		}},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			rule_testing.ExpectClean(t, correctnessNoDiscardedOutcomeRun(t, correctnessNoDiscardedOutcomeSource(testCase.lines...)))
		})
	}
}

// The same alias, name and arms, declared by a file that is not Nexus's. The declaring file is half
// of what makes a type ours, so the call stays silent; the Nexus write beside it is the control.
func TestCorrectnessNoDiscardedOutcomeIgnoresTheSameAliasOutsideNexus(t *testing.T) {
	t.Parallel()

	lookalike := strings.Join([]string{
		"export type JsonFileWriteOutcomeType =",
		"    | { outcome: 'Written' }",
		"    | { outcome: 'Unwritable'; message: string };",
		"export declare function writeJsonFile(path: string, value: unknown, context: string): JsonFileWriteOutcomeType;",
	}, "\n")
	subject := strings.Join([]string{
		"import { writeJsonFile } from '../structured-text/json/JsonFile';",
		"import { writeJsonFile as writeNexusJsonFile } from '../../libraries/structure/libraries/nexus/source/structured-text/json/JsonFile';",
		"export function save(path: string) {",
		"    writeJsonFile(path, {}, 'lookalike');",
		"    writeNexusJsonFile(path, {}, 'nexus');",
		"}",
	}, "\n")
	files := map[string]string{
		"/repository/source/structured-text/json/JsonFile.ts": lookalike,
		"/repository/source/save/Save.ts":                     subject,
	}
	for fileName, fileText := range correctnessNoDiscardedOutcomeNexusFiles {
		files[fileName] = fileText
	}
	result := rule_testing.RunTypedFiles(t, CorrectnessNoDiscardedOutcome, files, "/repository/source/save/Save.ts")
	correctnessNoDiscardedOutcomeExpect(t, result, []correctnessNoDiscardedOutcomeFinding{
		correctnessNoDiscardedOutcomeWrite("writeNexusJsonFile(path, {}, 'nexus');"),
	})
}
