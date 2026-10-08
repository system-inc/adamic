# Destructuring lowering evidence

Base: area/compiler b410340dc8f889b5799c3bc519117c63def3aa24. Replay merged at 9a1f14c5. Checked non-null merged at c41c0e06 (merge ebfbfe43).

## Computed field names (6 roots)

Lowered literal string concatenations and safe integral const enum keys. Runtime enum keys retain evaluation before the value. Computed __proto__ remains an own field in both backends. Dynamic keys and numeric concatenation remain explicit NotYet.

Replays use the exact adapted source manifest from stage3-latent-full 9d534d3a31814f1a192a528e701f6c2ea7c910bc (81/81 matching hashes). Evidence: computed-replay.json. Parser reaches a generic function expression; scanner and sys report no next finding; diagnostics reaches a census statement panic on ComputedPropertyName; visitor reaches its next structural-method stop. These are measurement results on a checker-rejected compiler corpus, not executable compiler builds.

Validation: three fixtures run through TestNativeAgreesWithNode with ADAMIC_GATE_UNCACHED=1 (source Node, native ASan/UBSan/leaks, emitted JavaScript Node), passed after non-null merge in 2.689s. Lower boundary tests passed 0.196s; JavaScript package has no tests. TestCountsAreRecorded -update-counts passed 209.148s and added three rows.

Mutants: string-key, enum-key, binding-key, runtime-key-evaluation, own-proto each caught by stdout differs. singleton-call and numeric-concatenation each caught by the lower boundary test. Run internal/oracle/testdata/run-notyet-computed-mutants.py with the toolchain environment sourced. Build failures were rejected and are not counted as mutant kills. Logs: /tmp/destructuring-computed-mutants/.

Toolchain setup: node .409s; Go .393s; clang 2.087s; markdown 4.299s; submodules 62.639s; Go build 2339.790s; deferred test binaries 2340.484s; cache warm 2340.491s; done 2340.731s. nproc 5, quota 4 CPUs. Setup log /tmp/destructuring-setup.log.
