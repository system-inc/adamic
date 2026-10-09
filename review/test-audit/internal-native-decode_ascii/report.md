Starting commit: b902a0ccc09e97940571388a1634450da3383559; all 94 names present.
94 functions group into five rows; none moved or vanished.
Clean scoped baseline: 76.942s; whole package exceeded 90s.
12 mutants and six probes; bounded uniqueness only.
Evidence branch: test-audit/internal-native-decode_ascii; nproc 5.

```json
[
  {
    "test": "TestDecodeASCII family",
    "package": "internal/native",
    "file": "internal/native/decode_ascii_test.go",
    "seconds": 71.027,
    "oracle": "Node Buffer UTF-8 output, independent decoder snapshot, self-written metadata, indexing expectations and completed-partition counts",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M01",
      "M02",
      "M03",
      "M04"
    ],
    "unique_kills": [
      "M01",
      "M02",
      "M03",
      "M04"
    ],
    "last_proven_fail": "M04: decode cache mismatch record=0 run=0 offset=0 ascii=0 units=4 bytes=3",
    "verdict": "slow-worthy",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestDecodeASCII family"
    ],
    "evidence": "ADAMIC_MUTANT=4 ADAMIC_BUILD_CACHE_DIR=/tmp/u046/cache/M04-native timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^TestDecodeASCIIUnit[0-9]+$ ; ADAMIC_MUTANT=4 ADAMIC_BUILD_CACHE_DIR=/tmp/u046/cache/M04-wasi timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^TestDecodeASCIIWASIUnit[0-9]+$ => decode cache mismatch record=0 run=0 offset=0 ascii=0 units=4 bytes=3",
    "members": [
      "TestDecodeASCIIUnit00",
      "TestDecodeASCIIUnit01",
      "TestDecodeASCIIUnit02",
      "TestDecodeASCIIUnit03",
      "TestDecodeASCIIUnit04",
      "TestDecodeASCIIUnit05",
      "TestDecodeASCIIUnit06",
      "TestDecodeASCIIUnit07",
      "TestDecodeASCIIUnit08",
      "TestDecodeASCIIUnit09",
      "TestDecodeASCIIUnit10",
      "TestDecodeASCIIUnit11",
      "TestDecodeASCIIUnit12",
      "TestDecodeASCIIUnit13",
      "TestDecodeASCIIUnit14",
      "TestDecodeASCIIUnit15",
      "TestDecodeASCIIUnit16",
      "TestDecodeASCIIUnit17",
      "TestDecodeASCIIUnit18",
      "TestDecodeASCIIUnit19",
      "TestDecodeASCIIUnit20",
      "TestDecodeASCIIUnit21",
      "TestDecodeASCIIUnit22",
      "TestDecodeASCIIUnit23",
      "TestDecodeASCIIUnit24",
      "TestDecodeASCIIUnit25",
      "TestDecodeASCIIUnit26",
      "TestDecodeASCIIUnit27",
      "TestDecodeASCIIUnit28",
      "TestDecodeASCIIUnit29",
      "TestDecodeASCIIUnit30",
      "TestDecodeASCIIUnit31",
      "TestDecodeASCIIUnit32",
      "TestDecodeASCIIUnit33",
      "TestDecodeASCIIUnit34",
      "TestDecodeASCIIUnit35",
      "TestDecodeASCIIUnit36",
      "TestDecodeASCIIUnit37",
      "TestDecodeASCIIUnit38",
      "TestDecodeASCIIUnit39",
      "TestDecodeASCIIUnit40",
      "TestDecodeASCIIUnit41",
      "TestDecodeASCIIUnit42",
      "TestDecodeASCIIUnit43",
      "TestDecodeASCIIUnit44",
      "TestDecodeASCIIWASIUnit00",
      "TestDecodeASCIIWASIUnit01",
      "TestDecodeASCIIWASIUnit02",
      "TestDecodeASCIIWASIUnit03",
      "TestDecodeASCIIWASIUnit04",
      "TestDecodeASCIIWASIUnit05",
      "TestDecodeASCIIWASIUnit06",
      "TestDecodeASCIIWASIUnit07",
      "TestDecodeASCIIWASIUnit08",
      "TestDecodeASCIIWASIUnit09",
      "TestDecodeASCIIWASIUnit10",
      "TestDecodeASCIIWASIUnit11",
      "TestDecodeASCIIWASIUnit12",
      "TestDecodeASCIIWASIUnit13",
      "TestDecodeASCIIWASIUnit14",
      "TestDecodeASCIIWASIUnit15",
      "TestDecodeASCIIWASIUnit16",
      "TestDecodeASCIIWASIUnit17",
      "TestDecodeASCIIWASIUnit18",
      "TestDecodeASCIIWASIUnit19",
      "TestDecodeASCIIWASIUnit20",
      "TestDecodeASCIIWASIUnit21",
      "TestDecodeASCIIWASIUnit22",
      "TestDecodeASCIIWASIUnit23",
      "TestDecodeASCIIWASIUnit24",
      "TestDecodeASCIIWASIUnit25",
      "TestDecodeASCIIWASIUnit26",
      "TestDecodeASCIIWASIUnit27",
      "TestDecodeASCIIWASIUnit28",
      "TestDecodeASCIIWASIUnit29",
      "TestDecodeASCIIWASIUnit30",
      "TestDecodeASCIIWASIUnit31",
      "TestDecodeASCIIWASIUnit32",
      "TestDecodeASCIIWASIUnit33",
      "TestDecodeASCIIWASIUnit34",
      "TestDecodeASCIIWASIUnit35",
      "TestDecodeASCIIWASIUnit36",
      "TestDecodeASCIIWASIUnit37",
      "TestDecodeASCIIWASIUnit38",
      "TestDecodeASCIIWASIUnit39",
      "TestDecodeASCIIWASIUnit40",
      "TestDecodeASCIIWASIUnit41",
      "TestDecodeASCIIWASIUnit42",
      "TestDecodeASCIIWASIUnit43",
      "TestDecodeASCIIWASIUnit44"
    ]
  },
  {
    "test": "TestDevirtualizedCalls",
    "package": "internal/native",
    "file": "internal/native/devirtualize_test.go",
    "seconds": 0.049,
    "oracle": "Self-written C fragments and call-case counts; expected adapter also uses production exactReceiverMethod, weakening independence",
    "oracle_kind": "self",
    "kills": [
      "M09",
      "M10",
      "M11",
      "M12"
    ],
    "unique_kills": [
      "M09",
      "M11",
      "M12"
    ],
    "last_proven_fail": "M12: devirtualize_test.go:66: interface receiver lost its exact allocation",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P02",
      "P03",
      "P04"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestDevirtualizedCalls",
      "TestExactReceiverRejectsAssignments"
    ],
    "evidence": "ADAMIC_MUTANT=12 ADAMIC_BUILD_CACHE_DIR=/tmp/u046/cache/M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestDevirtualizedCalls|TestExactReceiverRejectsAssignments)$ => devirtualize_test.go:66: interface receiver lost its exact allocation"
  },
  {
    "test": "TestExactReceiverRejectsAssignments",
    "package": "internal/native",
    "file": "internal/native/devirtualize_test.go",
    "seconds": 0.009,
    "oracle": "Self-written zero class identity for an assigned synthetic receiver",
    "oracle_kind": "self",
    "kills": [
      "M10"
    ],
    "unique_kills": [],
    "last_proven_fail": "M10: devirtualize_test.go:92: assigned receiver taken as exact: 1",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestDevirtualizedCalls"
    ],
    "mutants_in_matrix": 4,
    "probe_kills": [],
    "subsumer_seconds": 0.049,
    "vacuous": true,
    "bounded": true,
    "matrix_rows": [
      "TestDevirtualizedCalls",
      "TestExactReceiverRejectsAssignments"
    ],
    "evidence": "ADAMIC_MUTANT=10 ADAMIC_BUILD_CACHE_DIR=/tmp/u046/cache/M10 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^(TestDevirtualizedCalls|TestExactReceiverRejectsAssignments)$ => devirtualize_test.go:92: assigned receiver taken as exact: 1"
  },
  {
    "test": "TestToExponentialAndToPrecisionMatchNode",
    "package": "internal/native",
    "file": "internal/native/dtoa_test.go",
    "seconds": 5.57,
    "oracle": "Node toExponential/toPrecision, all answers compared byte for byte",
    "oracle_kind": "external-run",
    "kills": [
      "M05",
      "M06"
    ],
    "unique_kills": [
      "M06"
    ],
    "last_proven_fail": "M06: dtoa_test.go:149: precision 1 3ff0000000000000 3eb0c6f7a0b5ed8d: native \"1e-6\", Node \"0.000001\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P05",
      "P06"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToExponentialAndToPrecisionMatchNode",
      "TestToExponentialAndToPrecisionOutOfRangePanic"
    ],
    "evidence": "ADAMIC_MUTANT=6 ADAMIC_BUILD_CACHE_DIR=/tmp/u046/cache/M06 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^TestToExponentialAndToPrecision => dtoa_test.go:149: precision 1 3ff0000000000000 3eb0c6f7a0b5ed8d: native \"1e-6\", Node \"0.000001\""
  },
  {
    "test": "TestToExponentialAndToPrecisionOutOfRangePanic",
    "package": "internal/native",
    "file": "internal/native/dtoa_test.go",
    "seconds": 0.593,
    "oracle": "Node stdout and RangeError text; self-written exit 70 and panic expectations from docs/0.1.md; stderr checked by prefix",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M05",
      "M07",
      "M08"
    ],
    "unique_kills": [
      "M07",
      "M08"
    ],
    "last_proven_fail": "M08: See raw log for failure diagnostic",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 4,
    "probe_kills": [
      "P05",
      "P06"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestToExponentialAndToPrecisionMatchNode",
      "TestToExponentialAndToPrecisionOutOfRangePanic"
    ],
    "evidence": "ADAMIC_MUTANT=8 ADAMIC_BUILD_CACHE_DIR=/tmp/u046/cache/M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/native/ -run ^TestToExponentialAndToPrecision => See raw log for failure diagnostic"
  }
]
```

