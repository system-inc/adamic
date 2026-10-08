# ecmascript/comments package claim

Branch: lint-helpers/ecmascript-comments; base origin/area/stage1-lint 334509ee.
Triage d7ab0bc4 rank 27; build the two unported remaining helpers LeadingRunFor
and isAdjacentGap, retaining the five already-landed comment helpers.
Consumer unblocked: structure/react-hook-require-effect-comment. Frozen complete
forecast: zero alone, one with earlier packages (136 cumulative), conditional on
its other shared prerequisites and actual rule parity.

Before code: fetched all origin heads, checked 1,534 refs (378 unique cohere
trees, 77 distinct comment modules), and all 36 distinct claim blobs across 46
lint-helpers/* and codex/lint-helpers* branches. No port or claim for these two
symbols. Higher eligible packages are claimed or complete; runtime regex packages
and tonight exclusions skipped. Directives is already ported in suppression;
strict options row covers per-rule adapters, not another generic helper.

No retained partial port for these two in triage. Use the existing shared scan,
compare actual consumer upstream captures with Go on Node, emitted JavaScript
and sanitized native, and catch one running mutant per helper on all three.
Prove the consuming rule in its own directory. Push claim before code, fetch
again, and yield to any earlier competing package claim. Finish with complete
helpers and lint gates using all required inputs and one implementation push.

Post-push fetch: only this package claim, 017bc6c1b (12:56:49Z).
Rule prerequisite files are reused without ownership changes: FileContextFor from
origin/lint-helpers/structure (2079ada9), and IsNamespacedMember, its identifier
predicate and AST projection from origin/lint-helpers/react (ee4069f5). These
previously certified files are extracted individually, not whole-branch merges.
Both missing comment helper bodies are fresh; no retained partial exists.

Finished validation: all helpers/... packages pass with zero skips. The comments
package matches 33,331 Go output rows across three runtimes; seven helper mutants
are caught on each. The proof rule matches all 22 upstream cases and both
witnesses with its mutant caught on each. Full lint passes (4,598 captured
source/rule/options combinations; 84 rule mutants), with all corpus, benchmark
and profile inputs supplied. Its sole existing TSGoError-dependent test remains
skipped, as documented in comments/README.md; no new helper or rule is stopped.
