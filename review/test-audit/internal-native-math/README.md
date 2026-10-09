# u049 native math audit

Starting origin/main: 83f3940ec77b8ba779da05ded113e9d9e7e710b6. All five requested names exist and remain in the stated files. No families among these five: the random normalization body constructs a distinct generator and harness, rather than just passing an input index to runNormalizeUnit. The 18 TestNormalizeMatchesNodePoints/Contexts wrappers are one additional family.

Code under test: C runtime Math and number formatting, optional-number packing, Buffer and SHA-256, Unicode normalization. Oracle: Node results for Math/Buffer/random normalization; self-written NaN packing constants and long-measurement exit-status policy. No oracle, test, or harness was mutated.

The fixed plan has 15 menu mutants plus 15 separate empty-answer probes. plan.json records origin locations and menu operations. M07 changes the sign condition, including negative zero. switch.patch and plant.py preserve the runtime selector. Each standalone M/P diff applies to the starting commit and compiles with the runtime release clang flags; validation.json has exact commands and results. The test matrix also exercised the switched runtime in release and sanitizer configurations. All source changes are reverted before the evidence commit.

Bounded scope: full package baseline timed out at 90.097 test-binary seconds, with no preceding Test failure event. Functional matrix runs the four completed requested rows; normalization mutations and its probe additionally run every member of the reached normalization sweep family. Other package rows remain unknown. Coverage records the Go functions actually reached by the four completed rows. functions.json lists definitions in the six selected runtime files, including some helpers not reached by these rows. It is not a complete dynamic C call-graph proof; allocation/string support routines were not fully inventoried before planting. That part of the brief is incomplete.

The runtime archive is keyed by all embedded source/header bytes, flags and compiler version (internal/native/library.go). All runs use the same switched source and runtime environment selector; no stale archive can preserve the original behavior. ADAMIC_BUILD_CACHE_DIR is distinct for each matrix run and controls the normalization family's product cache, but the runtime archive itself uses os.UserCacheDir and the content key. The selector adds a getenv/strcmp call, so mutant timings are not clean performance measurements. The clean row medians use the original runtime, three separate -count=1 runs.

Survivors: M05 changes power(16,0.25) from 2 to 4; Node produces 2. M15 changes a second digest from a finalized-hash panic to successful length 64; Node reports ERR_CRYPTO_HASH_FINALIZED. witness.json and witness-*.log contain commands/output. Both are survivors only in the bounded matrix; repo/package-wide coverage is unknown.

Budget and brief issues:
* The supplied file/commit example names 8de93800f4, but the required fresh origin/main was 83f3940ec77b8ba779da05ded113e9d9e7e710b6. Locations use the fresh commit.
* A full native package run exceeds 90 seconds. Timeout is treated as a narrowing trigger, not a red baseline. No whole-package uniqueness is claimed.
* The opt-in long row is 610 serial measurements and cannot finish within the limit. One clean attempt and attempts with P09 and M12 each reached the 90-second timeout. The stop-on-cooked rule prevented two more clean attempts; seconds and vacuity are null. The driver records Node output but never compares it with native output or asserts performance bounds. Its first/last measurements use different units (bytes versus code points/UTF-16 units). Any nonzero exit on oversized cases satisfies the policy, including unrelated failures.
* The long probe's completed positive native measurements all returned zero exit status despite an empty normalization result. The oversized negative cases were not reached before timeout, so whole-row vacuity cannot be established.
* The long baseline began alone; later compile/matrix work overlapped its latter portion. Family timing runs overlapped other audit work. These are observed wall medians under that load, not isolated performance estimates.
* One initial switch control failed to compile because M07's declaration was scoped inside conditional branches. The corrected switch control passed before mutant runs. The error log is retained and its failures never count as kills.
* The separate survivor witness initially lacked a final C newline and failed compilation; it was corrected and rerun. The error log is retained.
* Mandatory npm ci is unrelated to the direct runtime rows but was run as requested. Warm toolchain setup was skipped.
* The five-row output schema does not specify how to represent a cooked row: this audit uses cannot-judge, over_budget true, null seconds/vacuity, and no timeout counted as a kill.

No other packages or repo-wide replay were run. No witness/setup/helper rows were found among the five requested names. No PR and no main push.
