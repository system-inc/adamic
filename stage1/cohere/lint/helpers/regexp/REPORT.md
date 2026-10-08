# ecmascript/regexp stopped at Compile

Claim f105782dec162f8d60ec0603e4bc90d496539500 was pushed before any probe or implementation. The branch starts from origin/area/stage1-lint at fb6cb5dbf79e247d74dccf16d8e3aefc5b31c039. Both fetches explicitly included every origin/lint-helpers/* and origin/codex/lint-helpers* ref because the checkout has a narrow default refspec. The post-push scan found no competing package claim; evidence/claim-scan.json records all checked branch tips and claim addition commits.

Read origin/lint-helpers/triage at d7ab0bc423f2c6e987a3a8651aa8a69925d2a59c, the whole TRIAGE.md and helpers/README.md. Rank 15 is the highest eligible unclaimed helper package: tonight's excluded packages were omitted; rules/react, ecmascript/text, ecmascript/nextjs, ecmascript/regexsyntax, rules/tailwind, rules/core, and rules/tailwind/collapse already have package claims. Rank 13 is per-rule strict-option contracts, not another missing generic helper. No branch has a complete regexp package; the triage retains 44 of 45 symbols with Compile missing.

## Fresh blocker

The pinned Go algorithm compiles a runtime rewritten pattern at cohere/internal/lint/ecmascript/regexp/regexp.go:154, in Compile (entry at line 121). The same source and flags cannot be encoded as fixed regex literals because they come from rule options. The current compiler refuses dynamic regex construction at internal/lower/regexp.go:39. With runtime pattern a, source Node and unchanged Go cohere both return true; sanitized-native build and emitted-JavaScript compilation both exit 1 with:

```
/workspace/adamic/stage1/cohere/lint/helpers/regexp/gaps/dynamic_pattern.a:3:27: stage 0 can't lower RegExp with a nonconstant pattern yet
```

See evidence/node.log, native.log, emitted.log and gap-test.log. This is a fresh observation on the selected base, not copied historical success. The reproducer and narrow overlay oracle are reused from origin/codex/lint-helpers-05 at cd24dd580a8e8d97cdc1cf0aa839b47bf333e9bc, slot05/batch34. Its boundary test is adapted to the owned directory. No production partial helper is extracted because the package stops at its missing compiler operation. Other potential source owners are recorded in the claim.

TestRegexpCompileDynamicPatternGap proves the actual Go/Node result and asserts the typed lower.NotYet diagnostic. The oracle is included through a temporary Go overlay and the cohere checkout stays untouched. This test is a blocker regression, not a green port or a semantic mutant. One Go/Node probe agrees; zero helper-use capture cases are certified across all backends. There is no executable native/emitted implementation to mutate, so zero Node/native semantic mutants and zero consuming-rule ports are claimed.

The blocked consumers remain @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. The complete-package conditional forecast was one rule alone, or cumulative 117 with the earlier complete packages; actual new rules unblocked is zero.

All requested external input variables were supplied in evidence/inputs.log, including the clean TypeScript checkout at 050880ce59e30b356b686bd3144efe24f875ebc8 and the WASI SDK. Both profile variables point to the same fresh directory. The existing helpers package plus this regression run is recorded in evidence/helpers.log. The lint package and rule port are not run after this unexpressible gap, following the instruction to stop; no package findings parity is claimed.

Final local check: `go test ./stage1/cohere/lint/helpers ./stage1/cohere/lint/helpers/regexp -count=1 -v -timeout=20m` passed. The existing helpers package has four top-level passes, zero failures/skips, 139.463 seconds; the blocker package has one pass, zero failures/skips, 0.390 seconds. Existing helper baseline, four compiled semantic mutants, ten message-refusal comparisons and explicit-gap checks passed. These inherited results are not regexp implementation coverage.
