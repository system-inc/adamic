Current landing certificate: [wave02-ready](../grouped-accessor-pairs/evidence/wave02-ready/REPORT.md). The material below records historical validation on the original branch. Its private probes are not part of this landing.

Built three .a rule candidates: array-callback-return, arrow-body-style and base/consistency-no-bare-throw, after recovering the earlier reservations.\
Commits: earlier recovery 08bff6e9/2f48129f; claim 1f3be00e; rules aed33f8c, bd2d3f77, d982fe58; branch codex/lint-wave1-02.\
Checks: 441/447 upstream cases, selected/all-rule witnesses, and 222 compiler/stage1 files (666 rule/file pairs) compare byte for byte on Node, emitted JavaScript and sanitized native with Go cohere.\
Mutants: omitted void exemption, missing object parentheses and reported AggregateError all compile, run cleanly and are caught only by output comparison on each runtime.\
Uncovered: native CFG construction (explicit Go CFG data adapter), two valid shared-parser refusals, four malformed Go inputs, production harness integration and the full repository gate.

Existing work was pushed first. The older no-restricted-types reservation now has its real rule implementation, automatic edits, independent suggestions and full 209-file evidence. Google font display now has its complete URL/message policy and a semantic mutant, compared on all 16 original Go JSX inputs through a declared Go JSX adapter. Shared JSX source parsing remains blocked. Those changes and raw evidence were pushed in 08bff6e9 and 2f48129f before this claim. The preceding Tailwind/comment and non-null/alias ports remain pushed with their published integration limits. No additional rule reservation is made here.

Selection fetched every origin head, read 335 origin refs and 52 claim Markdown files, and matched complete public names including sentence punctuation. Main was ef3d907ecdc4c771b016f7d9c52372def057a340. All original 46 helper-ready entries were already on main or named in claims. The first three unclaimed syntax-only inventory entries were array-callback-return, arrow-body-style and base/consistency-no-bare-throw. Claim 1f3be00e was pushed before implementation. Inventory source was origin/codex/lint-inventory; all three needs_type_information and binding_only flags are false.

Each rule owns its descriptor, .a implementation/messages, original Go adapter, raw witnesses and mutant. No shared registration generator, test harness, parser, compiler, runtime or submodule source changed. The private validation tools remain inside directories owned by this worker. Their Go overlays patch scratch copies, not production files. Generated registries remain ignored. No new Adamic .ts file and no pull request was created.

Array callback return ports static/computed methods, typed-array from and Array.fromAsync positions, async/generator asymmetry, logical/conditional/parenthesized/IIFE ancestry, return boundaries, function names/head ranges, every option/report branch and each independent suggestion edit. The missing dependency is control_flow_graph.Graph.EndReachable. No inspected helper branch provides it. The private oracle parses each actual source with Go cohere and serializes only that boolean for each callback head, converted to UTF-16 positions. Adamic still parses the original source and computes callback selection, returns, findings and suggestions itself. No Go diagnostic is supplied as input. This explicitly certifies rule logic with a Go CFG adapter, not native CFG construction. Without data, checked braced value-return callbacks refuse with exit 70 on all three runtimes. ForEach and concise judgments do not require CFG data. The real rule is never replaced with a no-findings placeholder.

Arrow body style ports all three modes, valid decoded/positional options, object-return exemptions, all five messages, comment-preserving disjoint edits, semicolon whitespace, ASI hazards, comma/in/object parentheses, nested in operators, forced object parentheses and multiline returns. Trivia is cached once per immutable file. Its complete automatic edits use the owned FixedFinding model and private reporter/fixer. The private fixer compares proposed edits individually and final converged text against Go's real edit engine; it does not normalize Go output. Ordinary production reporting still needs a compatible multi-edit contract. The existing owned guard refuses unsupported ordinary reporting instead of dropping edits. origin/codex/lint-harness-dot-a at 2650ad59 supplies .a and suggestion support but is not merged here.

Bare throws implements all capture, declaration-time, lower-vocabulary and test path exemptions, normalized separators, seven built-in constructors, the AggregateError exemption, direct-new/bare-identifier restrictions, exact new-expression ranges and policy wording. The Go adapter returns the unmodified real rule. It has no repairs.