| ID | origin/main file:line | Change | Failed grouped rows |
|---|---|---|---|
| M01 | internal/native/runtime/input.c:47 | `lead >= 0xc2 -> lead >= 0xc3` | TestDecodeASCII family |
| M02 | internal/native/runtime/input.c:56 | `upper = 0x9f; -> upper = 0xbf;` | TestDecodeASCII family |
| M03 | internal/native/runtime/input.c:129 | `string->units = length + 1; -> string->units = length + 2;` | TestDecodeASCII family |
| M04 | internal/native/runtime/input.c:111 | `bytes[offset] < 0x80 -> bytes[offset] <= 0x80` | TestDecodeASCII family |
| M05 | internal/native/runtime/dtoa.c:975 | `int exponent = decimal_point - 1; -> int exponent = decimal_point;` | TestToExponentialAndToPrecisionMatchNode, TestToExponentialAndToPrecisionOutOfRangePanic |
| M06 | internal/native/runtime/dtoa.c:992 | `exponent < -6 -> exponent < -5` | TestToExponentialAndToPrecisionMatchNode |
| M07 | internal/native/runtime/dtoa.c:1054 | `digits < 0 || -> digits < -1 ||` | TestToExponentialAndToPrecisionOutOfRangePanic |
| M08 | internal/native/runtime/dtoa.c:1072 | `digits < 1 || -> digits < 0 ||` | TestToExponentialAndToPrecisionOutOfRangePanic |
| M09 | internal/native/devirtualize.go:18 | `!class.Static && -> class.Static &&` | TestDevirtualizedCalls |
| M10 | internal/native/devirtualize.go:49 | `declarations == 1 && !written -> declarations == 1 && written` | TestDevirtualizedCalls, TestExactReceiverRejectsAssignments |
| M11 | internal/native/class_inheritance.go:48 | `len(targets) == 1 -> len(targets) == 0` | TestDevirtualizedCalls |
| M12 | internal/native/devirtualize.go:68 | `method.Name == name -> method.Name != name` | TestDevirtualizedCalls |

