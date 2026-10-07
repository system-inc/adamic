Built: three rule-local .a candidate ports; normal integration and Tailwind's JSX case remain blocked.
Commits: pre-code claim 01b7fb9a; module e19fd35b and default-parameter 10bab9b6; Tailwind/evidence bf7c6799 on codex/lint-wave1-03.
Commands and outputs: scratch parity PASS, 183 upstream cases plus 200 corpus files; three mutants PASS; registry/vet/filtered oracle PASS; overlay setup PASS in 12s, nproc 5.
Mutants: module_binding_ignored, required_parameter_ignored and direction_exemption_ignored compile/run and are caught only by Go output comparison on Node, emitted JavaScript and sanitized native.
Not covered: normal .a registration, the one upstream JSX case, arbitrary JSX, .a self-lint, and the complete repository gate.

## Selection and claim

All previous work was already pushed. A nonrecursive fetch of every origin head
updated main to ef3d907e and helpers to 6769b88e. The helper report links the
46-name ordered list in HELPERS.md. Its first 45 names occur in claim Markdown
under stage1/cohere/lint/claims on origin branches; the last does not.

Selected structure/tailwind-no-physical-direction, then the first two available
names in inventory.json's syntax ready for AST/API adaptation queue:
@next/next/no-assign-module-variable and @typescript-eslint/default-param-last.
Selection follows the new main-plus-claims criterion. Historical implementations
on other branches are not substituted for main ports or treated as claims outside
the named claims directory. The origin snapshot, 27 distinct claim blobs across
297 refs, and exclusion ledger are in evidence/selection.json. Own-branch claim
availability is evaluated at its previously pushed 0b44e1a tip.

The claim update was committed and pushed at 01b7fb9a before implementation.
No pull request was opened. Every authored Adamic source is .a; raw witnesses
use the inherited .ts.txt convention. No shared source or compiler file was edited.

## Implementations

Module-variable: visit variable statements, descend declaration lists, inspect
only plain identifier names, report the first module binding once against the
whole statement. Includes ambient and nested statements, while for-loop lists,
parameters and destructuring remain outside this rule's listener/decision.

Default-parameter: preserve seven function-like listeners, the body guard,
required/default/optional/rest categories and source-order reports. Classification
reads punctuation between parser children rather than scanning inside a binding
pattern or conditional type. Parameter modifiers are included in the report range.

Tailwind: preserve all 20 physical-to-logical mappings, exact and negatable
families, variant prefixes and rtl/ltr exemptions. Use the same class-shape regex
as Go, and split by Go Unicode White_Space rather than JavaScript's different
whitespace set. Visit every string and template's static pieces, preserving
whole-literal ranges and the .ts/.tsx filename gate. Render the exact catalog
message. All three rules have no fix or suggestion; their fixed text stays identical.

## Integration and parser limits

Default go run ./cmd/lint-registry exits 1 opening
rules/next-no-assign-module-variable/rule.ts. The existing registry requires
rule.ts, emits imports to it and restricts mutant modules to .ts. The test
harness copies only .ts, excludes .a from corpus discovery, does not compare
emitted JavaScript and loses captured file extensions. Renaming a JavaScript
fixture to TypeScript hides Tailwind's file gate.

The owned compatibility.patch proposes extension discovery, module copying,
.a corpus inclusion and emitted-JavaScript comparisons. validate.py applies it
only to scratch copies through a Go overlay, adds owned tests and preserves
capture filename extensions. No shared repository source is changed. The patch
and runner adapt the proposal already published by slot 12, scoped to the two
shared files present on this branch. Each capture rewrite checks a unique anchor.
An asynchronous request for a narrow shared-file territory exception remains
unanswered; silence is not approval. Normal integration is not certified.

One of Tailwind's 54 captured cases is a JSX attribute. The parser cannot read
it, so the validation output explicitly names it as a GAP and compares the other
53. A separate proving test runs unchanged Go's a_jsx_attribute test, then feeds
that exact source to Node and sanitized native. Both exit 70 with the same
parser panic: expected GreaterThanToken, got Identifier at 22. This is an explicit
refusal, not a clean finding result. See evidence/jsx.txt and testdata/jsx-gap.txt.
It prevents reporting this Tailwind port as complete over its upstream corpus.

## Commands and results

Source /workspace/adamic-tools/env.sh for every shell. TypeScript v6.0.3 is
050880ce59e30b356b686bd3144efe24f875ebc8, cloned into scratch. Go cohere remains
715ba94f3608a6500086b1076ce5cb7e51b836db. Go 1.27.1, Node 24.19.0,
clang 20.1.8, nproc 5, CPU quota 400000/100000. All test output goes directly to
log files; raw logs are preserved in evidence with .txt suffixes.

