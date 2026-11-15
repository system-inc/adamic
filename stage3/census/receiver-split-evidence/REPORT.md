# Receiver split build checkpoint

Status: generated and guarded; full proof pending. No adaptation is certified by this checkpoint. No Loom submission.

Base: `8ada94542fbdcb5c1d1689e27564e471eae241b8`, fresh detached scratch worktree. Recursive shallow submodules initialized. Upstream clone: v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`. Locked stage3/api npm ci passed. Baseline stage3/apply.sh passed.

The unmodified split-codemod.cjs from accepted `24c61e74d26caa79c0f313112902e325d090afcd` generated all 19 files. Its complete rewrite receipt is byte-for-byte identical to the accepted receipt: 79 roots, 19 sites, 46 parser uses, 52 transformer arrows. A second run with --verify reproduced all 19 generated files byte-for-byte. The safety scan found no new failure. The public-site inventory supplied to it was reconstructed from the accepted split-rewrite.json (file, site, kind); the generator reads no member inventory for its edits.

The adapters pin complete enclosing owners and changed namespace imports, before/after text and sha256. They preserve homogeneous source line endings and reject mixed endings. Real-adapter virtual-filesystem controls pass for creation, zero-write replay, exact-byte restoring mutants, duplicates, drift, partial owners, and mixed endings; transformers additionally reject a half-adapted multi-file tree before writing. Output matches normalized generator output; source replacements were never hand-edited.

The required prior owner grep is retained in owner-grep.json. None of the 19 enclosing owners has an existing whole-owner pin. Existing nested-owner pins remain unchanged. Each transformer destructure is directly in its transformer/serializer function body; none is in a nested per-node visitor. Individual scope decisions and dynamic counts remain to be proved.

Remaining: full single-file CommonJSBundle censuses and required pairing, stock and Adamic type checks, real apply integration, JavaScript scope comparison, declarations, complete unfiltered lanes, dynamic receiver controls and counts, ledger entries, and the three adaptation branch pushes.

Raw sources and logs remain in scratch; raw-manifest.json records sizes and sha256. This checkpoint adds no Go files.

## Census checkpoint

The guarded new-base census tool built, and its continuation, attribution, mutant and no-output audit passed. The first full CommonJSBundle pair, es2021.ts, has 27 base / 25 changed findings: all 25 surviving records pair by (kind, phase, reason, message with source positions stripped), with 0 new and 0 changed outside the target. The two removals are the target method-destructuring finding and its boundary. The printer baseline completed with 1,430 findings and zero checker diagnostics; its changed run is pending.

The parser audit was expanded to the 34 nearest function owners containing the 46 rewritten calls, in addition to the namespace. No existing whole-owner amendment was found there either. Stronger nested owner pins have been prepared; their final adapter control rerun is pending. Checkpoint recipe files retain their original hashes; stronger recipe versions use new v2 names.

## Parser and printer full pairs complete

Parser: 3,183 base / 3,181 changed findings, all 3,181 survivors paired and exactly the two target destructuring records removed. Printer: 1,430 base / 1,428 changed findings, all 1,428 survivors paired and exactly the two target var records removed. Both have 0 new and 0 changed findings outside targets. The complete printer record has 0 findings at the readonly view and all ten aliases.

The first pairing wrapper conservatively stopped on printer var multiplicity attribution, not on a new or changed finding. The position-stripped key correctly paired all survivors but left indistinguishable later var records as removals. An independent unchanged-source line map proves all other createPrinter var records retain their source sites, and that the two removals are precisely line 1215. The corrected attribution keeps the required pairing key unchanged. Both wrapper versions and all completed raw files remain in scratch.

AST scope audit: all 17 transformer destructures are directly in top-level factory bodies, with context introduced by that factory parameter. All 52 new arrows are directly in the same body, and all 17 factory property bindings remain. No nested per-node or visitor-call placement is eligible to lift. Dynamic counts remain pending.

The remaining 16 transformer file pairs and all runtime/checker/integration/declaration proofs are still pending; none of the three adaptations is yet certified or pushed.
