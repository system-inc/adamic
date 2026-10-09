u114 at ce1c5a2fd91e: all four listed tests exist; no families, helpers or vanished names.
Bounded verdicts: one sacred, two subsumed, one local-comparison witness.
All four production mutants are caught; each standalone diff builds natively.
All three production rows reject the empty-main probe; witness was not probed.
Production source restored; evidence on test-audit/stage1-cohere-lint-helpers.

```json
[
  {
    "test": "TestHelpersMatchCohere",
    "package": "stage1/cohere/lint/helpers",
    "file": "stage1/cohere/lint/helpers/helpers_test.go:78",
    "seconds": 12.769,
    "timing_samples": [
      13.542,
      12.545,
      12.769
    ],
    "oracle": "Go encoding/json, cohere optionschema, generated Go target decoders and policy.Messages.Render; byte comparison. 5923 Expected seeds also check archived ESLint acceptance/refusal labels. Checked [{}] for require-description against the ESLint 10.8.1/plugin 4.8.1 archive; no fresh ESLint execution.",
    "oracle_kind": [
      "external-run",
      "external-authority"
    ],
    "kills": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "unique_kills": [
      "M1",
      "M3"
    ],
    "last_proven_fail": "M4: helpers_test.go:98: node [--disable-warning=ExperimentalWarning /workspace/adamic/oracle/node.mjs /workspace/adamic/stage1/cohere/lint/helpers/main.ts /tmp/adamic-gate/TestHelpersMatchCohere1805568307/006/cases.json /workspace/adamic/stage1/cohere/lint/helpers/testdata/catalog.json]: exit status 70",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "probe_kills": [
      "P"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestHelpersMatchCohere",
      "TestMessageRefusalsMatchGo",
      "TestKnownGapsAreExplicit"
    ],
    "evidence": "ADAMIC_MUTANT=M4; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/helpers/ -run ^(TestHelpersMatchCohere|TestMessageRefusalsMatchGo|TestKnownGapsAreExplicit)$; helpers_test.go:98: node [--disable-warning=ExperimentalWarning /workspace/adamic/oracle/node.mjs /workspace/adamic/stage1/cohere/lint/helpers/main.ts /tmp/adamic-gate/TestHelpersMatchCohere1805568307/006/cases.json /workspace/adamic/stage1/cohere/lint/helpers/testdata/catalog.json]: exit status 70; log=M4-matrix.log",
    "probe_evidence": {
      "command": "ADAMIC_MUTANT=P; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/helpers/ -run ^(TestHelpersMatchCohere|TestMessageRefusalsMatchGo|TestKnownGapsAreExplicit)$",
      "failure": "helpers_test.go:99: output line 1: got \"\", Go \"invalid\"",
      "entry": "main.ts",
      "diff": "P.diff"
    }
  },
  {
    "test": "TestHelperMutants",
    "package": "stage1/cohere/lint/helpers",
    "file": "stage1/cohere/lint/helpers/helpers_test.go:118",
    "seconds": 27.224,
    "timing_samples": [
      27.224,
      27.235,
      27.06
    ],
    "oracle": "Independent Go answers and archived ESLint seeds against compiled builtin mutants, using its own direct byte-equality check. W1 proves that local check; WShared+M1 shows this witness does not protect the production compare helper.",
    "oracle_kind": [
      "external-run",
      "external-authority"
    ],
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": "W1: helpers_test.go:142: compiled mutant survived",
    "verdict": "witness",
    "subsumed_by": [],
    "mutants_in_matrix": [],
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [
      "TestHelperMutants"
    ],
    "evidence": "ADAMIC_MUTANT=W1; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/helpers/ -run ^(TestHelperMutants)$; helpers_test.go:142: compiled mutant survived; log=W1.log",
    "probe_evidence": null
  },
  {
    "test": "TestMessageRefusalsMatchGo",
    "package": "stage1/cohere/lint/helpers",
    "file": "stage1/cohere/lint/helpers/helpers_test.go:195",
    "seconds": 6.134,
    "timing_samples": [
      6.131,
      6.787,
      6.134
    ],
    "oracle": "Go policy.Messages.Render panic text; own adamic prefix and exit 70. Full stderr rejects M4 despite the same exit code. Supplemental SOut shows stdout is unchecked.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M4"
    ],
    "unique_kills": [],
    "last_proven_fail": "M4: helpers_test.go:236: guard 2 /tmp/adamic-gate/TestMessageRefusalsMatchGo2931154253/001/helpers: exit status 70 stderr \"adamic: panic: policy/messages: nexus/consistency-no-stuttering-name \\\"stutteringName\\\" picks from 0 phrases, and was given 0 options\\n\", Go \"adamic: panic: policy/messages: nexus/consistency-no-stuttering-name \\\"stutteringName\\\" uses the values [name], and was given 0\\n\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestHelpersMatchCohere"
    ],
    "mutants_in_matrix": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "probe_kills": [
      "P"
    ],
    "subsumer_seconds": 12.769,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestHelpersMatchCohere",
      "TestMessageRefusalsMatchGo",
      "TestKnownGapsAreExplicit"
    ],
    "evidence": "ADAMIC_MUTANT=M4; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/helpers/ -run ^(TestHelpersMatchCohere|TestMessageRefusalsMatchGo|TestKnownGapsAreExplicit)$; helpers_test.go:236: guard 2 /tmp/adamic-gate/TestMessageRefusalsMatchGo2931154253/001/helpers: exit status 70 stderr \"adamic: panic: policy/messages: nexus/consistency-no-stuttering-name \\\"stutteringName\\\" picks from 0 phrases, and was given 0 options\\n\", Go \"adamic: panic: policy/messages: nexus/consistency-no-stuttering-name \\\"stutteringName\\\" uses the values [name], and was given 0\\n\"; log=M4-matrix.log",
    "probe_evidence": {
      "command": "ADAMIC_MUTANT=P; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/helpers/ -run ^(TestHelpersMatchCohere|TestMessageRefusalsMatchGo|TestKnownGapsAreExplicit)$",
      "failure": "helpers_test.go:236: guard 0 /tmp/adamic-gate/TestMessageRefusalsMatchGo2481035693/001/helpers: <nil> stderr \"\", Go \"adamic: panic: policy/messages: missing/rule has no message \\\"missing\\\"\\n\"",
      "entry": "main.ts",
      "diff": "P.diff"
    },
    "limitations": "Subsumption rests on 1 caught mutant. It is a bounded hint, not deletion advice."
  },
  {
    "test": "TestKnownGapsAreExplicit",
    "package": "stage1/cohere/lint/helpers",
    "file": "stage1/cohere/lint/helpers/helpers_test.go:244",
    "seconds": 1.788,
    "timing_samples": [
      1.788,
      1.886,
      1.671
    ],
    "oracle": "Handwritten NotYet labels and valid recovery answer, checked on source Node and sanitized native. No outside authority for those labels.",
    "oracle_kind": [
      "external-run",
      "self"
    ],
    "kills": [
      "M2"
    ],
    "unique_kills": [],
    "last_proven_fail": "M2: helpers_test.go:260: output line 1: got \"NotYet: unsupported schema keyword type\", Go \"NotYet: unsupported schema keyword pattern\"",
    "verdict": "subsumed",
    "subsumed_by": [
      "TestHelpersMatchCohere"
    ],
    "mutants_in_matrix": [
      "M1",
      "M2",
      "M3",
      "M4"
    ],
    "probe_kills": [
      "P"
    ],
    "subsumer_seconds": 12.769,
    "vacuous": false,
    "bounded": true,
    "matrix_rows": [
      "TestHelpersMatchCohere",
      "TestMessageRefusalsMatchGo",
      "TestKnownGapsAreExplicit"
    ],
    "evidence": "ADAMIC_MUTANT=M2; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/helpers/ -run ^(TestHelpersMatchCohere|TestMessageRefusalsMatchGo|TestKnownGapsAreExplicit)$; helpers_test.go:260: output line 1: got \"NotYet: unsupported schema keyword type\", Go \"NotYet: unsupported schema keyword pattern\"; log=M2-matrix.log",
    "probe_evidence": {
      "command": "ADAMIC_MUTANT=P; timeout 120 go test -json -count=1 -timeout 90s ./stage1/cohere/lint/helpers/ -run ^(TestHelpersMatchCohere|TestMessageRefusalsMatchGo|TestKnownGapsAreExplicit)$",
      "failure": "helpers_test.go:260: output line 1: got \"\", Go \"NotYet: unsupported schema keyword pattern\"",
      "entry": "main.ts",
      "diff": "P.diff"
    },
    "limitations": "Subsumption rests on 1 caught mutant. It is a bounded hint, not deletion advice."
  }
]
```

