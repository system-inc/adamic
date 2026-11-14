# Receiver split build checkpoint

Status: generated and guarded; full proof pending. No adaptation is certified by this checkpoint. No Loom submission.

Base: `8ada94542fbdcb5c1d1689e27564e471eae241b8`, fresh detached scratch worktree. Recursive shallow submodules initialized. Upstream clone: v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`. Locked stage3/api npm ci passed. Baseline stage3/apply.sh passed.

The unmodified split-codemod.cjs from accepted `24c61e74d26caa79c0f313112902e325d090afcd` generated all 19 files. Its complete rewrite receipt is byte-for-byte identical to the accepted receipt: 79 roots, 19 sites, 46 parser uses, 52 transformer arrows. A second run with --verify reproduced all 19 generated files byte-for-byte. The safety scan found no new failure. The public-site inventory supplied to it was reconstructed from the accepted split-rewrite.json (file, site, kind); the generator reads no member inventory for its edits.

The adapters pin complete enclosing owners and changed namespace imports, before/after text and sha256. They preserve homogeneous source line endings and reject mixed endings. Real-adapter virtual-filesystem controls pass for creation, zero-write replay, exact-byte restoring mutants, duplicates, drift, partial owners, and mixed endings; transformers additionally reject a half-adapted multi-file tree before writing. Output matches normalized generator output; source replacements were never hand-edited.

The required prior owner grep is retained in owner-grep.json. None of the 19 enclosing owners has an existing whole-owner pin. Existing nested-owner pins remain unchanged. Each transformer destructure is directly in its transformer/serializer function body; none is in a nested per-node visitor. Individual scope decisions and dynamic counts remain to be proved.

Remaining: full single-file CommonJSBundle censuses and required pairing, stock and Adamic type checks, real apply integration, JavaScript scope comparison, declarations, complete unfiltered lanes, dynamic receiver controls and counts, ledger entries, and the three adaptation branch pushes.

Raw sources and logs remain in scratch; raw-manifest.json records sizes and sha256. This checkpoint adds no Go files.
