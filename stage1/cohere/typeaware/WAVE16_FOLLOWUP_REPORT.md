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

## Limits and remaining work

The passing run predates formatting and the isolated registration map refactor.
The mutant replacement strings were updated for formatting. These final files
need integration of the shared dispatch hook, a fresh full comparison, independent
raw-question tests, foundation bridge regressions, and quiet native versus Go
timing measurements. No follow-up native performance measurement is claimed.
No full repository gate or all upstream option matrices were run. Existing setup
and timing evidence for the original finished rules is in WAVE16_REPORT.md.
The correction's instruction to stop on a shared-file blocker prevents completing
these remaining steps here. No more rules are claimed.
