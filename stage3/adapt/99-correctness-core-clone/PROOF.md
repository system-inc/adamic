# Proof

Base: f0271c6a7b2e92900eb7046d8246cfb89f52cfb6. TypeScript 6.0.3 source pin:
050880ce59e30b356b686bd3144efe24f875ebc8. Both complete applies exit 0.
The checkpoint branches preserve the completed applies before the lane run.
Both active adapter files, adapt.cjs and owners.json, are byte-identical to
those checkpoints. Later proof files change product keys without changing the
frozen transformation.

| Check | Result |
| --- | --- |
| Full apply, main and main plus this directory | Both exit 0 |
| Ordinary Adamic types, 79 sorted compiler roots together | Both exit 0, zero diagnostics |
| Owner contract | Pass; idempotent, zero replay writes, drift refusal before writes, LF/CRLF preserved |
| Complete source scope | 746 files compared; only the pinned owner differs |
| Complete core.ts JavaScript, comments kept and removed | Byte-identical |
| All 9 emitted JavaScript files | Byte-identical; zero hunks |
| typescript.d.ts and tsserverlibrary.d.ts | Byte-identical |
| typescript.internal.d.ts | Byte-identical; both functions occur only here |
| Every adapter paired replay | 84 adapters; only this adapter's outcome differs |
| Full core.ts site census | 59 distinct sites before, 58 after; target gone, zero new sites |
| Original-owner mutant | Exact main core.ts bytes restored; original site set returns |
| Unchanged whole lane control | Pass with the sanctioned API failure |
| Counted lane union | 398 guarded shards in 20 bounded pieces; same test identities and baseline bytes |
| Counted behavior equivalence | Zero counterexamples; selected JavaScript is byte-identical |
| Node witnesses | Equal before/after results and exceptions on all retained input classes |

The paired replay retains these pre-existing, identical baseline refusals:
30-indexed-reads, 32-indexed-reads-program, 33-indexed-reads-emit, 40-explicit-any, 41-explicit-any-remaining, 43-any-returns, 47-host-errors, 65-temporary-node-builtins, 76-truthful-casts, 93-builtin-imports, 95-checker-axolotl-b, 95-checker-axolotl-h, 98-axolotl-e, 99-correctness-axolotl-g. They have identical exit status, normalized stdout,
normalized stderr, and source edits before and after this adaptation. No refusal
is caused by this edit. In particular adaptation 30 stops at its historical
core.ts:992 declined-site label in both trees. The independent AST check in
evidence/untouched-site.json confirms that its groupBy values[i]! remains the
second of three matching reads in all three completed source trees.

The census follows the core-sites-4186581e full meter: the unmodified
measurement overlay generator, LATENT_FULL=1, LATENT_ONLY_FILE set to the
complete core.ts path, LATENT_ASSERT_NO_OUTPUT=1, and a timing JSONL file. It
probes the completed compiler root set and retains all raw findings and deferred
units. Each tree has 271 units and 208 deferred generic units; this proof claims
the specific refusal removal, not successful native lowering of all units.
The ordinary types binary has no compiler-source overlay.

Setup used source /workspace/adamic-tools/env.sh, bash cloud/setup.sh,
npm ci --prefix stage3/api, and ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source.
Every command's output goes to a retained log. Exact commands, root lists, tool
hashes, source manifests, patch tables, native census rows and timings, mutants,
paired replay rows, whole-lane task records, and each counted shard's guards
are packed in evidence.tar.gz with SHA256SUMS. Generated Go overlays are .go.txt.
The LF display diff is source.patch; evidence/raw-scope.json verifies the exact
raw-byte owner-only delta including CRLF. No JavaScript hunk exists to display.
