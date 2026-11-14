> Historical stop from the first turn. The resumed real-file census proves the restoring mutant; see REPORT.md for the completed work.

# Stopped: exact-body restoring mutant is masked by an earlier wall

Base main: f2546f1fcb7c76f8dd549094401246296b6998a7.
Workers fallback, fetched by SHA: e8a1529e6f6addd4ae8dc28ee07c217bcb7b0f0e.
Upstream: 050880ce59e30b356b686bd3144efe24f875ebc8, exact tag v6.0.3, package version 6.0.3.

No adaptation was applied, committed, or pushed. The mutation programs preserve the exact flushMappingBuffer body in a concrete enclosing function. They differ only at the requested clear statement. Both main and Workers exit 1 at line 6:25, before the clear:

    stage 0 can't lower apply without a dense argument literal (length, presence and argument representations must be proven) yet

This is NotYet for the preceding String.fromCharCode.apply call, not the requested Refused diagnostic at the length assignment. Workers contains the required refusal implementation in internal/lower/never_refused.go, but these executions do not reach it. No simplified storage-only mutant is substituted for the required proof. This does not establish that a separate latent census could not expose the assignment; no such census was run.

The 99s README.md, adapt.cjs, owners.json, contract.cjs, witness.cjs, direct-mutant.cjs, concrete mutants, and generic-clear contract/proof were read. Searching all stage3/adapt for flushMappingBuffer and createSourceMapGenerator found no whole-owner claim. 33's required-values.json targets only JSON.stringify(toJSON()) at upstream line 83. 98's planSourcesContent inspects appendSourceMap and edits its sourcesContent guard/payload. Neither owns the proposed nested function. No conflicting whole-owner pin caused this stop.

The source array is private to createSourceMapGenerator. It is captured by its nested appendMappingCharCode and flushMappingBuffer functions; the only external argument use is String.fromCharCode.apply, which consumes its argument list synchronously. No array alias is assigned, stored or returned in this source. This is a source audit, not a dynamic alias census.

Counts, Node equivalence witness, ledger amendment, completed apply, emission comparison, type check and lane are NOT proved or claimed. The baseline apply was deliberately cancelled at 32-indexed-reads-program after the exact-body proof blocker was confirmed. Its child was terminated by SIGTERM; apply and the recorder then exited 1. This is cancellation, not an observed adaptation failure. No generated shard files changed.

Only this evidence directory is committed on the evidence branch. Scratch amendment/contract/witness drafts were prepared but never applied. No Loom submission was made.
