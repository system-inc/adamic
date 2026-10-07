# Wave 16 follow-up, awaiting shared bridge integration

The original three wave 16 rules are complete and pushed through d94b805f.
The premature follow-up was claimed in pushed commit 34ec5511 before code.
It selected nexus/correctness-no-global-listener-target-assertion,
nexus/correctness-no-leaked-number-render and
nexus/correctness-no-mock-on-module-namespace. Each has zero findings in both
ranked corpora. No further rules were claimed.

## Correction and blocker

Ahra's correction arrived after the first complete follow-up comparison passed.
It says: "Keep your changes inside your own rule directories" and "If anything
else blocks you, say exactly what it is and stop, rather than editing shared
files." The new raw checker questions require a dispatch hook in shared
bridge/tsgo/checker/facts.go. My three-line hook has been removed from that file;
it is preserved only as wave16_followup_registration.patch for integration.
No shared registration generator or shared test harness was changed. All shared
source files now match this branch's committed base. The new files are work in
progress, not registered ports: running their suite on the current tree refuses
unsupported checker questions. This is an explicit failure, not a silent empty
port. The report does not claim that this corrected tree passes the final gate.

Each new compiler question and Adamic decoder has its own file. The Go modules
register in a new additional_questions.go map. They return signature declaration
ancestry, emit module kind, symbol declarations' origins, asserted type/base
edges plus non-nullable assignability, and render operand flags/literal values/
constituents/constraints. None invokes a lint rule or returns findings or edits.
Native Adamic owns every rule judgment and diagnostic.

The base parser lacks JSX. Four isolated .a dependency files reuse the JSX
worker's parser, scanner, lookahead and JSX descent from
origin/codex/stage1-jsx-lint, with their exact originating SHA in each file.
Shared parser/scanner sources are untouched. The JSX controls are TypeScript
.tsx fixture data, not Adamic modules; every new Adamic module is .a.

## Observations before the correction

After sourcing /workspace/adamic-tools/env.sh:

```
ADAMIC_WAVE16_FOLLOWUP_ARTIFACTS=/workspace/wave16-artifacts/followup3 \
ADAMIC_WAVE16_COMPILER_MANIFEST=/workspace/wave16-artifacts/compiler.manifest \
ADAMIC_WAVE16_REPOSITORY_MANIFEST=/workspace/wave16-artifacts/repository.manifest \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave16-corpus/typescript \
go test ./stage1/cohere/typeaware -run '^TestWave16FollowupAgreementAndMutants$' \
-v -count=1 -timeout=20m > /workspace/wave16-artifacts/followup3-test.log 2>&1
```

PASS, 80.376 seconds. Independent production Go rules and native output agree
on complete findings, fixes and suggestions, preserving duplicates and sorted
canonical strings. Controls: 22 findings, 11,321 identical bytes. Frozen 287-file
repository: 0 findings, 18,485 identical bytes. TypeScript's 77 compiler roots:
0 findings, 5,780 identical bytes. All comparisons passed both normal and
ASan/UBSan/LeakSanitizer builds with empty native stderr.

Every rule mutant compiled, exited 0 with empty stderr, and was caught only by
independent Go bytes:

- listener-general: stop exempting HTMLElement, byte 699, 23 findings.
- render-number: treat number/bigint as silent, byte 6997, 14 findings.
- mock-range: add one to the report end, byte 3468, still 22 findings.

The released program probe panicked with exit 70 and
invalid or released checker handle. Retaining the handle in the registry made
the mutant exit 0 and the required refusal caught it.

An earlier run exposed an invalid Types() query on a non-union type, explicitly
refused by the checker. The implementation now reads constituents only for
unions/intersections. A second run's byte oracle caught a type-only namespace
import false positive; the native check now reads the parser's ImportClause
semantic field. Both failure logs are preserved, not counted as passes.

## Initial pause (superseded below)

