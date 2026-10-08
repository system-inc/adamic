Ported no-unsafe-finally from batch 2 into a standalone .a descriptor directory on the unified harness.
Base area bb2ece564 (witness-options 29c41e102 included); the rule source commit is followed by this evidence commit.
Registry, gofmt and full vet pass; TestRulesAgree 49.12s, TestMutants 775.56s, TestOwnedWitnesses 21.63s; total 846.333s.
All 41 compiling mutants caught, including finally continue wrongly absorbed by switch on Node, emitted JavaScript and sanitized native.
Step 1 skipped: existing wave-29 implementations remain blocked from unified certification; no blocked implementation was carried into this branch.

The rule is based on batch2:no_unsafe_finally.ts, the source named by DEDUP_LEDGER.md.
Its four descriptions were moved verbatim into messages.a. The listener subscribes
to ReturnStatement, ThrowStatement, BreakStatement and ContinueStatement, consumes
the handed ParseNode, and uses existing RuleContext ancestry/function helpers.
No primary-node relevance filter, extra dispatch or shared helper was introduced.
Jump stopping behavior follows the actual upstream Go rule, including its labeled
continue loop stopping point and the break/continue switch asymmetry. No fixes
or suggestions are produced by the upstream rule or this port.

Upstream prefix TestNoUnsafeFinally captures all five real functions:
TestNoUnsafeFinallyReportsEscapingControlFlow,
TestNoUnsafeFinallyReportsThroughNestedStatements,
TestNoUnsafeFinallyReportsLabeledJumpsLeavingFinally,
TestNoUnsafeFinallyReportsContinueThroughSwitch, and
TestNoUnsafeFinallyStaysSilentWhenTheJumpIsAbsorbed.
The complete shared comparison includes their positive and negative cases.

The upstream rule declares no options type and ignores its options argument.
The adapter decodes supplied field-5 JSON into map[string]json.RawMessage and
returns that non-nil value; malformed JSON panics through json.Unmarshal.
The witness's escaping.options.json supplies {} on the updated unified harness.
A separate invocation of the actual unified Go oracle with field 5 {} also
reports unsafeReturn, exit zero and empty stderr (preserved options logs).
No options guard was bypassed. The witness covers return, throw, continue
through switch, and a labeled break leaving finally through an inner loop.

Commands, with stdout/stderr redirected to preserved compressed logs:

- bash cloud/setup.sh: cache warm 127s, total 127s, nproc 5, quota 4, 17.6 GB.
- source /workspace/adamic-tools/env.sh
- go run ./cmd/lint-registry
- gofmt -w stage1/cohere/lint/rules/no-unsafe-finally/oracle.go
- gofmt -l stage1/cohere/lint/rules/no-unsafe-finally/oracle.go: empty output.
- go vet ./stage1/cohere/lint/... and go vet ./...: empty output, exit zero; full vet repeated after rebase.
- go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestMutants)$' -count=1 -timeout=30m -v

The first comparison caught an implementation mistake: this parser consumes the
finally keyword without retaining a keyword child. The corrected rule identifies
the final child Block of a TryStatement, distinct from its first try Block and
its CatchClause. The initial mutation run was interrupted after this comparison
failure; no green result is claimed for that attempt. The subsequent full run
passed on d3a37422c in 924.317s. Integration advanced to bb2ece564 while it ran;
the port rebased cleanly and the entire requested gate reran, including the new
options witness, in 846.333s. Final agreement is 13066438 identical bytes across
Go, source Node, emitted JavaScript and ASan/UBSan native; owned witnesses agree
on 96465 bytes. The compiling finally mutant is caught only by diagnostic output:
it drops unsafeContinue at line 2 of the owned witness while preserving later
findings. All 41 registered mutants pass the independent comparison checks.
Inherited malformed method-signature recovery cases remain explicit tested
refusals in the unchanged harness; no check was skipped or relaxed by this port.

Disk reached zero free space during the second run. Verified obsolete ELF and
ar outputs in three owned wave-29 scratch directories were removed, preserving
sources and logs. Space remained stable and the complete final run passed.
The full repository gate, 17 external-input correctness checks and lint throughput
benchmarks were not run. No known checker, regex, parser or Tailwind blocker was
encountered by this rule after correcting the port's parser-layout assumption.

Step 1 parking record: existing implementations remain on codex/typeaware-wave-29.
Their concrete shared blocker is RuleContext's lack of checker program/project
and file identity, rather than missing diagnostic serialization. The original
six implementations use bridge-based source/binding facts and are not unified
factory registrations; the JSX and hook implementations are test-only kernels.
The following reproductions are on that branch and were already published there:

- id-denylist, nexus/concurrency-no-check-then-write, no-restricted-globals,
  no-setter-return and no-shadow-restricted-names: the wave_29_test.go suite and
  wave-29-next/check.py exercise native checker-backed streams; the same bridge
  entry points cannot be acquired from shared RuleContext. These remain source
  integration blockers, not claims that every rule intrinsically needs a checker.
- id-match: wave-29-configured/check_regex_contract.py compiles a runtime
  new RegExp(pattern, 'u') probe; native refuses the nonconstant pattern.
- react/jsx-fragments and react/jsx-no-undef: wave-29-fourth/prove_gaps.py and
  rules/react-jsx-no-undef/testdata/check_resolution.py show source JSX parsing
  and Go findings, with checker identity/symbol acquisition still test-only.
- react/jsx-no-constructed-context-values: wave-29-fourth/prove_gaps.py has the
  constructed JSX provider witness; native memo/escape/capture analysis is absent.
- react-hooks/set-state-in-effect, react-hooks/set-state-in-render and
  react-hooks/static-components: wave-29-third/prove_gaps.py has three positive
  upstream sources; no native source-to-HIR/SSA/capture analysis entry exists.

All that blocked work is absent from this dedicated landing branch. No separate
wave-29 landing branch is claimed green on the unified harness.
