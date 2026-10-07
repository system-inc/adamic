Built: rebased all 34 wave-15 commits onto current main b8fb957aa; no new claims or implementation changes.
SHAs: old remote dbf53f1a7; tested rebased code 8eeec71ef; evidence commit follows in git log.
Checks: nine type-aware suites PASS 978.160s; bridge PASS 75.547s, checker 0.642s; uncached Node PASS 1.067s; vet clean; JSX adapter scripts PASS.
Mutants: 22 suite finding mutants plus seven JSX byte mutants and three JSX refusal mutants caught; released-registry, bridge ownership and Node byte guards pass.
Uncovered: full repository gate, complete native JSX/regex/HIR integration and current partial memo/capture/Unicode work; no new rules claimed.

Main advanced from c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06 to b8fb957aa839a9e8cb0b54279dd9864fa317bd30. Its four-file change adds the inherited static-field lookup fix in emit_objects.go and the corresponding oracle fixture/count. All 34 branch commits rebased without conflicts; protected compiler files were inherited from main, not edited by this unit. Rechecked main after the gates; it remains b8fb957aa. The old/new commit map is validation/relanding-b8/rebased-commits.txt. Only codex/typeaware-wave-15 is published, using an exact lease against the old remote tip; no main or area branch is pushed.

Commands, with /workspace/adamic-tools/env.sh sourced and TMPDIR=/tmp/wave15-b8-tmp:

    go test -v -count=1 -timeout=30m ./stage1/cohere/typeaware -run '^TestWave15|^TestCoverageAgreementAndMutants$'
    ADAMIC_TSGO_CORPUS=/workspace/wave-15-typescript go test -v -count=1 -timeout=15m ./bridge/tsgo/...
    ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=10m ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(class_layouts|inherited_static_field_read|override_same_representation)\.a$'
    go vet ./bridge/tsgo/... ./stage1/cohere/typeaware
    go build -o /tmp/wave15-b8-adamic ./cmd/adamic
    python3 stage1/cohere/typeaware/rules/react-jsx-fragments/testdata/check.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_indirect.py

Both wave and inherited coverage corpus settings explicitly used /workspace/wave-15-compiler.manifest (77 files) and /workspace/wave-15-repository.manifest (287 files), with ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-15-typescript. Separate artifact variables point to /tmp/wave15-b8/{original,continuation,leaked,core,regex,throw,arrow,backreference,coverage}. JSX scripts used the freshly rebuilt compiler via ADAMIC_WAVE15_SIXTH_COMPILER, with artifacts in /tmp/wave15-b8-sixth and /tmp/wave15-b8-indirect. All test output was written to logs, never piped. Logs, byte-stream hashes, selected deterministic canonical stream archives and mutant/timing lines are retained in validation/relanding-b8.

Inherited agreement: compiler 16,589 findings and 7,120,921 bytes; repository 180 findings and 85,151 bytes. Ordinary and ASan/UBSan/LSan streams match Go. Original three: compiler two findings/5,629 bytes, repository one finding/18,704 bytes. Every suite retained its declared scope, including partial regex and synthetic leaked-render controls. Released-handle questions panic 70; a compiling released-registry mutant exits zero and is caught by that expected refusal.

All 22 suite finding mutants compile and exit zero before independent Go bytes catch them: before, cast, methods, coercion, caller, parameter, invariant, optional, alias, unused, arrow, backreference, listener, mock, label, throw, leaked, flags, suggestion, default, find, sort. JSX rechecks retain three original byte mutants, three refusal mutants, and four expanded context byte mutants (branch-order, provider-factory, class-owner, function-kind). Original JSX controls match 17 findings and two option variants; expanded controls match 19 findings/6,726 bytes under sanitizers. Production Go satisfies panic exit 2 and native explicit refusal exit 70 are still outside successful byte agreement. The partial adapters do not own live checker handles; bridge and legacy suites supply the released-handle evidence.

Native versus Go, original three end-to-end: compiler 2.029899s versus 0.359171s; repository 0.308000s versus 0.123902s. Arrow compiler 2.229984s versus 0.351515s. These are measurements, with no speedup claimed. Expanded JSX supplied-fact controls took 1.648ms versus Go source/checker/all-three-rule 72.569ms; those workloads differ and are not a comparable speed ratio.

Setup passed in 82s: Go/clang/Node/submodules readiness 0s each, cache warm 81s; nproc 5, CPU quota 4, memory 17.6GB. Removed only 79 old owned wave-15 ELF/archive artifacts, 3,119,589,228 bytes, preserving sources, logs and finding streams; exact paths are retained. Node oracle has 10 native and seven Node misses, zero hits. No full go test ./... was run.

This is landing verification, not completion of parked or partial claims. Registry listeners use named AST kinds, as corrected by Ahra; missing numeric kinds are not a blocker. The custom JSX snapshot adapters still need shared-context factories/visitors, live checker facts and owned registry oracle descriptors. Context memo/capture analysis and Unicode rendering remain unfinished. Earlier partial regex scanners are not promoted to complete ports under the RegExp requirement. No shared generator or harness was edited.
