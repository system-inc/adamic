Audited internal/childguard at origin/main 7b18d0576930caca4e22ce2eef92fcf563af52d0.
Seven top-level tests: six sacred parent rows and one helper, no families or skips.
Eighteen production mutants, including one supplemental mutant; one empty-answer probe.
Three survivors have behavior-change witnesses; all six parent rows fail their entry probe.
Evidence is on test-audit/internal-childguard under review/test-audit/internal-childguard/.

CODE UNDER TEST, declared before mutants: Run, watched.Write, (*Error).Error and Run's writer wrapper, restoration and wait closures in internal/childguard/childguard.go. CombinedOutput is not reached by the scoped tests; it is used only in a supplemental survivor witness.
ORACLE, declared before mutants: self-written output, elapsed-time, error-field, diagnostic-text and exit-status expectations. The child Go executable and sh supply controlled inputs, not independent reference answers. No external authority value was checked.

Scope command: go test -list . ./internal/childguard/. The seven names in list.log all exist at the starting commit. No requested slice names were supplied; the whole package was audited. No moved or vanished rows. Bodies have distinct assertions, so sharing Run does not make them a family.

Every diff is standalone against the starting origin/main. Each passed go vet ./internal/childguard/ and git apply --check. switch.diff records the compiled scratch instrumentation. Production code was restored before commit. M13 changes error construction, broader than a literal constant/option change, so it is conservatively supplemental and excluded from kills, unique_kills and verdicts. Its observations are supplemental_kills.

