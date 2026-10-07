Built: three default-option native React ports in .a; all eighteen earlier ports re-greened.
Commits: native ports c06def68 on main c01907a7; claim originally pushed as 57ba4d2f, rebased as 1a7ec93d.
Checks: 345 controls, 163 findings, 58,921 exact bytes; both frozen corpora have zero findings and match under sanitizers.
Mutants: three native rule mutants caught by complete bytes, two raw-question mutants by contracts, retained handles by required panic.
Not covered: shared registration and checker-linked emitted JavaScript; full gate, configurable options and every upstream fixture matrix.

The rules are `react/button-has-type`, `react/checked-requires-onchange-or-readonly`
and `react/display-name`. Their predicates, component detection, report spans,
messages and quoting execute in native Adamic. Go supplies raw syntax roles,
every resolved symbol declaration and existing type/literal facts. It supplies
no verdict, finding or repair. Unicode printability and single-rune capitalization
use native tables generated from the pinned Go toolchain's Unicode data.

The audited selection is in ../claims/wave-26-next-refresh.json: 417 origin refs,
25 ranked baseline/main ports, 148 claimed names, 24 available. These three were
the first available names under descending combined-volume and lexical tie order,
all with zero original corpus findings. The claim push was verified before code.
This batch adds only two shared checker/facts.go lines, one dispatch line per new question.
The shared registration generator and test harness are untouched.

Run after sourcing /workspace/adamic-tools/env.sh:

```sh
export TMPDIR=/workspace/wave-26-bridge-scratch
mkdir -p /workspace/wave-26-attributes
go build -o /workspace/wave-26-attributes/adamic ./cmd/adamic
go build -buildmode=c-archive -o /workspace/wave-26-attributes/checker.a ./bridge/tsgo/archive
go build -o /workspace/wave-26-attributes/checker-server stage1/cohere/typeaware/wave_26_react_attributes/testdata/checker_server.go
python3 stage1/cohere/typeaware/wave_26_react_attributes/validate.py > /workspace/wave-26-attributes/validation.log 2>&1
python3 stage1/cohere/typeaware/wave_26_react_attributes/validate_dependencies.py > /workspace/wave-26-attributes/dependencies.log 2>&1
python3 stage1/cohere/typeaware/wave_26_react_attributes/validate_question_mutants.py > /workspace/wave-26-attributes/question-mutants.log 2>&1
python3 stage1/cohere/typeaware/wave_26_react_attributes/validate_cost.py > /workspace/wave-26-attributes/cost.log 2>&1
```

The independent oracle builds through an overlay inside pinned cohere and calls
its unchanged production rules with their defaults. It imports no bridge code.
Go, native and ASan/UBSan/LeakSanitizer streams agree completely on file headers,
UTF-8 spans, full rule IDs, message IDs/text and all repair/suggestion fields.
All three production rules have no fixes or suggestions. Controls contain
79 button findings, 38 checked-input findings and 46 display-name findings.

Literal fixture extraction retains 345 parseable controls and excludes 16
nonsource/incomplete snippets through independent Go parse checks. It does not
claim every fixture's option matrix. Controls include merged declarations, all
React factory binding forms, shadowing, nested wrappers, static and inferred
display names, conditional and checker-proven button types, Unicode, CRLF and
nonprintable literal messages. The two frozen manifests remain 77 compiler files
and 287 repository files. TypeScript v6.0.3 is 050880ce59e30b356b686bd3144efe24f875ebc8;
cohere is 715ba94f3608a6500086b1076ce5cb7e51b836db and its typescript-go is
8d550c837c90bd1805b047b7eeccc2baac2d5e7a. Source hashes and logs accompany this report.

Each native rule mutant compiled, exited 0 and wrote empty stderr. Only the
production Go complete-byte comparison caught it:

