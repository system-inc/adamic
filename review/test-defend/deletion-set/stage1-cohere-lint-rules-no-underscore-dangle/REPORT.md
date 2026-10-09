Deletion-set replay at b8bcadb2c493173855f19d7e5c508b34f5eeb5b6.

Seven production diffs were inspected: audit M1-M4 and defender D1-D3. All recorded candidate failure lists were empty, so none qualified for replay. The eligible list was recorded before the baseline in mutant-list.json. P1 is a probe and was excluded. No diffs were applied, rebased or evaluated for staleness; no mutant tests, witness-failure classifications or panic retries were necessary.

The exact required uncached baseline skipped TestCompileProfiles, using ADAMIC_GATE_UNCACHED=1 and fresh ADAMIC_BUILD_CACHE_DIR=/tmp/deletion-no-underscore-dangle/cache/baseline. It passed in 1.783 command wall seconds, 0.007 seconds on the binary ok line (JSON package elapsed 0.008). Current listing contains only TestCompileProfiles; after skipping it the binary reports no tests to run. No demonstrated mutant catch is lost by this set. The compilation-only gate itself disappears, but this experiment establishes only the requested finite-evidence deletion criterion, not that the row has no possible value.

Main was detached, tracked production clean, submodules updated at their pins. Warm env.sh worked; setup skipped, nproc 5, Go 1.27.1. Earlier /tmp scratch/cache directories were removed after the disk check. /tmp totals 8.8 GB, so 15 GB free is impossible; /workspace stayed at 14 GB free. No full-disk failure occurred. Unrelated prior untracked evidence was preserved. Only this evidence directory is committed and pushed. Production was never mutated and the detached main tree is restored after publication.

Audit and defender branch SHAs and original recorded failures are retained in mutant-list.json and the two source-results files. No other package was tested and no new mutations were invented.
