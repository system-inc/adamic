Built: three additional native Nexus rules: no-process-exit-after-output, no-uncleared-race-timeout, require-blocking-standard-streams.
Commits: first batch 41b6c3e1; second claim f3f30a65 pushed before code; implementation is the commit containing this report.
Commands and outputs: isolated gate PASS 121 projects/160 findings, 287 repository and 77 compiler roots, normal and ASan/UBSan/LSan; checker, vet, formatting, lint, filtered Node passed.
Mutants: three rule decisions, three checker answers, and released-program registry retention; individual results below.
Not covered: the full repository gate and shared-harness JavaScript integration; production fixes and suggestions are empty for these rules.

The initial three claims were completed and pushed before fetching and claiming
this batch. The claim scan and exclusions are recorded in
`../claims/wave-20.md`; no additional claims were made after this batch.

The unit uses new `.a` files only. Native decisions and control-flow/import walks
live under the three owned rule directories. The new checker questions live in
three isolated Go files and three isolated Adamic decoder files. The only shared
code edit is one logical registration entry per question in checker/facts.go
(two physical lines each after gofmt). No generator, shared test harness, parser,
compiler, or forbidden native/lower/oracle source was edited.

Environment setup was `bash cloud/setup.sh`, then
`source /workspace/adamic-tools/env.sh`. Its timing lines were Go ready 0s,
clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 163s,
done 163s. `nproc` is 5, with a four-CPU quota. Versions: Go 1.27.1,
clang 20.1.8, Node 24.19.0. This batch reuses that successful setup.

Validation is isolated from the shared harness. `README.md` gives its replay
command and the pinned compiler revision. The Go oracle invokes the unmodified
production cohere rules, with its own checker/program loader and traversal and
no bridge imports. All canonical finding fields, fixes, and suggestions are
compared as bytes. These three production rules have no fixes or suggestions;
empty fields are verified, and no nonempty fix/suggestion coverage is claimed.

The committed fixture tape contains all 121 upstream input projects. The first
capture included 107 projects. A resolved-callee answer mutant survived that
subset. Inspection found 14 more helper-call projects behind
RunTypedFilesWithSetup; both original test entry points were run with their
original expected findings, and their inputs were added. A helper declared
never exposed an incorrect type-bit assumption, corrected to the pinned checker
Never bit 262144 and asserted in the checker tests. The expanded byte oracle
then killed the originally surviving answer mutant.

The first compiler-corpus attempt exposed a Go panic from asking Text() of an
object binding pattern in declaration ancestry. Text extraction now accepts only
textual name kinds; structured names retain their kind/span with empty text.
The checker regression explicitly probes destructuring and computed names.
The failed runs are preserved as evidence rather than reported as passes.

Stage 0 refused a nested construction shape; factories outside constructors
avoid it within owned code. Parser differences in anonymous-function bodies,
initializer tokens, for slots, and top-level await are handled locally; the
existing public awaitContext knob is used before file parsing. None required
changing the shared compiler or parser.

Source validation used the pinned Go cohere formatter/linter against in-memory
virtual TypeScript names for the `.a` sources, without physical `.ts` ports.
All 25 files format idempotently and configured lint reports zero findings.

The final full gate command was:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/no-process-exit-after-output/verify.py \
  --repository /workspace/adamic \
  --artifacts /workspace/wave20-validation/next/passing \
  --compiler /workspace/wave20-typescript \
  --cases /workspace/wave20-validation/next/restored \
  > /workspace/wave20-validation/next/passing-gate.log 2>&1
```

It passed all 121 projects, totaling 160 findings. Normal and ASan/UBSan/LSan
runs have identical full bytes and empty sanitizer stderr. The frozen repository
manifest has 287 roots, 18,485 output bytes, zero findings; the compiler manifest
has 77 roots, 5,241 output bytes, zero findings. Both have identical Go/native
bytes in normal and sanitized runs. Positive controls for each rule are required,
so the zero-finding corpora cannot make the gate vacuous.

Additional commands:

```sh
go test ./bridge/tsgo/checker -count=1
go vet ./...
go test ./internal/oracle -run 'TestNativeAgreesWithNode/.*/(classes|functions)|TestTheOracleCatchesOneByte' -count=1 -timeout 5m
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(classes|functions|closures)[.]a$' -count=1 -v -timeout 5m
```

Checker PASS 0.410s; vet clean. The first oracle command exercises the one-byte
oracle failure probe; it passed in 5.245s. The second runs functions, generic
functions, classes, closures, and method closures against Node plus the existing
backend/native/sanitizer comparisons; PASS 16.281s. Gofmt and git diff checks
are clean. Formatting and configured lint logs cover all 25 new Adamic files.
The entire repository test gate was not run.

Every deliberate rule and checker-answer mutant exits 0 with empty stderr;
only the independent Go/native byte comparison catches it:

| Mutant | Fixture | First differing byte | Detector |
| --- | --- | ---: | --- |
| output | case-1055269108 | 121 | Full finding bytes |
| timer | case-1049730460 | 119 | Full finding bytes |
| blocking | case-1105167505 | 1942 | Full finding bytes |
| declaration-chain | case-1049730460 | 119 | Full finding bytes |
| resolved-callee | case-1148031436 | 121 | Full finding bytes |
| module-records | case-1170353498 | 1415 | Full finding bytes |

The released-handle probe for each of the three new questions returns panic 70
with `invalid or released checker handle`. The registry-retention mutant returns
0 for each question and is caught by the required panic assertion. Registry
mutation is a Go overlay, never a repository edit.

Timings exclude builds and run three alternating native/Go rounds without
concurrent builds or tests. Every timed output is byte-identical and its SHA-256
is saved in validation/bench.json. Medians in seconds:

| Corpus | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| repository | 0.314631 | 0.169244 | 1.859 |
| compiler | 1.972463 | 0.415907 | 4.743 |

Observation: native is slower on these isolated entries. No speedup is claimed;
shared profile compilation is outside this unit. No emitted-JavaScript rule
comparison or shared `.a` harness integration is claimed. There is no remaining
rule-local blocker, and no fourth new rule has been claimed.

Validation includes compressed complete per-command stdout/stderr logs, final
and failed-attempt summaries, benchmark hashes/phases, and mutant evidence.
The compressed upstream fixture tape and isolated gate are sufficient to rerun
the required Go/native comparisons without modifying shared files.
