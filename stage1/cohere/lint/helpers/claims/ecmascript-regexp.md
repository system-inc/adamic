# ecmascript/regexp package claim

Branch: lint-helpers/ecmascript-regexp. Base: origin/area/stage1-lint at fb6cb5db.
Triage: d7ab0bc4, rank 15. Earlier eligible packages are claimed; rank 13 is per-rule option contracts, not a new generic helper package. Tonight exclusions honored.
44 retained / 45 required symbols; no complete port on any audited branch. Missing Compile has a historical nonconstant RegExp compiler blocker that must be rechecked on this base.
Partial sources: codex/lint-helpers-01, -02, -04, -05; from-codex-lint-wave1-14; from-codex/lint-wave1-10; from-codex/lint-wave1-15.
Conditional forecast: one rule alone when complete, cumulative 117 with earlier complete packages. No readiness credit before current Go/Node/emitted-JavaScript/sanitized-native and semantic-mutant gates.

Consumers:

- @next/next/no-html-link-for-pages
- @typescript-eslint/no-empty-object-type
- no-restricted-exports
- no-restricted-imports

Claim pushed before implementation. Stop on an unexpressible compiler gap with the upstream Go line and fresh compiler diagnostic.
