# rules/typescript package claim

Branch: lint-helpers/rules-typescript. Base: origin/area/stage1-lint at fb6cb5db. Triage d7ab0bc4, rank 17; two missing symbols, no partial or complete port on any audited branch. Earlier eligible packages are claimed, excluded tonight, or parked; runtime regex compilation packages are skipped. The two shared helper bodies perform no runtime regex compilation.

Helpers: isTypeScriptSourceFile (no_explicit_any.go:160), nonNullAssertionOperatorRange (no_extra_non_null_assertion.go:115).
Conditional forecast: three rules alone / 123 cumulative with preceding complete packages.
Frozen remaining consumers unblocked: @typescript-eslint/no-this-alias, @typescript-eslint/triple-slash-reference, @typescript-eslint/no-non-null-asserted-optional-chain.
All helper consumers additionally include @typescript-eslint/no-explicit-any, @typescript-eslint/ban-ts-comment, @typescript-eslint/no-extra-non-null-assertion and @typescript-eslint/no-non-null-assertion. Capture every use in their pinned upstream tests and compare Node, emitted JavaScript and sanitized native, with semantic mutants on Node and native.

Claim pushed before code; fetch again and yield to any earlier competing package claim. Build fresh; no production partial helper to reuse. Prove one consuming rule in its own directory and run the helpers and lint packages with every required input supplied before the finished-unit push.

Shared prerequisite discovered before implementation: internal/types/sourcename.TreatedAs. Build this smaller helper first in helpers/sourcename/treated_as.a, rather than duplicating its .a mapping inside the TypeScript predicate.

Status: complete. Both TypeScript helpers and the sourcename.TreatedAs prerequisite have Go capture parity and individual running mutants on Node, emitted JavaScript and sanitized native. The no-non-null-asserted-optional-chain consuming rule matches 24 upstream cases and its witness/corpus, with a running mutant. Complete helpers and lint suites passed; evidence is in helpers/typescript/evidence and rules/typescript-no-non-null-asserted-optional-chain/evidence. No dependency or language blocker remains.
