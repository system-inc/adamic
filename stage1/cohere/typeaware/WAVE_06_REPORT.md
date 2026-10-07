Built: three native Adamic rules, a promised-return checker fact, and an independent Go comparison harness.
Commits: claim `bf0b933b50b0aed3a2cb98e7262a6e012ab7c981`; implementation is the commit adding this report on `codex/typeaware-wave-06`.
Commands and outputs: final wave test PASS in 222.014s; bridge PASS; eight Node comparisons PASS; vet and gofmt clean.
Mutants: all three rule mutants and the promised-type mutant exit 0 and lose byte agreement; released-registry mutant loses the required panic; bridge and Node mutants caught below.
Not covered: the full root test matrix, nondefault rule configuration, arbitrary inputs outside these populations, and the pinned cohere CLI's unsupported `.a` lint inputs.

# Type-aware wave 06

## Claim and pins

Branched from `origin/codex/tsgo-c-library` at
`0d540f413625f016f20fea39761c7b184f335de6`, overriding the general unit's
`origin/main` base with its specific type-aware base. Read `CLAUDE.md`, the
language and memory documents, and the complete type-aware README, volume
report and coverage report before implementation.

The checkout initially exposed only `origin/main`. Fetching all origin heads
made the specified base available. All 268 fetched origin references were
searched for native ports and claims before the claim was committed and pushed.
The unneeded recursive submodule fetch was stopped after the origin heads were
available; no submodule pin changed. No rule was already ported or claimed, and
none was skipped.

Positions 16 through 18 in the linked combined volume counts, excluding the
base's 26 ports:

| Position | Rule | Compiler | Repository | Total |
| --- | --- | ---: | ---: | ---: |
| 16 | adamic/nominal-class | 18 | 2 | 20 |
| 17 | @typescript-eslint/no-duplicate-type-constituents | 19 | 0 | 19 |
| 18 | no-useless-return | 16 | 0 | 16 |

The claim is [claims/wave-06.md](claims/wave-06.md), pushed before code.
Cohere remains at `715ba94f3608a6500086b1076ce5cb7e51b836db`, its TypeScript-Go
submodule at `8d550c837c90bd1805b047b7eeccc2baac2d5e7a`, and the compiler corpus
at TypeScript v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.

## Implementation boundary

Each judgment is in its own `.a` file:

- `nominal_class.a` follows the existing flow walker and asks for existing
  metadata, class bases, signatures, property shapes and type identity. Class
  identity, derivation, generic arguments, type-parameter constraints, fresh
  expressions and nested flow paths are judged in native Adamic.
- `no_duplicate_type_constituents.a` compares constituent token text before
  checker type identity, recurses into parenthesized same-kind types, and emits
  the production rule's distinct separator and constituent removal edits in
  their original order.
- `no_useless_return.a` builds statement events and control-flow edges using
  numeric block identities. It preserves dead continuations, jump targets,
  try/catch/finally behavior, loop/finally exemptions, annotation requirements,
  and the rule's comment and statement-list fix declines.

The new `promised-shape` fact lives in `bridge/tsgo/checker/promised_shape.go`
and `promised_shape.a`. It exposes an annotated function's declared return,
unwrapping async promises with the checker. The native rule decides whether a
return is required. The shared Go dispatch changes by exactly one line; every
other unrecognized question retains its prior error. The dedicated new native
suite imports the new helper and rules, requiring no shared native registry
edit. No protected compiler file, existing rule source, previous report or
submodule was edited. All five new Adamic source files use `.a`.

`testdata/oracle_wave_06.go` is an overlay-built driver inside pinned cohere.
It invokes the three unmodified production `Rule.Run` implementations with
nil/default options, an independent program loader and the production shared
file cache. It imports no bridge code. The native suite uses its own parser,
walker and judgments, with the bridge supplying checker facts only.

## Byte comparison and sanitizers

| Population | Files | Nominal class | Duplicate constituents | Useless return | Total | Identical bytes | Fix edits |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Controls | 166 | 14 | 81 | 34 | 129 | 35,110 | 192 |
| TypeScript src/compiler | 77 | 18 | 19 | 16 | 53 | 22,060 | 53 |
| Frozen repository corpus | 287 | 2 | 0 | 0 | 2 | 19,880 | 0 |

These compare complete sorted finding records, preserving duplicates, byte
ranges, rule names, message IDs and descriptions, every fix range and text in
order, and every suggestion record. All three production rules emit zero
suggestions in these populations. These are complete output comparisons,
including per-file headers and totals, not count-only comparisons.

The controls include literal sources extracted from the pinned Go return and
duplicate rule tests plus targeted nominal, annotated/async/generator/getter,
control-flow, comments, parentheses, Unicode and CRLF cases. Each control gets
`export {};` to isolate its declarations. They run under default settings;
this is not a claim that every upstream option-specific assertion ran.

Both the checker C archive and native executable were built with
AddressSanitizer and UndefinedBehaviorSanitizer. Linux LeakSanitizer is enabled
by AddressSanitizer. Normal and sanitized output agree with Go on all three
populations; sanitizer stderr is empty. The released-handle question must exit
70 with `adamic: panic: invalid or released checker handle`.

[validation-wave-06/results.json](validation-wave-06/results.json) records
SHA-256, byte lengths, findings by rule and control-source hashes. Its compressed
stdout files retain complete Go, native, sanitizer, timed and mutant streams.
The relative manifests are exact copies of the base coverage corpus manifests.
Absolute runtime paths in finding headers account for differences from older
report stream lengths. Build products remain outside the repository in
`/workspace/wave-06-verified`.

