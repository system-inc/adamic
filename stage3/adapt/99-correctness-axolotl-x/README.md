# 99-correctness-axolotl-x: blocked evidence

Task kpp0f8s. Base origin/main: 3e395f9e5a7038d859ec8930be72cde68d8be3ff.

Status: not implemented. javascriptChanging true is the requested classification
of the proposed adaptation, not a claim that this branch changes JavaScript.

Adamic's ruling, supplied in the task on November 6, 2026, authorizes a stage 3
rewrite of maybeBind(X, X.m) to X.m?.bind(X) for a single identifier X.
Both forms read X, then X.m, then X again with nothing between. A function
returns bound to X; undefined returns undefined; any other truthy value throws
TypeError at .bind in both. core.ts's maybeBind body must remain unchanged.

Requested inventory: 46 sites across builder.ts (3), checker.ts (2),
program.ts (9), tsbuildPublic.ts (8), watch.ts (16), watchPublic.ts (8).
These are task-supplied counts, not a measured AST inventory. All 46 requested
sites remain as written because implementation never began. Rewritten: 0.
No import specifier was removed. No casts were added.

The user requires reading both reference adapters in full first. The k adapter
was read in full. The w adapter does not exist at the fetched main tip or in
the three related adaptation heads inspected. See prerequisites.log.

Check 1's baseline full apply with --build-product exited 1 before cloning or
running any adaptation: ADAMIC_TYPESCRIPT_SOURCE is unset. apply.py requires
the gate inputs' pinned TypeScript v6.0.3 checkout at
050880ce59e30b356b686bd3144efe24f875ebc8. See check1-before.log.
The task explicitly says to stop and push on a check 1 failure.

This directory contains evidence only. It has no adapt.cjs and is not a
runnable adaptation. Do not treat this branch as ready to merge or apply.
See PROOF.md for the exact stopping point and unrun checks.
