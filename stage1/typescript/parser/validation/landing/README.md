The landing report names every final command and result. Raw logs preserve trailing tabs in canonical tree records. Only this directory's `.log` data has Git's blank-at-end-of-line whitespace check disabled.

Instruction evidence uses Callgrind's `Ir` event over one process parsing and counting all 77 compiler files. The same area compiler, release flags and manifest were used for every branch snapshot. It includes process startup, scanning, parsing, allocation and the node-count traversal. It excludes canonical tree printing because the driver runs with `--whole --count`.

The common-base, area and recovery snapshots are recorded in `measurements.json`; its initial landing row is merge commit cd958cf, before the scanner-message correction. `final-measurements.json` records the committed landing and actual extra-parse mutant. Compressed profiles retain the raw event records; `stage1/typescript/scanner/profile.py` verifies that their self costs sum to the summary.

To replay the final measurement, prepare an artifact directory containing decompressed `base.callgrind`, `area.callgrind` and `recovery.callgrind`, the pinned 77-path `compiler.manifest`, and the fresh Go oracle's `go-tree.stdout`. Then source the tool environment and run:

```sh
VALGRIND_LIB=/path/to/valgrind/libexec/valgrind python3 stage1/typescript/parser/validation/landing/measure.py /path/to/repo /path/to/artifacts --valgrind /path/to/valgrind
```

The replay rebuilds both programs from source with the repository's native default release flags, compares 44,766,682 tree bytes with Go, checks 887,803 nodes, measures both programs and asserts that only the actual extra parse breaches the recovery budget. Child output goes directly to artifact files. It requires the Go tree output and does not substitute hashes for byte comparisons.

`lint-recovery-contract.diff` is the only scratch overlay change used to compare all captured recovered findings positively. The checked-in lint harness is byte-for-byte the area version. Its old refusal assertion fails as recorded in `lint-oracle.log`; the overlay is not used to claim that unchanged test passed.
