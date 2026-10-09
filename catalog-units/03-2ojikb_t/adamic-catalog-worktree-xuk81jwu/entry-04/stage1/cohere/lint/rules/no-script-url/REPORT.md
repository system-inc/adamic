# no-script-url unified port

Ported batch2:no_script_url.ts from the dedup ledger into a node-based descriptor on origin/area/stage1-lint. Only the two registered literal kinds enter the rule. The handed node is used directly; the parent is consulted only for the direct tagged-template exemption. Messages are copied verbatim. No shared harness, registry generator, or RuleContext files were edited.

The upstreamTest prefix TestNoScriptUrl captures all seven upstream test functions (29 unique rows). Upstream NoScriptUrl has no options type or fields and ignores its options argument. The adapter nonetheless decodes the complete JSON value and supplies a non-nil value, preserving the shared guard against dropped options. The optional witness.options.json sidecar is included for the witness-options integration; the initial area base did not consume sidecars, so explicit options were also tested independently. The landing branch now includes the merged witness-options harness and its sidecar path is rechecked.

Validation commands (output in evidence):

```sh
source /workspace/adamic-tools/env.sh
go run ./cmd/lint-registry
gofmt -l stage1/cohere/lint/rules/no-script-url/oracle.go
go vet ./stage1/cohere/lint/...
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript TMPDIR=/workspace/wave19-f801-scratch go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestMutants|TestRulesAgree)$' -count=1 -timeout=30m -v
```

Independent option rows used no options, {}, {"unused":true}, and whitespace-padded null. Go, source Node, emitted JavaScript, and sanitized native output agreed on 9,620 bytes, four findings per row, and no fixes or suggestions. A separate dotted-I boundary probe agreed on 9,788 bytes; javascrİpt:x stays clean. Source and emitted JavaScript ran through oracle/node.mjs, and native was built with --sanitize. The witness also covers mixed case, a direct tagged template, a nested template, leading whitespace, long s, and a decoded escape.

The no-script-url-case-fold mutant changes toLowerCase to toUpperCase, suppressing valid findings while still compiling. The unified comparison caught it on source Node, emitted JavaScript, and native.

The complete initial gate passed: TestRulesAgree (166.56s, 2,046 cases and 13,085,952 identical bytes), all 41 TestMutants (1,254.35s), and TestOwnedWitnesses (47.76s, 96,554 identical bytes). After rebasing onto the merged witness-options harness, the full upstream and witness comparisons plus the affected no-script-url and no-bitwise mutants passed with:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave19-typescript TMPDIR=/workspace/wave19-f801-scratch go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree|TestMutants)$/^(no-script-url-case-fold|int32_hint_option_ignored)$' -count=1 -timeout=15m -v
```

Scope: this port does not need the type checker, dynamic RegExp, Tailwind inputs, or new parser support. The full repository gate and a rule-specific performance benchmark were not run. Wave19 checker-dependent algorithms remain parked on codex/typeaware-wave-19 with named reproducers; none enters this landing branch.

Final rebased gate: PASS, 178.222s total; TestOwnedWitnesses compared 97,837 identical bytes. Registry generation and vet exited zero; gofmt reported no changes. All native comparisons used the harness sanitizer build. Evidence includes the complete initial suite and the final affected checks.
