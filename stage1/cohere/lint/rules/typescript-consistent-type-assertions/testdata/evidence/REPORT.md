Built: consistent-type-assertions, with exact messages, style fixes and structured suggestions; converted the other three owned rules to .a.
Commits: harness integration cea28392; exact committed dependency restoration 8cf9f2a0; implementation and evidence in the containing commit.
Checks: TestAssertions PASS (25.464s), TestAssertionCorpus PASS (243.535s), TestRuleContracts PASS (58.418s), registry PASS (0.058s).
Mutants: equal assertion precedence, removed rtl/ltr exemption, removed description reason exemption, permitted display=block; all compile/run and fail output comparison on all three backends.
Not covered: independent JSX parsing and a passing shared TestRulesAgree; details below.

The assertion fixtures capture 196 distinct upstream tests and guards. Go AST projection matches 44,254 output bytes on source Node, emitted JavaScript and sanitized native. All 193 non-JSX fixtures also match when Adamic independently parses the source. Three TSX fixtures deliberately refuse with NotYet: assertion JSX parser on each backend. The mutant changes strict precedence > to >= and is caught only by output comparison after compilation and execution succeed.

The corpus is all 253 stage1 source files plus 77 TypeScript src/compiler files, pinned to TypeScript v6.0.3 commit 050880ce59e30b356b686bd3144efe24f875ebc8. All 330 files match 13,146,510 output bytes per backend using full projected Go ASTs. This certifies rule behavior on that AST contract, not independent parser parity for the entire corpus. Fixture output includes finding ids, messages, spans, fixes, converged fixed source, structured suggestion ids/messages/edits and suggestion-applied source.

Run from the repository with the setup environment sourced:

    go test ./stage1/cohere/lint/rules/typescript-consistent-type-assertions -run TestAssertions -timeout 15m -count=1 -v
    go test ./stage1/cohere/lint/rules/typescript-consistent-type-assertions -run TestAssertionCorpus -timeout 30m -count=1 -v
    go test ./stage1/cohere/lint/rules/structure-tailwind-no-physical-direction -run TestRuleContracts -timeout 15m -count=1 -v
    go test ./stage1/cohere/lint/registry -count=1 -v
    go test ./stage1/cohere/lint -run TestRulesAgree -timeout 15m -count=1 -v

Every command redirects its output to a log; retained logs are in this directory. TestRulesAgree fails before exercising the rules: lint.ts imports missing ./volume.ts, ./messages.ts and ./comments.ts. No shared implementation was changed to work around this. The harness branch was merged; eight merge conflicts took its exact committed versions. Its deleted-in-our-history volume_test.go was restored verbatim to resolve missing volumeGenerated/checkRecoveryRefusal test dependencies. Setup initially passed in 21s; after integration failed on those undefined symbols; after restoration passed in 28s. nproc=5.

Fixture throughput (findings/s including process startup): native 254.420, Node 655.505, Go 9808.253, from 111 findings in 0.436285s/0.169335s/0.011317s. Go parses raw source; these Node/native measurements consume projected ASTs, so they are not comparable parser benchmarks. The other three rules retain their prior rates and corpus evidence under structure-tailwind-no-physical-direction/testdata/evidence; their .a revalidation and all three semantic mutants pass in recertify.log. The older temporary-.ts report is historical and is superseded by this .a conversion.
