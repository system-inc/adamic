Current recut: origin/main 698965fe66aed945d18efa320a594a8a8b4df2f8.
The earlier f0271c6a7 results below are historical until replaced by the recut proof.

# Proof status

Frozen base: f0271c6a7, origin/main at task start. Upstream source:
TypeScript 6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8. The ordinary
compiler is built from this main without an overlay. The separate census
binary uses only the guarded measurement overlay.

| Check | Result |
| --- | --- |
| Full apply, main | Exit 0 |
| Full apply, main plus this adaptation | Exit 0 |
| Owner contract | Pass, 12 no-write refusals and LF/CRLF zero-write replay |
| Exact source scope | All 746 files compared; only the two pinned core.ts owners differ |
| Adamic types, all 79 sorted compiler roots together | Both exit 0, zero diagnostics |
| Ordinary compiler builds | Both exit 0 |
| All nine emitted compiler JavaScript files | Byte-identical, zero diff hunks |
| Complete core JavaScript, comments kept and removed | Byte-identical |
| typescript.d.ts and tsserverlibrary.d.ts | Byte-identical, zero diff hunks |
| Full core site census | Exactly F6-F11 gone; no new finding anywhere in core.ts |
| Original-owner mutant | Restores exactly the complete baseline finding set |
| Node owner witnesses | All 14 owner/class pairs agree |
| Observer contract | Shape seeds pass; getters untouched; return and exception identities preserved |
| Unchanged whole stage3/lane/run.sh control | Pass, 106366 passes, 1 sanctioned failure, 0 skips |
| Every-adapter paired replay | Pass, all 84 outcomes match apart from this adapter; 14 existing refusals match |
| Final whole-inventory runtime shape count | In progress, bounded pieces with frozen observer/source/harness |

The whole control uses an isolated checkout of exactly the frozen main and
restores the baseline keyed product produced by the first full apply. It runs
the unchanged lane script and all normal API, baseline, count and identity
guards. The final instrumented count uses the unchanged 398-unit lane plan and
shard runner. Its test identities and raw baseline diff will be independently
reconciled with the completed whole control.

The observer adds wrappers only in a scratch tree. It records own symbol keys,
user accessors, inherited enumerable strings, and plain/non-plain prototypes
without reading getter values. The ordinary inherited Object.prototype
__proto__ accessor is excluded from the user-accessor category. Standard
Object.prototype is recognized across VM realms; null-prototype records are
listed separately. The complete 80520-file fixture corpus and all 81504
non-build inputs were independently verified before the accepted rerun. Every invocation is recorded, including throws. Seeds,
the zero-call view-setup failure, the superseded prototype-classifier prefix,
and the incomplete-fixture-copy attempt are excluded from accepted totals and retained separately.

Both bodies remain deferred_generic_instantiation, instantiated=false, in the
full site meter. See CENSUS.md for the limits of this scan result. No claim of
successful generic-body or whole-program lowering is made.

All commands source /workspace/adamic-tools/env.sh. Setup follows 99w PROOF.md:
bash cloud/setup.sh, npm ci --prefix stage3/api, and
ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source at the exact pin.
Reproduction programs are in reproduce/. Full logs, source manifests, binary
hashes, complete census streams, empty diff files, and emitted-file hashes are
packed in core-proof.tar.gz, covered by its internal SHA256SUMS and adjacent
SHA-256 file. Bulky lane pieces are packed separately before final closeout.
