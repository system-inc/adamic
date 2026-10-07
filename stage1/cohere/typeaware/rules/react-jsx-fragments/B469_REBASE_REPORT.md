Built: rebased wave-15 onto origin/area/stage1-lint b46914832, including unchanged main c7991b900; no new claims.
SHAs: previous remote 3eff48caa; tested rebased code 63a70b2ef; this evidence commit follows in git log.
Checks: five owned adapter scripts PASS with native/Go bytes, sanitizers and emitted-JavaScript quoting; registry PASS 0.056s; uncached Node one-byte PASS 0.254s; setup 37s, nproc 5.
Mutants: nine adapter finding mutants caught by independent Go/Node bytes and four adapter refusal mutants caught; registry rejection and Node one-byte checks passed.
Uncovered: full repository gate, mandatory external-input comparisons, complete integrated JSX ports and memo/capture analysis; unchanged legacy corpus suites were not rerun this turn.

The lint area moved with extraction of syntax rules into owned registry directories, shared context helpers and associated lint tests. Main did not move. The diff contains no changes under internal, bridge, stage1/typescript or stage1/cohere/typeaware. The compiler, runtime, bridge, parser, owned legacy rule code and frozen corpus inputs are unchanged. Consequently the previous nine-suite 943.892s PASS and bridge/released-handle results in B84_REBASE_REPORT.md remain their latest verification. This turn reran the five current JSX/Unicode/quoting adapters against their independent Go oracles and the changed registry. No shared code, other worker rule, harness or generator was edited.

Commands with /workspace/adamic-tools/env.sh sourced, all test output logged:

    bash cloud/setup.sh
    nproc
    python3 stage1/cohere/typeaware/rules/react-jsx-fragments/testdata/check.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_indirect.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_unicode.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_quoted.py
    python3 stage1/cohere/typeaware/rules/react-jsx-no-constructed-context-values/testdata/check_quoted_messages.py
    go run ./cmd/lint-registry
    go test -v -count=1 -timeout=10m ./stage1/cohere/lint/registry -run 'TestDeterministicRegeneration|TestDescriptorRejections|TestAdamicRuleModule'
    ADAMIC_GATE_UNCACHED=1 go test -v -count=1 -timeout=10m ./internal/oracle -run '^TestTheOracleCatchesOneByte$'

All commands passed. Adapter scripts used ADAMIC_WAVE15_SIXTH_COMPILER=/tmp/wave15-b84-adamic, the unchanged compiler built on the prior landing base. Their sixth/indirect/unicode/quoted artifact variables use /tmp/wave15-b469-{sixth,indirect,unicode,quoted}. The message script uses /tmp/wave15-quoted-messages. Exact logs and inventory are committed under validation/b469-rebase. No skips or failures were observed in these scoped runs. Registry generation is ignored output, not a shared-file edit.

Three-rule controls match 17 findings and the fragment/undef option variants. Indirect constructions match 22 findings/7,770 bytes. All agree under ASan/UBSan/LSan. Uppercase classification matches Go/Node over all 1,114,112 code points; quoting matches 1,112,064 scalar values plus seven strings/13,409,957 bytes under native, sanitized native and emitted JavaScript on Node. Four actual Unicode identifier findings match 1,520 bytes. Production SatisfiesExpression panic exit 2 and native refusal exit 70 remain outside successful agreement.

All nine successfully executing finding mutants are caught by independent bytes: fragment property, intrinsic predicate, construction label, branch order, provider factory, class owner, function kind, Lu-to-Ll and quoted escape prefix. Four refusal mutants: missing fragment facts, missing undef facts, omitted constructed-value refusal and unpaired-surrogate guard. Each compiles and exits zero; expected bytes or expected panic 70 catches it. These snapshots own no live checker handle, so no released-handle result is inferred from them.

Scalar quoting native 0.858s, sanitized 6.332s. Those are not comparable to Go source/checker whole-process work. Last comparable original-three compiler run remains native 2.134604s versus Go 0.348194s; no speedup claimed. Setup: Go/clang/Node readiness 0s, submodules 1s, cache warming 36s, total 37s; nproc 5, quota four CPUs, memory 17.6GB.

Fresh all-origin inventory: 644 refs, 135 distinct claim blobs under typeaware/claims and lint/claims, all 197 ranked names covered, zero unmentioned candidates. No new reservation was made. The inventory uses full-name boundaries and excludes the findings summary row. Live shared checker Program/declaration facts and memo stability/capture analysis remain actual blockers. No completed integrated ports or registry certification are claimed. No full gate or mandatory external-input check was skipped, relaxed or claimed green; they were not run. The own-branch rebase is pushed with an exact lease against 3eff48caa. Main and all area branches remain untouched.
