Built: rebased all 42 own commits onto origin/area/stage1-lint d3a37422c, including origin/main b6b1538b0; no new claims or rule implementation edits.
SHAs: prior remote f9c3e7681; tested rebased code 358ebe7a9; this evidence commit follows in git log.
Checks: nine type-aware suites PASS 923.928s; bridge PASS 129.426s/checker 1.506s; affected Node fixtures PASS 1.197s and typeof mutants PASS 5.006s; registry PASS 0.061s; all five adapter retries PASS; vet clean.
Mutants: 31 independent finding-byte mutants, four adapter refusal mutants and released-handle/registry guards caught; four inherited typeof mutants, Node one-byte and registry rejection checks pass.
Uncovered: full repository gate with 17 mandatory external-input checks, integrated JSX checker wiring, memo/capture analysis and registry certification; existing regex ports remain partial.

The new base changes typeof lowering, tagged unions and runtime slot/dispatch support. Rebase was clean; no shared compiler, bridge, parser, harness, generator or rule was edited. Both required bases remain ancestors. Additional changes in this unit are the owned report and logs only.

Verification with /workspace/adamic-tools/env.sh sourced, every test writing to a log:

    bash cloud/setup.sh
    nproc
    go test -v -count=1 -timeout=30m ./stage1/cohere/typeaware -run '^TestWave15|^TestCoverageAgreementAndMutants$'
    go build -o /tmp/wave15-d3-adamic ./cmd/adamic
    python3 stage1/cohere/typeaware/rules/react-jsx-fragments/testdata/check.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_indirect.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_unicode.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_quoted.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_quoted_messages.py
    ADAMIC_TSGO_CORPUS=/workspace/wave-15-typescript go test -v -count=1 -timeout=15m ./bridge/tsgo/...
    ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=10m ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestTypeof|^TestNativeAgreesWithNode$/internal/oracle/testdata/typeof_'
    ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=10m ./internal/oracle -run '^TestTypeOf'
    go run ./cmd/lint-registry
    go test -v -count=1 -timeout=10m ./stage1/cohere/lint/registry -run 'TestDeterministicRegeneration|TestDescriptorRejections|TestAdamicRuleModule'
    go vet ./bridge/tsgo/... ./stage1/cohere/typeaware ./stage1/cohere/lint/registry

Final runs all passed with no skips. The first five adapter runs failed building native controls because the root runtime cache could not mkdir: no space left on device. These initial failure logs are retained. Cleanup removed only obsolete identified ELF/archive files from own wave-15 scratch roots: 91 files/3,291,741,990 bytes in /tmp and 91 files/3,661,634,618 bytes in /workspace. Sources, logs and canonical streams were preserved. All five scripts were rerun after cleanup and passed; no check was weakened. Setup itself passed: Go readiness 0s, clang/Node/submodules 1s each, build cache and total 129s; nproc 5, CPU quota four, memory 17.6GB.

TMPDIR=/tmp/wave15-d3-tmp, mode 1777. ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-15-typescript. Both wave and inherited coverage manifest variables used /workspace/wave-15-compiler.manifest (77 files) and /workspace/wave-15-repository.manifest (287 files). Artifact roots are /tmp/wave15-d3/{original,continuation,leaked,core,regex,throw,arrow,backreference,coverage}. Adapter scripts use the fresh /tmp/wave15-d3-adamic compiler and /tmp/wave15-d3-{sixth,indirect,unicode,quoted} artifact variables. The quoted-message script uses its fixed /tmp/wave15-quoted-messages scratch root. Logs and inventory are in validation/d3-rebase.

Compiler 16,589 findings/7,120,921 bytes and repository 180 findings/85,151 bytes match production Go under ordinary and ASan/UBSan/LSan native runs. Original three compiler two findings/5,629 bytes and repository one finding/18,704 bytes, including complete findings, fixes and suggestions. All 22 successfully executing legacy finding mutants are caught only by independent Go bytes: before, cast, methods, coercion, caller, parameter, invariant, optional, alias, unused, arrow, backreference, listener, mock, label, throw, leaked, flags, suggestion, default, find, sort. Released checker handles panic 70; the released-registry mutant compiles/exits zero and is caught by the expected panic.

Nine adapter finding mutants: fragment property, intrinsic predicate, construction label, branch order, provider factory, class owner, function kind, Lu-to-Ll and quoted escape prefix. Four adapter refusal mutants: fragment facts, undef facts, unresolved construction and unpaired-surrogate guard. Every mutant compiles and executes successfully before expected bytes or expected panic 70 catches it. These snapshots own no live checker handles.

JSX controls match 17 findings and both option variants; indirect context matches 22 findings/7,770 bytes. Unicode component predicates match all 1,114,112 code points. Quoting matches 1,112,064 scalar values plus seven strings/13,409,957 bytes against Go, sanitized native and emitted JavaScript on Node. Four complete Unicode identifier findings match 1,520 bytes. The known production SatisfiesExpression panic exit 2 versus native refusal exit 70 remains outside successful byte agreement.

Affected typeof dispatch/null fixture comparisons passed, and inherited constructor, string-literal, null and null-slot-presence mutant tests passed. The first filter selects fixtures but not the case-sensitive TypeOf mutant names; the second command explicitly ran all four mutant tests. No oracle fixture or check was changed.

Original-three compiler native 2.072824s versus Go 0.351864s; repository native 0.301499s versus Go 0.139892s. Arrow compiler native 2.210175s versus Go 0.321002s. Native is slower. Scalar quoting native 0.885s/sanitized 6.434s is a different workload from Go full-source checking; no comparable speed ratio asserted.

Fresh claim inventory: 658 origin refs, 136 distinct claim documents under typeaware/claims and lint/claims, all 197 ranked rule names covered and no unmentioned candidates. No new rules reserved. Remaining blockers are the shared RuleContext live checker Program/declaration-fact provider, registry factory/visitor integration and constructed-context memo/capture/reference analysis. Historical claim-layout certification limits and partial regex scanner limits remain. The full repository gate and its 17 required external-input comparisons were not run, skipped, relaxed or claimed green. Only the own-branch rebase is published, with an exact lease against f9c3e7681; main and all area branches remain untouched.