The private capture preserves each original filename's path fragments and suffix and includes filename in deduplication. This matters for bare-throw path gates. Capture ran the actual Go tests and collected 1,030 distinct registered cases; this unit selected 447: array 288, arrow 111, bare throws 48. The successful subset is 441: array 288, arrow 108, bare throws 45. Six exclusions are explicit, not passing cases:

- Arrow source rejected by both parsers and Go's fix engine: `var foo = () => { return bar } /* c */ + 1;`.
- Valid arrow source refused by the shared Adamic parser at offset 29: `for (a = b => { return c = d in e } ;;);`.
- Valid arrow source refused by the shared Adamic parser at offset 25: `for (a = b => { return c in d ? e : f } ;;);`.
- Malformed bare throw rejected by Go before fixes: `export function f(): void { throw new Error( }`.
- Malformed bare throw rejected by Go before fixes: `export function f(): void { throw }`.
- Malformed bare throw rejected by Go before fixes: `export function f(): void { throw; }`.

The Node classifier temporarily throws on parser panic only to enumerate coverage. Every parity, mutant and refusal execution uses the unchanged Adamic runtime. Shared parser or recovery behavior is not edited. No findings-plus-fixes claim is made for the six excluded sources.

The compiler corpus is all 77 TypeScript v6.0.3 compiler files at 050880ce59e30b356b686bd3144efe24f875ebc8 plus all 145 .ts/.a files in this branch's stage1 tree. Inputs are copied to a frozen temporary snapshot while preserving filename fragments before comparison. There are no compiler/stage1 exclusions. The final release comparison passed 97.188s with 36,861,977 identical output bytes across Go, source Node, emitted JavaScript and ASan/UBSan native. Reports, descriptions, IDs, byte ranges, every proposed automatic edit, every suggestion ID/message/edit, applied suggestions on witnesses/upstream, rejection metadata and final sources are compared. Compiler suggestions compare their full edit data without printing each repeated whole source. Array judgments disclose the Go CFG adapter throughout.

Final combined validation passed 193.257s: earlier non-null/alias witnesses (15,111 identical bytes), restricted-type upstream/witness parity (15,833 bytes), new selected witnesses (5,067 bytes), 441 upstream cases, the three semantic mutants, missing-CFG refusal, witness throughput and source throughput. Path lengths make byte totals differ between temporary runs. A separate selected-plus-all dispatch check with defaults passed 19.014s, 9,199 identical bytes. Passing the array rule's checkForEach options to every rule initially failed the existing strict comment-option decoder; all-mode now correctly uses per-rule defaults. No decoder was weakened.

A final review found the private fixer's no-progress rejection wording did not match Go's ReasonNoProgress. The wording was corrected, and a real Foo-to-Foo automatic replacement was added alongside the restricted-type witness options. Its comparison passed 26.971s, 16,565 identical bytes, including the exact rejection reason and unchanged source on all three runtimes. The final compiler release run above includes that correction. Earlier successful scope evidence remains separate from this follow-up.

Every credited mutant completed with exit 0 and empty stderr, including sanitized native; compiler errors, panics or sanitizer failures are not credited:

| Rule | Mutation | What comparison caught |
| --- | --- | --- |
| array-callback-return | remove allowVoid exemption for concise void body | extra function-head finding |
| arrow-body-style | remove object-return parentheses | changed automatic edits and final source |
| base/consistency-no-bare-throw | include AggregateError in banned constructors | extra new-expression finding |

Final TestFourthMutants passed 48.38s (15.91s, 14.86s, 15.32s). Each mutation was caught independently on source Node, emitted JavaScript and sanitized native. Missing-CFG refusal passed 15.19s with the exact panic text and exit 70 on each. That refusal is gap evidence, not a semantic mutant.

One optimized-native count run per cell, including startup, file reads, source parsing and rule work; formatting, repair and compilation are excluded:

| 222 compiler/stage1 files | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| Array callbacks, provided Go CFG data on native/Node | 2 | 1.54 | 2.06 | 9.68 |
| Arrow bodies | 46 | 18.85 | 36.66 | 173.37 |
| Bare throws | 82 | 59.47 | 82.88 | 383.10 |

Array native/Node receives precomputed Go CFG data while Go builds CFGs, so that row is not a like-for-like whole-pipeline speed comparison. The low finding count does not establish useful positive visitor throughput. Arrow seconds: native 2.440466405, Node 1.254858953, Go .265324624. Bare throws: native 1.378893950, Node .989340687, Go .214042215. Array: native 1.299108707, Node .972512018, Go .206522890.

