# Wave 23 landing evidence

The completed upstream matrices and their raw sources, captured options, command outputs, checker transcripts, elapsed times and hashes are in `isolated-cases.tar.gz` and `max-nested-cases.tar.gz`. Native is the full gate's ASan/UBSan build (its ELF defines `__asan_init`); binary and emitted-module hashes are in `isolated-build-hashes.json`. Binaries are build products, not committed sources. Build them from this branch using the ordinary unchanged harness.

`landing-jsx-inventory.patch.gz` is an unapplied, formatted two-row proposal for integration. The shared guard and all old counts are unchanged in this branch.

The first complete retained wave-23 suite run found stale listener metadata; every rule suite passed. `landing-listeners-final.log` records the corrected metadata oracle and all eighteen caught mutants. The initial full lint attempt was stopped after a nominal scratch-copy bug; its events are kept in `landing-lint-before-copy-fix.jsonl.gz`. `landing-copy-fix.log` proves the owned adapter relocation fixed the cache-byte control. These earlier logs are observations, not claimed full-gate passes.

The final full-package events and result summary are added after the all-input run finishes. Inputs and commands are in the detached runner. Nothing is filtered out of that full package run.

Final package: 123 pass / 3 fail / 1 skip including subtests, all 88 registered mutants caught, no input skips. See landing-lint-result.json and landing-lint-complete.jsonl.gz. Every failing or blocked case is retained; this is not a green full-package claim.
