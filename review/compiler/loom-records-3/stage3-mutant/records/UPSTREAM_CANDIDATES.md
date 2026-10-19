# Record upstream candidates

Observed with stock TypeScript 6.0.3 on Node v24.19.0. No issue was sent. The requested `stage3/upstream/LEDGER.md` does not exist in the fetched `origin/codex/stage3-regex-captures`; entries use the requested location, counterexample, expected/actual result and drafted-issue fields. No shared ledger was created or edited.

## RECORDS-01: Own __proto__ path mapping is lost

- **Status:** confirmed stock tsc CLI wrong diagnostic; candidate upstream bug.
- **Location:** `src/compiler/commandLineParser.ts:2515` (`result[keyText] = value` in JSON conversion); ownership enumeration at `commandLineParser.ts:3320` / `utilities.ts:10239`, followed by matching at `moduleNameResolver.ts:3170-3177`. This is a construction bug, not evidence of an unguarded inherited read at 3177.
- **Counterexample:** [08_paths-own___proto__](input-fixtures/08_paths-own___proto__/), authored source copied unchanged from main.a/target.a to main.ts/target.ts by the runner.

```json
{"compilerOptions":{"noEmit":true,"module":"esnext","moduleResolution":"bundler","types":[],"paths":{"__proto__":["./target"]}},"files":["main.ts"]}
```

```typescript
// main.ts
import { value } from "__proto__";
const n: number = value;
// target.ts
export const value = 1;
```

- **Command:** `node <typescript@6.0.3>/lib/tsc.js --pretty false --project tsconfig.json`.
- **Expected:** own exact path key resolves ./target.ts, stdout/stderr empty, exit 0. Replacing the key and import with constructor, toString or hasOwnProperty produces exactly that result. Package typesVersions also correctly resolves an own __proto__ key.
- **Actual:** exit 2, stderr empty; exact stdout:

```text
main.ts(1,23): error TS2307: Cannot find module '__proto__' or its corresponding type declarations.
```

- **Additional observation:** stock `parseConfigFileTextToJson` reports no own __proto__ key in the parsed paths object, while the other three names remain own keys. JSON.parse of the same JSON preserves the own key. This supports the conversion-write explanation; no native compiler behavior is inferred.
- **Drafted issue:**

**Title:** tsconfig paths loses an own __proto__ mapping and reports TS2307

With TypeScript 6.0.3 on Node 24.19.0, the tsconfig and two files above should resolve an exact string-keyed paths entry. Instead tsc reports TS2307. All three ordinary Object.prototype spellings in the control cases resolve, and an own __proto__ package typesVersions entry also resolves. The config JSON converter writes properties onto `{}` by assignment, so __proto__ appears to invoke the inherited setter rather than create an own field. Please preserve JSON own fields, including __proto__, when building the config record. No prototype mutation is requested by the JSON input.

## RECORDS-02: compareDataObjects considers different empty paths equal

- **Status:** confirmed stock internal helper/API wrong result; full CLI/watch wrong behavior not demonstrated. This is not counted as a second failing tsc CLI run.
- **Location:** `src/compiler/utilities.ts:8154:45`, `src[e]` in compareDataObjects; compiler caller `program.ts:1260` checks whether a program is up to date. The same source read is verdict b for nonempty values and c for empty values.
- **Counterexample:** [29_compare-old](input-fixtures/29_compare-old/tsconfig.json) and [30_compare-new](input-fixtures/30_compare-new/tsconfig.json), with paths `{"constructor":[]}` versus `{"other":[]}`. [19_compare_empty_objects.a](19_compare_empty_objects.a) retains the complete source helper and covers all four prototype spellings with own-key controls.
- **Stock API command:** `STOCK_TYPESCRIPT=<typescript@6.0.3> node stage3/fixtures/records/probe-guards.cjs`. It reads the two real config files with stock ts.readConfigFile and compares their compilerOptions with stock ts.compareDataObjects.
- **Expected:** false: the maps have different own keys.
- **Actual:** true. Exact observation: `stock API compareDataObjects distinct empty paths: true`. The source loop reads missing src.constructor as Object.prototype.constructor, then recursion compares the empty array to that function. Both have zero own keys, so recursion returns true. For an own __proto__ miss it compares against Object.prototype instead of a function and likewise returns true.
- **CLI limitation:** both configs contain invalid empty substitution arrays, and stock CLI correctly reports TS5066 for their respective own key. Config text comparison later in isProgramUptoDate can also prevent reuse. No incorrect emitted output, lost diagnostic, or full watch-mode reuse was established. This is a confirmed helper result and an upstream candidate, not a demonstrated compiler end-to-end failure.
- **Drafted issue:**

**Title:** compareDataObjects returns true for records with different prototype-named own keys

TypeScript 6.0.3's internal helper returns true for compilerOptions loaded from the two configs above, despite different own paths keys. A minimal stock-package reproduction is `ts.compareDataObjects({paths: JSON.parse('{"constructor":[]}')}, {paths:{other:[]}})`, which returns true. The destination-key loop does not check that the source owns the key before reading it; recursion accepts an inherited constructor function as equivalent to an empty array. Please check source ownership before comparing record values. The helper is used by isProgramUptoDate, but this report does not demonstrate an incorrect full CLI/watch result; the empty-path inputs themselves correctly produce TS5066.

## Unresolved routes and scope

No requested real-input route was established for the two generic core compareProperties accesses or process.env[name]. The stock host API returns a function for getEnvironmentVariable("toString"), but compiler callers choose internal environment names; JSON/CLI inputs cannot select that parameter in the traced calls. groupBy's generic helper has the inherited-key crash in existing fixture 14, but its sole compiler caller selects only true/false. These are not invented end-to-end reproductions or new CLI bug entries.

Exact observations are in [input-fixtures/observations.json](input-fixtures/observations.json), [logs/input-probes.log](logs/input-probes.log) and [logs/stock-helper-probes.log](logs/stock-helper-probes.log). Guarded/proven-domain input tests verify those input forms; they do not assert that every language-service/API-only module-specifier site was executed by CLI. The per-site table states its source proof separately from these observations.
