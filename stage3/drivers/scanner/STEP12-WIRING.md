# Step 12 scanner evidence wiring

`scratch-run.sh NEW_OUTPUT REF...` retains the original scratch integration,
source construction and both scanner compilation modes. This branch imports
that command and its helpers from origin/codex/stage3-scanner-native-3, without
merging that branch's compiler changes. The comparator is copied unchanged
from the scout at ebaf1bc0. It compares every byte and both EOFs, hashes the full
streams, and reports zero-based offset and one-based byte line/column.

The command requires a clean source checkout. adamic_sha is the merged scratch
compiler commit; source_sha is the caller checkout commit containing the driver,
apply profile, adaptations and slice recipe. It compares the full-tree Node
reference against each slice Node stream, then against each native stream.
Each successful mode retains a self-contained evidence-split-N directory with
both complete stdout files, comparator stdout/stderr and a copied native output
whose first byte is XORed with 1. The same comparator must reject that mutant
at offset 0. Existing native exit, full-tree and split-mode checks still apply.

After both modes and their measurements succeed, scanner-native-evidence.json
contains an `evidence.scanner_native` block with all fields named by
stage3/progress.json's evidence_required. The command does not edit progress.json.
The meter must commit the retained evidence directory and set run_directory to
its committed repository path when copying the block into progress.json. The
normal command's top-level block is published only after both modes succeed.
A mismatching native stream produces a first_difference report and exits 1.
Compilation failures and merge conflicts produce no top-level evidence block.

## Stand-in proof

The test-only options `--stand-in-native EXECUTABLE --stand-in-node NODE_STDOUT
--adamic-sha FULL_SHA --source-sha FULL_SHA` exercise this command without setup,
fetch, compilation or a scanner source claim. The stand-in executes as a process.
It produces the same block shape, with stand_in true and milestone_eligible false.
These blocks cannot establish scanner_native. Matching stdout must pass; a
single changed byte must fail and name its position; neither an unsuccessful
process nor empty output can produce a block. The tests generate expected bytes
with Node independently. No new Adamic oracle fixture is added, so counts.md
needs no change.

Targeted checks:

```sh
source /workspace/adamic-tools/env.sh
python3 -B -m unittest discover -s stage3/drivers/scanner -p 'test_scratch*.py' -v > /tmp/step12-wiring-tests.log 2>&1
go test ./stage3/scouts/step12/compare -count=1 -v > /tmp/step12-wiring-compare-tests.log 2>&1
```

The retained wiring evidence includes test and setup logs. Matching stand-in
passes and emits the complete block; its native-output mutant is caught at
byte 0. The changed executable's output fails at offset 16, line 2, column 2,
Node byte 79 and native byte 78. Comparator tests also reject four single-byte
changes including block boundaries and the final row, plus truncated/appended
EOFs. Summary tests reject a missing evidence block, a missing split mode,
a surviving native mutant and a merge-conflict falsely labeled pass.

No real native scanner execution or full gate is claimed by this wiring unit.