| Rule | Mutant | Catcher |
| --- | --- | --- |
| button-has-type | Reverse the checker-permitted exemption | Complete diagnostic bytes |
| checked-requires-onchange-or-readonly | Reverse the checked-property gate | Complete diagnostic bytes |
| display-name | Reverse the named-component exemption | Complete diagnostic bytes |
| jsx-structure | Swap tag and attributes roles | Independent AST contract assertion |
| symbol-locations | Keep declaration zero only | Merged-declaration contract assertion |
| Both new questions | Keep a released registry entry | Expected panic 70 becomes exit 0 |

Unchanged .a source also executes on Node with a raw checker dependency server
and matches all control bytes. The server sends compiler facts and never
diagnostics. The CLI's emitted-JavaScript path exits 1 with
`Adamic 0.1 refuses an unlinked typescript-go library call; build with --tsgo <checker archive>`.
This shared CLI/linking gap and shared registration remain for integration; no
shared implementation was changed to bypass them. Source Node agreement is not
represented as emitted-JavaScript agreement.

The previous eighteen ports were re-greened on current main. The five three-rule
Go suites recompiled native and sanitized programs and reran their rule, question
and released-handle mutants. Four passed in the first concurrent attempt; the
first failed during a mutant archive link with `no space left on device`, after
its controls and three rule mutants passed. Its sequential retry passed in
105.738 s. The three prior React source suites re-greened all 129 controls,
69 findings, both frozen populations, sanitizers and all three rule mutants.
Every prior mutant and its specific catcher is recorded in the earlier wave-26
reports and ../wave_26_react_reporting/SOURCE_REPORT.md; current logs are retained
in this unit's validation evidence. This yields 21 native ports, 692 controls
and 457 findings across the separate control suites, including cross-rule reports.

The full bridge suite on the initial main base passed. The current-main rerun
ran through input/output length mutants, stale handle, incorrect source position,
link opt-in and missing frees, but exhausted scratch while creating the final
region probe. After removing only named obsolete binaries and archives from this
worker's own scratch, the complete sequential retry passed in 71.63 s. It
compared 1,600 positions across four compiler files and 54,982 bytes under
ASan/UBSan/LeakSanitizer, plus all seven foundation mutants. All source
files and logs were preserved. The original concurrent failure logs are retained.

The setup from this session succeeded: Go ready 0 s, clang ready 1 s, Node ready
1 s, submodules ready 1 s, build cache warm 135 s, total 135 s. nproc=5; cgroup
quota=4 cores; memory=17.6 GB; Go 1.27.1, clang 20.1.8, Node 24.19.0, Linux amd64.
All test output went directly to log files. The full uncached repository gate
was not run. Configurable option decoding, edit application, suppression, every
JSX/JavaScript project and unpaired-surrogate button messages are not claimed.

Quiet alternating process timings are recorded in validation/cost.log.gz and
validation/cost-runs.json.gz. Every timing sample checks complete diagnostic bytes.
These measure this three-rule suite, not a combined twenty-one-rule runner.

| Quiet process median | Native | Production Go |
| --- | ---: | ---: |
| Compiler | 7.555482 s | 0.576423 s |
| Repository | 1.212682 s | 0.212778 s |

Native is slower for this three-rule suite. Three alternating rounds ran without
competing builds/tests and checked every complete stream. This is an observation,
not a claim of Go-speed parity. The current-main filtered Node oracle passed in
2.156 s, including the one-byte comparison mutant. Vet, compiler formatting and
owned Go-file formatting passed with empty logs.

Prior mutant details: [first three](../WAVE_26_REPORT.md),
[next three](../WAVE_26_NEXT_REPORT.md), [third batch](../WAVE_26_THIRD_REPORT.md),
[fourth batch](../WAVE_26_FOURTH_REPORT.md), [fifth batch](../WAVE_26_FIFTH_REPORT.md),
and [the React HIR source ports](../wave_26_react_reporting/SOURCE_REPORT.md).
Current-main raw HIR agreement again covered 22 controls and 1,046,120 production
bytes, with the omitted-SSA and retained-handle mutants caught. The prior pure
reporting helpers again matched six records under native/Node/emitted JS/sanitizers;
all four diagnostic mutants and three no-input-refusal mutants were caught.
Those helper checks do not substitute for the separate native source analyses.
