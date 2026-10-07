Built: rebased the existing fifteen rule ports onto current origin/main and reran their byte oracles; no new rules claimed.
Commits: previous pushed tip ca76e25f; rebased tested source c5f50597b694e5c7ef604c39795ed9722aa42c56; this report is committed separately.
Commands and outputs: all five wave gates PASS in 518.361s; bridge PASS in 90.017s, checker PASS in 0.298s; filtered Node oracle PASS in 11.671s; vet, gofmt and diff checks exit 0.
Mutants: all fifteen rule mutants, seven additional semantic mutants, five released-registry mutants and the malformed callback-question mutant caught by their intended checks.
Not covered: full repository gate or three unfinished React ports; current main still lacks JSX grammar, with the published shared frontend integration outside this unit's territory.

Rebased codex/typeaware-wave-23 onto origin/main
(e011f8f60899586d6373a5ccb07335ad82cfbf3c) without conflicts. This is the only
branch this worker pushed; the local work branch is the inherited checkout.
The user's explicit rebase-and-push instruction overrides CLAUDE.md's default
history-rewrite prohibition for this worker-owned branch. Push uses an exact
lease on the previously verified remote tip, ca76e25f29619ef68ae581313626ec8695c48575.
No new rule or bridge implementation changed in this landing pass.

Every batch runs independent Go findings/fixes/suggestions comparisons, release
and ASan/UBSan/LeakSanitizer controls, each rule's successfully compiled mutant,
both frozen corpora in normal and sanitized native, and released-handle probes.
The compiler corpus has 77 roots and the repository corpus has 287 frozen roots.
These fifteen rules have zero findings on each frozen corpus; positive controls
are required separately. Corpus streams match at 5,318 and 18,485 bytes.

| Batch | Positive control findings | Identical bytes | Gate seconds |
| --- | ---: | ---: | ---: |
| Original three TypeScript ESLint rules | 39 | 13,510 | 107.48 |
| Three Nexus rules | 120 | 56,201 | 107.50 |
| eval, extend-native, func-assign | 208 | 124,352 | 103.92 |
| Three constructor rules | 54 | 25,157 | 66.91 |
| throw-literal, backreference, arrow-callback | 200 | 75,679 | 132.55 |

Control-stream byte sizes differ from older reports because the artifact paths
are serialized and changed in this replay. No output was normalized; native and
Go bytes are compared directly. Nondefault options and custom declarations in
these existing gates also passed. All output mutants below exit 0 with empty
stderr; only the byte comparison catches them. Released-registry mutants exit 0
where the original must exit 70. The malformed-question guard overlay exits 1
at the intended accepted-malformed-question assertion, after the normal checker
test passed.

```
wave_23_behavior_test.go:183: throw mutant: exit 0, empty stderr, byte oracle catches byte 56503
wave_23_behavior_test.go:183: backreference mutant: exit 0, empty stderr, byte oracle catches byte 3892
wave_23_behavior_test.go:183: arrow mutant: exit 0, empty stderr, byte oracle catches byte 336
wave_23_behavior_test.go:193: callback symbol origin mutant: exit 0, empty stderr, Go bytes catch byte 39601
wave_23_behavior_test.go:243: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_constructor_test.go:157: func mutant: exit 0, empty stderr, byte oracle catches byte 8955
wave_23_constructor_test.go:157: nonconstructor mutant: exit 0, empty stderr, byte oracle catches byte 1851
wave_23_constructor_test.go:157: wrappers mutant: exit 0, empty stderr, byte oracle catches byte 126
wave_23_constructor_test.go:209: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_core_test.go:174: eval mutant: exit 0, empty stderr, byte oracle catches byte 4102
wave_23_core_test.go:174: extend mutant: exit 0, empty stderr, byte oracle catches byte 8753
wave_23_core_test.go:174: func mutant: exit 0, empty stderr, byte oracle catches byte 71677
wave_23_core_test.go:203: eval-option mutant: exit 0, empty stderr, byte oracle catches byte 49957
wave_23_core_test.go:203: extend-option mutant: exit 0, empty stderr, byte oracle catches byte 3124
wave_23_core_test.go:242: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_next_test.go:153: collection mutant: exit 0, empty stderr, byte oracle catches byte 1853
wave_23_next_test.go:153: outcome mutant: exit 0, empty stderr, byte oracle catches byte 28878
wave_23_next_test.go:153: pure mutant: exit 0, empty stderr, byte oracle catches byte 17115
wave_23_next_test.go:167: ancestry-count mutant: exit 0, empty stderr, Go byte oracle catches byte 28878
wave_23_next_test.go:167: awaited mutant: exit 0, empty stderr, Go byte oracle catches byte 31935
wave_23_next_test.go:205: released handle: normal exit 70; registry mutant caught by required released-handle error
wave_23_test.go:123: throw mutant: exit 0, empty stderr, byte oracle catches byte 94
wave_23_test.go:123: reject mutant: exit 0, empty stderr, byte oracle catches byte 5411
wave_23_test.go:123: reduce mutant: exit 0, empty stderr, byte oracle catches byte 11061
wave_23_test.go:137: alias-argument mutant: exit 0, empty stderr, Go byte oracle catches byte 6558
wave_23_test.go:137: rest-callback mutant: exit 0, empty stderr, Go byte oracle catches byte 4554
wave_23_test.go:175: released handle: normal exit 70; registry mutant caught by required released-handle error
callback guard mutant: TestCallbackSymbolFacts FAIL, accepted malformed question; exit 1
```

