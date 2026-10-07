# Corrected next claim and dependency blocker

Corrected the overly narrow previous queue audit and reserved consistent-return, constructor-super and default-case.
Claim ae8e7213 was pushed on codex/lint-wave1-07 before any new rule code; the prior completion remains pushed at c0c4ff37.
Fresh all-head fetch: 346 origin refs, 16 distinct claim trees; setup PASS in 30s, nproc 5; unmodified upstream trio tests PASS in 0.019s.
No mutant or Adamic parity test is claimed for the new trio: no new implementation was written; the four previous executable mutants remain recorded.
Stopped at missing compatible stage1 consistent-return judgment/control-flow support; constructor-super and default-case are reserved but not started.

The previous audit incorrectly narrowed the inventory fallback to its ready queue. The existing October 7 reservation explicitly says syntax-only includes waiting-on-helper entries while excluding needs_type_information and binding_only. This continuation restores that definition, preserving inventory order. The original 46 helper-ready rules are already claimed. The next three names are absent from main's executable selections, all fresh origin claim Markdown, and a fresh origin executable-source scan. The claim was committed and pushed before implementation.

The first rule, consistent-return, calls consistentreturn.Judge at cohere/internal/lint/rules/core/consistent_return.go. The shared judgment calls canRunOffEnd, which builds the Go control_flow_graph and reads Graph.EndReachable. Counting return statements cannot reproduce it: a trailing throw, an else-throw, an infinite loop, and try/catch/finally all affect whether the function can reach its end. Go's pinned graph also has a documented try/finally over-report, which byte parity must preserve rather than improve away.

A scan of 102 distinct origin stage1 trees found two graph modules. Neither supplies the required independently built stage1 graph:

- typeaware-wave-25's syntax_control_flow.a reads graph blocks from SyntaxProjection.ask(root, 'syntax-control-flow').
- typeaware-wave-01's syntax_flow_graph.a reads frames from bindings.rules.ask(root, 'syntax-flow-graph').

Both depend on external type-aware bridge contexts. This unit's stage1 lint RuleContext holds the numeric Parser/Scanner tree and has no ask bridge or consistentreturn judgment. Their source is preserved in wave1-07-next-dependency-evidence/. No compatible shared judgment/build/end-reachability API was found. A partial listener pretending its syntactic judgment were the whole rule would violate the required Go parity; a stub would also break all-rule runs. No such implementation is committed.

This is a shared-helper dependency gap, not the shared harness's old .a loading or suggestion serialization gap. Under the earlier instruction, “If anything else blocks you, say exactly what it is and stop, rather than editing shared files,” this continuation stops and does not edit shared context, helpers, parser, harness, or compiler. The reserved constructor-super and default-case are not reported implemented or blocked on this same helper; they were not started after the first dependency stop. Further claims are not taken.

The current shared harness tip f4d98cab retains JSX witness extensions and still explicitly documents no JSX parsing. It therefore does not close the previous Google Font Display parser refusal. That owned selector/listener, its successful bounded comparison and mutant, and its demonstrated three-runtime exit-70 blocker remain pushed. Earlier assertions/alias and original three ports remain pushed with their existing evidence; no re-certification of those unchanged sources was claimed here.

Commands wrote output directly to logs:

- git fetch --no-recurse-submodules origin '+refs/heads/*:refs/remotes/origin/*'; git push origin codex/lint-wave1-07: existing completion up to date before selection.
- bash cloud/setup.sh: Go 0s, clang 0s, Node 0s, submodules 0s, build cache warm 30s, done 30s on 5 processors; nproc printed 5.
- In cohere, go test ./internal/lint/rules/core -run '^(TestConsistentReturn|TestConstructorSuper|TestDefaultCase)' -count=1 -v: 31 top-level tests passed, package PASS in 0.019s. These prove the upstream reference tests run, not that an Adamic port exists.

The full repository gate and new-trio Node/emitted-JavaScript/sanitized-native parity, mutants and findings/s measurements were not run and are not claimed. Raw setup, tests, selection and dependency evidence are retained beside this report. The previous statement that nothing remained unclaimed is superseded by this correction.


## Subsequent environment resume

Rechecked the claimed trio and compatible shared dependency; no new rule implementation was added.
Existing claim ae8e7213 and blocker report 88417969 were already pushed; this evidence is committed separately.
Fresh all-head fetch inspected 355 origin refs and 103 distinct stage1 trees; setup PASS in 30s, nproc 5.
No new mutant, parity test or throughput measurement is claimed because no executable rule source changed.
The same consistent-return shared-judgment/graph dependency remains unavailable; no further claims taken.

The new environment was checked with the cloud-environment-runtime skill. HTTP policy is enforced and the runtime is ready. The existing branch is clean and its push reported Everything up-to-date before work. Main remains ef3d907e and shared harness remains f4d98cab.

The refreshed scan now finds a third graph projection, origin/codex/typeaware-wave-07's wave07_control_flow.a. It calls bindings.rules.ask(root, 'wave07-control-flow') and reads externally supplied graph blocks/events, just as the two earlier candidates read their external frames. All three sources and their exact branch commits are preserved in resume-audit.json. None supplies compatible consistentreturn.Judge or an independently built control_flow_graph.Build/EndReachable result through this lint unit's RuleContext. The current RuleContext was reread and its source hash recorded; it still has no graph-query bridge.

Setup printed Go ready 0s, clang ready 0s, Node ready 0s, submodules ready 0s, build cache warm 30s, done 30s on 5 processors; nproc separately printed 5. Setup compiles the existing package/test graph without executing rule parity tests. No upstream-test rerun or new-trio certification is claimed. Previous parity, mutant and throughput evidence belongs to the earlier completed rules.

The earlier user instruction to stop on other dependency gaps still applies. Shared context, helpers, parser, harness and compiler were left untouched. consistent-return remains blocked; constructor-super and default-case remain reserved and unstarted after that dependency stop. The current source state also leaves the prior Google Font independent JSX-parser limitation unchanged. The unit is not declared complete and the syntax-only queue is not declared exhausted.