```json
[
  {
    "test": "TestChild",
    "package": "internal/childguard",
    "file": "internal/childguard/childguard_test.go:14",
    "seconds": 0.002,
    "oracle": "Subprocess entry, activated only by CHILDGUARD_TEST; no independent oracle.",
    "oracle_kind": "self",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "helper",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": false,
    "matrix_rows": [
      "TestChild",
      "TestProgress",
      "TestStalled",
      "TestCeiling",
      "TestExitIsNotGuardError",
      "TestKillsProcessGroup",
      "TestNoFirstOutput"
    ],
    "evidence": "go test -list . ./internal/childguard/; TestChild",
    "supplemental_kills": [],
    "vacuous_subcases": [],
    "parents": [
      "TestProgress",
      "TestStalled",
      "TestCeiling",
      "TestExitIsNotGuardError",
      "TestNoFirstOutput"
    ]
  },
  {
    "test": "TestProgress",
    "package": "internal/childguard",
    "file": "internal/childguard/childguard_test.go:49",
    "seconds": 8.008,
    "oracle": "Self-written success and count of five progress newline occurrences; extra unrelated output is not rejected.",
    "oracle_kind": "self",
    "kills": [
      "M03",
      "M05",
      "M11",
      "M16"
    ],
    "unique_kills": [
      "M11"
    ],
    "last_proven_fail": "M11 childguard_test.go:57: output lost: \"\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestChild",
      "TestProgress",
      "TestStalled",
      "TestCeiling",
      "TestExitIsNotGuardError",
      "TestKillsProcessGroup",
      "TestNoFirstOutput"
    ],
    "evidence": "ADAMIC_MUTANT=M11 timeout 120 go test -json -count=1 -timeout 90s ./internal/childguard/ -run . > M11.log 2>&1; childguard_test.go:57: output lost: \"\"",
    "supplemental_kills": [
      "M13"
    ],
    "vacuous_subcases": []
  },
  {
    "test": "TestStalled",
    "package": "internal/childguard",
    "file": "internal/childguard/childguard_test.go:62",
    "seconds": 0.206,
    "oracle": "Self-written stalled error fields and text, 200 ms to 1200 ms timing bound, and exact stderr once newline.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M03",
      "M04",
      "M05",
      "M09",
      "M12",
      "M14",
      "M16"
    ],
    "unique_kills": [
      "M12"
    ],
    "last_proven_fail": "M12 childguard_test.go:76: stderr lost: \"\"",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestChild",
      "TestProgress",
      "TestStalled",
      "TestCeiling",
      "TestExitIsNotGuardError",
      "TestKillsProcessGroup",
      "TestNoFirstOutput"
    ],
    "evidence": "ADAMIC_MUTANT=M12 timeout 120 go test -json -count=1 -timeout 90s ./internal/childguard/ -run . > M12.log 2>&1; childguard_test.go:76: stderr lost: \"\"",
    "supplemental_kills": [],
    "vacuous_subcases": []
  },
  {
    "test": "TestCeiling",
    "package": "internal/childguard",
    "file": "internal/childguard/childguard_test.go:81",
    "seconds": 0.404,
    "oracle": "Self-written ceiling error reason and prefix, with upper timing bound only; no lower bound.",
    "oracle_kind": "self",
    "kills": [
      "M02",
      "M03",
      "M06",
      "M09"
    ],
    "unique_kills": [
      "M02"
    ],
    "last_proven_fail": "M02 childguard_test.go:86: want ceiling, got limit: 400.269866ms",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestChild",
      "TestProgress",
      "TestStalled",
      "TestCeiling",
      "TestExitIsNotGuardError",
      "TestKillsProcessGroup",
      "TestNoFirstOutput"
    ],
    "evidence": "ADAMIC_MUTANT=M02 timeout 120 go test -json -count=1 -timeout 90s ./internal/childguard/ -run . > M02.log 2>&1; childguard_test.go:86: want ceiling, got limit: 400.269866ms",
    "supplemental_kills": [],
    "vacuous_subcases": []
  },
  {
    "test": "TestExitIsNotGuardError",
    "package": "internal/childguard",
    "file": "internal/childguard/childguard_test.go:92",
    "seconds": 0.005,
    "oracle": "Self-written concrete exec.ExitError and exit status 7, excluding guard errors; exit code alone does not establish cause.",
    "oracle_kind": "self",
    "kills": [
      "M06",
      "M08"
    ],
    "unique_kills": [
      "M08"
    ],
    "last_proven_fail": "M08 childguard_test.go:98: wrong exit error: ceiling: 263.27\u00b5s",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestChild",
      "TestProgress",
      "TestStalled",
      "TestCeiling",
      "TestExitIsNotGuardError",
      "TestKillsProcessGroup",
      "TestNoFirstOutput"
    ],
    "evidence": "ADAMIC_MUTANT=M08 timeout 120 go test -json -count=1 -timeout 90s ./internal/childguard/ -run . > M08.log 2>&1; childguard_test.go:98: wrong exit error: ceiling: 263.27\u00b5s",
    "supplemental_kills": [
      "M13"
    ],
    "vacuous_subcases": []
  },
  {
    "test": "TestKillsProcessGroup",
    "package": "internal/childguard",
    "file": "internal/childguard/childguard_test.go:104",
    "seconds": 0.207,
    "oracle": "Self-written stalled guard result and 2 s pipe-closure deadline. Descendant survival is inferred from inherited pipes, not independently checked by PID.",
    "oracle_kind": "self",
    "kills": [
      "M04",
      "M09",
      "M14",
      "M16",
      "M17"
    ],
    "unique_kills": [
      "M17"
    ],
    "last_proven_fail": "M17 childguard_test.go:115: descendant kept output pipes open after group kill",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestChild",
      "TestProgress",
      "TestStalled",
      "TestCeiling",
      "TestExitIsNotGuardError",
      "TestKillsProcessGroup",
      "TestNoFirstOutput"
    ],
    "evidence": "ADAMIC_MUTANT=M17 timeout 120 go test -json -count=1 -timeout 90s ./internal/childguard/ -run . > M17.log 2>&1; childguard_test.go:115: descendant kept output pipes open after group kill",
    "supplemental_kills": [],
    "vacuous_subcases": []
  },
  {
    "test": "TestNoFirstOutput",
    "package": "internal/childguard",
    "file": "internal/childguard/childguard_test.go:120",
    "seconds": 0.305,
    "oracle": "Self-written first-output flag, 300 ms window, diagnostic text, and 300 ms to 1300 ms timing bound.",
    "oracle_kind": "self",
    "kills": [
      "M01",
      "M09",
      "M14",
      "M15",
      "M16"
    ],
    "unique_kills": [
      "M15"
    ],
    "last_proven_fail": "M15 childguard_test.go:125: want first-output stall, got stalled: no first output for 100ms after 301.434443ms, load 0.02 (/tmp/adamic-gate/go-build855092530/b001/childguard.test)",
    "verdict": "sacred",
    "subsumed_by": [],
    "mutants_in_matrix": 18,
    "probe_kills": [
      "P01"
    ],
    "subsumer_seconds": null,
    "vacuous": false,
    "bounded": false,
    "matrix_rows": [
      "TestChild",
      "TestProgress",
      "TestStalled",
      "TestCeiling",
      "TestExitIsNotGuardError",
      "TestKillsProcessGroup",
      "TestNoFirstOutput"
    ],
    "evidence": "ADAMIC_MUTANT=M15 timeout 120 go test -json -count=1 -timeout 90s ./internal/childguard/ -run . > M15.log 2>&1; childguard_test.go:125: want first-output stall, got stalled: no first output for 100ms after 301.434443ms, load 0.02 (/tmp/adamic-gate/go-build855092530/b001/childguard.test)",
    "supplemental_kills": [],
    "vacuous_subcases": []
  }
]
```

