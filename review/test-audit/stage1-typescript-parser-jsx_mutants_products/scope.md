CODE UNDER TEST: JSX survivor and ownership checks, union construction, and build-product recipes. ORACLE: executed unmodified typescript-go parser, plus self-written ownership and coverage/build assertions.

Local check/construction functions: jsxMutantsOracle, jsxMutantsLower, jsxMutantsNative, jsxMutantsRows, jsxMutantsUnion, jsxMutantsSelection, jsxMutantsRun, jsxMutantsPlantedProof, jsxMutantsCommand, jsxMutantsFetch, jsxMutantsSource.

Preparation functions read whole: jsxManifest, jsxMutantsChanges, copyPort, execute, difference, wholeNode. The external oracle adapter was read whole and not edited. No claim of full transitive compiler/runtime coverage is made.

Fixed edits before results: W1 disable the bytes.Equal survivor condition; S1 rename program.c output; S2 drop seen[id] assignment. Empty-entry probes P1 runner, P2 union, P3 oracle, P4 lower, P5 native are separate. No production parser mutant was planted because the selected rows are witnesses and product construction.

All 31 named Tests exist; none moved or vanished. The family rule combines nine leaves with their union and groups all 21 product recipes through jsxMutantsFetch. Complete members and individual outcomes are preserved. Matrix scope is exactly the requested slice; all outside rows are unknown.
