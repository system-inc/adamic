# Narrowing filter copy investigation — #ev14ch9

Work in progress; no accepted adaptation or landing commit.

Base: 49f1d643064a9231f7f343d8bbd02ba718866e6a.
Upstream: TypeScript v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.
Freeze evidence: df886e406b521bacf030d7e8ca76d6e828fed249.
Count evidence: 888d2a68c6418d3c760ef550f9b1d49d83377b15.

First gates: unchanged adapted tree, full single-file core latent census with
concrete helper calls, and isolated complete-filter-owner scan-all controls for
both remaining readonly narrowing declarations. A generic unit deferred without
a concrete entry is not successful body lowering.

The frozen site list has 16 entries: five overload-1 and eleven overload-5 calls.
Exact-call search across every owners.json, targets.json and adapt.cjs finds the
transformer call inside receiver-transformers' whole transformTypeScript pin and
the command-line call in axolotl-j's script. Imports and containing-owner overlap
still require checking before any landing edits. No pins have been amended.

Not run yet: lane, public declaration build, diagnostics, product test, mutants.
No acceptance claim is made from the freeze lane, which preserved identity.
