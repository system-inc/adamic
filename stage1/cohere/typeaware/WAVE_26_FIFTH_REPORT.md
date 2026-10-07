Built: native .a ports of prefer-promise-reject-errors, prefer-regex-literals and prefer-rest-params.
Commits: claim d5216097; implementation 5d8f17d1; evidence is committed separately.
Commands and outputs: rule suite PASS 218.781s; bridge PASS 118.470s; Node oracle PASS 44.417s.
Mutants: three rules and four questions caught only by finding/suggestion bytes; registry, seven ABI and Node mutants caught too.
Not covered: full gate, nondefault options, all upstream fixtures, emitted JavaScript rule runner; pinned CLI refuses .a.

# Wave 26 fifth batch

All twelve preceding claims were implemented, tested and pushed at 135bbbc7
before selection. This batch brings this branch's completed claims to fifteen.
Claim d5216097 was pushed before implementation. No sixth batch is claimed.

## Selection

Fetched all origin branches with `git fetch --prune --no-recurse-submodules origin
'+refs/heads/*:refs/remotes/origin/*'`. The audit inspected 356 remote refs,
base/main ports, and Markdown claims on every origin branch. Base remains
5afbdb83da2ed7ad9815657cd3f6ececd5294bf6; main remains
ef3d907ecdc4c771b016f7d9c52372def057a340. Conservatively excluding 132 ranked
names mentioned in claims and the existing base/main ports left 40 candidates.
The first three by combined compiler/repository volume descending and lexical
ties were prefer-promise-reject-errors, prefer-regex-literals and
prefer-rest-params, each with zero combined findings. Earlier candidates were
skipped as ported or mentioned in origin claims. The exact refs, claim blobs,
excluded names and selected subjects are in
[validation-wave-26-fifth/selection.json](validation-wave-26-fifth/selection.json).

## Implementation

Each rule is in its own `.a` file. Native Adamic performs every lint predicate,
message, span, repair and suggestion decision. The independent Go oracle loads
its own program, walks its own AST and calls unchanged production registry rules
with default options; it imports no bridge implementation.

Promise rejection distinguishes the global Promise by requiring every declaration
ambient. It recognizes static reject methods through parentheses/type wrappers,
checks optimistic error-producing syntax, and treats the bare global undefined
separately. Executor calls match opaque parameter symbols, including duplicate
parameter spellings, nested closures, defaults and local shadows. The production
asymmetry of logical-and, assignment and comma values is preserved.

Rest parameters distinguishes a present implicit arguments symbol with zero
declarations from unresolved top-level arguments and declared shadows. It exempts
the direct receiver of dotted access while reporting computed access and the
parenthesized receiver form, matching Go's immediate-parent test. It has no fixes
or suggestions.

Regex literals follows global RegExp references through aliases, assignments,
defaults, global-object members, destructuring and pass-through expressions.
It preserves flow-insensitive tracing, cycle protection, duplicated traces,
modified-global refusal, computed constant keys and actual read-symbol identity.
Accepted constant arguments are cooked strings/templates and the raw text of
String.raw templates on the resolved global String. Native checks arity, flags,
printable patterns, comments and preceding tokens, balances quantifiers/groups,
rewrites exact character spellings, pads neighboring tokens and emits suggestions.
It preserves the production grammar behavior, including withheld suggestions
for noncapturing/named-group and lazy-quantifier controls rather than silently
substituting a different JavaScript validity test.

Four new raw questions and separate Go/.a files supply syntax roles/declaration
names, ordinary/read symbol identities and declaration-file flags, regex character
and escape/class extents, and comment byte ranges. Go grammar/comment helpers
are isolated copies of pinned cohere's raw syntax machinery. Go supplies no
regex-rule verdict or replacement text. Their schemas, provenance and ownership
are detailed in [WAVE_26_FIFTH_FACTS.md](WAVE_26_FIFTH_FACTS.md).

Only four one-line switch registrations changed shared facts.go. The existing
binding-structure schema and decoder are untouched. No shared registration
generator, test harness, suggestion serializer or profile compilation file was
changed; no protected compiler file was touched. The own batch runner and test
file use existing type-aware helpers. New Adamic files are all .a. No native
shared-harness gap blocked these ports.