The default setup failed in cache warming because lint TestMain calls the
.ts-only registry. Go, clang, Node and submodules were ready at 0s; no final
warm/done line was printed. Workaround:

```
GOFLAGS=-overlay=/tmp/lint-wave1-03-corpora/overlay.json bash cloud/setup.sh
```

Exit 0: Go ready 0s; clang ready 0s; Node ready 0s; submodules ready 0s;
build cache warm 12s; done 12s on 5 processors, 17.6 GB. Setup logs record both
runs. The workaround validates the proposal, not installed default registration.

From the repository root, the owned validator is invoked as:

```
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate.py \
  --scratch /tmp/lint-wave1-03-next-retry --run '^TestContinuationCorpus$'
```

It writes test output to the scratch test.log and returns the test's status.
Additional calls use a distinct scratch directory, --typescript with the pinned
checkout, and the following --run values:

| Filter | Observation |
| --- | --- |
| ^TestContinuationCorpus$ | PASS 39.690s; 13 module, 117 default-parameter, 53 Tailwind cases; one explicit JSX gap. Capture contains 402 unique cases including inherited rules. |
| ^(TestCompilerAndStage1Agree\|TestContinuationShapes)$ | PASS 67.592s; 200 files, 12,488,778 identical output bytes; added corner cases 4,123 identical bytes. |
| ^TestMutants$/(module_binding_ignored\|required_parameter_ignored\|direction_exemption_ignored)$ | PASS 39.058s; all three compiling semantic mutants caught on all three paths. |
| ^TestContinuationThroughput$ | PASS 27.303s; release native, source Node and Go counts match in every round. |
| ^TestOwnedWitnesses$ | PASS 18.560s; 8,440 identical bytes across Go, source Node, emitted JavaScript and sanitized native using the finalized validator. |
| ^TestContinuationJSXRefusal$ | PASS 21.570s; Go's unchanged finding test passes, Node/native explicitly refuse JSX with exit 70. |

The first corpus attempt failed because Go dependency-download notices appeared
on stderr although its command succeeded. Dependencies then warmed and the retry
passed. The first JSX proving attempt used an incorrect cohere directory, failed
before running the test, and was corrected. Both failed attempts are preserved.
Neither failure counts as passing evidence.

Other checks:

- go test -overlay=<scratch overlay> ./stage1/cohere/lint/registry -count=1 -v:
  PASS 0.014s, including all existing metadata rejection probes.
- go vet -overlay=<scratch overlay> ./...: exit 0, empty log.
- ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run
  '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout 5m:
  PASS 0.770s; one native miss, one Node miss, no cache hits.
- git diff --check: exit 0.

Native parity and mutants use ASan/UBSan with normal Linux leak checking.
Every compared successful process must exit 0 with empty stderr. Compile errors,
sanitizer errors and panics cannot count as caught semantic mutants.

## Semantic mutants

| Rule | Change | Independent witness |
| --- | --- | --- |
| Module variable | Match moduleName instead of module | Go reports the whole let statement; mutant reports nothing. |
| Default parameter | Treat rest rather than required parameters as the last plain parameter | Go reports a = 1 before required b; mutant reports nothing. |
| Tailwind | Disable the rtl/ltr exemption except for an empty base | Go reports ml-4 only; mutant also reports rtl:mr-4. |

Each mutant completes successfully on source Node, emitted JavaScript and
sanitized native before its output is compared to unchanged Go. Raw first
mismatches are in evidence/mutants.txt.

## Findings per second

Best of three interleaved rounds, startup and parsing included. Each row measures
77 compiler files plus one file with 1,000 planted violations. This addition is
explicit because module and Tailwind have zero compiler findings; default
parameters have 24. Figures are observations on this worker, not compiler-only
throughput or a claim that native beats Go. Sanitizers are off for timing only.

| Rule | Findings | Native/s | Node/s | Go/s |
| --- | ---: | ---: | ---: | ---: |
| Module variable | 1000 | 858.06 | 1206.40 | 5366.82 |
| Default parameter | 1024 | 847.80 | 1134.71 | 5389.37 |
| Tailwind | 1000 | 808.37 | 1160.35 | 4540.36 |

Full repository tests, full Tailwind JSX parity, default integration, .a self-lint
and application of the shared compatibility proposal were not completed.
