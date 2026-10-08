# Predicate lowering slice

Built independent named-helper proofs, direct named callback contracts, and checked `.ts` narrowings for exact members of closed readonly discriminated unions.
Implementation SHA: recorded in the commit containing this report; branch `codex/notyet-predicates`.
Focused lower and oracle tests pass; both reduced programs emit C; counts refreshed.
Nine distinct mutants are caught by fixture failures or explicit refusal assertions; see evidence.
Verified coverage of the corrected refusal table: **0 of 122**. The structural predicate unit is incomplete.

## Scope and observations

Base: `b410340dc8f889b5799c3bc519117c63def3aa24`, the resolved compiler-area head when work began. Census replay `9a1f14c5d994aa855625e7cfa295677060348fec` was merged and pushed as `0de311f931132a534f4e6d5582f08f82a8981b1a`. Existing evidence-only commits are `50446255` and `c137f65f443eeb0bf9a704af6b90ccc1e85ef6ee`. Corrected table pin: `d35a81d36fdafccf827bad0f572d311b2a0d4deb`.

The raw CSV has exactly 80 findings for `test` receiving `isExpression`, 23 unproven returns on `node`, and 19 findings for `test` receiving `isStatement`. `coverage.csv` selects exactly these reasons and marks all 122 unverified. It excludes the separate three `nodeTest` findings and other similarly named predicates.

Two reductions expose the mechanism: `predicate_callback_contract.a` initially stops on its callback's bodyless signature; `predicate_helper_return.a` initially stops on `return isExpression(node)`. On this earlier compiler base, the callback signature stops before the newer table's unproven-argument diagnostic. Therefore these are mechanism reductions, not an exact replay of those 80/19 diagnostics. After this change, both `adamic c` commands succeed and produce C with no named stop. Their Node, JavaScript backend, release native, and sanitized native observations agree.

The three hatch witnesses are authored as `.a` and copied to temporary `.ts` paths by their own test. Each `.a` remains refused. Their `.ts` forms run on in source Node and stop with exit 70 and `predicate narrowing failed` in both backends, including release and sanitized native. True, false, and callback narrowings are tested separately. Only the two proven `.a` fixtures enter the ordinary fixture registry and counts table; the hatch tests independently assert the source/backend divergence.

## Soundness boundaries

A helper's claimed predicate is used only after its named declaration's body independently proves both directions. Recursive helpers do not establish their own proof. Callback signatures are accepted only for plain, unchanged parameters on direct named functions; every reference to that function must be a direct call with a named predicate argument. The checker already rejects assignment to named function declarations. A `.ts` argument may instead have an exact closed-union runtime contract. Aliases, overloads, generic callback contracts and callback escapes remain outside this slice.

Checked reads validate the tag against the consumed narrowed type, including the complement used after a false return. Every source member must have a unique readonly literal tag, and every target must be an identical source member. Optional payload refinements, additional fields, and open structural targets stay refused, including in `.ts`. Tests pin these refusals.

The real compiler's `isExpression` delegates through kind helpers and partially emitted expressions. `isStatement` additionally depends on parent context. Most return findings refine nested payloads, generated names, initializers, or other semantic conditions. These were not unlocked by replacing them with closed unions. No total of 122 is claimed from the category names.

The reference `codex/proven-predicates-2` has structural checked views and additional predicate summaries absent from this compiler-area base. It was inspected, not merged wholesale. A follow-up needs those contracts or a minimal equivalent before checking the real `.ts` structural narrowings. For `.a`, accepting a kind comparison as proof of additional payload fields would require a proven construction invariant or an explicit language ruling; the refusal is preserved here. Backend and runtime permission is no longer a blocker. No runtime C helpers were added.

## Validation

Commands used `source /workspace/adamic-tools/env.sh` and `GOPROXY='https://proxy.golang.org|direct'`. Output went directly to files.

- `go test ./internal/lower -run 'TestPredicate|TestUnprovenPredicateReturnsAreRefused' -count=1`: pass.
- `go test ./internal/oracle -run 'TestPredicateCheckedNarrowing|TestNativeAgreesWithNode/internal/oracle/testdata/predicate_(callback_contract|helper_return)' -count=1`: pass.
- `go test ./internal/oracle -run TestPredicateOpenContractsStayRefused -count=1`: pass.
- `go run ./cmd/adamic c internal/oracle/testdata/predicate_callback_contract.a`: pass, emitted C.
- `go run ./cmd/adamic c internal/oracle/testdata/predicate_helper_return.a`: pass, emitted C.
- `go test ./internal/oracle -run TestCountsAreRecorded -args -update-counts`: pass, adds the two fixture rows.
- `git diff --check`: pass.

No whole package suite or full gate was run. Touched packages are lower and oracle. Existing IR `CheckedCast` supplies both backends without changing their packages.

| Mutant | Catcher |
| --- | --- |
| Disable helper-call proof | Helper-return fixture refuses instead of lowering |
| Disable callback signature admission | Callback fixture refuses instead of lowering |
| Ignore callback parameter writes | Reassignment refusal assertion sees successful lowering |
| Bypass `.a` hatch refusal | `.a` hatch assertions see successful lowering |
| Remove true narrowing check | All three generated executions run on with exit 0 |
| Remove false narrowing check | All three generated executions run on with exit 0 |
| Disable checked callback admission | Callback hatch fixture cleanly refuses instead of lowering |
| Admit open source without closed contract | Open-payload refusal assertion sees successful lowering |
| Admit a target outside exact source members | Extra-refinement refusal assertion sees successful lowering |

The callback-admission mutant was rerun after adding the bodyless-signature guard; its final evidence is a clean refusal, not a compiler panic. All mutation edits were restored. Scripts and final logs are in `evidence/`.

Toolchain setup retry completed in 69.904 seconds: Go ready 0.038s, Node 0.051s, submodules 0.146s, markdown skipped 0.019s/ready 0.162s, clang 0.392s, Go build 69.677s, test binaries deferred 69.812s, cache warm 69.814s. `nproc` was 5, with CPU quota 4. The first setup attempt overlapped checkout and saw missing compiler symbols; retry succeeded. Its original logs remain under `stage3/notyet-predicates/evidence/`.

Ownership check fetched `origin/codex/notyet-*` and inspected their non-ancestor changes to the touched lower files. Other workers changed `enumNeverValue`, `prefix`, `combine`, and `callClosure`; this slice adds a three-line hook in `expression` and a one-line predicate dispatch in `refuse`. It does not change those workers' functions.
