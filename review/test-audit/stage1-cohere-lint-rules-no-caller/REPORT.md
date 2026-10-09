u120 stopped on a red clean baseline.
Starting origin/main: ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2.
Default discovery failed; no test rows could be enumerated.
No mutants, probes, timing medians or quality verdicts were produced.
Evidence branch: test-audit/stage1-cohere-lint-rules-no-caller.

```json
[]
```

Commands and evidence

`go test -list . ./stage1/cohere/lint/rules/no-caller/` failed with:
`package github.com/system-inc/adamic/stage1/cohere/lint/rules/no-caller: build constraints exclude all Go files in /workspace/adamic/stage1/cohere/lint/rules/no-caller`

`timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/rules/no-caller/ -run .` produced the same build failure and `[setup failed]`. See list.log and baseline.log. This is a build failure, not a 90-second timeout or a skipped test.

Mutant table: empty. Survivors: none. No source was changed.

Scope problem and uncovered work

Both oracle.go:1 and oracle_options_test.go:1 require `lintoracle`. The only source-discovered Test is TestNoCallerOptionsPayload at oracle_options_test.go:7; it was not enumerated by the required command. It tests oracleNoCallerOptions, a Go cohere oracle adapter that decodes explicit JSON options, preserves absent options and panics on malformed JSON. This adapter is an oracle, not Adamic's native rule port. Its handwritten assertions would be self expectations if audited independently, but no oracle quality verdict was established here.

The actual port source is rule.a, whose functions are Rule.constructor, Rule.visit and create. Native agreement tests live in the parent stage1/cohere/lint harness according to this directory's REPORT.md. They are outside the requested package scope. No claim from that historical report was treated as session-produced audit evidence. The brief names a rule-source directory as a default Go test package; that scope cannot execute. Adding a build tag would also require the cohere overlay for Go internal-package imports described by the existing report, and would audit the oracle adapter rather than port agreement. The brief explicitly says to stop on a red baseline, so neither changing scope nor modifying the adapter was justified.

Timing and preparation

Warm /workspace/adamic-tools/env.sh worked, so setup was skipped (0 seconds). nproc=5. npm ci in stage3/api succeeded: added 3 packages in 361ms. Discovery plus npm invocation took 0.287 seconds of tool-reported command wall time; npm's own 361ms report is retained verbatim, and these clocks are not treated as comparable measurements. The baseline command and inspection invocation took 0.017 seconds; the JSON baseline reports failed build and package elapsed 0 seconds, not a valid test timing. No native build, mutant build or three-run test median exists. Exact standalone setup and npm wall times were not instrumented. No other packages were run, no port or oracle was mutated, and no PR was opened.