## Byte agreement and mutants

| Population | Roots | Findings | Suggestions | Identical bytes |
| --- | ---: | ---: | ---: | ---: |
| Generated controls | 65 | 94 | 45 | 44486 |
| Frozen repository | 287 | 0 | 0 | 18485 |
| TypeScript src/compiler | 77 | 0 | 0 | 5318 |

Ordinary native and ASan/UBSan/LeakSanitizer runs match production Go over every
population. Native comparison stderr is empty. Canonical streams include file
headers, UTF-8 ranges, full rule/message IDs and descriptions, all fix counts,
suggestion IDs/descriptions and every proposed edit range/text, duplicate findings
and final totals. Automatic fix counts are zero; the regex rule offers 45
suggestions in controls. Positive findings are 27 Promise rejection, 59 regex
literal and 8 rest-parameter findings, so zero corpus findings are not the sole
verification evidence. Full compressed streams and per-run hashes are preserved.

Controls cover static/optional/computed reject calls; type-only wrappers,
undefined shadows, arithmetic/logical/assignment/conditional/comma reasons;
executor symbols, defaults, duplicate parameters, nested closures and shadows;
implicit arguments across arrows and function boundaries, catches, local shadows,
dotted/computed accesses and shorthand uses; RegExp aliases, cycles, assignments,
conditionals, comma expressions, type wrappers, global-object aliases, destructuring
and constant computed keys; modified globals; static/dynamic/raw/cooked arguments;
comments, hashbangs, preceding tokens and neighbor padding; slashes and controls,
empty/non-ASCII patterns, invalid flags/groups/classes/quantifiers, unicode sets,
Unicode trivia and CRLF spans. Exact control text/config is in controls.json.

The direct question contract tests compare binding metadata with actual compiler
operations, syntax population with the compiler walk, pattern characters with
hardcoded extents and comments with source offsets. Malformed/wrong-kind requests
are refused. Checker tests pass separately in 0.904s and in the bridge gate in
0.975s. Both frozen populations' source hashes were reverified, all 77 compiler
and 287 repository roots. Original absolute file headers are preserved; relocating
roots changes those bytes.

Every following mutant compiles, exits 0 and emits empty stderr. Only the
independent production diagnostic-byte oracle detects the difference:

| Mutant | Change | First differing byte |
| --- | --- | ---: |
| prefer-promise-reject-errors | Remove logical-and from right-value error propagation | 3151 |
| prefer-regex-literals | Leave slash spelling unescaped in the suggested literal | 12183 |
| prefer-rest-params | Exempt computed access instead of dotted access | 8356 |
| preference-structure | Omit invocation arguments | 1217 |
| preference-binding | Return zero for symbol identities | 55 |
| regex-pattern | Return an incomplete raw character scan | 10946 |
| source-comments | Omit raw comment ranges | 25546 |

A preference-binding request after program release panics 70 with
`invalid or released checker handle`. The retained-registry mutant instead exits
0 and fails that required-refusal expectation.

The full bridge gate passes 100 C ABI queries, strings surviving release,
zero/stale rejection, distinct subsequent handles and Unicode ownership. Its
independent checker oracle compares 162 positions and 3261 identical bytes under
ASan/UBSan/LeakSanitizer. All seven existing ABI mutants are caught: input/output
lengths plus one by ASan heap-buffer-overflow; retained release by the stale-handle
assertion; wrong source-file type by byte mismatch at 6; removed link opt-in by
refusal; omitted C-buffer frees and heap allocation of a region result by
LeakSanitizer. These sanitizers do not instrument the Go heap.

The filtered Node oracle passes eight selected native/JavaScript/Node compiler
fixtures and catches its one-byte mutant. These are compiler regressions, not an
emitted-JavaScript execution of the checker-linked rule runner.

## Native time against Go

After all builds and regression tests finished, three isolated alternating rounds
measured the complete default suite's count-only throughput. All rounds agreed
on zero corpus findings. Medians are seconds; separate phase medians need not
sum to process medians.

