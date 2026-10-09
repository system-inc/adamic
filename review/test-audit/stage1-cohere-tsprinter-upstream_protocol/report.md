Unit u142: both requested rows exist at origin/main 60f24a6ef736a6016edeca63df075678186bba8d.
The whole-package baseline timed out at 90.026 test seconds; no row failure preceded the timeout.
The bounded baseline passed both rows without skips in 1.781 test seconds.
Three-run medians: 0.527s and 1.416s; nproc=5.
Both verdicts are cannot-judge under the permitted-code mutation rule; no mutants or probes planted.

[
  {
    "test": "TestTSCCorpusUpstreamDifferences",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/upstream_protocol_test.go",
    "seconds": 0.527,
    "oracle": "Pinned tsc-upstream-differences.json Records checked byte-for-byte against npm Prettier 3.9.6 and the embedded Prettier fork running on Node. Includes exact parser error strings, not just exit codes. No Adamic port output is checked.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "ADAMIC_TS_PRETTIER=/workspace/u097-prettier ADAMIC_TYPESCRIPT_SOURCE=/workspace/u111-typescript timeout 120 go test -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run ^TestTSCCorpusUpstreamDifferences$; ok  \tgithub.com/system-inc/adamic/stage1/cohere/tsprinter\t0.527s; no failing line: no permitted Adamic code under test is reached.",
    "samples": [
      0.507,
      0.53,
      0.527
    ]
  },
  {
    "test": "TestStatementUpstreamDifferences",
    "package": "stage1/cohere/tsprinter",
    "file": "stage1/cohere/tsprinter/upstream_test.go",
    "seconds": 1.416,
    "oracle": "Five pinned statement-upstream-differences.json records checked against external Go cohere Format and embedded/npm Prettier 3.9.6. The npm parent checks exit/stderr, but statementDifferences.mjs compares exact text and requires disagreement with Go. No Adamic port output is checked.",
    "oracle_kind": "external-run",
    "kills": [],
    "unique_kills": [],
    "last_proven_fail": null,
    "verdict": "cannot-judge",
    "subsumed_by": [],
    "mutants_in_matrix": 0,
    "probe_kills": [],
    "subsumer_seconds": null,
    "vacuous": null,
    "bounded": true,
    "matrix_rows": [],
    "evidence": "ADAMIC_TS_PRETTIER=/workspace/u097-prettier ADAMIC_TYPESCRIPT_SOURCE=/workspace/u111-typescript timeout 120 go test -count=1 -timeout 90s ./stage1/cohere/tsprinter/ -run ^TestStatementUpstreamDifferences$; ok  \tgithub.com/system-inc/adamic/stage1/cohere/tsprinter\t1.416s; no failing line: no permitted Adamic code under test is reached.",
    "samples": [
      1.445,
      1.385,
      1.416
    ]
  }
]

Mutant table: empty. No permissible Adamic implementation function is reached.
Survivors: none tested. This is not evidence of complete coverage.

Scope and brief issues

1. The assigned rows validate external formatter pins, not Adamic's port. upstream_test.go:11 explicitly describes the test as a small pin check on external formatters. Its executable call at line 23 runs Go cohere in the cohere submodule; lines 51 and 61 run embedded/npm Prettier. upstream_protocol_test.go:121 and :124 run only npm/embedded Prettier. All line numbers refer to the starting origin/main commit. Mutating these implementations would violate the brief's prohibition on oracle mutation. A harness comparison mutant would also be unauthorized: neither row is a witness asserting that another check can fail, nor a shard/setup construction check. Therefore there is no honest production mutant, standalone diff, kill, unique kill, or vacuity result to give. Nulls and empty arrays preserve that distinction. Nothing supports an untrue or sacred verdict.

2. The requested base description names 8de93800f4. The required fetch produced 60f24a6ef736a6016edeca63df075678186bba8d, which is the actual starting origin/main. Both names exist in list.log and neither moved or vanished. No grouping applies: their bodies call different checkers and assert different things. They remain two rows.

3. The whole package is much larger than this slice. Its baseline timed out at the test binary's 90-second limit. The whole-run log is retained, with no test fail action before the timeout. Narrowing to exactly the two named rows passed. No mutation matrix was run, so matrix_rows is empty even though timing and baseline runs used these two rows. No package or repository uniqueness is claimed.

4. The initial whole baseline used the installed npm formatter but did not set ADAMIC_TYPESCRIPT_SOURCE. Its skip list is saved in skipped.json: 56 expression shard wrappers and TestProduct_TSPrinterStatementsNPM_012 skipped. The bounded run then set the available pinned TypeScript checkout as well; neither assigned row uses that checkout, and neither assigned row skipped. Unrelated skipped rows were not rerun after narrowing.

5. /usr/bin/time is absent in this environment. Two initial npm attempts therefore did not execute; their logs were replaced by successful real npm ci runs using Bash time. Dependency installs took 0.521s in stage3/api and 0.685s in /workspace/u097-prettier. The toolchain environment worked, so setup was skipped. Go 1.27.1, Node v24.19.0, npm Prettier 3.9.6. No port native rebuild was performed because the rows never call one.

6. The protocol row checked 15 expression and 14 statement cases against each external Prettier: exact text differences and exact parser refusal messages. The statement row checked five exact Go fork differences and five npm differences; its npm script enforces exact text before returning success. These runs establish current pin agreement, not Adamic semantic correctness. Both tests load their test inputs from their own pinned records, so they also do not independently establish that the pinned set exhausts possible upstream differences.

Timing and omissions

Warm standalone Go test-binary build: 2.491s. Whole baseline command: 92.098s wall, 90.026s test time. Bounded baseline command: 3.948s wall, 1.781s test time. Six isolated timing commands: 19.140s total wall. Samples are package ok-line times, not child Go cohere timings. See individual .time and .log files.

No allowed mutants, probes, witness edits, survivor witnesses, or standalone diffs were produced. Thus truth, worthiness, subsumption, and vacuity remain unmeasured. No other package was run as an audit target; the statement row itself necessarily runs its external Go cohere oracle package. Production source is unchanged, verified against origin/main. Evidence-only commit and push to the requested branch; no PR or main push.