## Mutants and negative observations

| Mutant | Observed result | What caught it |
| --- | --- | --- |
| Nominal class symbol equality reversed | Compiles, exit 0, empty stderr | Go finding bytes differ at 11,032 |
| Duplicate finding end advanced one byte | Compiles, exit 0, empty stderr | Go finding bytes differ at 5,148 |
| Useless-return fix end advanced one byte | Compiles, exit 0, empty stderr | Go fix bytes differ at 285 |
| Async promised-type unwrapping removed | Compiles, exit 0, empty stderr | Go finding bytes differ at 24,114 |
| Released registry deletion removed | Compiles, exit 0, empty stderr | Required released-handle panic 70 is absent |

The question mutant initially survived `Promise<void>`: the original void
check looked through the entire shape graph. Adding `Promise<undefined>` caught
the missing unwrap. Adding synchronous `Promise<void>` and an embedded-void
control then caught an incorrect extra finding at byte 24,562. The final native
check follows only the root and union/intersection members, matching Go's
`maybeTypeOfKind`. The final complete run catches the question mutant again.
The retained negative logs are labeled separately from the successful run.
An earlier control comparison also caught incorrect assumed checker flag masks;
the final pinned-flags test guards the values used by these rules.

The existing bridge regression was run in full on its default sample. It also
runs these mutants:

| Bridge mutant | What caught it |
| --- | --- |
| Input string length advanced one byte | ASan heap-buffer-overflow |
| Output string length advanced one byte | ASan heap-buffer-overflow |
| Released handle kept live | Stale-handle assertion |
| Type queried at source-file position | Independent Go oracle mismatch at byte 6 |
| Link opt-in guard removed | Unlinked build/C/JavaScript refusal check |
| C output free removed | LeakSanitizer |
| Region result allocated on heap | LeakSanitizer |

The filtered Node oracle additionally mutates a dedication string by appending
`!`; Node/native stdout disagreement catches it. Eight fixture comparisons
passed: functions, generic functions, closures, method closures, maps and text,
lone surrogates, sorting and string indexing. The filter matches the additional
generic and method fixtures because Go matches slash-separated test names.

## Timing observations

One isolated final observation per population, fresh programs and processes,
with builds outside the measurement and complete output compared again. The
preceding comparison runs warmed the filesystem cache. Times are for the
three-rule suite together, including load, native parsing, bridge calls,
judgments, formatting and output. They are not per-rule medians.

| Population | Native wall | Go wall | Native / Go | Native queries | Native query time |
| --- | ---: | ---: | ---: | ---: | ---: |
| Compiler | 24.476589s | 3.433777s | 7.128 | 1,121,072 | 8.934170s |
| Repository | 0.857843s | 0.266614s | 3.218 | 25,287 | 0.333453s |

Native is slower in both observations. Native compiler load was 0.311948s and
run 24.105288s; Go load was 0.279904s, listener time 3.040601s and run 3.118931s.
Native repository load was 0.089371s and run 0.760962s; Go load was 0.076203s,
listener time 0.118475s and run 0.175298s. The large query count is observed;
no profiling or optimization conclusion is claimed from these timings.

## Commands, setup and limits

Exact commands and complete logs are in
[validation-wave-06/commands.txt](validation-wave-06/commands.txt) and
[validation-wave-06/agreement.txt](validation-wave-06/agreement.txt).

- `bash cloud/setup.sh`: success; Go/clang/Node/submodules ready in 0s each,
  build cache warm in 77s, done in 77s. Go 1.27.1, clang 20.1.8, Node 24.19.0.
  `nproc` prints 5; cgroup quota is 4 CPUs (`400000 100000`), memory 17.6 GB.
  Environment is `/workspace/adamic-tools/env.sh`.
- Final wave test with both corpus environment variables: PASS, 222.014s.
- `go test ./bridge/tsgo/... -count=1 -timeout=15m -v`: PASS; bridge 114.333s,
  checker 0.482s, including the new question's identity and rejection tests.
- Filtered `go test ./internal/oracle ...`: PASS, 39.810s, eight Node fixture
  comparisons and the one-byte mutation.
- `go vet ./...`: exit 0, no output. `gofmt -l cmd internal bridge/tsgo
  stage1/cohere/typeaware`: no output. `git diff --check`: clean.

The pinned cohere CLI lint command exits 1 before linting, stating that `.a` is
not a TypeScript or JavaScript program input, despite this repository's
`sourceExtensions`. Its format-only command exits 0, but that does not prove
`.a` lint or formatting coverage and is not counted as a source gate pass.
The exact rejection is retained in `cohere-lint.txt`. Changing the pinned
oracle or introducing `.ts` copies would violate this unit's constraints, so
neither was done.

The full root `go test ./...` matrix was not run; the touched bridge packages,
new stage-1 wave tests, filtered independent Node oracle and root vet gate were
run. The stage-1 suite has no user-settings or suppression dispatcher;
`ignoreUnions`/`ignoreIntersections` configurations are not implemented here.
The comparison does not execute the generic fix-application engine or prove
agreement on arbitrary JavaScript/JSX, malformed syntax or all TypeScript
inputs outside the recorded controls and frozen corpora. No website, PR,
protected compiler change or change to the 26 previous ports is included.
