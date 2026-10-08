# use-isnan landing evidence

Base: origin/lint-checker/facts, d845dccde413c89643293e808626344d12e3f023.
Cohere oracle: 7945d102a6c18dd36adf9114a758ce646e8b2359, unchanged.

Only use-isnan is registered in this landing unit. Its 316 unique upstream cases and firing witnesses agree across Go, source Node, emitted JavaScript and sanitized native. Five upstream option variants were captured. The descriptor omits programReads, matching cohere. It uses ordinary node-level checker.ask and does not call askFile.

TestRulesAgree and TestOwnedWitnesses passed on the final landing graph. The source-shadowing mutant compiled, ran and disagreed with Go on all three runtimes; see mutants.log. Formatting and vet logs are empty and successful.

The full lint package was run once with all inputs enabled and -timeout 3h. It recorded 120 passing checks, one failure (TestRulesAgree, the parked no-throw-literal recovery row) and one unconditional skip (TestCheckerBridgeRefusalPending, awaiting C error envelopes). Its package elapsed time was 2014.417 seconds; shell wall time was 2016.493 seconds. nproc was 5, with four effective cgroup CPUs. Load before was 0.90/0.65/0.48; after was 1.62/2.20/2.41. setup.sh succeeded with total=41.517s.

A second filtered attempt discovered the callback parser gap and was stopped after failure. Both blocked candidates were removed from discovery and retained outside the landing branch. BLOCKERS.md names all 18 stopped claims and their exact calls or reproducers. no-new-wrappers and no-new-native-nonconstructor are already ported in the facts base.

Typed corpus: the original 554 TypeScript files, with defaults and enforceForIndexOf=true, produced 1108 rows and 29171974 identical output bytes. Release native took 5.599s, Go 1.742s, sanitized native 24.604s. These include program creation, parsing, findings, fixes, suggestions and output. There were no refusals or skips. The corpus hash is in corpus.json. The full package's broader 887-file plain corpus comparison passed; a separate attempt to open .a files in Go's checker project was refused because the Go builder omitted those extensions. That typed .a coverage remains unproven; the guard was not changed.

Commands (run from /workspace/adamic after sourcing /workspace/adamic-tools/env.sh):

- go run ./cmd/lint-registry
- gofmt -l stage1/cohere/lint/rules/use-isnan/oracle.go
- go vet ./stage1/cohere/lint/...
- ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave16-corpus/typescript ADAMIC_LINT_BENCH=1 ADAMIC_LINT_PROFILE_DIR=/workspace/wave16-artifacts/facts-profile ADAMIC_LINT_PROFILE_SNAPSHOTS=/workspace/wave16-artifacts/facts-profile go test -json -count=1 -timeout 3h ./stage1/cohere/lint
- go test -json -count=1 -run '^(TestRulesAgree|TestOwnedWitnesses)$' -timeout 3h ./stage1/cohere/lint
- go test -json -count=1 -run '^TestMutants$/^source_shadows_ignored$' -timeout 3h ./stage1/cohere/lint
- go run ./cmd/adamic build stage1/cohere/lint/main.ts -o /workspace/wave16-artifacts/facts-landing-release --tsgo /workspace/wave16-artifacts/facts-final-checker.a
- python /workspace/wave16-artifacts/landing-corpus.py

The raw test logs are preserved here as gzip files. results.json distinguishes the full attempted run from the final landing checks. No harness, parser, shared helper or bridge sources were edited.
