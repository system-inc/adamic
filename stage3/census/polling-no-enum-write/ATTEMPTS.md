# Preliminary source attempt

An isolated owner application to the pinned upstream working copy exited 0
(pristine-owner-apply.txt). The following preliminary source CommonJS witness
exited 1 before system initialization: the pristine checkout lacked
src/compiler/diagnosticInformationMap.generated.ts, so require of
../diagnosticInformationMap.generated.js failed with MODULE_NOT_FOUND.
pristine-after-custom.txt is its empty stdout, not a passing witness.
The accepted before/after witness runs use the complete adapted trees, including
the normal generated inputs, and are recorded separately by verify.py.

Both full applies clone the pinned Git commit's objects, not this working copy's
modified sys.ts. Thus the preliminary owner edit is not an apply input.

# Initial stock harness input omission

The first stock checks exited 1 on both trees (64 diagnostics on the changed
tree) because stock.cjs supplied only compiler roots and omitted the Node
library. Adamic's loader automatically adds its bundled @types/node 25.3.3 and
undici-types 7.18.2 plus the existing host_axioms.json corrections. The corrected
stock harness explicitly supplies that same Node input; prepare-node.py retains
the exact source package and row-by-row preparation method. No source edit,
checker option relaxation or TypeScript checker modification was used.
The original stock logs and verification-exits.json retain those failures;
*-stock-corrected.log and final-checks.json are the acceptance runs.
