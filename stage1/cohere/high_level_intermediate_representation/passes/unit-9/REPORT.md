# Unit 9 stopped at the shared Scope terminal boundary

Base: 9addb0e8a, branch hir/unit-9. Cohere: 7945d102a6c18dd36adf9114a758ce646e8b2359.

No pass implementation or census certificate is claimed. Node: 0/1,465 pass
records; emitted JavaScript: 0/1,465; sanitized native: 0/1,465. The 23 Flow
records are not dropped or classified as successes. Semantic mutants are not
run because an unmutated pass cannot yet consume its required input.

## Exact blocker and owner request

Go terminal.go:324 defines Scope with Scope, Block, Fallthrough and Order.
FlattenReactiveLoops (flatten_reactive_loops.go:51, Scope switch at :100)
requires it. FlattenScopesWithHooksOrUse (flatten_scopes_with_hooks.go:38,
Scope checks at :75 and :86) also reads it and can rewrite it to Label.

Shared core.ts:87 TerminalType excludes Scope; replay/decode.ts:109
readTerminal reaches :134 and panics with unknown terminal Scope. The shared
ScopeIndex already exists. Request from hir-01: add the Scope terminal using
ScopeIndex and checked scope identity lookup, BlockIndex body/fallthrough,
and existing terminal order; decode and re-encode it without changing Go IDs
or construction dumps. Include shared copy/graph/terminal visitors needed to
preserve Scope through replay. No private index mint, string identity or
parallel terminal model is introduced here.

The five tree subpasses also need the published unit-6 ScopeIdentity,
unit-7 ScopeDependencies and unit-8 reactive arena record contracts. Request
those public schema/import paths through the owner; they are absent on this
base. Their algorithms are not prerequisites for Go-input replay. Unit 9
will own its state sidecars once these contracts can be read by reference.

Exact Go consumers: prune_non_escaping_scopes.go:95
PruneNonEscapingScopesWithScopes; prune_non_reactive_dependencies.go:38
PruneNonReactiveDependencies; prune_unused_scopes.go:36 PruneUnusedScopes;
merge_invalidating.go:328 MergeReactiveScopesThatInvalidateTogether;
prune_always_invalidating.go:37 PruneAlwaysInvalidatingScopes.

## Reproducer and retained work

scope_replay.a imports ../../replay/index.ts and calls the existing shared
readTerminal with a Go Scope payload. Node refuses with
`adamic: panic: unknown terminal Scope`. This is a missing shared record,
not evidence of a compiler language gap.

oracle_test.go is a lintoracle tagged adapter, overlaid beside real Go HIR.
Its top-level test calls t.Parallel. It exports the exact graph immediately
before/after real FlattenReactiveLoops on a loop/scope witness and its
returned flattened identities. This is a boundary witness, not a census
exporter or an integrated pipeline. No shared files are changed.

Setup log: /workspace/hir-unit-9-setup.log. Go ready 0.023s; Node 0.026s;
markdown 0.070s; clang 0.229s; nproc 5. Setup reported that the reused
cohere dependency is a symlink. The pinned physical dependency workspace is
used for the Go overlay; no tracked dependency or workspace file is changed.

Validation logs: /workspace/hir-unit-9-go.log,
/workspace/hir-unit-9-node.log, /workspace/hir-unit-9-native-build.log.
Stopped before pass implementation; no sampled comparisons or fake no-op
certification. Resume after the requested shared records land, merging the
owner branch without rebasing.

Go boundary witness passed (package 0.006s). Sanitized native build was
interrupted when the workspace filled; it did not execute. Native and emitted
JavaScript boundary checks are unrun. The initial logical-symlink GOWORK
invocation refused the package; the physical GOWORK rerun passed.

The local TestSharedScopeBoundary regression passed (0.188s). Validation
logs are retained in validation/. Raw Go graph fixtures intentionally retain
trailing spaces in empty params/context rows for byte identity.