| ID | origin/main file:line | Change | Failed rows |
|---|---|---|---|
| M01 | internal/childguard/childguard.go:39 | `if e.FirstOutput { -> if !e.FirstOutput {` | TestStalled, TestNoFirstOutput |
| M02 | internal/childguard/childguard.go:44 | `return fmt.Sprintf("ceiling: %s", e.Elapsed) -> return fmt.Sprintf("limit: %s", e.Elapsed)` | TestCeiling |
| M03 | internal/childguard/childguard.go:61 | `w.activity.last = time.Now() -> w.activity.last = time.Time{}` | TestProgress, TestStalled, TestCeiling |
| M04 | internal/childguard/childguard.go:62 | `w.activity.seen = true -> w.activity.seen = false` | TestStalled, TestKillsProcessGroup |
| M05 | internal/childguard/childguard.go:69 | `return w.dst.Write(p) -> return len(p), nil` | TestProgress, TestStalled |
| M06 | internal/childguard/childguard.go:77 | `options.FirstOutput = DefaultFirstOutput -> options.FirstOutput = time.Nanosecond` | TestCeiling, TestExitIsNotGuardError |
| M07 | internal/childguard/childguard.go:80 | `options.Stall = DefaultStall -> options.Stall = time.Nanosecond` |  |
| M08 | internal/childguard/childguard.go:83 | `options.Ceiling = DefaultCeiling -> options.Ceiling = time.Nanosecond` | TestExitIsNotGuardError |
| M09 | internal/childguard/childguard.go:91 | `cmd.SysProcAttr.Setpgid = true -> cmd.SysProcAttr.Setpgid = false` | TestStalled, TestCeiling, TestKillsProcessGroup, TestNoFirstOutput |
| M10 | internal/childguard/childguard.go:94 | `defer func() { cmd.Stdout, cmd.Stderr = originalOut, originalErr }() -> (drop statement)` |  |
| M11 | internal/childguard/childguard.go:103 | `cmd.Stdout = wrap(originalOut) -> cmd.Stdout = wrap(originalErr)` | TestProgress |
| M12 | internal/childguard/childguard.go:108 | `cmd.Stderr = wrap(originalErr) -> cmd.Stderr = wrap(originalOut)` | TestStalled |
| M13 | internal/childguard/childguard.go:120 | `return err -> return fmt.Errorf("%v", err)` | TestProgress, TestExitIsNotGuardError |
| M14 | internal/childguard/childguard.go:132 | `firstOutput := !a.seen -> firstOutput := a.seen` | TestStalled, TestKillsProcessGroup, TestNoFirstOutput |
| M15 | internal/childguard/childguard.go:135 | `window = options.FirstOutput -> window = options.Stall` | TestNoFirstOutput |
| M16 | internal/childguard/childguard.go:140 | `if elapsed >= options.Ceiling { -> if elapsed < options.Ceiling {` | TestProgress, TestStalled, TestKillsProcessGroup, TestNoFirstOutput |
| M17 | internal/childguard/childguard.go:146 | `syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) -> syscall.Kill(cmd.Process.Pid, syscall.SIGKILL)` | TestKillsProcessGroup |
| M18 | internal/childguard/childguard.go:153 | `load = fields[0] -> load = "unavailable"` |  |
| P01 | internal/childguard/childguard.go:75 | `func Run(cmd *exec.Cmd, options Options) error { -> func Run(cmd *exec.Cmd, options Options) error {  if true { return nil }` | TestProgress, TestStalled, TestCeiling, TestExitIsNotGuardError, TestKillsProcessGroup, TestNoFirstOutput |

