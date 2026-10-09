Run bash review/test-audit/internal-regexp-parser/replay.sh M7 from the evidence branch. The script applies one standalone diff, runs vet and the package, and reverses the diff on exit. Tests log to replay-M7.log. A killed mutant gives a nonzero exit, as expected.

plant-selector.py reproduces the predeclared selector and original diffs; the submitted M2 standalone diff drops the whole exclusion statement, and submitted menu.json labels M6 supplemental. Do not overwrite the submitted evidence while replaying. scratch-*.txt preserve the exact switched sources used in the matrix.

For survivor demonstration, copy survivor-witness.go.txt to survivor-witness.go in this directory, apply M7.diff, run go run on that file, then revert the diff and delete the temporary Go file. The file is stored as .go.txt so normal repository package discovery does not compile a witness program.

Probe P1 should be replayed on each Parse row alone because nil AST inspection panics. P2 and P3 target TestSharedCanonicalize. Original probe logs already run each intended row separately.
