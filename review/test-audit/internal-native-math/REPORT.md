u049: all five requested names present at origin/main 83f3940ec77b.
Bounded verdicts: 3 sacred, 1 subsumed on 3 mutants, 1 cannot-judge.
15 production mutants: 13 caught, 2 demonstrated survivors.
15 empty-answer probes caught by their completed owning rows; long row cooked.
Evidence branch: test-audit/internal-native-math; nproc 5.

```json
[
  {
    "test": "TestMathAndToFixedMatchJavaScript",
    "package": "internal/native",
    "file": "internal/native/math_test.go",
    "seconds": 0.94,
    "oracle": "Node Math.round/sign/max/min, exponentiation, Number.toFixed; exact answers including negative zero",
    "oracle_kind": "external-run",
    "kills": [
      "M01",
      "M02",
      "M03",
      "M04",
      "M06",
      "M07"
    ],
    "unique_kills": [
      "M01",
      "M02",
      "M03",
      "M04",
      "M06",
      "M07"
    ],
    "last_proven_fail": "M07 math_test.go:145: 10 of 69132 answers differ",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "P01",
      "P02",
      "P03",
      "P04",
      "P05",
      "P06"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMathAndToFixedMatchJavaScript",
      "TestMaybeNumbersPackIntoOneDouble",
      "TestNodeBufferRuntimeWithoutDeclarations",
      "TestNormalizeRandomMatchesNode"
    ],
    "evidence": "ADAMIC_MUTANT=M07 ADAMIC_BUILD_CACHE_DIR=/tmp/u049/cache/M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestMathAndToFixedMatchJavaScript|TestMaybeNumbersPackIntoOneDouble|TestNodeBufferRuntimeWithoutDeclarations|TestNormalizeRandomMatchesNode)$'; math_test.go:145: 10 of 69132 answers differ"
  },
  {
    "test": "TestMaybeNumbersPackIntoOneDouble",
    "package": "internal/native",
    "file": "internal/native/maybe_test.go",
    "seconds": 0.21,
    "oracle": "Self-written reserved and canonical NaN bits, presence flag and exact round-trip bits; Go math.IsNaN only classifies inputs",
    "oracle_kind": "self",
    "kills": [
      "M08",
      "M09",
      "M10"
    ],
    "unique_kills": [
      "M08",
      "M09",
      "M10"
    ],
    "last_proven_fail": "M10 maybe_test.go:100: 20015 of 20015 values pack wrong",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "P07",
      "P08"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMathAndToFixedMatchJavaScript",
      "TestMaybeNumbersPackIntoOneDouble",
      "TestNodeBufferRuntimeWithoutDeclarations",
      "TestNormalizeRandomMatchesNode"
    ],
    "evidence": "ADAMIC_MUTANT=M10 ADAMIC_BUILD_CACHE_DIR=/tmp/u049/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestMathAndToFixedMatchJavaScript|TestMaybeNumbersPackIntoOneDouble|TestNodeBufferRuntimeWithoutDeclarations|TestNormalizeRandomMatchesNode)$'; maybe_test.go:100: 20015 of 20015 values pack wrong"
  },
  {
    "test": "TestNodeBufferRuntimeWithoutDeclarations",
    "package": "internal/native",
    "file": "internal/native/node_buffer_test.go",
    "seconds": 0.85,
    "oracle": "Node v24.19.0 Buffer and crypto.createHash; exact stdout in release and sanitizer builds, plus JavaScript backend agreement",
    "oracle_kind": "external-run",
    "kills": [
      "M14"
    ],
    "unique_kills": [
      "M14"
    ],
    "last_proven_fail": "M14 --- FAIL: TestNodeBufferRuntimeWithoutDeclarations (0.59s)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "P10",
      "P11",
      "P12",
      "P13",
      "P14",
      "P15"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMathAndToFixedMatchJavaScript",
      "TestMaybeNumbersPackIntoOneDouble",
      "TestNodeBufferRuntimeWithoutDeclarations",
      "TestNormalizeRandomMatchesNode"
    ],
    "evidence": "ADAMIC_MUTANT=M14 ADAMIC_BUILD_CACHE_DIR=/tmp/u049/cache/M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestMathAndToFixedMatchJavaScript|TestMaybeNumbersPackIntoOneDouble|TestNodeBufferRuntimeWithoutDeclarations|TestNormalizeRandomMatchesNode)$'; --- FAIL: TestNodeBufferRuntimeWithoutDeclarations (0.59s)"
  },
  {
    "test": "TestNormalizeLongMeasurements",
    "package": "internal/native",
    "file": "internal/native/normalize_benchmark_test.go",
    "seconds": null,
    "oracle": "Self-written expected exit-status schedule. Node measurements are recorded but outputs and times are never compared; any nonzero oversized-case exit is accepted, including unrelated failures",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 1,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestNormalizeLongMeasurements"
    ],
    "evidence": "Opt-in clean baseline, P09 and M12 attempts timed out at 90 seconds. No completed functional failure observed.",
    "over_budget": true,
    "reason": "610 serial measurements cannot complete within 90 seconds; no completed mutation result or three-run median."
  },
  {
    "test": "TestNormalizeRandomMatchesNode",
    "package": "internal/native",
    "file": "internal/native/normalize_random_test.go",
    "seconds": 10.83,
    "oracle": "Node String.normalize in all four forms, exact WTF-8 hex for 100000 independently generated input strings and output line count",
    "oracle_kind": "external-run",
    "kills": [
      "M11",
      "M12",
      "M13"
    ],
    "unique_kills": [],
    "last_proven_fail": "M13 normalize_random_test.go:103: 94683 of 100000 lines differ",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestNormalizeMatchesNode family"
    ],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "P09"
    ],
    "subsumer_seconds": 26.401,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestMathAndToFixedMatchJavaScript",
      "TestMaybeNumbersPackIntoOneDouble",
      "TestNodeBufferRuntimeWithoutDeclarations",
      "TestNormalizeRandomMatchesNode",
      "TestNormalizeMatchesNode family"
    ],
    "evidence": "ADAMIC_MUTANT=M13 ADAMIC_BUILD_CACHE_DIR=/tmp/u049/cache/M13 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run '^(TestMathAndToFixedMatchJavaScript|TestMaybeNumbersPackIntoOneDouble|TestNodeBufferRuntimeWithoutDeclarations|TestNormalizeRandomMatchesNode)$'; normalize_random_test.go:103: 94683 of 100000 lines differ",
    "subsumption_mutants": 3
  }
]
```