Separate synthetic measurement, 1,000 repeated owned witnesses per rule:

| Witness workload | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| Array callbacks, same CFG caveat | 4,000 | 37,244.71 | 22,256.64 | 146,652.25 |
| Arrow bodies | 3,000 | 40,748.55 | 18,345.37 | 81,128.81 |
| Bare throws | 2,000 | 51,532.54 | 14,566.90 | 92,877.73 |

These are observations from one run, not extrapolations to representative application speed. Raw times are in evidence/wave02-fifth-final.log. Source measurement passed 30.53s and witness measurement 7.02s. The no-progress wording follow-up does not change visitor counts.

Commands from the repository after `source /workspace/adamic-tools/env.sh`; every test wrote directly to its log:

```sh
python3 stage1/cohere/lint/rules/base-consistency-no-bare-throw/prepare-validation.py /tmp/wave02-fifth-final
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-02-typescript GOFLAGS=-overlay=/tmp/wave02-fifth-final/overlay.json go test ./stage1/cohere/lint -run '^(TestFourthWitnesses|TestFourthUpstream|TestFourthMutants|TestFourthCFGRefusal|TestFourthThroughput|TestFourthSourceThroughput|TestThirdWitnesses|TestRestrictedTypesParity)$' -count=1 -v -timeout 20m > /tmp/wave02-fifth-final.log 2>&1
python3 stage1/cohere/lint/rules/base-consistency-no-bare-throw/prepare-validation.py /tmp/wave02-fifth-no-progress
GOFLAGS=-overlay=/tmp/wave02-fifth-no-progress/overlay.json go test ./stage1/cohere/lint -run '^TestRestrictedTypesParity$' -count=1 -v -timeout 10m > /tmp/wave02-fifth-no-progress.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-02-typescript GOFLAGS=-overlay=/tmp/wave02-fifth-no-progress/overlay.json go test ./stage1/cohere/lint -run '^TestFourthCompiler$' -count=1 -v -timeout 20m > /tmp/wave02-fifth-compiler-release.log 2>&1
python3 stage1/cohere/lint/rules/base-consistency-no-bare-throw/prepare-validation.py /tmp/wave02-fifth-all-defaults
GOFLAGS=-overlay=/tmp/wave02-fifth-all-defaults/overlay.json go test ./stage1/cohere/lint -run '^TestFourthWitnesses$' -count=1 -v -timeout 10m > /tmp/wave02-fifth-all-defaults.log 2>&1
GOFLAGS=-overlay=/tmp/wave02-fifth-final/overlay.json go test ./stage1/cohere/lint/registry -count=1 -v > /tmp/wave02-fifth-registry.log 2>&1
go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -v -timeout 10m > /tmp/wave02-fifth-oracle.log 2>&1
```

Registry deterministic generation and 11 rejection mutants passed .050s. The filtered input oracle passed .245s with six probe cache hits. This is a bounded worker gate, not the full repository gate or uncached integration gate. Source and adapter formatting/diff checks passed. Toolchain setup with the prior owned overlay passed: Go, clang, Node and submodules ready in 0s; build cache warm 16s; done 16s; nproc 5, cgroup four CPUs. Go 1.27.1, clang 20.1.8, Node 24.19.0. The setup log lives in the earlier restricted-type evidence directory.

Raw failures are preserved under evidence/: the wrong bare-throw test prefix initially captured zero cases and was corrected; malformed sources stopped the initial upstream comparison, then both parsers were classified explicitly; a first source-corpus run failed because an owned source file changed between its Go and emitted-JavaScript reads, leading to frozen input snapshots; all-mode initially supplied another rule's options to a strict decoder. None of these failures is counted as passing coverage. The previous no-restricted-types/font report records its own raw failures and limited adapter evidence.

Not certified: native CFG building, source-level JSX font parsing, two valid for-initializer parser cases, malformed-source recovery parity, exhaustive invalid-option rejection/error wording, arbitrary generic parse-failing fixes, fuzzing, production multi-edit integration, or the full repository gate. All available rule logic is ported; the exact shared boundaries are stated rather than filled by changing files outside this worker's territory.