The passing run predates formatting and the isolated registration map refactor.
The mutant replacement strings were updated for formatting. These final files
need integration of the shared dispatch hook, a fresh full comparison, independent
raw-question tests, foundation bridge regressions, and quiet native versus Go
timing measurements. No follow-up native performance measurement is claimed.
No full repository gate or all upstream option matrices were run. Existing setup
and timing evidence for the original finished rules is in WAVE16_REPORT.md.
The correction's instruction to stop on a shared-file blocker prevents completing
these remaining steps here. No more rules are claimed.


## Final validation under the renewed instruction

The subsequent instruction permits continuing past a shared-harness gap. Final
formatted .a sources were tested using a Go build overlay containing exactly the
pending dispatch hook. Shared repository sources and the shared harness remain
untouched. The rule-specific test owns this overlay; it is explicit in the test,
not an implied production integration. These three native ports and all their
non-shared implementation work are complete. The production shared registration
remains an integration dependency, with the patch included for its owner.

The same corpus command above with ADAMIC_WAVE16_FOLLOWUP_ARTIFACTS set to
/workspace/wave16-artifacts/followup-final passed in 84.112 seconds. Controls:
22 findings / 11,416 bytes (longer absolute path headings); repository:
0 / 18,485 bytes; compiler: 0 / 5,780 bytes. Normal and sanitizer outputs agree
byte for byte, including complete fixes and suggestions (these three rules have
none), with empty native stderr. The final normal-exit rule mutants were caught
at byte 709 (listener-general, 23 findings), 7067 (render-number, 14 findings),
and 3513 (mock-range, unchanged 22 findings). The released handle panic and
retained-registry mutant passed their expectations again.

```
go test ./bridge/tsgo/checker -run '^TestWave16QuestionFacts$' -v -count=1
# PASS 0.049s, direct identities, ancestry, origins, narrowing, bigint values,
# suffix and kind refusals; output in followup-question-test.log.
python3 /workspace/wave16-artifacts/followup-question-mutants.py
# Six compiling mutants caught by direct fact assertions or suffix refusal.
go test ./bridge/tsgo/... -v -count=1 -timeout=15m
# PASS: bridge 66.590s, checker 0.138s; output in followup-bridge-test.log.
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware
# Empty log.
gofmt -l bridge/tsgo/checker/*.go stage1/cohere/typeaware/wave16_followup_test.go
# Empty log.
```

The additional mutants change emit-module-kind by +1, signature ancestor name,
symbol identity by +1, assignability to its inverse, render flags by +1, and
accept a question suffix. Each Go test compiles and fails on its intended fact
assertion, not a build failure. Foundation bridge checks cover 100 C ABI queries,
162 independent checker positions / 3261 bytes under ASan/UBSan/LSan, and all
seven established length, released-handle, position, link, output-free and
region-ownership mutants. Logs are saved directly to files.

Three quiet alternating timing rounds use the final overlay-registered native
binary and independent Go oracle, with matching full-stream SHA-256 in every
round. Median seconds:

| Corpus | Native load | Native run | Native process | Go load | Go run | Go process |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Compiler | 0.235142 | 1.740387 | 1.998714 | 0.231178 | 0.042437 | 0.294769 |
| Repository | 0.060041 | 0.202823 | 0.279019 | 0.057853 | 0.042771 | 0.110760 |

Native is 6.78 times slower on the compiler and 2.52 times slower on the
repository. These zero-finding corpora measure traversal and checking overhead,
not a positive-finding throughput benchmark. Controls supply the positive cases.
Raw nanoseconds and timing stderr are in validation-wave16-followup-final.

An archive built without the overlay demonstrates the current gap exactly:
its native runner exits 70, with empty stdout and
`adamic: panic: unsupported checker question: emit-module-kind`.
This refusal log is preserved. The patch must be integrated before using the
ports outside their rule-specific validation runner. No shared files were edited.
Full repository testing, the complete upstream per-rule/options matrix, and
emitted-JavaScript comparison remain uncovered; the shared harness owner is
adding that comparison. nproc remains 5; setup evidence is unchanged from the
original wave 16 report. No toolchain or submodule pins changed.