| ID | Origin file:line | One-line change | Failed rows |
|---|---|---|---|
| M1 | stage1/cohere/lint/helpers/options_json.ts:136 | `a.number === b.number` → `a.number !== b.number` | TestHelpersMatchCohere |
| M2 | stage1/cohere/lint/helpers/option_schema.ts:23 | `!known.includes(key)` → `known.includes(key)` | TestHelpersMatchCohere, TestKnownGapsAreExplicit |
| M3 | stage1/cohere/lint/helpers/strict_options.ts:14 | `if(value.kind === 'null') { return true; }` → `if(value.kind === 'null') { return false; }` | TestHelpersMatchCohere |
| M4 | stage1/cohere/lint/helpers/policy_message.ts:46 | `chosen.size !== count` → `chosen.size !== count + 1` | TestHelpersMatchCohere, TestMessageRefusalsMatchGo |

Survivors: none in the bounded production matrix. Native witnesses are in native-witnesses.json: M1 changes valid/invalid at line 15159; M2 emits unsupported schema keyword type at line 1; M3 rejects null at line 15171; M4 exits 70 with an erroneous phrase-count panic at line 23350.

Separate witness and supplemental observations:

- W1, helpers_test.go:141: replace the witness equality comparison of got/want with got/got. All four builtin-mutant subcases fail at helpers_test.go:142, compiled mutant survived. W1 is not a production kill.
- WShared, helpers_test.go:105: disable the production compare helper by comparing got/got. With admissible M1 also present, TestHelpersMatchCohere and TestHelperMutants both pass. This shows the witness does not protect the production helper. Its witness verdict applies only to its own direct byte-equality comparison.
- SOut, main.ts:7: insert console.log of audit unexpected stdout. This insertion is outside the fixed menu, explicitly supplemental, and supports no verdict. Helpers agreement and gap labels fail; TestMessageRefusalsMatchGo passes. A native refusal run prints the unexpected line and still exits 70 with the expected stderr.

