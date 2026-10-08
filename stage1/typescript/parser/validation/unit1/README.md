# Unit 1 evidence

`../../UNIT1_LANDING_REPORT.md` explains the change and coverage. Every test writes directly to its `.log`; canonical byte comparisons retain their trailing tabs. `coverage.json` records pinned refs, flags, source pin and tree hash. `compiler-files.txt` lists all 77 paths relative to the TypeScript checkout.

The three controlled profiles are `base.callgrind.gz` (area parser db2ecc004), `before.callgrind.gz` (landing parser daf277b01), and `after-final.callgrind.gz` (the committed fast path). Every binary is built with the current merged compiler and the same native release options. `extra-parse-final.callgrind.gz` is a real parse-work mutant, not a fabricated count. `final-measurements.json` gives the instruction totals and node counts. `profile-hashes.json` hashes the uncompressed profiles. `function-costs.json` ranks normalized self-cost differences; `top-function-costs.json` contains the three largest total self costs and three largest added self costs with base/before/after columns.

`measure.py` replays the final source and extra-parse mutant. Prepare a writable scratch artifact directory with decompressed `base.callgrind` and `before.callgrind`, a `compiler.manifest` containing absolute paths to the pinned checkout's 77 sources, and fresh external-Go `go-tree.stdout` for `--manifest <manifest> --whole`. The Go protocol adapter is `../../testdata/oracle.go`, built inside cohere's pinned `TypeScript/tsc` module using the ordinary overlay described in `../../parser_test.go:goOracle`. The typescript-go parser itself is unmodified.

Then source the tool environment and run the replay, directing all output to a log:

```sh
VALGRIND_LIB=/path/to/valgrind/libexec/valgrind python3 stage1/typescript/parser/validation/unit1/measure.py /path/to/repo /path/to/artifacts --valgrind /path/to/valgrind > /path/to/artifacts/measurement.log 2>&1
```

The replay builds current source and an actual extra-parse mutant, compares their 44,766,682 tree bytes to Go, checks 887,803 nodes, measures `--whole --count`, and requires the current parser to cost less than the original landing while the mutant exceeds that budget. It also corrupts the profile summary by one instruction and requires the accounting reader to reject it.

To rebuild the controls rather than consume recorded profiles, materialize only the tracked parser/scanner `.ts` and `.a` files from the exact `coverage.json` refs under two separate scratch roots. Build each root's `stage1/typescript/parser/main.ts` with the same current `go run ./cmd/adamic build <entry> -o <binary>` command. Byte-compare `--manifest <manifest> --whole` output with Go before collecting `valgrind --tool=callgrind --callgrind-out-file=<control>.callgrind <binary> --manifest <manifest> --whole --count`. No older compiler binary is used for a control.

`area-incomplete.log` and `incomplete.log` are failing gates, retained without filtering or deadline changes. `case1778-results.json` records a separate direct diagnostic probe which terminates and matches all three runtimes. This is not a claim that the entire malformed-input corpus passes. The retained large source lives at the scratch path cited by `incomplete.log`; new raw minimal witnesses are under `../../testdata/recovery/`.
