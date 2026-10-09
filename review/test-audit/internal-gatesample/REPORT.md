u021: five rows, all sacred within the whole-package matrix.
Base: 7b18d0576930caca4e22ce2eef92fcf563af52d0; baseline passed, no skips.
14 fixed-menu mutants plus supplemental M04; one survivor, M15.
Primary empty-answer probes caught by all rows; no row vacuous.
Evidence: test-audit/internal-gatesample, review/test-audit/internal-gatesample/.

```json
[
  {
    "test": "TestSelection",
    "package": "github.com/system-inc/adamic/internal/gatesample",
    "file": "internal/gatesample/sample_test.go:13",
    "seconds": 0.003,
    "oracle": "Self-written exact paths, offsets, log text, determinism and rotation count; malformed-switch rejection only asserts a non-nil error, not its diagnostic.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M04",
      "M05",
      "M08",
      "M09",
      "M10",
      "M11"
    ],
    "unique_kills": [
      "M01",
      "M05",
      "M08",
      "M09",
      "M10",
      "M11"
    ],
    "last_proven_fail": "M11 sample_test.go:119: accepted malformed switch \"gggggggggggggggggggggggggggggggggggggggg\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "PSelect",
      "PLog",
      "PValidate"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestSelection",
      "TestUnsetIsWholeCorpus",
      "TestBadChangedList",
      "TestHexOffset",
      "TestGeneratedSelection"
    ],
    "evidence": "ADAMIC_MUTANT=M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/gatesample/ -run . > M11.log 2>&1; sample_test.go:119: accepted malformed switch \"gggggggggggggggggggggggggggggggggggggggg\""
  },
  {
    "test": "TestUnsetIsWholeCorpus",
    "package": "github.com/system-inc/adamic/internal/gatesample",
    "file": "internal/gatesample/sample_test.go:125",
    "seconds": 0.002,
    "oracle": "Self-written full-corpus order and Sample=false expectation.",
    "oracle_kind": "self",
    "kills": [
      "M02"
    ],
    "unique_kills": [
      "M02"
    ],
    "last_proven_fail": "M02 sample_test.go:144: unset changed full corpus: {Paths:[] Sample:false Total:2 Stride:3 Offset:0}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "PSelect"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestSelection",
      "TestUnsetIsWholeCorpus",
      "TestBadChangedList",
      "TestHexOffset",
      "TestGeneratedSelection"
    ],
    "evidence": "ADAMIC_MUTANT=M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/gatesample/ -run . > M02.log 2>&1; sample_test.go:144: unset changed full corpus: {Paths:[] Sample:false Total:2 Stride:3 Offset:0}"
  },
  {
    "test": "TestBadChangedList",
    "package": "github.com/system-inc/adamic/internal/gatesample",
    "file": "internal/gatesample/sample_test.go:149",
    "seconds": 0.004,
    "oracle": "Self-written rejection expectations, checking only err != nil. A different error can satisfy the oracle.",
    "oracle_kind": "self",
    "kills": [
      "M06",
      "M07"
    ],
    "unique_kills": [
      "M06",
      "M07"
    ],
    "last_proven_fail": "M07 sample_test.go:163: accepted non-repository path \"../outside.md\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "PSelect"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestSelection",
      "TestUnsetIsWholeCorpus",
      "TestBadChangedList",
      "TestHexOffset",
      "TestGeneratedSelection"
    ],
    "evidence": "ADAMIC_MUTANT=M07 timeout 120 go test -json -count=1 -timeout 90s ./internal/gatesample/ -run . > M07.log 2>&1; sample_test.go:163: accepted non-repository path \"../outside.md\""
  },
  {
    "test": "TestHexOffset",
    "package": "github.com/system-inc/adamic/internal/gatesample",
    "file": "internal/gatesample/sample_test.go:169",
    "seconds": 0.003,
    "oracle": "Self-written hexadecimal offset 1 and exact selected path b.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M04"
    ],
    "unique_kills": [
      "M03"
    ],
    "last_proven_fail": "M03 sample_test.go:178: offset must parse hexadecimal: {Paths:[a] Sample:true Total:3 Stride:3 Offset:0}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "PSelect"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestSelection",
      "TestUnsetIsWholeCorpus",
      "TestBadChangedList",
      "TestHexOffset",
      "TestGeneratedSelection"
    ],
    "evidence": "ADAMIC_MUTANT=M03 timeout 120 go test -json -count=1 -timeout 90s ./internal/gatesample/ -run . > M03.log 2>&1; sample_test.go:178: offset must parse hexadecimal: {Paths:[a] Sample:true Total:3 Stride:3 Offset:0}"
  },
  {
    "test": "TestGeneratedSelection",
    "package": "github.com/system-inc/adamic/internal/gatesample",
    "file": "internal/gatesample/sample_test.go:183",
    "seconds": 0.002,
    "oracle": "Self-written indices/count/rotation. Expected keys call the same GeneratedKey implementation as production: M15 and PKey pass despite changed or empty key content.",
    "oracle_kind": "self",
    "kills": [
      "M12",
      "M13",
      "M14"
    ],
    "unique_kills": [
      "M12",
      "M13",
      "M14"
    ],
    "last_proven_fail": "M14 sample_test.go:188: indexed selection: gatesample.Selection{Paths:[]string{\"generated-index/0\", \"generated-index/1\", \"generated-index/4\"}, Sample:true, Total:8, Stride:3, Offset:1}",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 15,
    "probe_kills": [
      "PGenerated"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestSelection",
      "TestUnsetIsWholeCorpus",
      "TestBadChangedList",
      "TestHexOffset",
      "TestGeneratedSelection"
    ],
    "evidence": "ADAMIC_MUTANT=M14 timeout 120 go test -json -count=1 -timeout 90s ./internal/gatesample/ -run . > M14.log 2>&1; sample_test.go:188: indexed selection: gatesample.Selection{Paths:[]string{\"generated-index/0\", \"generated-index/1\", \"generated-index/4\"}, Sample:true, Total:8, Stride:3, Offset:1}"
  }
]
```

