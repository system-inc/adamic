Built: rebased wave-15 onto origin/area/stage1-lint d65a8f931, including main 39638d9e2; no new claims or rule implementation changes.
SHAs: previous remote e5ea66ba0; tested rebased code cd2ebcfdd; this evidence commit follows in git log.
Checks: nine type-aware suites PASS 945.411s; bridge PASS 70.873s/checker 0.367s; Node PASS 0.521s; runtime PASS 7.909s; registry PASS 0.023s; three adapter scripts PASS; vet clean.
Mutants: 22 legacy finding mutants and eight JSX/Unicode finding mutants caught by independent bytes; three JSX refusal mutants and released-registry guard caught; Node one-byte and registry rejection checks passed.
Uncovered: full repository gate and integrated JSX checker wiring, memo/capture analysis and Go quoted identifier rendering; existing regex ports remain partial.

The requested lint area advanced with allocation, string construction and string search runtime changes. All 37 own commits rebased cleanly. Main and area remain ancestors. No shared file or protected compiler file was edited. The inherited runtime changes were accepted. The only additional branch edits are this report and verification logs inside the owned fragment rule directory.

Commands used with /workspace/adamic-tools/env.sh sourced; every test wrote to a log:

    bash cloud/setup.sh
    nproc
    go test -v -count=1 -timeout=30m ./stage1/cohere/typeaware -run '^TestWave15|^TestCoverageAgreementAndMutants$'
    go build -o /tmp/wave15-d65-adamic ./cmd/adamic
    python3 stage1/cohere/typeaware/rules/react-jsx-fragments/testdata/check.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_indirect.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_unicode.py
    go test -v -count=1 -timeout=15m ./bridge/tsgo/...
    ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=10m ./internal/oracle -run '^TestTheOracleCatchesOneByte$|TestRuntimeLastIndex'
    go test -v -count=1 -timeout=10m ./internal/native -run '^TestRuntimeReleasePaths$|^TestRuntimeStringEquality$'
    go vet ./bridge/tsgo/... ./stage1/cohere/typeaware ./stage1/cohere/lint/registry
    go run ./cmd/lint-registry
    go test -v -count=1 -timeout=10m ./stage1/cohere/lint/registry -run 'TestDeterministicRegeneration|TestDescriptorRejections|TestAdamicRuleModule'

All commands passed. TMPDIR=/tmp/wave15-d65-tmp, mode 1777. ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-15-typescript. Both wave-15 and inherited coverage manifest variables used /workspace/wave-15-compiler.manifest (77 files) and /workspace/wave-15-repository.manifest (287 files). Wave artifact variables used /tmp/wave15-d65/{original,continuation,leaked,core,regex,throw,arrow,backreference,coverage}. The Python scripts used ADAMIC_WAVE15_SIXTH_COMPILER=/tmp/wave15-d65-adamic and their respective artifact variables /tmp/wave15-d65-sixth, /tmp/wave15-d65-indirect and /tmp/wave15-d65-unicode. Exact test output is in validation/d65-rebase.

Inherited corpus: compiler 16,589 findings/7,120,921 bytes and repository 180 findings/85,151 bytes, ordinary and sanitized. Original three: compiler two findings/5,629 bytes and repository one finding/18,704 bytes. All findings, fixes and suggestions serialized by these comparisons remained identical. This does not assert full new JSX parser/checker corpus agreement.

Every successfully executing legacy finding mutant was caught by independent Go bytes: before, cast, methods, coercion, caller, parameter, invariant, optional, alias, unused, arrow, backreference, listener, mock, label, throw, leaked, flags, suggestion, default, find, sort. Released handles panic 70; the released-registry mutant compiles and exits zero and is caught by the expected panic. JSX adapter byte mutants: fragment property, intrinsic predicate, construction label, branch order, provider factory, class owner and function kind. Unicode Lu-to-Ll mutant caught by Go and Node bytes. All eight compile and execute successfully. Three compiling missing-fact/refusal mutants exit zero instead of required 70. These adapters own no live checker handles.

Snapshot controls: 17 findings plus fragment/undef option variants; expanded context 22 findings/7,770 bytes. Unicode predicate matches Go and Node over all 1,114,112 code points, including 1,886 uppercase points/10,634 bytes; native 0.250s and sanitized 2.806s. Upstream Go SatisfiesExpression panic exit 2 versus native refusal exit 70 remains explicitly outside byte agreement.

Native against Go, original three compiler: 2.104930s versus 0.436505s; repository 0.327501s versus 0.134633s. Arrow compiler: 2.186459s versus 0.355161s. Native is slower. Snapshot and Unicode timings are different workloads from Go whole-source checking, so no speedup ratio is asserted.

Setup: Go, clang, Node and submodule readiness each 0s; cache warming 94s; total 94s. nproc 5, quota four CPUs, 17.6GB memory. Removed only obsolete own ELF/archive artifacts from named wave-15 scratch directories: 82 files, 3,171,144,827 bytes. Logs, sources and canonical streams were retained.

Remaining shared gap was rechecked in stage1/cohere/lint/context.ts: RuleContext carries source, parser, scanner and parent information but no checker Program or declaration-fact provider. The existing JSX adapters consume manually supplied declaration facts and have not been wired into the registry's supplied-node factory/visitor contract. Constructed-context also needs memo stability/capture analysis and Go-compatible quoted identifier output. The earlier claim-path certification failure remains documented in AREA_REBASE_REPORT.md: the shared validator requires stage1/cohere/lint/claims, while this historical reservation is under stage1/cohere/typeaware/claims. No receipt, completed integrated port or new claim is asserted. No validator was bypassed or changed. The user-requested existing partial rebase is pushed only to codex/typeaware-wave-15 with an exact lease against e5ea66ba0. No main or area push.