Survivors: none in the bounded matrix.

The brief says six rows but lists 94 functions. All 90 decoder wrappers call the same checker with only target and index inputs. The family rule therefore produces five rows. Treating targets as separate rows would split that shared checker. members.json records every member.

The brief cites 8de93800f4; fetched origin/main was b902a0cc. All names were checked through go test -list and remain in the cited files. All diff locations refer to fetched main.

The whole-package clean baseline timed out at 90.024s without an earlier reported failure. The scoped clean baseline passed in 76.942s with WASI enabled. The combined inactive switch baseline also timed out at 90.017s without an earlier reported failure. Splitting decoder execution by target gave clean inactive runs of 58.791s native and 18.792s WASI. Matrix parts observe every member and count as one family. Tests outside the recorded bounded sets remain unknown. No other package was tested.

Warm env.sh lacked WASI SDK. SDK 27 was installed and ADAMIC_TEST_WASI=1 enabled for scoped baseline, costs and matrices. The initial whole-package baseline skipped WASI. No scoped row skipped after installation. Core setup was skipped, zero seconds. SDK installation and npm ci were not timed separately.

ADAMIC_BUILD_CACHE_DIR controls the decoder cache. library.go's runtimeLibrary instead uses a content-hashed user cache. Each selector run has its own decoder cache directory, seeded with hard links to immutable products from the identical switched source. The C selector is read per process. This reuses a valid switched product, not a stale product from before mutation. The emission rows inspect generated text and do not compile it into native products.

The first cost run for each row included coverage instrumentation; all three used -count=1 and the binary's reported package seconds. This limits comparisons to purely uninstrumented timings. The decode family median measures all 90 members, not the near-zero scheduling time of a parallel wrapper.

The Go reach inventory contains 142 covered function-location records. The C inventory names decoder and formatting entries and routines, but transitive allocation, output, startup and string helpers were not exhaustively instrumented. Those paths were not targeted. Twelve code-derived mutants, four per area, fall below the aspirational three per grouped row. The budget favored observing all family members and validating each standalone diff.

No test or oracle was changed. The receiver assignment row passes its own empty-answer probe because zero is its expected negative result. This is recorded as vacuous, separately from its observed production kill. Subsumption is a hint from the recorded mutant set, not grounds for deletion. Expected adapter computation in the devirtualization row shares production exactReceiverMethod, limiting oracle independence.

No package-wide or repository-wide uniqueness is claimed. Central replay must settle that. No PR was opened and main was not pushed. Baseline, timing, matrix, probe, compile and replay-check logs are retained with the standalone diffs.

Matrix and probe wall total: 268.418s. Timing runs: 246.555 binary seconds. Individual compile times and inactive build phases are in driver.log and inactive logs.