| Corpus / implementation | Load | Run | Whole process |
| --- | ---: | ---: | ---: |
| Compiler native | 0.522705 | 6.303651 | 6.691446 |
| Compiler Go | 0.372084 | 0.145204 | 0.554959 |
| Repository native | 0.112120 | 0.629746 | 0.756329 |
| Repository Go | 0.149584 | 0.092112 | 0.341200 |

Native process medians are about 12.06 times Go's compiler time and 2.22 times
its repository time. Rounds vary: compiler native 5.593 to 8.484s and repository
native 0.690 to 1.789s. No build/test from this unit ran during these rounds.
Native issues 255 compiler and 313 repository questions, including complete raw
syntax transport per root and relevant binding/comment queries. Complete AST
transport and decoding remain costly; this is not Go-speed parity. Raw rounds,
phase stderr and exact metrics are in measurements.json and benchmark.txt.
Agreement-test timing loops were not isolated and are not used for these claims.

## Commands and environment

Persisted toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc 5.
Source /workspace/adamic-tools/env.sh. Original setup log is
validation-wave-26/setup.txt: Go ready 0s, clang/Node ready 1s, submodules 1s,
cache warm 135s, done 135s. Cohere remains at
715ba94f3608a6500086b1076ce5cb7e51b836db; TypeScript corpus v6.0.3 remains at
050880ce59e30b356b686bd3144efe24f875ebc8. No submodule pin changed. Test
output went directly to logs without pipes.

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE26_FIFTH_ARTIFACTS=/workspace/wave-26-fifth-validation \
ADAMIC_WAVE26_REPOSITORY_MANIFEST=/workspace/wave-26-repository.manifest \
ADAMIC_WAVE26_COMPILER_MANIFEST=/workspace/wave-26-compiler.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-26-typescript \
TMPDIR=/workspace/wave-26-bridge-scratch \
go test ./stage1/cohere/typeaware -run '^TestWave26FifthAgreementAndMutants$' -count=1 -v -timeout 30m > /tmp/wave-26-fifth-agreement.log 2>&1
# PASS 218.781s.
go test ./bridge/tsgo/checker/... -count=1 -v > /tmp/wave-26-fifth-checker.log 2>&1
# PASS checker 0.904s; helper subpackages have no standalone tests.
TMPDIR=/workspace/wave-26-bridge-scratch go test ./bridge/tsgo/... -count=1 -v -timeout 15m > /tmp/wave-26-fifth-bridge.log 2>&1
# PASS bridge 118.470s, checker 0.975s.
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' -count=1 -v -timeout 10m > /tmp/wave-26-fifth-node.log 2>&1
# PASS 44.417s; eight selected cases and one-byte mutant.
go vet ./... > /tmp/wave-26-fifth-vet.log 2>&1
# Exit 0, empty output.
# gofmt -l cmd internal and all new Go files: exit 0, empty output.
python3 bridge/tsgo/profile/volume_bench.py /workspace/wave-26-fifth-validation/wave26 /workspace/wave-26-fifth-validation/wave26-oracle /workspace/wave-26-fifth-bench --corpus compiler /workspace/wave-26-typescript/src/compiler/tsconfig.json /workspace/wave-26-compiler.manifest --corpus repository /workspace/adamic/tsconfig.json /workspace/wave-26-repository.manifest > /tmp/wave-26-fifth-bench.log 2>&1
```

## Limits

The full repository gate and preceding type-aware suites were not rerun. Touched
bridge packages, the complete new default-rule suite and filtered Node oracle
passed. Nondefault allowEmptyReject/disallowRedundantWrapping options, all
upstream fixture permutations, arbitrary JavaScript/JSX populations, suppressions
and applying edits are not covered. This runner ports the production defaults.
The raw syntax projection does not verify Adamic's independent parser on these
corpora. The frozen selection/validation population excludes the new port files.

The emitted-JavaScript backend lacks the checker bridge, so this rule runner's
JavaScript could not be compared. No shared harness was changed to address that
integration gap. The pinned /workspace/wave-26-cohere --no-cache --no-fix command
on all ten new .a files exits 1 and refuses them as not TypeScript or JavaScript
inputs, recorded in cohere-refusal.txt. This is not a lint pass. Native builds,
findings, fixes/suggestions, mutants, released handles and sanitizer comparisons
are complete despite those integration limits.
