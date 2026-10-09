The sameMap draft is rejected and not activated. See ../../TRUTHFUL-SITES.md.

Fresh main and topic worktrees both start at origin/main 89ac4a8c. The topic
checkout adds only adaptation 76; no compiler merge or compiler source edit is
included. The proof branch merges current main normally. Full lanes use
NODE_OPTIONS=--max-old-space-size=1536 and stage3/lane/run.sh in each checkout;
that runs apply.sh and the complete upstream oracle. Complete logs and reports
are retained. No new public API sanctions are added.

The rejected scratch copies the adapted compiler source and links the baseline's
pinned node_modules. Each trial runs stock TypeScript 6.0.3:
node <stock-typescript>/bin/tsc -p <trial>/src/compiler --noEmit --pretty false.
All five trials intentionally fail; their failures are not lane passes.
Zero-context patches preserve the type-only proposal while ignoring CRLF-only
noise. Core and builder transpileModule output is byte-identical with comments
both retained and removed. Public types.ts is byte-identical to the baseline.
The built JavaScript/API byte identity covers the ACTIVE adapter, not the draft.

Fixture: node --disable-warning=ExperimentalWarning oracle/node.mjs
stage3/adapt/76-truthful-casts/same-map-chain-boundary.a.
Expected stdout is four false rows, stderr empty, exit 0. It models the general
alias contract, not an observed occurrence in a real builder caller.
Contract check: NODE_PATH=<stock-api-node_modules> node
stage3/adapt/76-truthful-casts/check-chain-boundary.cjs.
All four lying-signature mutants admit the forbidden public assignment and are
caught by the negative contract check. Their assertions exist only in virtual
mutants and erase to the same JavaScript. The first version unnecessarily added
parentheses, failed exact erasure, and was corrected; both logs are retained.

Active check.cjs passes on the fresh topic tree, including its existing key and
sameMap-signature mutants. 32 local stage3/lane tests pass. No native parser run
is attempted in this type-only source-contract unit. No public declarations are
edited, and no partially repaired sameMap signature is pushed into apply.

Isolated topic commit, based directly on main: c581cf81e24f072dd2ddf5d4e296d83d3aa3babd.
It is local and not pushed. Its staged paths are all under adaptation 76; the
existing proof branch receives this same unit without rewriting pushed history.
