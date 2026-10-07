Built: rebased all 40 own commits onto origin/area/stage1-lint b84a9d931, including current origin/main c7991b900; no new claims.
SHAs: prior remote a4f9263de; tested rebased code 824c2abd1; this evidence commit follows in git log.
Checks: nine type-aware suites PASS 943.892s; bridge PASS 78.216s/checker 0.413s; Node one-byte PASS 2.835s; five proven-type Node fixtures PASS 0.862s; registry PASS 0.026s; five owned adapter scripts PASS; vet clean.
Mutants: 31 finding-byte mutants caught by independent Go/Node bytes, four adapter refusal mutants and released-handle/registry guards caught; Node one-byte and registry rejection checks passed.
Uncovered: full repository gate and its 17 mandatory external-input checks, live JSX checker wiring, memo/capture analysis and registry certification; existing regex ports remain partial.

The requested landing base advanced with proven type predicates, cast relations, record runtime support and associated Node fixtures. Rebase completed without conflicts. No shared generator, harness, parser, compiler or protected file was edited. Both origin/main and origin/area/stage1-lint are ancestors. The only additional changes in this unit are the owned report and verification evidence.

Read-only reservation inventory fetched every origin branch: 636 origin refs, 134 distinct Markdown/JSON claim blobs under both typeaware/claims and lint/claims. All 197 ranked checker-dependent rule names occur in these claim documents; there are no unmentioned candidates. Matching uses full-name boundaries, not substring matches. The counts' findings summary row is excluded. No new rule was reserved or implemented. Existing partial and parked claims were not promoted to complete ports.

Commands with /workspace/adamic-tools/env.sh sourced, each writing to a log:

    bash cloud/setup.sh
    nproc
    go test -v -count=1 -timeout=30m ./stage1/cohere/typeaware -run '^TestWave15|^TestCoverageAgreementAndMutants$'
    go build -o /tmp/wave15-b84-adamic ./cmd/adamic
    python3 stage1/cohere/typeaware/rules/react-jsx-fragments/testdata/check.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_indirect.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_unicode.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_quoted.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_quoted_messages.py
    ADAMIC_TSGO_CORPUS=/workspace/wave-15-typescript go test -v -count=1 -timeout=15m ./bridge/tsgo/...
    ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=10m ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestProven'
    ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=10m ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/proven_(guards|class_guards|assertions|satisfies|upcasts)\.a$'
    go vet ./bridge/tsgo/... ./stage1/cohere/typeaware ./stage1/cohere/lint/registry
    go run ./cmd/lint-registry
    go test -v -count=1 -timeout=10m ./stage1/cohere/lint/registry -run 'TestDeterministicRegeneration|TestDescriptorRejections|TestAdamicRuleModule'

All commands passed. The first Node filter ran the one-byte test only; the corrected fixture subtest filter then verified all five inherited fixtures. The latter recorded native 15 misses, Node 10 misses, zero hits. No skips or failures occurred in these scoped logs. TMPDIR=/tmp/wave15-b84-tmp, mode 1777. ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-15-typescript. Wave and coverage compiler/repository variables use /workspace/wave-15-compiler.manifest (77 compiler files) and /workspace/wave-15-repository.manifest (287 repository files). Wave artifact roots are /tmp/wave15-b84/{original,continuation,leaked,core,regex,throw,arrow,backreference,coverage}. Adapter scripts use ADAMIC_WAVE15_SIXTH_COMPILER=/tmp/wave15-b84-adamic and sixth/indirect/unicode/quoted artifact variables /tmp/wave15-b84-{sixth,indirect,unicode,quoted}. The quoted-message script uses its fixed /tmp/wave15-quoted-messages scratch root. All logs and inventory are retained in validation/b84-rebase.

Compiler 16,589 findings/7,120,921 bytes and repository 180 findings/85,151 bytes match unmodified production Go under normal and sanitized native. Original three compiler two findings/5,629 bytes and repository one finding/18,704 bytes. Comparisons retain findings, fixes and suggestions. All 22 legacy successfully executing byte mutants are caught: before, cast, methods, coercion, caller, parameter, invariant, optional, alias, unused, arrow, backreference, listener, mock, label, throw, leaked, flags, suggestion, default, find, sort. Released checker handles panic 70; the compiling released-registry mutant exits zero and is caught by the expected refusal.

Eight existing JSX/Unicode finding mutants remain caught: fragment property, intrinsic predicate, construction label, branch order, provider factory, class owner, function kind and Lu-to-Ll. Three missing-fact/refusal mutants are caught. Quoted-name escape-prefix mutant compiles/exits zero and independent Go bytes catch it; unpaired-surrogate guard mutant compiles/exits zero and required panic 70 catches it. Total 31 finding mutants plus four adapter refusal mutants. The adapters own no live checker handles, so their controls are not presented as handle-lifetime tests.

Snapshot controls still match 17 findings plus two option variants; expanded context matches 22 findings/7,770 bytes under normal and sanitizer builds. Unicode uppercase classification matches Go/Node over all 1,114,112 code points; quoting matches 1,112,064 scalar values plus seven mixed strings/13,409,957 bytes on Go, native, sanitized native and emitted JavaScript executing on Node. Four complete Unicode identifier findings match 1,520 bytes. Production SatisfiesExpression panic exit 2 and native explicit refusal exit 70 remain outside successful byte agreement.

Original three compiler native 2.134604s versus Go 0.348194s; repository native 0.337615s versus Go 0.152074s. Arrow compiler native 2.139553s versus Go 0.337747s. Native remains slower. Scalar quoting native 0.960s/sanitized 7.015s; Go go-run includes compilation, so no comparable speed ratio is asserted.

Setup readiness: Go, clang, Node and submodules 0s each; build cache and total 121s; nproc 5, quota four CPUs, memory 17.6GB. To prevent disk exhaustion, removed 145 obsolete identified own ELF/archive artifacts totaling 5,584,717,066 bytes from named wave-15 scratch roots; logs, sources and canonical streams retained.

The shared RuleContext still has no live checker Program/declaration-fact provider. Supplied JSX snapshots are not integrated registry factory/visitor ports, and constructed-context memo stability/capture/reference analysis remains unported. Historical claim-layout certification limitations remain in AREA_REBASE_REPORT.md. No successful registry certification is asserted. Regex partial scanners remain incomplete under the RegExp requirement. The complete repository gate and its 17 mandatory external-input checks were not run, skipped, relaxed or claimed green. All actually invoked scoped comparisons had their inputs; pinned compiler population was mandatory in the coverage test. No main or area branch was pushed; only the own-branch rebase is published with an exact lease against a4f9263de.
