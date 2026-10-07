Rules: rebased seven retained ports onto current area; no rule or shared source edits.
Commits: old pushed tip 91bf139a536b327bf24245f3e6396dd8cad89add; rebase implementation tip b7563b2c24f098401ed0ee2aaa90120053d76e7a; evidence commit follows.
Checks: fresh unified supported upstream/corpus/const controls, six mutants, registry, vet, external one-byte and landed runtime oracle; raw logs in d65_evidence/.
Mutants: all six retained representable rule mutants comparison-killed on Node, emitted JavaScript and ASan/UBSan native; const annotation mutant still blocked.
Limits: shared two-edit reporting/oracle and malformed computed-key parser recovery remain; no new helper claim, full parity assertion or throughput measurement.

# Rebase and ownership

Fetched every origin branch. Current main remains 39638d9e278d38bb5aeae887f46d55a70e47aaad. Current area/stage1-lint is d65a8f931c98655936ae04c6899f38f14862b73e and contains that main. Its changes since the previous base include native heap/string runtime work and profiling artifacts. Rebased the owned rule branch cleanly onto it, accepting inherited runtime changes without editing them. No main or area branch is pushed.

The dedup ledger is unchanged, SHA256 4c39ec0cb129b05a3971ff257c26296d0d4ce06545b42e53dc14bbf0828a526f; its full content was read in the prior integration and identity checked on this fetch. The seven retained and eleven retired copies remain exactly as listed in AREA_LANDING.md. No new batch rule assignment or claim is inferred. No new implementation file or shared harness edit is made in this refresh.

The existing helper branch stays at a7ef0e6bb4dd72334ef028c989fa756f8f4638ad, already rebased onto unchanged current main and freshly green in its prior pushed commit: all four Go/Node/emitted-JavaScript/sanitized-native baselines and four comparison-only mutants PASS 26.323s, vet zero, one-byte oracle PASS 0.677s. No new helper validation is claimed here. Those four helpers serve six Tailwind consumers as documented in helpers/from_wave08/REPORT.md and BREAKPOINTS_REPORT.md; none alone removes all final blockers.

# Fresh observations

Commands use source /workspace/adamic-tools/env.sh, ADAMIC_GATE_UNCACHED=1 and the unmodified shared comparison. The owned test overlay and pinned compiler checkout are the same paths and commands documented in AREA_LANDING.md. Fresh log prefixes are /tmp/wave08-d65- instead of /tmp/wave08-unified-. The overlay still maps only the owned unified_test.go.txt to a virtual new test file and replaces no shared source.

Supported upstream gate PASS 130.27s: all 414 selected cases, including all JSX cases, produce identical Go/Node/emitted-JavaScript/sanitized-native output, 149400 bytes. Only the two explicitly named malformed computed-key cases remain excluded.

Six semantic mutants PASS 236.724s. Each compiles and executes normally, with only byte comparison detecting its semantic change: core_default_suppressed suppresses default-param-last; computed_replacement_wrong corrupts the fix; foreign_read_suppressed suppresses foreign propTypes; gating_invalid_suppressed suppresses invalid configuration; constraint_suggestion_wrong corrupts a suggestion; enum_second_suggestion_wrong corrupts the second enum suggestion. The const_append_wrong annotation mutant remains blocked by the baseline multi-edit guard; its earlier rich-transport proof is historical, not fresh unified-harness evidence.

Registry PASS 0.114s. Whole-repository go vet ./... exits zero with empty log. Filtered uncached TestTheOracleCatchesOneByte PASS 0.379s, no native/Node cache hits. Separately ran the newly inherited TestRuntimeLastIndexOfMatchesNode: PASS 47.691s, Node/emitted-JavaScript/release-native/sanitized-native match on 758 output bytes, no cache hits. This checks the runtime change inherited by the rebase without modifying it.

Full TestOwnedWitnesses FAIL 80.374s: actual Go exits 2 with panic: unexpected fix shape, still refusing the const annotation's two independent fixes. Direct source Node probes freshly confirm panic: a finding carries at most one automatic edit for const and parser slice expected CloseBracketToken, got CloseBraceToken at 8 for the malformed computed key; both exit 70. The shared multi-edit limitation and parser recovery gap keep full parity and the strict landing cap blocked. No new helper is claimed or written.

bash cloud/setup.sh PASS. Timing lines: Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 2s; build cache warm 124s; done in 124s on five processors. nproc prints 5; cgroup cpu.max is 400000 100000. Tool versions remain Go 1.27.1, clang 20.1.8 and Node 24.19.0. No full repository test suite, fresh throughput benchmark or integration push is claimed.

Final selected gate PASS 390.566s. The compiler/stage1 test PASS 222.46s on all 364 .ts/.a files and 2184 rule cases; all four outputs match at 80203185 bytes. The three const single-edit/negative controls PASS 37.62s, all four outputs match at 1011 bytes. Together with the 414 upstream cases, this certifies the supported scope against the actual Go oracle after the runtime rebase, without asserting the excluded cases or full const fixture parity.

The separate runtime command was ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestRuntimeLastIndexOfMatchesNode$' -count=1 -v -timeout=10m, and the filtered one-byte command used -run '^(TestTheOracleCatchesOneByte|TestRuntimeLastIndexOf)$'; only the first of that latter pattern exists and runs. All output was redirected directly to log files. The exact inherited-runtime test was run separately as documented above. No test run was piped.
