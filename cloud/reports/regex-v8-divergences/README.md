# V8 RegExp compatibility unit

## Node 24.19.0 recheck, October 6

Rechecked directly on Linux x64 against explicit spec witnesses; integration
reports identical output on official darwin-arm64 Node 24.19.0 with the same
V8 13.6.233.17-node.51. The exact program and `node --version` output are in
[integration-reproduction.js](integration-reproduction.js) and
[integration-node2419.txt](integration-node2419.txt). All five drafts reproduce;
none of the existing refusals was dropped. The uppercase empty-result witness
is a real mismatch, but does not isolate ordering from singleton folding.

Added the missing scoped legacy negated-class alternative refusal. Lowering
pins both integration modifier witnesses. No test262 patterns or attributed
files are lost; the complete 127,369-row and 10,000-random Go oracles still pass.
C runs all 127,369 test262 rows and 9,911 accepted random rows in metered and
unlimited modes with zero disagreements. Go package: 11.392s; C: 48.013s.

Commands (output redirected to the archived `node2419-*.txt` files):

```sh
go test ./internal/regexp ./internal/lower -run 'TestV8|TestRegExpV8' -count=1 -v
go test ./internal/regexp -count=1 -v -timeout=20m
go test ./internal/native -run '^TestRegExpBytecode' -count=1 -v -timeout=20m
go vet ./internal/regexp ./internal/lower
```

A real mutant omitting the added legacy-alternative refusal fails the explicit
Node/spec recheck, naming `(?i:x|[^a-z])` with a missing refusal. The mutation
was run in an isolated copied package; the working tree was not mutated.
The current native-fixes environment's setup was 122s total/cache, Go 0s,
clang/Node/submodule 1s each; `nproc` 5, quota 4 CPUs. Native fixes continue on
their separate branch from current main after this recheck commit.

Branch: `codex/regex-v8-divergences`, cut from `origin/main` at `5d4c801`.
This unit follows the already pushed literal-object fixture commit `50a1dc8`;
it does not merge the matcher/protocol branches into main. No runner verdict,
classifier, cohere corpus, protected emitter/lowerer, or `constantPattern`
function was edited. No V8 reports were filed.

## Built

The shared stage-0 bytecode emission gate refuses visible incompatible shapes,
with a typed diagnostic naming V8's behavior and ECMA-262 sections. Its resolved
set comparison preserves masked singleton folding, safe multi-character class
strings, modifier resets, and ordinary operands. A separate conservative rule
rejects mixed empty/single/multi class strings, following the ruling.

The Go and C interpreters reproduce Node's input-dependent Unicode assertion
behavior inside surrogate pairs, while rejecting consuming half a pair. They
preserve initial lastIndex rewind and sticky interior retry. New source
fixtures check both backends, C sanitizer/release output, and allocation counts.

Five factual draft reports, minimal reproductions, and exact spec references
are in [docs/regexp-v8-divergences.md](../../../docs/regexp-v8-divergences.md).
Direct empty-alternative `exec` ordering **agrees** with the published/current
spec; its alleged ordering divergence was not reproduced. A replacement
slow-path **hang** was reproduced and bounded by a two-second subprocess limit.

The exact integration fixture is **refused** at its third regex, `[\q{Ss|x}]/iv`:
the given input works but `X` exposes V8's singleton bug. The accepted companion
uses `[\q{Ss}]/iv`, retaining both `AB`/`ab` cases and the long-s fold. It catches
the actual `copyS[j] = c` mutation via Node/native stdout disagreement.

## Before and after

Here “before” means the checked-in extraction's baseline acceptance; these
tables do not claim a fresh 1,879-test runner survey. Baseline reference/random
agreement is also preserved by the full current Go oracle. This main tip has
no DFA: actual C configurations are metered VM and unlimited VM.

| Corpus/configuration | Before accepted | After accepted | Loudly refused | Disagreements after |
| --- | ---: | ---: | ---: | ---: |
| test262 execution, Go | 127,369 | 127,369 | 0 | 0 |
| test262 execution, C metered | 127,369 | 127,369 | 0 | 0 |
| test262 execution, C unlimited | 127,369 | 127,369 | 0 | 0 |
| Fixed-seed random, Go | 10,000 | 10,000 | 0 | 0 |
| Fixed-seed random, C metered | 10,000 | 9,911 | 89 | 0 |
| Fixed-seed random, C unlimited | 10,000 | 9,911 | 89 | 0 |
| Extracted parser patterns, extra compatibility refusals | 0 | 0 | 0 | — |

The execution corpus has 2,791 distinct pattern/flag pairs. The parser corpus
has 5,746 rows. Compatibility refusals cost **zero test262 files or patterns**
in both extractions. Random collateral is **89 executions / 83 unique pairs**;
all agree with Node on their sampled `exec` input. Every refused pair and its
diagnostic is in [random-refusals.json](random-refusals.json). This includes
dead `{0}` occurrences; the mixed-empty rule is explicitly conservative.