Fresh single-run times, including loading and canonical serialization, are below.
They are observations from this gate, not best-of-five performance claims; the
bridge/Node checks overlapped part of the rule gate.

| Batch | Compiler native / Go seconds | Repository native / Go seconds |
| --- | --- | --- |
| Original | 2.030296 / 0.326052 | 0.313025 / 0.141347 |
| Nexus | 2.901176 / 0.903199 | 0.395021 / 0.168150 |
| Core | 3.972838 / 0.400059 | 0.356767 / 0.150075 |
| Constructor | 1.620590 / 0.307000 | 0.258632 / 0.140172 |
| Behavior | 2.419850 / 0.409184 | 0.300894 / 0.191057 |

Commands, after source /workspace/adamic-tools/env.sh:

```sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-23/typescript
export ADAMIC_WAVE23_COMPILER_MANIFEST=/workspace/wave-23/compiler.manifest
export ADAMIC_WAVE23_REPOSITORY_MANIFEST=/workspace/wave-23/repository.manifest
export ADAMIC_WAVE23_ARTIFACTS=/workspace/wave-23/landing/first
export ADAMIC_WAVE23_NEXT_ARTIFACTS=/workspace/wave-23/landing/next
export ADAMIC_WAVE23_CORE_ARTIFACTS=/workspace/wave-23/landing/core
export ADAMIC_WAVE23_CONSTRUCTOR_ARTIFACTS=/workspace/wave-23/landing/constructor
export ADAMIC_WAVE23_BEHAVIOR_ARTIFACTS=/workspace/wave-23/landing/behavior
go test ./stage1/cohere/typeaware -run '^TestWave23(Next|Core|Constructor|Behavior)?AgreementAndMutants$' -count=1 -timeout=30m -v > /workspace/wave-23/landing/rules.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/wave-23/typescript go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-23/landing/bridge.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|functions|closures|sorting)\.a$' -count=1 -timeout=10m -v > /workspace/wave-23/landing/node.log 2>&1
go test -overlay /workspace/wave-23/callback-guard-overlay.json ./bridge/tsgo/checker -run '^TestCallbackSymbolFacts$' -count=1 -v > /workspace/wave-23/landing/callback-guard-mutant.log 2>&1
go vet ./... > /workspace/wave-23/landing/vet.log 2>&1
gofmt -l cmd internal bridge/tsgo stage1/cohere/typeaware > /workspace/wave-23/landing/gofmt.log 2>&1
git diff --check > /workspace/wave-23/landing/diff.log 2>&1
```

The callback guard command intentionally fails; every normal gate exits 0.
Setup was already complete: Go 0s, clang 0s, Node 0s, submodules 0s,
cache warm 83s, done 83s; nproc remains 5. No setup rerun is claimed.

After rebase, a fresh React capability probe was built using the gate's rebuilt
stage0 and checker archive. It exits 0 on ordinary TypeScript, but exits 70 on
the three JSX controls at bytes 53, 101 and 83. Main, the bridge branch and the
harness branch still lack the JSX parser import. Published implementation
e715ef4a on codex/stage1-jsx-lint changes shared parser/scanner/lint files;
those remain outside this unit's allowed territory. React rule decision and
lowering ports remain unfinished, as documented in WAVE_23_REACT_BLOCKER_REPORT.md.
No React parity, mutant, sanitizer or performance completion is claimed.
No new rules were claimed and no PR was opened.

[validation-wave23-landing](validation-wave23-landing) retains compressed raw logs
and all batch stdout/stderr streams with SHA-256 fingerprints, plus exact fresh
React probe errors. Previous batch reports remain historical evidence; their
pre-rebase commit IDs are not the current branch IDs.