Code under test, declared before mutation: Select, Validate, Selection.Log, Selection.Generated, GeneratedKey, all in internal/gatesample/sample.go. TestSelection reaches Select, Validate, Log; TestUnsetIsWholeCorpus reaches Select; TestBadChangedList and TestHexOffset reach Select and Validate; TestGeneratedSelection reaches Generated and GeneratedKey. Oracle: self-written expected answers; no external authority or external process compares results. This is the selector implementation, not a port or native compiler. No native rebuild or per-mutant native cache applies.

All file:line references use the base commit. Matrix command for each id: `source /workspace/adamic-tools/env.sh; ADAMIC_MUTANT=<id> timeout 120 go test -json -count=1 -timeout 90s ./internal/gatesample/ -run . > review/test-audit/internal-gatesample/<id>.log 2>&1`. Timing command, three separate count=1 runs: `go test -count=1 -timeout 90s ./internal/gatesample/ -run ^<row>$ > <row>-<round>.log 2>&1`. Standalone verification: restore base source, `git apply --check diffs/<id>.diff`, `git apply diffs/<id>.diff`, `go vet ./internal/gatesample/ > <id>-vet.log 2>&1`, restore base. Every diff applied and every vet exited zero. No other package test was run.

| Id | Origin file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/gatesample/sample.go:25 | len(s.Paths), s.Total -> s.Total, len(s.Paths) | TestSelection |
| M02 | internal/gatesample/sample.go:34 | Paths: append([]string(nil), paths...) -> Paths: nil | TestUnsetIsWholeCorpus |
| M03 | internal/gatesample/sample.go:44 | strconv.ParseUint(sha[:8], 16, 32) -> strconv.ParseUint(sha[:8], 10, 32) | TestHexOffset |
| M04 supplemental | internal/gatesample/sample.go:45 | int(prefix % uint64(stride)) -> int(prefix % uint64(stride)) * 0 | TestSelection, TestHexOffset |
| M05 | internal/gatesample/sample.go:48 | changed[name] = true -> changed[name] = false | TestSelection |
| M06 | internal/gatesample/sample.go:54 | return Selection{}, fmt.Errorf("ADAMIC_GATE_CHANGED: %w", err) -> return Selection{}, nil | TestBadChangedList |
| M07 | internal/gatesample/sample.go:61 | if filepath.IsAbs(line) \|\| strings.HasPrefix(filepath.ToSlash(filepath.Clean(line)), "../") \|\| filepath.Clean(line) == ".." { -> if false { | TestBadChangedList |
| M08 | internal/gatesample/sample.go:65 | changed[name] = true -> changed[name] = false | TestSelection |
| M09 | internal/gatesample/sample.go:71 | if part == "testdata" \|\| part == "fixtures" { -> if part == "testdata-disabled" \|\| part == "fixtures-disabled" { | TestSelection |
| M10 | internal/gatesample/sample.go:103 | sort.Strings(sorted) -> drop sort.Strings(sorted) and unused sort import | TestSelection |
| M11 | internal/gatesample/sample.go:132 | if _, err := hex.DecodeString(sha); err != nil { -> if _, err := hex.DecodeString(sha); err != nil && false { | TestSelection |
| M12 | internal/gatesample/sample.go:144 | keep[index] = true -> keep[index] = false | TestGeneratedSelection |
| M13 | internal/gatesample/sample.go:148 | !s.Sample \|\| index%s.Stride == s.Offset \|\| keep[index] -> s.Sample \|\| index%s.Stride == s.Offset \|\| keep[index] | TestGeneratedSelection |
| M14 | internal/gatesample/sample.go:147 | index < count -> index < count-1 | TestGeneratedSelection |
| M15 | internal/gatesample/sample.go:156 | "generated-index/%d" -> "generated-key/%d" |  |

M04 adds a multiplier to the offset expression, rather than changing an existing constant. It is supplemental, outside the fixed menu. Its observed failures appear in kills but establish no verdict or uniqueness. M10 drops the entire sort statement and the unused sort import so the standalone diff compiles. All other mutants use the fixed menu. The menu was fixed before matrix outcomes were inspected.

Survivor M15: `go run ./review/test-audit/internal-gatesample/witness > key-before.txt 2>&1` printed `GeneratedKey(4)="generated-index/4"`; the same command with `ADAMIC_MUTANT=M15` printed `GeneratedKey(4)="generated-key/4"`. The whole package still passed. This is unguarded key-prefix behavior, not an equivalent candidate. TestGeneratedSelection computes expected key strings through GeneratedKey too. PKey also makes all keys empty and the row still passes. This key-content portion cannot distinguish working keys from no keys. Generated is the row's primary entry and its PGenerated probe fails, so vacuous=false; vacuous_subcases: ["GeneratedKey content expectations (no named Go subtest)"].

Probes: PSelect returns Selection{}, nil at Select entry; PGenerated returns Selection{}; PLog and PKey return empty string; PValidate returns nil. PSelect fails the four Select rows; PGenerated fails TestGeneratedSelection; PLog fails TestSelection; PKey fails none. PValidate panics on TestSelection's short malformed SHA. The aborted whole run was not used to infer later results: every row was rerun alone (`ADAMIC_MUTANT=PValidate timeout 120 go test -json -count=1 -timeout 90s ./internal/gatesample/ -run ^<row>$ > PValidate-<row>.log 2>&1`). TestSelection alone fails with `panic: runtime error: slice bounds out of range [:8] with length 0`; all four other isolated rows pass. This supplemental validation probe is not used for primary-entry vacuity. No actual Go subtests are present.

Brief issues and time costs:

- The brief refers to "Your rows" but supplies no row list. The explicit whole-package scope and clean go test -list output determined the five rows. There is no prior row list from which to establish moves or vanished names; every discovered row was audited.
- Warm env.sh provided Go 1.27.1 and Node 24.19.0 but did not mean repository submodules were initialized. Discovery initially failed with missing cohere/TypeScript/tsc/go.mod; no baseline or mutants ran then. SSH initialization failed with port 22 refused; CLAUDE.md's documented HTTPS URL worked. This was dependency recovery before baseline, not an audit on a red baseline.
- *.log is ignored by repository policy, while the brief requests pushed raw test evidence. Logs are explicitly force-added only under the requested evidence directory.
- Strict fixed-menu treatment excludes M04 from verdict evidence.
- Scratch switch generation initially failed syntax and type checking; corrected before all matrix runs. No compiler failure counts as a kill.
- The brief asks for wall seconds of the binary's own ok line. Those exclude shell/compiler cost; both binary times and shell matrix wall times are recorded. Sub-millisecond test PASS lines round to zero, so timings use the ok package line.

Timing and coverage: nproc=5. Full toolchain setup skipped because env.sh worked. Toolchain probe completed in the recorded shell call's 0.00794s; npm ci reported 947ms (tool call 0.964s). Submodule recovery duration was not instrumented separately and cannot be precisely reported. Clean baseline binary: 0.157s; restored binary: 0.004s. Isolated medians: Selection 0.003s, Unset 0.002s, BadChangedList 0.004s, HexOffset 0.003s, GeneratedSelection 0.002s; all three rounds are in timings.json. Matrix/probe commands consumed 5.071720s shell wall total and 0.098s reported binary elapsed total; first switched command took 0.530734s including build/startup. All 20 standalone vet checks consumed 2.448311s wall total. Restored cached test-binary build (`go test -c -o /tmp/u021-gatesample.test ./internal/gatesample/`) took 0.318414s. Exact setup total and separate initial compile-only time were not measured. No timeout, opt-in skip, family, helper, witness, or bounded slice occurred. No repo-wide replay, native build, external oracle validation, or exhaustive mutation coverage was attempted. Production source is restored to base; only review evidence is committed.
