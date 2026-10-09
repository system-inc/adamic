Closed u057 M4 with existing guard TestDefaultTaggedSourceViews/default-literal; no new test was needed.
Delivery: evidence-only commit on compiler/downcast-write-guard, based on 62286991694debbb14fdb981d8f0957636c90dbc.
Clean focused command passed (0.376 s); M4 focused command failed (0.315 s), stdout other\n, exit 0.
M4 was caught by the existing pinned exit-70 literal-field expectation, then restored.
Not covered: full oracle/repository replay, new alias-write coverage, three-row floors; the brief's existing-guard stop condition applies.

The requested audit was fetched at d2a2c67b3d84c6f58d43f42dc5847f53def2d292, base 68db8ddd145281a62655452496bdc32ef848bdf3. Its M4.diff, observation source (.go.txt), clean and mutant observation logs are preserved here verbatim. M1.diff and its observation/bridge are also preserved as non-compilable evidence. No cohere source was copied.

Main moved from the initial 09fe4b54913753188a9357982bfd47cdf36ef97c to 62286991694debbb14fdb981d8f0957636c90dbc. The branch fast-forwarded before committing. M4 applied unchanged at both tips with git apply; no diff rebase or hunk adaptation was necessary.

Observation: M4 replaces e.values(property.ViewAllowed) with an empty allowed-value list in internal/javascript/view_fields.go. The audit source writes raw.kind after creating view=source as Child, then reads view.kind. It is an alias write followed by a checked read, not a write through the readonly view. The saved observation logs show JavaScript's clean exit 70 and M4's other\n/exit 0; they contain no native run.

Existing guard: internal/oracle/view_fields_test.go, TestDefaultTaggedSourceViews/default-literal. Its source stage3/interface-downcasts/default-literal.a has a field typed as literal wanted but holding other; reading through a downcast must stop. It compares native sanitized, native release and JavaScript against exit 70, empty stdout, and:

    adamic: panic: field read failed: (node as Identifier).name expected "wanted", found string other

Both native runs remain correct under this JavaScript-only mutant. JavaScript prints other\n with empty stderr and exit 0; the existing test fails at view_fields_test.go:196, a semantic kill rather than a build or sanitizer failure. This satisfies the explicit instruction to name an existing failing guard and stop. No production or test code changes, fixture, counts row or three-row floors were added. Evidence is wholly under review/ and remains in the test-only lane.

M1: benign for the observed admitted-source scope; it flips checkLazyViewReads on synthetic dictionary IR, while the corresponding source is refused earlier, so the audit establishes no admitted-source miscompile (not a general proof of harmlessness).

Commands (all test output saved to logs):

- export GOPROXY='https://proxy.golang.org|direct'; timeout 240 bash cloud/setup.sh > setup.log 2>&1. Outer limit reached during cold Go warming; the wrapper later printed nproc=5. No setup completion is claimed for this invocation.
- Retried the same setup command into setup-retry.log: exit 0. Timing lines: node 0.019 s, Go 0.021 s, submodules 0.056 s, markdown 0.070 s, clang 0.153 s, go build 28.570 s, test binaries deferred 28.672 s, build cache warm 28.673 s, done 28.697 s. nproc=5, CPU quota 400000/100000 (4 CPUs). Environment sourced from /workspace/adamic-tools/env.sh.
- First existing-guard attempt used missing /opt/adamic-tools/env.sh and the system's unrelated go; no tests ran. Corrected to setup's printed environment path. First corrected timeout 120 invocation reached its outer limit during cold test compilation, without test output; production diff restored.
- Under M4: timeout 180 go test -v -count=1 -timeout 90s ./internal/oracle -run '^TestDefaultTaggedSourceViews$' > existing-M4.log 2>&1: exit 1, default-literal failed at 19.46 s, package 20.916 s. Restored M4.
- Clean: timeout 120 go test -v -count=1 -timeout 90s ./internal/oracle -run '^TestDefaultTaggedSourceViews$' > existing-clean.log 2>&1: exit 0, literal leaf 0.52 s, package 0.895 s.
- On final main, clean: ADAMIC_GATE_UNCACHED=1 timeout 120 go test -v -count=1 -timeout 90s ./internal/oracle -run '^TestDefaultTaggedSourceViews$/^default-literal$' > final-clean.log 2>&1: exit 0, leaf 0.37 s, package 0.376 s.
- On final main, applied M4.diff and repeated that command into final-M4.log: exit 1, leaf 0.31 s, package 0.315 s. Restored with git apply -R. git diff --check clean; no production diff remains.

Lane checks are recorded in lane-checks.log after the evidence commit. The checkout's remote fetch mapping names only main, so explicit remote-tracking refs for devtools/fast-gate and cloud/merge-tree were fetched before running the prescribed command.

Roadmap contribution: closes task #4bnpyxj's u057 M4 audit gap by proving the existing outside-slice guard catches it. No broader coverage-map or uniqueness claim is made.
