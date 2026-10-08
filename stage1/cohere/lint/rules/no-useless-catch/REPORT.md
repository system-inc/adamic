Built: no-useless-catch migrated from batch2 onto the unified registry, as rule.a/messages.a with node:true and an owned Go oracle adapter; all fourteen wave-30 rules parked, so step 1 landing skipped.
Commits: validated rule 46d04d359b3f25a2e0393ac3c45586d828e13886 on area d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898; evidence follows on codex/lint-port-no-useless-catch. Parking note pushed separately on codex/typeaware-wave-30.
Commands and outputs: registry/gofmt/vet PASS; TestRulesAgree 53.17s, TestMutants 733.21s and TestOwnedWitnesses 20.52s PASS; required 397-file compiler parity PASS 82.102s; configured rows match 6536 bytes.
Mutants: compiled finally-range reversal caught by production Go on source Node, emitted JavaScript and sanitized native; compiled nil-options adapter caught by the unchanged shared guard; all 41 registered mutants pass.
Not covered: a full repository gate, unrelated packages' required external inputs, or landing of parked checker/HIR/SSA/capture rules. Native 2.197473ms versus Go 8.130512ms is a five-sample whole-process median on four small configured rows only.

Step 1 disposition: every retained wave-30 rule needs missing infrastructure to
enter the unified harness. Eight standalone ports require the checker program
and parser/checker-node correlation on RuleContext; three JSX ports require
checker symbol facts; three React analyses require native HIR/SSA/capture.
Each is named with its reproducer in the pushed parked/wave-30.md note on
codex/typeaware-wave-30, whose parking commit is 5cc89b852. That branch is a
parked archive, not a landing candidate. Under the explicit all-blocked exception
no step-1 landing branch or unified certification for those ports is claimed.
This new branch starts directly from the lint area and contains none of that
blocked wave-30 source or evidence. It only adds this rule directory.

Source: DEDUP_LEDGER.md names batch2:no_useless_catch.ts. The pinned branch/blob
is recorded in evidence/inputs.json. The rule's two descriptions are moved
verbatim, independently checked against Go's concatenated message literals:
322 bytes for unnecessaryCatch and 240 for unnecessaryCatchClause.
No shared helper, context, harness, generator, dispatch or oracle list was edited.
The only node subscription is CatchClause. The listener takes the supplied
ParseNode; it does not refetch that node or compare its kind for relevance.
Its structural checks inspect catch binding/body, erased wrappers and the parent
try. With finally it reports only the catch; without finally it reports the
whole try. Upstream offers no fix or suggestion, and their empty fields are
compared rather than omitted.

The real cohere test names are:

- TestNoUselessCatchReportsBareRethrows
- TestNoUselessCatchAcceptsClausesThatDoWork
- TestNoUselessCatchReportsTheNodeMatchingTheRepair

The upstreamTest prefix is TestNoUselessCatch, covering all three names. Shared
TestRulesAgree captures their cases with the real upstream rule; no copied Go
judgment decides the native result. Owned witnesses cover plain/finally rethrows,
comments, parenthesis/non-null/as wrappers, unreachable trailing statements,
work before rethrow, destructuring and optional bindings, and non-ASCII/astral
source text with Unicode catch names. TestOwnedWitnesses verifies a finding
for each source and compares selected-rule and all-rule outputs. All owned
witnesses across the 41-rule registry match 95030 bytes in that gate. The broader
TestRulesAgree comparison matches 13058977 bytes.

Upstream NoUselessCatch has no options type and ignores its options argument.
The owned adapter nevertheless decodes manifest field 5 with json.Unmarshal,
preserving supplied JSON values and returning a non-nil empty value for null.
Absent/blank input remains the unconfigured upstream default. No semantic option
or validation restriction is invented. A separate test uses four configured
rows: {}, an ignored scalar field, an ignored nested object/array and null.
Go, source Node, emitted JavaScript and sanitized native match all 6536 bytes.
This exercises decoding without depending on the not-yet-integrated witness
options sidecar transport. Replacing return options with return nil compiles;
the unchanged oracle refuses its {} row with
"no-useless-catch: options {} reached an adapter that decodes none". The guard
was never weakened or bypassed. That is an adapter-guard proof, distinct from
the source-rule mutant.

The source-rule mutant changes the finally test from children.length === 3 to
=== 2. It builds and executes on all three port paths. Only output comparison
rejects it: at owned witness case 145, line 2738, the mutant reports column 17
where Go reports column 1, also choosing the wrong description and range.
All 41 registered mutation checks pass, not only this rule's check.

Mandatory compiler/stage1 parity receives
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript, pinned TypeScript 6.0.3.
It executes over all 397 files on this clean area-based branch and matches
20516103 bytes across Go, source Node, emitted JavaScript and sanitized native.
The smaller count than the old wave-30 landing reflects that none of the parked
wave-30 modules are on this branch. No correctness check was skipped, relaxed
or removed. Existing malformed parser-recovery cases remain explicit refusals,
not successful lint parity. No known blocker was hit by no-useless-catch.

Commands after source /workspace/adamic-tools/env.sh, with all outputs logged:

- gofmt -w stage1/cohere/lint/rules/no-useless-catch/oracle.go, followed by
  gofmt -l of that file (empty output).
- go run ./cmd/lint-registry.
- go vet ./stage1/cohere/lint/....
- go test ./stage1/cohere/lint -run
  '^(TestOwnedWitnesses|TestRulesAgree|TestMutants)$' -count=1 -timeout 30m -v.
  Combined PASS 806.924s.
- ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-30-typescript go test
  ./stage1/cohere/lint -run '^TestCompilerAndStage1Agree$'
  -count=1 -timeout 30m -v. PASS 82.102s.
- python3 /workspace/no-useless-catch-options/check.py (preserved in evidence),
  which builds the discovered Go oracle via its standard overlay, native under
  ASan/UBSan/LSan and emitted JavaScript, compares configured rows, then proves
  the nil-options guard mutant. It writes exact stdout/stderr/exit artifacts.
- go run ./cmd/adamic build stage1/cohere/lint/main.ts -o
  /workspace/no-useless-catch-options/native-fast; five whole-process samples
  each of that unsanitized native driver and the unchanged Go oracle, checking
  identical output, clean exit and empty stderr on every sample.
- git diff --check; git diff --cached --check.

The toolchain was already prepared by the preceding cloud/setup.sh run
(110 seconds, nproc 5, CPU quota 4 cores); no setup change was made for this
area-based branch. Timing samples are preserved, not a large-corpus speed claim.
Logs are gzip-compressed and verified against their uncompressed hashes in
streams.json. Generated registry files and binaries remain untracked. No PR,
main push or area push is part of this unit.
