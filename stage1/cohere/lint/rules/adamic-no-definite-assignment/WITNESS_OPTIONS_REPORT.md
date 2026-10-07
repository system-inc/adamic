Built: rebased the owned rule branch onto witness-options 29c41e102; existing security and path-alias sidecars now work, and security options are scoped consistently to selected rows.
Commits: shared fix 29c41e102, rebased worker 0821606d7; this report accompanies the owned adapter/port fix and logged evidence commit.
Commands and outputs: TestOwnedWitnesses PASS 26.90s, 96,379 identical bytes; TestMutants PASS all 45; TestRulesAgree FAIL on a decorated-class parser refusal. Combined gate exit 1 after 1088.662s.
Mutants: all 45 discovered mutations compile and execute and are caught by byte comparison on Node, emitted JavaScript and sanitized native, including all five retained owned rule mutants.
Not covered: complete upstream agreement remains blocked by the parser; Tailwind full-rule providers remain pending. No new claims, broader correctness gate or fresh throughput.

The existing testdata/witness.options.json files in base-security-require-context-access and nexus-import-require-path-alias already use the documented names and upstream options shapes. Both Go adapters decode field 5 into their upstream options types. No duplicate sidecars were added, and no shared source file was authored. The nil-adapter guard remains unchanged.

The initial three-gate run passed TestOwnedWitnesses, but TestRulesAgree and every mutant failed before comparison because generated legacy all-rule rows carry unrelated options such as allowLoop. The strict security decoder rejected those as unknown fields. The owned security adapter now decodes only selected security rows, and the port uses the same unconfigured security defaults for all rows. Selected security rows still use the original strict upstream decoder and owned checked decoder; invalid selected options are not accepted. After this fix, all 45 mutant subtests and the witness gate passed. The full agreement run captured 2,173 unique upstream source/rule/options combinations but failed at its Node execution on the retained parser case; that count is not certified parity coverage.

Run from the repository root, with output redirected to a log:

    source /workspace/adamic-tools/env.sh
    ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave11-typescript go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants|TestRulesAgree)$' -count=1 -v -timeout=30m > /tmp/wave11-witness-options-fixed-gates.log 2>&1

The supplied TypeScript checkout identifies as version 6.0.3. No requested test was skipped, filtered out individually or relaxed. An initial gofmt call before sourcing the tools environment found no binary; rerunning with that environment succeeded. The failed initial gate is retained and earns no mutant credit.

Parser blocker, reproduced once with the original source:

    const A = @RequireSessionAccess() class { m(@InjectRequestContext(AccountRequestContextKey) a: string) {} };

The shared Node driver exits 70: parser slice unsupported primary AtToken at 10. Source and original options are retained in evidence/witness-options-29c41e102/decorated-class.ts.txt and decorated-class.options.json. Reproduce using:

    source /workspace/adamic-tools/env.sh
    python3 stage1/cohere/lint/rules/adamic-no-definite-assignment/evidence/witness-options-29c41e102/reproduce.py > /tmp/wave11-parser-reproducer.log 2>&1

This produces exit 70; it is not a passing parity check. Full raw gate logs, the initial failure, rebase and reproduction output are in the same evidence directory. No parser or Tailwind provider was changed. Work stops here as requested, with no further helper claim and no push to main or area.
