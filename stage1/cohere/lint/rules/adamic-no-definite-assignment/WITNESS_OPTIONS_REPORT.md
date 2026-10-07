Built: rebased the owned rule branch onto witness-options 29c41e102; security sidecars now work, the path-alias sidecar has overlapping directories, and security options are scoped consistently to selected rows.
Commits: shared fix 29c41e102, rebased worker 0821606d7; this report accompanies the owned adapter/port fix and logged evidence commit.
Commands and outputs: final TestOwnedWitnesses PASS 25.11s, 96,507 identical bytes; full TestMutants initially 44 PASS / 1 survived, corrected alias mutant PASS on focused rerun; TestRulesAgree FAIL on decorated-class parser refusal. Full combined gate exit 1 after 1088.662s; focused recheck exit 0 after 51.761s.
Mutants: 44 discovered mutations passed the full run; path_alias_longest_directory_reversed initially survived its single-alias witness, then was caught by byte comparison on all three backends after strengthening that witness. All five retained owned rule mutations are now caught; the complete 45-mutation command was not repeated after the witness-only correction.
Not covered: complete upstream agreement remains blocked by the parser; Tailwind full-rule providers remain pending. No new claims, broader correctness gate or fresh throughput.

The existing testdata/witness.options.json files in base-security-require-context-access and nexus-import-require-path-alias already use the documented names and upstream options shapes. Both Go adapters decode field 5 into their upstream options types. The alias sidecar now includes both . / @project and adamic-gate / @gate beneath repositoryRoot /tmp, using the supplied TMPDIR=/tmp/adamic-gate. This makes ../../foundation/Thing match two directories and proves longest-match selection. No duplicate sidecars were added, and no shared source file was authored. The nil-adapter guard remains unchanged.

The initial three-gate run passed TestOwnedWitnesses, but TestRulesAgree and every mutant failed before comparison because generated legacy all-rule rows carry unrelated options such as allowLoop. The strict security decoder rejected those as unknown fields. The owned security adapter now decodes only selected security rows, and the port uses the same unconfigured security defaults for all rows. Selected security rows still use the original strict upstream decoder and owned checked decoder; invalid selected options are not accepted. After this fix, 44 mutant subtests passed and the alias mutation survived. Its original sidecar had only one alias, making the ordering mutation unobservable. Added a second matching directory; the focused mutation and all-witness rerun passed. The complete run and its surviving-mutant failure remain recorded, with no false green credit. The full agreement run captured 2,173 unique upstream source/rule/options combinations but failed at its Node execution on the retained parser case; that count is not certified parity coverage.

Run from the repository root, with output redirected to a log:

    source /workspace/adamic-tools/env.sh
    ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave11-typescript go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants|TestRulesAgree)$' -count=1 -v -timeout=30m > /tmp/wave11-witness-options-fixed-gates.log 2>&1

The focused recheck command was: ADAMIC_TYPESCRIPT_SOURCE=/tmp/wave11-typescript go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants)$/^path_alias_longest_directory_reversed$' -count=1 -v -timeout=10m, with output redirected to /tmp/wave11-witness-options-alias-final.log. TestMutants PASS 26.61s, including its alias subtest PASS 23.68s.

The supplied TypeScript checkout identifies as version 6.0.3. All three requested tests ran unfiltered initially and after the security fix. Only the affected mutation was filtered for the final witness-only correction; all witnesses ran again. No comparison was relaxed. An initial optional subtest regex accidentally selected extra mutations; that redundant attempt was terminated and retained without pass credit. An initial gofmt call before sourcing the tools environment found no binary; rerunning with that environment succeeded. The failed initial gate is retained and earns no mutant credit. The first evidence revision incorrectly said all 45 passed; this revision corrects it to the observed full-run failure and focused recheck.

Parser blocker, reproduced once with the original source:

    const A = @RequireSessionAccess() class { m(@InjectRequestContext(AccountRequestContextKey) a: string) {} };

The shared Node driver exits 70: parser slice unsupported primary AtToken at 10. Source and original options are retained in evidence/witness-options-29c41e102/decorated-class.ts.txt and decorated-class.options.json. Reproduce using:

    source /workspace/adamic-tools/env.sh
    python3 stage1/cohere/lint/rules/adamic-no-definite-assignment/evidence/witness-options-29c41e102/reproduce.py > /tmp/wave11-parser-reproducer.log 2>&1

This produces exit 70; it is not a passing parity check. Full raw gate logs, the initial failure, rebase and reproduction output are in the same evidence directory. No parser or Tailwind provider was changed. Work stops here as requested, with no further helper claim and no push to main or area.
