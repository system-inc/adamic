# JSDoc tag cache table

JavaScript-changing, #7ez1s16. The cache is one private strong Map keyed by the exact JSDoc array. Clones that share the array share its cached tags. An uncached request bypasses the table, a cacheable miss initializes `node.jsDoc ??= []` before storing, and parser reuse deletes the entry. `JSDocArray` is replaced by `JSDoc[]`. The helper is imported directly by the two readers and stays out of the `ts` barrel.

## Premises

- The pinned TypeScript version is 6.0.3. The owners are matched through that version's AST against each complete before and after owner text at this adaptation's position. Drift, partial adaptation and mixed line endings are refused before writes.
- In a short-lived tsc run the extra retention is for cached arrays that would otherwise become unreachable mid-run, including arrays left behind by dropped transform clones. The keys and cached tags are freed when the process exits. A long-lived language service needs a weak table first. Memory should verify ownership of array keys and tag values, including parent links, before submission.
- The round-five meter found 17 representation sites across 18 units and the distinct `no-optional-widening` refusal at adapted `parser.ts:1827:40`. The #ybz2259 scout counted 27 sites on an earlier main. Sixteen of today's sites share their reason. The other 11 cannot be named from any surviving record, so they remain unverified and do not bear on this adaptation. The paired census names every site this adaptation moves.
- Correctness ruled that the receiver-parser and MultiMap pins run before this adapter. The MultiMap finished-tree replay was a landing-time proof with no gate caller. This adaptation is separate. The eventual before/after census and the Memory ownership check belong to the maintainer.

## Clone test

Run `go test -count=1 -timeout 60m -v -run '^TestJSDocTagsAreSharedByClones$' ./stage3/lane/adaptedtree/` from the repository root. The test runs the clone fixture against stock TypeScript 6.0.3 and the lane's adapted product. The function clone shares the parsed function's JSDoc array before either read. The second read must reuse the first cached tag array. Keying the table by node makes the clone's tag read empty.
