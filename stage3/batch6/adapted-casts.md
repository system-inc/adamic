# Batch 6 adapted casts on current main

Rebased 61038c44 onto bdb89962. Changes remain entirely under stage3/.

Ran the branch README classification, historical source measurement, observation,
header checks, coverage checks and their mutants. The full current stage3/lane/run.sh
completed in 462.672 seconds: apply exit 0, oracle exit 1, 106366 passing,
one known Public APIs failure, zero pending. The stage3 baseline is unchanged
between the measured baseline main c4c59914 and this base bdb89962.

Historical snapshots use 855bcfaa plus adaptation 43 from 643639ea. Initial
prepare.py hit its 600-second deadline during concurrent oracle work. Completed
snapshots and six API receipts were retained; the interrupted 70 tail was restored
from its before snapshot and the unchanged 70, 71, 75 and 76 adapters rerun with
idempotence checks. RESUME metadata is in adapted-casts-proof.json. Classification
has 31 exact AST rows; 21 proposed type-only changes have no semantic diagnostics
and byte-identical JavaScript. The remaining ten are not proven Adamic upcasts.

Node observations pass on all three existing fixtures. Their three source mutants
change the goldens. Actual Gate.aCheck from fbac28c6 passes all three truthful
refusal headers, catches six missing/wrong-header mutants, and passes restored.
The missing-row and wrong-owner record mutants fail their assertions. The runtime
edit mutant is required to fail the emitted-JavaScript guard, independently of rows.

The isolated checkout lacked pinned Node declarations and compiler submodules.
Installed stage3/api with npm ci; compiler production files and gitlinks are identical
to the initialized nonnull checkout. An external GOWORK resolves those same modules
there for the actual gate build, without changing repository compiler files.
Scratch reference corpora from the completed oracle run were removed after the disk
filled; reports, local baselines and built artifacts remain. No new fixtures or
native success claims are introduced. No Go tests were added or changed.

Commands: classify.py; historical prepare.py plus bounded tail resume;
measure.cjs; observe.py; a-check.py; verify.py, --drop-row, --wrong-owner;
measure.cjs --runtime-edit; bash stage3/lane/run.sh. All output was written to logs.
Lane checks are run after this commit, before the foreground push.