M09's whole-package run exceeded 90 s and aborted during TestStalled. It was rerun over all seven rows individually with the same timeout, so no row is inferred from an aborted package run. TestChild, TestProgress and TestExitIsNotGuardError passed; TestKillsProcessGroup explicitly failed its 2 s assertion; TestStalled, TestCeiling and TestNoFirstOutput each failed with `panic: test timed out after 1m30s`. These are observed timeout kills, not ordinary assertion kills. Each command was `ADAMIC_MUTANT=M09 timeout 120 go test -json -count=1 -timeout 90s ./internal/childguard/ -run '^ROW$' > M09.ROW.log 2>&1`. Every row was observed in isolation, so this is not a package slice. Reruns ran concurrently; their 90 s windows overlapped. Four living orphan processes from kill mutants were identified and terminated after the runs.

P01 is Run returning nil at entry, implemented with a constant-true block in the standalone probe to pass vet. It fails all six parent rows. TestChild does not call Run and gets vacuous=null. There are no subtests or positive/negative subcase splits.

Survivors, with session-produced before/after witnesses:
M07: output="ready\ndone\n" error=<nil> -> output="ready\n" error=stalled: no output for 1ns after 1.598727ms, load 0.27 (/usr/bin/sh)

M10: stdout_restored=true output="ready" error=<nil> -> stdout_restored=false output="ready" error=<nil>

M18: reason=stalled load="0.32" -> reason=stalled load="unavailable"

Witness command: `ADAMIC_MUTANT= go run ./review/test-audit/internal-childguard/witness MID > MID.witness-before.log 2>&1`, then the same command with `ADAMIC_MUTANT=MID` and an after log, while the switch was present. The witness source is saved. These are changed, unguarded behaviors, not equivalent candidates. No mutant or probe edits the tests or subprocess helper.

Brief ambiguities, costs and limits:

- Warm env.sh worked, but go.work referenced missing cohere/TypeScript/tsc. The first list/baseline invocation failed to load the workspace; no tests ran. Initialized the pinned recursive submodule and retried. The first executable clean baseline passed. A distinction between an infrastructure invocation failure and a red test baseline would help.
- Dropping an error return left its case binding unused. Vet rejected that attempt before the matrix. The replacement M13 passed vet, but is supplemental because the menu's option boundary is unclear for returned error construction. No verdict rests on it.
- A direct unconditional empty return makes vet report unreachable code. A constant-true block preserves the empty result while satisfying the required standalone vet check.
- The instruction to compile once and the mandated repeated go test commands coexist here: one switch source/binary was built, and unchanged source reused Go's build cache across selectors. Standalone vet validation necessarily compiled each diff. No native build or ADAMIC build cache is involved.
- TestProgress counts five expected lines without rejecting arbitrary extra output. TestCeiling checks an upper bound but no lower bound; it passed M16 even though an immediate progress-triggered ceiling can happen before 400 ms. The M16 log demonstrates that weakness. The process-group test infers descendant termination from inherited pipe closure rather than checking every descendant PID.
- Sacred means package uniqueness in this finite matrix, with named self oracles. It does not prove external correctness or repository-wide uniqueness. Central replay has every standalone diff.
- The matrix does not cover other packages. CombinedOutput, negative-duration rejection, start failures, same-writer serialization, restoration of stderr, and all default first-output/stall/ceiling semantics are not exhaustively covered by these six rows or this mutant set. No witness or setup-check rows were present.

Timing and coverage:
Warm toolchain setup: skipped (0 setup work); go1.27.1 linux/amd64, nproc=5. npm ci reported 1 s and installed three packages. Recursive submodule initialization completed below 90 s but was not independently timed. Clean baseline binary line: 9.121 s. Three standalone timing trials per row: 32.255 wall seconds including Go command overhead. Whole-package mutant/probe commands: 250.839 wall seconds including compilation overhead. M09 individual reruns: concurrent, longest test-binary budget 90 s; exact aggregate wall was not instrumented. Clean restored binary build: 0.359 wall seconds. Initial switch build was not independently timed; final successful standalone vet validation and switch compilation took about 2.6 s from saved log timestamps. No native rebuilds. Total unit work about 10 minutes, including environment startup and evidence preparation. Production restored, no main push and no pull request.
