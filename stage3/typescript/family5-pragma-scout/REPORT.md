# Family 5 pragma scout · #pncjpqy

The checkpoint scout started from the observed origin/main tip `384c6daa3d230929b7b90a97ae44f9781282153a`. The recursive submodule pins are in [submodules.txt](evidence/submodules.txt). The TypeScript input was v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`, as pinned by `stage3/source.json` and fetched by `stage3/meter/twice-daily.sh`.

The prior scout at `464502ff430b1c88b2e79ebf6b21badcb09e02af` identified G43 in `addPragmaForMatch` and G44 in `extractPragmas`. C13 and C14b declared only the first absent optional member, `types`. The trial moved the refusal to `lib` and exposed a string widening into the union of pragma records. The structural-widening rule is at `docs/rulings.md:71`.

## Arguments by pragma

Line references in this table are to upstream v6.0.3. A span is `{ value: string; pos: number; end: number }`.

| Pragma | Constructed own keys and values | Spec |
| --- | --- | --- |
| `reference` | Optional `types`, `lib`, `path` spans. Optional `no-default-lib`, `resolution-mode`, `preserve` strings, with explicit `undefined` possible for empty single quotes. | types.ts:10241–10250 |
| `amd-dependency` | Required own `path`, optional own `name`. Each present value is a string or `undefined`. | types.ts:10252–10255 |
| `amd-module` | Required own `name`, a string or `undefined`. | types.ts:10256–10259 |
| `ts-check` | No keys, even with trailing arguments. | types.ts:10260–10262 |
| `ts-nocheck` | No keys, even with trailing arguments. | types.ts:10263–10265 |
| `jsx` | `factory` string when written. It can be absent. | types.ts:10266–10269 |
| `jsxfrag` | `factory` string when written. It can be absent. | types.ts:10270–10273 |
| `jsximportsource` | `factory` string when written. It can be absent. | types.ts:10274–10277 |
| `jsxruntime` | `factory` string when written. It can be absent. | types.ts:10278–10281 |

For the current table, the name lookup and flag check select exactly the applicable spec. In the positional branch, only that spec's `argument.name` is written. The early return on absent payload occurs before required-argument validation. In the XML branch, `extractPragmas` performs its own flag check and obtains each key and capture mode from the selected spec. A foreign attribute cannot add a key. The XML expression `matchResult[2] || matchResult[3]` yields `undefined` for empty single quotes. For spans, `value.length` throws before the write. For string entries, the parser performs the write with the undefined value. These guarantees concern the current runtime table. `as const` does not freeze it. See `parser.ts:10714–10764,10767–10797`.

## Readers

`processCommentPragmas` populates a map whose repeated names become arrays (`parser.ts:10591–10612`). `processPragmasIntoFields` reads the six reference fields, AMD dependencies and AMD module names. The check pragmas use only the range. The JSX entries are read directly (`parser.ts:10619–10699`). The source-file parser invokes both passes at `parser.ts:1822–1823`, and the services preprocessor invokes them at `services/preProcess.ts:421–422`. The JSX checker reads factory values at `checker.ts:2444–2447,2490–2493,51352–51355` and checks presence only at `checker.ts:33779–33780`. JSX runtime and import-source selection read the factory strings at `utilities.ts:9413–9425`. The classifier uses only table membership (`services/classifier.ts:941–945`).

## Trial

The scratch generator adds internal type-only `PragmaNamesWithKind` and `ParsedPragmaArguments` in `types.ts`, and changes the arguments member of the internal type alias `PragmaPseudoMap`. It preserves the existing entry and map declarations, which pick up the new payload type through that alias. The parser adds type-only imports and local aliases `XmlPragmaName`, `PositionalPragmaName`, `PragmaBuilder`, `XmlPragmaBuilder`, and `PositionalPragmaBuilder`. It changes erased annotations at the name boundaries, construction locals, indexed writes, the two arguments expressions feeding the entry casts, and the return of `getNamedPragmaArguments`. The finite builders cover the eligible keys while mutating. The asserted payloads express the eligible kind records. The existing discriminated entry assertion still bridges the correlation between the runtime spec, the name, and the payload. The trial edits no public declaration directly.

The producer contract must include missing JSX factories and present undefined AMD strings. Those facts expose existing consumer assumptions. `parseIsolatedEntityName` takes a string, and `AmdDependency.path` is a public string property. A type assertion that simply promises a string at either boundary would contradict the witnesses. See [trial generator](evidence/trial.py) and [runtime witnesses](evidence/witness.json).

## Measurement at origin/main

The measurement starts from checkpoint `43bf21686c28679c4782c3d1bff7cf89cbba67cc`. The baseline compiler, adaptations and census overlay all come from the fetched `origin/main` tip `491e020de2aa138d7b608611c40a641571a5a1eb`. TypeScript was cloned by tag from the repository named in that tip's `stage3/source.json`, following `stage3/meter/twice-daily.sh`. The tag resolved to v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. The recursive submodules were initialized at their gitlinks and the compiler worktree had a clean tracked status. Both trees came from the completed `stage3/apply.sh` run. The treatment is a scratch copy with the checkpoint's `evidence/trial.py` applied verbatim. [Pins](evidence/measurement/submodules.txt) · [Apply log](evidence/measurement/apply.log.gz) · [Trial diff](evidence/measurement/trial-source.diff)

### Stock tsc

The compiler project selected exactly the 79 recursive TypeScript roots under `src/compiler` on both sides. The strict stock run used those same files with the options in `internal/load/load.go`, including ES2024, forced modules, bundler resolution, verbatim module syntax, `types: []`, strict mode, exact optional properties and unchecked indexed access. The stock binary uses its usual libraries and file system. In this configuration it has 64 baseline diagnostics from the absence of Adamic's host declarations, plus their cascades. All 64 diagnostic blocks recur verbatim in the trial. The actual Adamic loader results are recorded separately in the census headers. [Commands](evidence/measurement/check.sh) · [Compiler config](evidence/measurement/stock-config.json) · [Loader options and roots](evidence/measurement/loader-tsconfig.json)

| Configuration | Baseline | Trial | Full output |
| --- | ---: | ---: | --- |
| Compiler's own project | Exit 0, 0 diagnostics | Exit 2, 4 diagnostics | [Baseline](evidence/measurement/before-stock.txt) · [Trial](evidence/measurement/after-stock.txt) |
| Adamic loader options with stock tsc | Exit 2, 64 diagnostics | Exit 2, 70 diagnostics | [Baseline](evidence/measurement/before-strict.txt) · [Trial](evidence/measurement/after-strict.txt) · [Six added diagnostics](evidence/measurement/strict-trial-diagnostics.txt) |

### String consumers

Locations refer to the adapted **trial** tree.

| Site | What it would need |
| --- | --- |
| `src/compiler/checker.ts:2451`, `:2497`, `:51382` | Each call to `parseIsolatedEntityName` needs a reader check and an explicit decision about the missing factory. The string parameter is at `src/compiler/parser.ts:1374`. The parser can produce a pragma with no own `factory` key, so a string assertion would be false. |
| `src/compiler/parser.ts:10641` and `src/compiler/types.ts:4279` | The dependency projection writes the trial's `string \| undefined` into public `AmdDependency.path: string`. Preserving a present undefined path across that boundary would require a public change to `path: string \| undefined`. Keeping the public type requires checking the value at projection or deriving a narrower type from the adapted XML producer's capture contract. A check that drops or substitutes a dependency with an undefined path would change upstream behavior. |
| `src/compiler/parser.ts:10654`, `:10658` and `src/compiler/types.ts:4361`, `:4473` | Direct writes of the AMD module's `string \| undefined` fail under exact optional property types. Check before writing with a defined policy for the missing case, or derive a narrower adapted-producer type. Preserving writes of explicit undefined with the broader producer would require widening the writable context property and the public `SourceFile.moduleName` declaration to admit it. |
| `src/compiler/transformers/module/module.ts:568`, `:572` | If public `AmdDependency.path` widens, both `createStringLiteral` readers need a path check before producing the dependency literal. The `name` check selects the alias branch and establishes nothing about `path`. |
| `src/compiler/emitter.ts:4322`, `:4325` | The interpolations accept undefined and would render it as the text `undefined`. If an undefined path is carried downstream, the emitter needs an explicit policy for that case. The module-name emitter already checks the name at `:4315`. |
| `src/services/services.ts:1081` | The internal `SourceFileObject` mirror spells out a string-only AMD dependency element type. It would need to follow a widened dependency contract instead of promising string paths. |

The JSX runtime and import-source readers already tolerate missing factories at `src/compiler/utilities.ts:9419`, `:9426` and `:9427`. Public `AmdDependency.name` already admits a present undefined value at `src/compiler/types.ts:4280`. The passing through of `name` produces no additional diagnostic.


### Emitted bytes and runtime witnesses

The only changed source files were `src/compiler/parser.ts` and `src/compiler/types.ts`. Their emitted JavaScript was byte-identical in both trees. Stock tsc emitted with the compiler's target and module settings and used `--noCheck` for the byte comparison. The TypeScript project declaration build also used `--noCheck` so the trial's checking failures could be measured independently of emission. The upstream `scripts/dtsBundler.mjs` then produced the full public bundle in each tree. Both public `typescript.d.ts` files were byte-identical, with SHA-256 `65c210e9d7553165606df3ac67544ea7ddcbccf2c322ed924e8e8f4c4622c5fd`. [Changed sources](evidence/measurement/changed-files.txt) · [Emit commands](evidence/measurement/emit.sh) · [JavaScript hashes](evidence/measurement/javascript.sha256) · [Public declaration hashes](evidence/measurement/public-dts.sha256)

The emitted baseline and trial both reproduce missing own factory keys for the four JSX pragmas. For empty single-quoted AMD dependency and module attributes, the measured trees preserve the own keys with the empty string as their value. An empty reference path also produces a zero-width span with an empty string. The checkpoint tested upstream's `||` expression, where the AMD values can be present as undefined. Stage 3 applies `46-fix-pragma-empty-argument` before this trial and uses the matcher alternatives' nullish fallback at `src/compiler/parser.ts:10724`. Any future narrowing to string would need to reflect that adapted capture contract, while missing JSX factories still require handling. The trial keeps the checkpoint's broader types. [Runtime script](evidence/measurement/runtime.cjs) · [Runtime results](evidence/measurement/runtime.json) · [Checkpoint witness](evidence/witness.json)

### Full latent parser census

Both runs loaded all 79 compiler roots with the overlay built from the pinned main worktree. `LATENT_ONLY_FILE` selected the exact absolute path to each tree's `src/compiler/parser.ts`. Both used `LATENT_FULL=1`, `LATENT_ASSERT_NO_OUTPUT=1` and `CENSUS_MODULE_FORMAT=CommonJSBundle`. The real Adamic loader accepted the baseline with zero diagnostics and rejected the trial with the same six trial-specific errors that stock tsc reported under the strict options. The measurement overlay continued the trial census with those diagnostics recorded. [Census command](evidence/measurement/census.sh) · [Baseline loader](evidence/measurement/before-adamic-loader.txt) · [Trial loader](evidence/measurement/after-adamic-loader.txt)

The grading aligns parser source text after normalizing line endings and projects finding and attempting-unit coordinates to the baseline. Within an aligned finding site and unit, it first pairs identical observations, then pairs changed observations by phase and kind when several share the site. A **reached** finding is the same observation on both sides. A **changed** finding has a paired site and a different census observation. **Gone** and **new** denote unpaired baseline and trial observations. A gone finding in a checker-rejected body can reflect lost coverage. The parser record's unit statuses show where that happened. [Grading script](evidence/measurement/grade.py)

| Run | Units | Findings | Checker diagnostics |
| --- | ---: | ---: | ---: |
| Baseline | 602 | 2,505 | 0 |
| Trial | 607 | 2,493 | 6 |

Every finding has one grade. There are **2,483 reached, 6 changed, 16 gone and 4 new**. Both streams have a completion trailer with no relation overflow or trailer findings. The five extra units are the new type aliases. `processPragmasIntoFields` is the sole existing unit with an eligibility change. It becomes `split_checker_body` because of the three AMD parser diagnostics. Thirteen findings in that unit disappear with the lost coverage. The two processes overlapped, so elapsed times include contention. [Baseline stream](evidence/measurement/before-latent.jsonl.gz) · [Trial stream](evidence/measurement/after-latent.jsonl.gz) · [All grades](evidence/measurement/parser-grades.jsonl) · [Full delta](evidence/measurement/parser-delta.txt)

| Site | Baseline | Trial |
| --- | --- | --- |
| G43, `getNamedPragmaArguments` result at the `addPragmaForMatch` push | `parser.ts:10760:45` refuses missing optional `types`. The outer cast at `:10760:18` is also refused in the scan and in lowering. | All three observations are gone. Lowering at `parser.ts:10770:20` now reports `NotYet reading name`, and the rollback boundary changes at `:10770:5`. The preceding cast of the narrowed table entry to `PragmaDefinition` is still refused at `:10763:20`, now for making readonly `kind` writable. |
| G44, `extractPragmas` XML push | `parser.ts:10728:53` refuses missing optional `types`. The outer cast at `:10728:26` is independently refused. | The inner payload assertion at `parser.ts:10738:53` now refuses an unchecked cast. The outer cast at `:10738:26` is still refused. |

Outside the two push sites, three new unchecked-cast findings occur on the XML and positional builder keys at `parser.ts:10727:34`, `:10734:34` and `:10787:16`. Four observations change outside the pushes. The mutable-array view refusal in `processCommentPragmas` changes at `:10596:42`. The cast from the narrowed comment pragma table entry to `PragmaDefinition` changes in both refusal scan and lowering at `:10763:20`, with the corresponding lowering boundary changing at `:10763:5`.

**Verdict.** The trial does not clear G43 and G44 without collateral changes. Both original missing-`types` refusals disappear, but G43 still has a lowering stop and G44 still has two unchecked-cast refusals. The trial adds checker errors at the JSX and AMD boundaries, changes or introduces findings away from the pushes, and loses coverage of `processPragmasIntoFields`. The byte comparisons pass. Stop this trial at the measured checkpoint.