An additional visible-shape sweep compares 4,650 cases from 310 pattern/flag
pairs. All 244 accepted pairs / 3,660 cases agree with Node. Refusals cover 66
pairs / 990 cases: 830 sampled inputs agree, 160 differ. Agreement on one input
does not make that pattern safe for both backends. A further 2,640 surrogate
assertion cases agree in Go; C's final 2,659 cases include those and 19 accepted
folding/modifier/set controls, in both configurations. Existing 890 lint input
cases, search tests, capture/backreference tests, and Unicode suites also pass.

## Commands and outputs

All test output went to files. Commands use:

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/regexp ./internal/lower -count=1 -v -timeout=20m
go test ./internal/native -run '^TestRegExp' -count=1 -v -timeout=20m
go test ./internal/native ./internal/lower -run 'TestRegExpBytecodeV8Node|TestRegExpV8Refusals' -count=1 -v -timeout=10m
go test ./internal/regexp ./internal/native -run 'TestV8CompatibilityAcceptedNode|TestRegExpBytecodeV8Node' -count=1 -v -timeout=10m
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -v -timeout=20m
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/regexp' -count=1 -v -timeout=20m
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/regexp|^TestCountsAreRecorded$' -count=1 -v -timeout=20m
go vet ./internal/regexp ./internal/lower ./internal/native ./internal/oracle
```

An earlier oracle filter, `TestCountsAreRecorded|TestNativeAgreesWithNode/.*regexp`,
selected the complete count gate but **no regex source subtests** because Go
filters slash-delimited segments separately. The corrected commands above
run the actual source fixtures (13 passing subtests, including one refusal).
The final combined run selects both fixtures and the complete count gate. Initial
attempts with `-timeout20m` were rejected as an unknown flag; corrected commands
above ran successfully. Initial modifier literal controls were rejected by
TypeScript TS18062; constant `new RegExp` controls now exercise Adamic's gate.

Allocation rows were added using the required complete update-counts run:
`go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout=20m -args -update-counts`.
Only the two new accepted fixture rows changed; the complete count gate was
then run without updates.

Package results: regexp **12.672s**, lower **27.885s**, regex native **75.618s**,
complete count gate **12.156s**; final combined source/count gate **6.742s**.
`go vet` exits 0. Raw outputs are archived alongside this report, including the
separate source oracle and final mutant results.

Setup: `bash cloud/setup.sh` succeeded, environment at
`/workspace/adamic-tools/env.sh`. Its timing lines: Go **0s**, clang **0s**,
Node **0s**, submodule **0s**, cache **121s**, total **121s**. `nproc` is **5**,
cgroup CPU quota is 4; memory 17.6 GB. Versions: Go 1.27.1, clang 20.1.8,
Node 24.19.0 / V8 13.6.233.17-node.51. Corpus test262 pin:
`7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd`.

## Real mutants

Each source mutation ran in an isolated checkout, failed on a semantic witness,
and was restored. All **10/10 caught**, rerun against the final implementation.
[run-mutants.py](run-mutants.py) reproduces them; use an isolated checkout with
the unit's files and toolchain/cohere workspace, never the working branch.

| Mutation | What caught it |
| --- | --- |
| Omit shared backend compatibility gate | Lowering expected V8/spec refusal |
| Expand V8 singleton q ranges, erasing refusal | Typed singleton refusal control |
| Restore modifier parser flags, erasing refusal | Typed modifier refusal controls |
| Omit mixed-empty refusal | Typed mixed-empty refusal control |
| `copyS[j] = c` in sets.go | Accepted folding source oracle: stdout differs from Node |
| Go search skips pair interiors | Go Node oracle: `\B` null versus `[2,2]` |
| C search skips pair interiors | Surrogate source oracle: stdout differs |
| Go consuming instructions accept pair halves | Go Node oracle: `(?!\W)` wrong index |
| C consuming instructions accept pair halves | Surrogate source oracle: stdout differs |
| Omit sticky interior retry | Go Node oracle: sticky `\B` null versus `[2,2]` |

## Limits

No full repository gate or fresh test262 runner survey was run; the complete
regexp and lower packages, all regex-native tests, complete counts gate,
filtered source oracle, and vet are the precise gates above. No new dynamic
pattern compilation or Symbol protocol work belongs to this branch. Scoped
modifier literals remain subject to the compiler's existing TypeScript target.
The compatibility model targets the recorded Node/V8 version, not future V8
releases. The replacement-hang control deliberately fails loudly if V8 fixes
that recorded behavior, prompting review of the corresponding refusal.
