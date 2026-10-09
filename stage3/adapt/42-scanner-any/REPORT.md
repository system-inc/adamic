Typed scanner private payload/map owners and the own driver callback.
Measured against main c966d4c4; native compiler is landing tip 28285421 alone.
Node retains 509,014 skipped-trivia tokens and the complete 1,369,432-token dump.
Wrong-type, restored-any, owner-guard and byte-output mutants are caught.
Native scanner execution remains blocked by earlier compiler features.

The permanent adapter changes no body or public declaration. The private
payload is string | number, optional at its owner; the driver adds undefined
explicitly. Stock TypeScript checks all 25 substitution callers, including
numberOfCapturingGroups and the string | undefined target-name return.
The private keyword map is Map<string, KeywordSyntaxKind> by its dictionary
values and get(tokenValue) consumer. No unknown or cast-back is introduced.

Pattern 40 already implements the same private helper narrowing in the full
pipeline. Pattern 41 was read from fetched 1e920a31 without merging its unrelated
adaptations. The full apply therefore records one line changed by 42; the small
scanner profile records three edits. The own driver annotation is separate.
Both full and slice adapter runs are idempotent. Missing, duplicate and drifted
owners fail instead of receiving a guessed edit.

## Validation

The complete main and 42 lanes use their unchanged apply/oracle scripts, all
runners, eight workers, no filter, and the existing declaration sanction.
Both lanes PASS: 106,366 passing, one sanctioned failing test, zero pending,
and the same API baseline diff and 222 sanctioned declarations. Main lane took
638.970s (with overlap); 42 alone took 467.499s. Their oracles took 531.303s
and 363.014s respectively; 42 tests took 331.803s. Final machine-readable
equality and timings are in evidence/validation.json.
The expected oracle API failure is preserved, rather than accepted as a new
baseline. No shared lane/oracle expectation or baseline is edited.

All 725 emitted JavaScript/declaration artifacts match main; public
built/local/typescript.d.ts SHA256 is
edd733bf256465ddcbbd753a76cfe14cf4ef0a872ce5da7aa53072e1c9f93a9d.
The independent stock API proof has zero compiler/driver diagnostics and
byte-identical whole scanner/driver JavaScript. It checks the public callback's
source identity and retains every inferred substitution type.

The complete applied tree with 42 and both native-mode slice runs match the
fixed full-tree Node dump: 81 files, 509,014 skipped-trivia tokens, 860,418
retained-trivia tokens and 466 errors. Total 108,019,935 bytes, SHA256
5cce1570354cc48b5d9db246daf2abe8db3c21edba3c35e613da812a26920182.
All three comparison controls pass and token-end mutants fail diff.

Per-site wrong-type mutants change the actual private diagnostic annotation
to string-only (TS2345 at the numeric caller), the map values to string
(TS2769/TS2322), and the driver's annotation to string-only (TS2345 against
ErrorCallback). Actual restored-any/inference native mutants separately
reproduce 608:87, 100:23 and own-driver 13:42. Three independent concrete-type
Node/native controls match; changing exactly one native output byte in each is
caught by diff. Owner drift and duplication fail the adapter guard.

## Landing compiler and stops

No compiler implementation change or merge was made. Rebuilt cloud/land-area-next
28285421, which contains compiler area-next 337aa466. Compiler build exits 0:
25.147s wall, 4.208s user, 1.902s system. This is a warm, contended measurement.
Pinned @types/node at the compiler checkout supplies its declaration loader.
Both ADAMIC_NATIVE_SPLIT=0 and split=1 with ADAMIC_NATIVE_JOBS=5 refuse
corePublic.ts:9:5, the unchanged MapLike index signature.

The earlier three any observations are gone in separate, throwing-placeholder
checkpoints. Restoring each old type brings back its exact old diagnostic.
The concrete controls next reach:

| Site | Next checkpoint stop | Limit |
|---|---|---|
| 10, private payload | scanner.ts:696:32, string/number BinaryExpression | Earlier unrelated features have discovery placeholders |
| 12, keyword map | scanner.ts:100:58, new Map input is not typed pairs | Dependent on removed MapLike index signature and throwing keyword object |
| 14, own driver | main.a:14:156, JSON.stringify string \| number \| null | Other unsupported scanner bodies/initializers throw |

These are source/compiler admission measurements, not successful scanner runs.
Independent `.a` type-admission controls build without scanner placeholders.
Their unused payloads do not prove scanner payload representation at runtime.
No native scanner binary, token diff, runtime timing or ownership proof exists.
No full Go package gate, new internal/oracle fixture or counts.md change was made.

## Execution and retries

Setup used GOPROXY=https://proxy.golang.org|direct and its printed environment.
nproc=5; cgroup quota=4 cores. Timing lines: Go .028s, Node .029s, submodules
.063s, markdown .088s, clang .222s, Go build 37.788s, deferred test binaries
37.958s, warm cache 37.960s, total 37.986s. Toolchain: Go 1.27.1,
Node 24.19.0, clang 20.1.8.

The initial 42 oracle overlapped main and lost a worker without a diagnostic;
cgroup recorded an OOM kill. Its zero totals and missing API baseline caused
the lane to fail. It is retained as failed evidence, and 42 was rerun alone.
The first two stock-proof setup attempts lacked upstream type roots; the second
also repaired the virtual driver's root location. They are retained separately.
The passing proof points its type roots at the locked upstream Node declarations.
An unavailable /usr/bin/time attempt was replaced by Python monotonic time and
RUSAGE_CHILDREN. These prerequisite failures are not source failures.

Exact commands, commits, final lane equality, native reports, raw compressed
logs and mutant diagnostics are in evidence and the driver's scanner-any
folder. Scratch refs/source placeholders stay uncommitted and never pushed.