| ID | Origin location | Change | Failed rows |
|---|---|---|---|
| M01 | internal/native/runtime/math.c:15 | `value - floored >= 0.5` -> `value - floored > 0.5` | TestMathAndToFixedMatchJavaScript |
| M02 | internal/native/runtime/math.c:27 | `value > 0 ? 1 : -1` -> `value < 0 ? 1 : -1` | TestMathAndToFixedMatchJavaScript |
| M03 | internal/native/runtime/math.c:35 | `signbit(left) && signbit(right)` -> `signbit(left) \|\| signbit(right)` | TestMathAndToFixedMatchJavaScript |
| M04 | internal/native/runtime/math.c:47 | `left < right ? left : right` -> `left > right ? left : right` | TestMathAndToFixedMatchJavaScript |
| M05 | internal/native/runtime/number.c:117 | `exponent == 0.5` -> `exponent == 0.25` |  |
| M06 | internal/native/runtime/number.c:152 | `exact[keep] >= '5'` -> `exact[keep] > '5'` | TestMathAndToFixedMatchJavaScript |
| M07 | internal/native/runtime/number.c:144 | `bool negative = value < 0;` -> `bool negative = signbit(value);` | TestMathAndToFixedMatchJavaScript |
| M08 | internal/native/runtime/maybe.c:34 | `bits = 0x7ff8000000000000u;` -> `bits = 0x7ff8000000000001u;` | TestMaybeNumbersPackIntoOneDouble |
| M09 | internal/native/runtime/maybe.c:29 | `if (value.present) {` -> `if (!value.present) {` | TestMaybeNumbersPackIntoOneDouble |
| M10 | internal/native/runtime/maybe.c:44 | `bits == ADAMIC_UNDEFINED_BITS` -> `bits != ADAMIC_UNDEFINED_BITS` | TestMaybeNumbersPackIntoOneDouble |
| M11 | internal/native/runtime/normalize.c:165 | `if (mapping == NULL \|\| (mapping->compatibility && !compatibility)) {` -> `if (mapping == NULL \|\| (mapping->compatibility && compatibility)) {` | TestNormalizeRandomMatchesNode, TestNormalizeMatchesNode family |
| M12 | internal/native/runtime/normalize.c:574 | `compatibility = true, composed = true;` -> `compatibility = false, composed = true;` | TestNormalizeRandomMatchesNode, TestNormalizeMatchesNode family |
| M13 | internal/native/runtime/normalize.c:572 | `compatibility = false, composed = false;` -> `compatibility = false, composed = true;` | TestNormalizeRandomMatchesNode, TestNormalizeMatchesNode family |
| M14 | internal/native/runtime/node_buffer.c:132 | `return (int)(c - 'a') + 26;` -> `return (int)(c - 'a') + 25;` | TestNodeBufferRuntimeWithoutDeclarations |
| M15 | internal/native/runtime/node_crypto.c:93 | `hash->slots[1].number = 1;` -> `/* dropped finalized flag assignment */` |  |

M05 survivor: power(16,0.25) control 2, mutant 4, Node 2.
M15 survivor: second digest control finalized panic, mutant length 64, Node ERR_CRYPTO_HASH_FINALIZED.

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

Timing evidence: {"origin_commit": "83f3940ec77b8ba779da05ded113e9d9e7e710b6", "nproc": 5, "setup_seconds": 0, "npm_seconds": 0.5594161699991673, "row_test_seconds": {"TestMathAndToFixedMatchJavaScript": 0.94, "TestMaybeNumbersPackIntoOneDouble": 0.21, "TestNodeBufferRuntimeWithoutDeclarations": 0.85, "TestNormalizeLongMeasurements": null, "TestNormalizeRandomMatchesNode": 10.83}, "subsumer_seconds": 26.401, "runtime_flags": "runtime native.Flags: release O2; sanitized O1 with address/undefined sanitizers", "matrix_wall_seconds": 590.9577626840146, "family_wall_seconds": 204.85271305199058, "validation_wall_seconds": 2.0082173779956065, "timing_wall_seconds": 158.8184005030016, "long_attempt_wall_seconds": 185.56182489199637, "final_checks": [{"label": "restored", "command": ["go", "test", "-json", "-count=1", "-timeout", "90s", "./internal/native/", "-run", "^(TestMathAndToFixedMatchJavaScript|TestMaybeNumbersPackIntoOneDouble|TestNodeBufferRuntimeWithoutDeclarations|TestNormalizeRandomMatchesNode)$"], "exit": 0, "wall": 12.904799391999404}, {"label": "vet", "command": ["go", "vet", "./internal/native/"], "exit": 0, "wall": 0.3982554340000206}, {"label": "test-binary-build", "command": ["go", "test", "-c", "-o", "/tmp/u049/restored.test", "./internal/native/"], "exit": 0, "wall": 2.327138731998275}]}
