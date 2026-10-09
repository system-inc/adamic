u083: 291 top-level tests grouped into three rows at 553ad06abb069736161b2d46532d9036e92f30b3.
Clean baseline and both switch controls passed; nproc = 5.
Verdicts: two sacred rows and one witness; no relevant empty-answer probe passed.
Nine production mutants were killed; six unique to CSSStrings and two unique to MultiPushGap.
Standalone diffs, exact member matrices, commands, probes and logs are in this directory.

```json
[
  {
    "test": "TestCSSStrings",
    "package": "stage1/cohere/cssstrings",
    "file": "port_test.go; shards_test.go; top_level_shards_test.go",
    "seconds": 24.931,
    "oracle": "Go cohere CSS adjustStrings through bridge; Node and JS/native port outputs compared to Go. Census and built-in witnesses use self checks.",
    "oracle_kind": "external-run",
    "kills": [
      "M1",
      "M2",
      "M3",
      "M4",
      "M5",
      "M6",
      "M8"
    ],
    "unique_kills": [
      "M1",
      "M2",
      "M3",
      "M4",
      "M5",
      "M6"
    ],
    "last_proven_fail": "M8: port_test.go:169: build cssstrings port lowering and backends: /workspace/adamic/stage1/cohere/cssstrings/strings.ts:23:13: stage 0 can't lower push with other than one value yet",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "vacuous": false,
    "evidence": "selector file /tmp/adamic-u083-mutant=; env {'ADAMIC_U083_MUTANT': 'M8', 'ADAMIC_U083_WITNESS': '', 'ADAMIC_NATIVE_SPLIT': 'u083-M8'}; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/cssstrings/ -run . > M8.log 2>&1; port_test.go:169: build cssstrings port lowering and backends: /workspace/adamic/stage1/cohere/cssstrings/strings.ts:23:13: stage 0 can't lower push with other than one value yet"
  },
  {
    "test": "TestMultiPushGap",
    "package": "stage1/cohere/cssstrings",
    "file": "gaps_test.go",
    "seconds": 0.117,
    "oracle": "Self-authored NotYet.What string; Node executes fixture and must print self-authored ab newline. Compiler refusal expectation has no outside authority.",
    "oracle_kind": "self",
    "kills": [
      "M7",
      "M8",
      "M9"
    ],
    "unique_kills": [
      "M7",
      "M9"
    ],
    "last_proven_fail": "M9: gaps_test.go:29: gap changed: /workspace/adamic/stage1/cohere/cssstrings/gaps/1_multi_push.ts:2:1: stage 0 can't lower  yet; update GAPS.md and remove the workaround if closed",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "vacuous": false,
    "evidence": "selector file /tmp/adamic-u083-mutant=; env {'ADAMIC_U083_MUTANT': 'M9', 'ADAMIC_U083_WITNESS': '', 'ADAMIC_NATIVE_SPLIT': 'u083-M9'}; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/cssstrings/ -run . > M9.log 2>&1; gaps_test.go:29: gap changed: /workspace/adamic/stage1/cohere/cssstrings/gaps/1_multi_push.ts:2:1: stage 0 can't lower  yet; update GAPS.md and remove the workaround if closed"
  },
  {
    "test": "TestCSSStringsPlantedDisagreement",
    "package": "stage1/cohere/cssstrings",
    "file": "shards_test.go",
    "seconds": 10.746,
    "oracle": "Self-authored synthetic case IDs and exactly-one-failure assertion on checker child.",
    "oracle_kind": "self",
    "kills": [
      "W1"
    ],
    "unique_kills": [],
    "last_proven_fail": "W1: shards_test.go:485: probe exit 0 stderr",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": 9,
    "vacuous": false,
    "evidence": "selector file /tmp/adamic-u083-mutant=; env {'ADAMIC_U083_MUTANT': '', 'ADAMIC_U083_WITNESS': 'W1'}; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/cssstrings/ -run . > W1.log 2>&1; shards_test.go:485: probe exit 0 stderr"
  }
]
```

F = TestCSSStrings family; G = TestMultiPushGap; W = TestCSSStringsPlantedDisagreement. All source positions below refer to the starting commit.

| ID | File:line | One-line change | Failed rows |
|---|---|---|---|
| M1 | stage1/cohere/cssstrings/strings.ts:4 | `const content = raw.slice(1, -1);` to `const content = raw.slice(1);` | F |
| M2 | stage1/cohere/cssstrings/strings.ts:35 | `parts.push(quote);` to `(drop statement)` | F |
| M3 | stage1/cohere/cssstrings/strings.ts:41 | `if(quote !== 34 && quote !== 39) return -1;` to `if(!(quote !== 34 && quote !== 39)) return -1;` | F |
| M4 | stage1/cohere/cssstrings/strings.ts:58 | `index = end - 1;` to `index = end;` | F |
| M5 | stage1/cohere/cssstrings/main.ts:12 | `index++;` to `(drop statement)` | F |
| M6 | stage1/cohere/cssstrings/main.ts:27 | `code === 10 ? '\\n'` to `code === 10 ? '\\r'` | F |
| M7 | internal/lower/object.go:1265 | `if len(arguments) != 1 {` to `if len(arguments) > 2 {` | G |
| M8 | internal/lower/object.go:1265 | `if len(arguments) != 1 {` to `if len(arguments) == 1 {` | F, G |
| M9 | internal/lower/diagnostics.go:33 | `return &NotYet{Where: l.program.Where(node), What: what}` to `return &NotYet{Where: l.program.Where(node), What: ""}` | G |
| E1 | stage1/cohere/cssstrings/strings.ts:50 | `export function adjustStrings(value: string, singleQuote: boolean): string {` to `export function adjustStrings(value: string, singleQuote: boolean): string { return '';` | F |
| E2 | internal/lower/lower.go:20 | `func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {` to `func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) { return nil, nil` | F |
| W1 | stage1/cohere/cssstrings/shards_test.go:374 | `func stringsOutputError(unit stringsUnit, side string, got, want []byte) error {` to `func stringsOutputError(unit stringsUnit, side string, got, want []byte) error { return nil` | W |
| W2 | stage1/cohere/cssstrings/port_test.go:255 | `if bytes.Equal(r.stdout, want.stdout) {` to `if true {` | F |

