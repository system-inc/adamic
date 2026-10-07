Brought seven canonical wave1-06 rule directories onto the unified handed-node harness; all seven remain in the landing branch.
Base origin/area/stage1-lint d3a37422; witness-options c650c0ab2 (cherry-pick of 29c41e102); ports 830eeeb46; replacement mutant a94061b64; evidence commit follows.
TestRulesAgree PASS 57.49s; TestOwnedWitnesses PASS 25.25s; 46 bulk mutants PASS and the one replaced mutant PASS on focused retry 27.394s; vet PASS; setup 40s, nproc 5.
Seven compiling semantic mutants are caught only by Go byte comparison on source Node, emitted JavaScript and sanitized native; the ineffective original restriction mutant is retained as a failed observation.
No kept rule stays out for a shared blocker; losing copies and baseline duplicates were not imported; helper branch and full repository/compiler corpus gates are outside this takeover.

Created codex/lint-wave1-06-land from origin/area/stage1-lint d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898, which includes current main b6b1538b0cebc4ba6741ac34f1aedb60293c1d06. Source directories came from origin/codex/lint-wave1-06 a5f44f6ba579a44544e828cad760f78aa8a70abb. Only the canonical directories were imported, avoiding its old shared harness commits. Both TypeScript ledger rows name wave1-06 canonical. Integration merges this landing branch; nothing was pushed to main, area, the abandoned worker branch or its helper branch.

Retained public rules:

- @next/next/no-before-interactive-script-outside-document
- @next/next/no-css-tags
- no-inner-declarations
- no-restricted-properties
- no-self-assign
- @typescript-eslint/no-unsafe-function-type
- @typescript-eslint/no-useless-empty-export

Not imported: google-font-display (wave1-01), no-document-import-in-page (wave1-12), require-description and tailwind-no-physical-direction (wave1-05), no-unnecessary-type-constraint (wave1-08). The area's eqeqeq, no-debugger, no-duplicate-case, no-empty and no-var copies were left intact; wave1-06's duplicates were not copied. Helper branch 6f2aff526107895a267896fdcbf182cd67c3c7a6 was inspected for context, not migrated in this rule-only landing.

Every descriptor declares its named syntax kinds and node: true. Every visit accepts the handed ParseNode; direct relevance guards for TypeReference and SourceFile were removed, and the restricted-properties and self-assignment visits no longer refetch that handed node through their helpers. All new Adamic source files are .a. Mixed-case duplicate mutant metadata was normalized and obsolete explicit rule.ts mutant paths now point to rule.a.

The three configurable core rules already decode manifest field 5 into their actual upstream types, preserving defaults where upstream supplies them: DefaultNoInnerDeclarationsSettings, NoRestrictedPropertiesOptions, DefaultNoSelfAssignOptions. The two Next.js and two TypeScript rules have no upstream options type and their Run ignores options; their nil adapters receive no options-bearing rows in the verified corpus. The shared nil-adapter guard remains unchanged, so unexpected options still panic rather than disappear.

Cherry-picked the explicitly supplied witness-options commit 29c41e102d38d8f51de995abc4e2c9464196f46c as c650c0ab2. Its upstream-authored shared changes let no-restricted-properties' witness.options.json reach Go and the port. No new shared harness edit was authored here. docs/parallel-work.md is absent. Cohere pin is 715ba94f3608a6500086b1076ce5cb7e51b836db.

Initial observations and fixes: historical Next.js reports said JSX was blocked, but isolated probes showed the current parser can parse the raw witnesses. Importing those copies exposed their obsolete guard, which panicked on any source containing < and no JSX nodes, including ordinary generic type sources, and their .ts.txt witnesses gave Go the wrong script kind and no finding. The initial logs contain NotYet: stage-1 JSX parser adapter and the no-finding witness failure. Removed that rule-owned guard rather than modifying the parser, and renamed both raw witnesses to .tsx.txt. On the unified final harness both witnesses fire and all captured upstream cases match, so neither rule is excluded. No projected AST comparison is counted as independent-parser parity.

Commands sourced /workspace/adamic-tools/env.sh. bash cloud/setup.sh passed: Go, clang, Node, submodules 0s; cache warm 40s; total 40s on five processors, cgroup cpu.max 400000 100000, 17.6 GB; nproc 5. go run ./cmd/lint-registry passed. The final baseline command was go test ./stage1/cohere/lint -run '^(TestOwnedWitnesses|TestRulesAgree)$' -count=1 -v -timeout 15m: PASS 82.757s. TestRulesAgree compared 13,186,480 identical bytes across Go, Node, emitted JavaScript and sanitized native in 57.49s; TestOwnedWitnesses compared 97,737 identical bytes in 25.25s. go vet ./stage1/cohere/lint/... passed. All output went directly to log files.

Full mutant command: go test ./stage1/cohere/lint -run '^TestMutants$' -count=1 -v -timeout 30m. It completed in 1069.804s with 46 passing cases and one failing case: the imported invent-unconfigured-restriction mutant survived on Node. It added a global bar restriction that the configured foo.bar restriction overrode, so the current configured witness did not distinguish it. This was an ineffective mutant, not accepted proof. Replaced it with configured-restriction-ignored, changing the Restrictions lookup to IgnoredRestrictions. Focused rerun go test ./stage1/cohere/lint -run '^TestMutants$/^configured-restriction-ignored$' -count=1 -v -timeout 10m passed in 27.394s. Its source, emitted and native processes all compile/run normally and differ from Go; the exact log is retained. Only this changed mutant was rerun; the entire 47-case command was not repeated. Thus all final mutant definitions have passing evidence, but the raw bulk run correctly remains marked FAIL.

The takeover's seven mutant proofs:

- no-before-interactive-script-outside-document: wrong diagnostic id (wrongStrategy), caught on all three runtimes.
- no-css-tags: wrong diagnostic id (wrongCssTags), caught on all three runtimes.
- no-inner-declarations: wrong-enclosing-scope, caught on all three runtimes.
- no-restricted-properties: configured-restriction-ignored, caught on all three runtimes after replacing the ineffective inherited mutant.
- no-self-assign: report-left-reference, caught on all three runtimes.
- no-unsafe-function-type: wrong-global-type-name, caught on all three runtimes.
- no-useless-empty-export: omit-empty-export-fix, caught on all three runtimes.

The inherited method-signature-style parser-recovery exclusions remain explicitly reported in baseline.log; excluded recovered findings are not parity coverage. The full repository gate, required compiler/stage1 corpus and other packages' required-input checks were not run for this takeover. No checks were skipped, relaxed or deleted to make a result green. No performance claim is made. Existing historical reports inside imported directories are superseded by this landing evidence where they discuss adapter and harness limits.
