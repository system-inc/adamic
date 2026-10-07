Ported nexus/consistency-no-long-line-comment onto the unified registry harness, with shared comment support on RuleContext.
Branch starts at lint area bb2ece56 and retains current main 4e0bfda5; only the dedicated worker branch is pushed.
Registry, gofmt, vet, upstream suite, TestOwnedWitnesses, owned TestMutants and TestRulesAgree PASS.
Ignoring the configured threshold compiles, exits successfully and is caught on source Node, emitted JavaScript and sanitized native.
No known blocker hit by this rule; full repository gate, broader input-dependent comparisons and the other 40 mutants were not rerun in this rule unit.

# Step 1 landing

The helper landing branch was refreshed and pushed before this rule branch was created. It owns 66 completed helpers and no rule descriptors. Its PARKED.md names four affected inventory consumers of the unclaimed dynamic RegExp prerequisites, with the concrete reproducer; no blocked rule implementation is registered there. All 40 inherited registered mutants passed on the new bases, in four disjoint groups, alongside TestOwnedWitnesses and TestRulesAgree. The landing evidence and parked note are on codex/lint-helpers-04 under stage1/cohere/lint/helpers/slot04_unified_landing/. Its SHA is reported once in the final response. Protected branches were untouched.

# Port and provenance

DEDUP_LEDGER.md names batch3:rules/no_long_line_comment.ts as the sole source. Its batch tip is fa9781c2c0911c52312002ba97ffd4b560d178ae. The port preserves the batch's run collection and safe fix, adapting its context and report API. The description is copied verbatim from cohere/policy/messages/consistency-no-long-line-comment.json, replacing only lineCount. The descriptor has node:true, kinds:[SourceFile], no order, and a concrete factory/listener. It performs no kind-based relevance dispatch inside the rule.

The adapter returns nexus.ConsistencyNoLongLineComment unchanged and decodes field 5 with json.Unmarshal into nexus.ConsistencyNoLongLineCommentOptions. It returns a typed value even for absent options, so the unified guard remains active. The optional options transport comes from 29c41e102, retained through the witness-options merge in bb2ece56.

The exact upstream prefix TestConsistencyNoLongLineComment covers all six actual functions: Fires, StaysSilent, HoldsGeneratedFilesToTheRule, Fixes, FixPreservesIndentation and WithholdsUnsafeFix. The combined capture grows by the rule's 16 cases to 2033. No upstream test is excluded. coverage.json records names and witness hashes. Four firing witnesses cover default threshold, an options-only report, unsafe text without a fix, and indentation/Unicode/blank comment lines. Their selected and all-rule outputs are compared independently against Go.

The shared additions are RuleContext.comments(node,index), its immutable-source cache and parser-anchor helpers, and Go-compatible RuleContext.trim. comment_ranges.a transports the batch's literal-aware comment model/scanner. Shared helper behavior is held by this rule's corpus and existing integrated comparisons; arbitrary parser recovery and non-UTF-8 source are not new claims. No existing rule, dispatch, shared oracle, registration generator, parser or compiler implementation was edited.

# Exact checks

Source /workspace/adamic-tools/env.sh before builds. bash cloud/setup.sh: Go ready 0s, clang 0s, Node 0s, submodules 0s, warm cache 100s, done 100s; nproc 5. All test output goes directly to logs.

- go run ./cmd/lint-registry: PASS 41 descriptors, evidence/registry.log.
- gofmt -w oracle.go followed by gofmt -l oracle.go: PASS empty evidence/gofmt.log.
- go vet ./stage1/cohere/lint/...: PASS empty evidence/vet.log.
- In cohere: go test -count=1 -v ./internal/lint/rules/nexus -run '^TestConsistencyNoLongLineComment': PASS 0.019s, evidence/upstream.log; all six top-level functions run.
- go test -count=1 -v -timeout=45m ./stage1/cohere/lint -run '^TestOwnedWitnesses$': PASS 75.428s; 94661 identical bytes across Go, source Node, emitted JavaScript and sanitized native, evidence/witnesses.log.
- go test -count=1 -v -timeout=45m ./stage1/cohere/lint -run '^TestMutants$/^nexus_long_line_comment_configured_threshold_ignored$': PASS 62.426s, evidence/mutants.log. This is the owned mutant; other descriptors' mutants were certified in Step 1.
- go test -count=1 -v -timeout=45m ./stage1/cohere/lint -run '^TestRulesAgree$': PASS 95.063s; all 2033 captured combinations and inherited generated cases, 13063694 identical bytes, evidence/rules.log. Existing malformed-method recovery rows are explicitly refused under the unchanged harness and are not successful recovery parity. No new-rule input hit those refusals.

# Mutant and initial failure

The official mutant changes the configured/default threshold selection to the constant four. Default witnesses remain unchanged. At case 147, the three-line witness configured with maximumLineCount:2 must report and fix; the mutant leaves it unchanged. Ordinary comparison catches that precise missing finding and fix on Node, emitted JavaScript and ASan/UBSan native. Each process exits successfully with empty stderr; no compilation or sanitizer failure is mutant credit.

The first native attempt inferred a conditional empty edit list as never[]. It was replaced by an explicitly typed SuggestionEdit array inside the rule, without changing shared lowering. The three initial logs record that failed attempt and are not passes or caught mutants. Final baseline comparisons and the compiling mutant pass.

This unit does not run the full repository gate or the 17 broader TypeScript/postcss/graphql/parser comparisons, change their required inputs, or relax/skip/delete any check. It has no dynamic regex, checker, parser or Tailwind blocker. No pull request is opened. The final response names each pushed branch SHA once.