Survivors: none among M1 through M9. E1, E2, W1 and W2 also failed. No survivor behavior witness was needed. E2 aborted in family setup with a nil dereference; unexecuted rows are unknown, rather than passes. E2-gap independently failed at gaps_test.go:29 with `gap changed: <nil>`.

The code under test was declared before planting: the native CSS quote port in strings.ts (printString, stringEnd, adjustStrings), its native command-line decode/encode functions in main.ts, and stage 0 lowering's multi-value push refusal. The oracle was declared separately: Go cohere for CSS outputs, Node execution plus a self-authored diagnostic literal for the gap, and a synthetic checker failure for the witness. Go cohere and Node were never mutated. W1 and W2 are the explicitly allowed witness comparison edits. Main.ts decode/encode are native product behavior in this audit, not the Go comparison harness; two mutants cover the product's escaping protocol rather than quote selection.

functions.txt lists the five port functions and all 316 internal/lower functions actually reached in clean package coverage before mutation. plan.log gives the fixed menu plan. Initial invalid instrumentation is kept in invalid-switch and contributes no kill or verdict. The revised M3 condition flip was written before the valid matrix began.

The family has 287 numbered shards, Setup and Union, with 19 batch, 261 raw, three built-in mutant witness and four throughput shard members. Every member was listed at the starting commit. No named Your rows list was provided, so the whole package was scoped; no moved or vanished name was inferred. W2 made all three built-in witness members fail, showing that their disagreement assertion can reject an always-agree comparison. These members remain inside their one family row. Exact failed member names are recorded in matrix.json.

Vacuous flags use the relevant entry point: E1 for the CSS family, E2-gap for the gap checker and W1 for the witness's comparison. E1 passed the unrelated gap and witness rows, which do not call adjustStrings as their checked behavior. That is not evidence that those rows accept an empty answer from their own checker.

The brief's .a instruction does not match this unit's actual source: these ports are .ts files compiled through Adamic to native products. The mixed family combines oracle agreement, census checks, throughput checks and three witnesses, so a single oracle label necessarily describes its main agreement check; its other checks are self assertions. The gap's Node execution does not externally justify the expected compiler refusal, which is why its oracle_kind is self.

The first planting command sourced the warm environment too late and could not find gofmt. The initial M3 switch also produced TS2367, because TypeScript rejects the contradictory condition. Both failures and invalid recorded runs are preserved, not scored. They cost 148.404 recorded wall seconds plus unrecorded interrupted work. Native compiler environment switches are absent from the cache key, so compiler probes use distinct existing ADAMIC_NATIVE_SPLIT values u083-M7/M8/M9/E2. Only literal 1 enables native split mode; these labels change the key without enabling it. Native source selector values reuse the same primary products. Temporary built-in mutant copies still cause lowering cache misses on each package run, even when their native products hit.

All valid runs stayed below the 90-second wall limit. E2's setup panic required the one narrow gap probe; otherwise the whole package was used. Repo-wide uniqueness, other packages, and the optional ADAMIC_CSSSTRINGS_LIBRARY comparison were not exercised. No external-authority value was claimed or checked.

Warm setup was skipped after the environment worked; npm ci in stage3/api reported 1 second. Baseline wall time was 42.353 seconds (binary 40.000), and clean lowering coverage took 41.214 seconds. The nine standalone timing runs cost 124.707 wall seconds. Matrix plus empty/witness probes cost 334.352 wall seconds. Recorded valid runs including controls total 526.660 seconds, excluding baseline and coverage. Successful product build misses total 66.410 seconds and are already included in these run costs; full per-product times are in build-times.json. The primary switch control built lowering/backends in 1.13 seconds, sanitized native in 0.26 and native-fast in 0.12. Primary compiler rebuilds: M7 1.05/0.27/0.12 seconds, M9 0.98/0.25/0.12; M8 and E2 stopped during lowering before native compilation, with whole-run wall times 5.304 and 3.587 seconds respectively. Total session elapsed was about 22 minutes, exceeding the approximate 20-minute budget because of rejected instrumentation and evidence preparation.

Raw diagnostics refer to scratch positions. rows.json normalizes positions to the starting commit: the witness's raw log line 488 is origin line 485; the M8 strings diagnostic raw line 26 is origin line 23. The production and test source changes were reverted after the green final switch control. Only review evidence is committed. Each standalone diff applies independently to the starting commit, without the switch. Reproduce with plant.py and matrix.py after sourcing the warm environment.