Brief ambiguities, mistakes, costs and limits:

- The complete cold baseline exceeded 90 seconds during the final builtin-mutant subcase, after the main agreement row passed in 63.26 seconds. This was a budget timeout, not an observed assertion failure. The remaining three rows passed a narrowed clean run in 35.179 seconds.
- The fetched origin/main is ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2. All four names exist in go test -list; the package has no additional top-level Tests and no grouping is needed.
- The production matrix is bounded to the three production rows. The builtin witness is excluded from production kill counts by the brief. Its guard is weakened separately; package-wide uniqueness is left for central replay.
- The main agreement test compares source Node before building/running its native product. Mutant runs can stop at that first mismatch. Every standalone mutant was therefore separately built and executed natively against unchanged independent Go output; none is a compiler-warning kill.
- The first planned fourth edit loosened equality to greater-than, which was not a literal fixed-menu operation. Before any mutant was applied, it was replaced by count+1, a strict off-by-one bound. The timing controller was interrupted, source cleanliness verified, and completed binary timing logs reused.
- The witness has its own equality check and does not call production compare. Disabling compare while numeric equality is broken makes both the agreement row and witness pass. Formal witness status here is local to its direct equality check; it must not be interpreted as protection of the production comparator.
- The refusal row checks exact stderr and code 70, so M4 is caught despite producing the expected exit code. It does not check stdout: the supplemental inserted output survives that row. This survivor is row-local and supplemental, not an unguarded whole-unit verdict.
- The gap row expects handwritten NotYet labels and a handwritten valid recovery result. These are self expectations, not Go diagnostic authority. Source Node is an additional implementation execution. M2 changes the refusal reason and proves the label check can fail.
- The Go oracle also checks 5923 pinned ESLint sample expectations. One valid [{}] require-description value was checked against the archived ESLint 10.8.1/plugin 4.8.1 sample. ESLint itself was not rerun or installed; these pins are external-authority evidence, separate from this session's Go execution.
- The current gzip corpus has 23539 cases and 23609 output lines. Message values can contain newlines. Historical REPORT.md numbers of 22347 cases are stale and were not used as scope or timing evidence.
- The declaration inventory lists 28 named functions/constructors including module entry and 11 callback sites. It is a source caller inventory, not dynamic branch coverage. A regex false positive for an if statement was removed from the inventory.
- The first witness-edit draft matched both equality conditions and stopped before mutations. The supplemental comparator draft likewise needed a newline anchor. Both controller errors are preserved; they support no verdict.
- The port selector is carried in an existing copied file, options_json.ts, so the witness's fixed copied-file list and the Go oracle remain unchanged. Switching /tmp/u114-mutant changes runtime behavior without changing source products. Standalone replay diffs contain no selector.
- The empty-main probe is a separate native-buildable empty entry. It is not mixed into the production mutant switch or counted as a kill. It rejects all three production rows. No empty probe was attributed to the witness.
- The subsumption of each smaller row rests on one caught mutant. Both smaller rows are faster than the 12.769-second subsumer. These are bounded hints, not advice to delete cheaper tests.
- All timings use three separate count=1 invocations with a single top-level row and no concurrent audit workload. The first witness timing completed after its controller was interrupted; its binary line is valid, while command wall time is unknown.
- No opt-ins or skips were found in this package. Warm tools did not excuse dependency setup: npm ci ran in stage3/api before the baseline. The Node loader needs no additional node_modules for this port.
- No other repository packages were tested, no full integration gate or exhaustive input fuzzer was run, and no central repo-wide uniqueness replay was attempted. Oracle and compiler executables were built only as required for this unit.
- Every standalone diff, including the empty probe, weakened checks and supplemental stdout insertion, applies to the starting commit and has its appropriate build or vet evidence. Production source is restored. No main push or pull request is made.

Setup/build/run measurements:

```json
{
  "base": "ce1c5a2fd91e40b05160e587ea5d88ed5bdfedb2",
  "nproc": 5,
  "toolchain_setup_seconds": 0,
  "toolchain": "Warm env.sh worked; setup skipped. API npm ci reported 428 ms. No other node_modules required.",
  "binary_baseline_seconds": 90.017,
  "narrowed_baseline_seconds": 35.179,
  "native_build_wall": {
    "M1": 1.5497365510000236,
    "M2": 1.6115707349999866,
    "M3": 1.6279866530003346,
    "M4": 1.6254646779998438
  },
  "compiler_build_wall": 5.995741160999387,
  "timing_command_wall_sum": 140.37033806599902,
  "timing_recovered": "The first witness timing completed after its controller received SIGINT during menu correction. Its binary line is complete and retained; command wall is null.",
  "audit_command_wall_sum": 334.6943979380012,
  "supplemental_command_wall_sum": 62.754304115000195,
  "cases": 23539,
  "go_output_lines": 23609,
  "elapsed_note": "Approximately 20 minutes through evidence preparation. Timings were serial, each count=1; no concurrent test workload was launched."
}
```
