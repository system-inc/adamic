**Both targets gone: yes. New findings: 1 compiler refusal, plus 9 recovery boundaries (10 total).**

# program.ts full census

Measurement and report only for #ty5myjt. No adaptation or compiler changes. Measured 2026-11-14 UTC.

Baseline: `7534c1d7575db02bb7b6512119e0685987b7e333`. Adapted commit: `5ead54d316a4a6606aa61a370adea226d66f3db9` (its parent is the baseline).

| Tree | Wall seconds | Peak RSS (KiB) | Exit | 100-minute bound reached |
|---|---:|---:|---:|---|
| base | 684.438143 | 1,938,048 | 0 | no |
| change | 207.009796 | 1,272,112 | 0 | no |

Each real census ran once, sequentially, with a 6,000-second bound. Wall time is monotonic time around process launch through `wait4`, polled every 0.2 seconds; peak RSS and exit are the individual child’s `wait4` results. Memory figures exclude scratch files and build processes. The runner stops after a timeout or nonzero exit.

## Paired findings

**Red observation:** one new compiler Refused signature at `src/compiler/program.ts:2781:9`, phase `lowering`: `overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics`. It appears in nine raw attempts. Compiler-only totals: **24 gone, 1 new, 0 changed**; recovery boundaries add 20 gone and 9 new.

Pairing uses `(kind, file, line, column, reason)`, deduplicated across attempts. Absolute adapted-tree prefixes are normalized to relative paths. Unchanged source lines are mapped with `difflib.SequenceMatcher(autojunk=False)` on each finding’s actual file; columns and reasons remain exact. An old line absent from the matching blocks is never equated with a candidate line. Findings in dependency files are included.

Unique signatures: **515 baseline**, **481 adapted**; **471 matched**, **44 gone**, **10 new**.

Changed reason groups at the same mapped kind/location: **0** (0 old and 0 new signatures, already included in gone/new). Excluding these groups: 44 gone and 10 new. Matched signatures with changed phase sets: **0**.

Counts include all emitted kinds, including recovery `Boundary` findings; compiler blocker counts are shown separately below. Every new signature and all its raw attempt records are reproduced below, with adapted-tree prefixes normalized. Broader raw evidence is retained in local scratch and is not part of this report-only push.

| Kind | Baseline unique | Adapted unique | Gone | New |
|---|---:|---:|---:|---:|
| Boundary | 232 | 221 | 20 | 9 |
| NotYet | 182 | 163 | 19 | 0 |
| Refused | 101 | 97 | 5 | 1 |

### Target findings

| Baseline site | Mapped adapted site | Gone |
|---|---|---|
| `src/compiler/program.ts:2796:16` | `src/compiler/program.ts:2806:16` | yes |
| `src/compiler/program.ts:2829:16` | `src/compiler/program.ts:2839:16` | yes |

### Every new finding, in full

#### New signature 1

```json
{
  "signature": [
    "Boundary",
    "src/compiler/program.ts",
    1500,
    1,
    "function signature or prologue failed; continued at next declaration"
  ],
  "records": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:1500:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "src/compiler/program.ts:1500:1",
      "reason": "function signature or prologue failed; continued at next declaration",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
      "start": 67052,
      "end": 67644
    }
  ]
}
```

#### New signature 2

```json
{
  "signature": [
    "Boundary",
    "src/compiler/program.ts",
    1515,
    1,
    "function signature or prologue failed; continued at next declaration"
  ],
  "records": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:1515:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "src/compiler/program.ts:1515:1",
      "reason": "function signature or prologue failed; continued at next declaration",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
      "start": 67644,
      "end": 68583
    }
  ]
}
```

#### New signature 3

```json
{
  "signature": [
    "Boundary",
    "src/compiler/program.ts",
    1516,
    1,
    "function signature or prologue failed; continued at next declaration"
  ],
  "records": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:1516:1",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "src/compiler/program.ts:1516:1",
      "reason": "function signature or prologue failed; continued at next declaration",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
      "start": 68814,
      "end": 254387
    }
  ]
}
```

#### New signature 4

```json
{
  "signature": [
    "Boundary",
    "src/compiler/program.ts",
    2779,
    5,
    "function signature or prologue failed; continued at next declaration"
  ],
  "records": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:2779:5",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "src/compiler/program.ts:2779:5",
      "reason": "function signature or prologue failed; continued at next declaration",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
      "start": 136096,
      "end": 136419
    }
  ]
}
```

#### New signature 5

```json
{
  "signature": [
    "Boundary",
    "src/compiler/program.ts",
    2784,
    5,
    "function signature or prologue failed; continued at next declaration"
  ],
  "records": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:2784:5",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "src/compiler/program.ts:2784:5",
      "reason": "function signature or prologue failed; continued at next declaration",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
      "start": 136419,
      "end": 136716
    }
  ]
}
```

#### New signature 6

```json
{
  "signature": [
    "Boundary",
    "src/compiler/program.ts",
    2789,
    5,
    "function signature or prologue failed; continued at next declaration"
  ],
  "records": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:2789:5",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "src/compiler/program.ts:2789:5",
      "reason": "function signature or prologue failed; continued at next declaration",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
      "start": 137012,
      "end": 137454
    }
  ]
}
```

#### New signature 7

```json
{
  "signature": [
    "Boundary",
    "src/compiler/program.ts",
    2806,
    9,
    "failed statement; state rolled back; continued at next statement or declaration"
  ],
  "records": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:2805:5",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "src/compiler/program.ts:2806:9",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
      "start": 137595,
      "end": 137696
    }
  ]
}
```

#### New signature 8

```json
{
  "signature": [
    "Boundary",
    "src/compiler/program.ts",
    2810,
    9,
    "failed statement; state rolled back; continued at next statement or declaration"
  ],
  "records": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:2809:5",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "src/compiler/program.ts:2810:9",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
      "start": 137854,
      "end": 138085
    }
  ]
}
```

#### New signature 9

```json
{
  "signature": [
    "Boundary",
    "src/compiler/program.ts",
    2839,
    9,
    "failed statement; state rolled back; continued at next statement or declaration"
  ],
  "records": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:2838:5",
      "phase": "lowering",
      "kind": "Boundary",
      "where": "src/compiler/program.ts:2839:9",
      "reason": "failed statement; state rolled back; continued at next statement or declaration",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance",
      "start": 139252,
      "end": 139355
    }
  ]
}
```

#### New signature 10

```json
{
  "signature": [
    "Refused",
    "src/compiler/program.ts",
    2781,
    9,
    "overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics"
  ],
  "records": [
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:1500:1",
      "phase": "lowering",
      "kind": "Refused",
      "where": "src/compiler/program.ts:2781:9",
      "reason": "overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:1515:1",
      "phase": "lowering",
      "kind": "Refused",
      "where": "src/compiler/program.ts:2781:9",
      "reason": "overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:1516:1",
      "phase": "lowering",
      "kind": "Refused",
      "where": "src/compiler/program.ts:2781:9",
      "reason": "overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:2779:5",
      "phase": "lowering",
      "kind": "Refused",
      "where": "src/compiler/program.ts:2781:9",
      "reason": "overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:2784:5",
      "phase": "lowering",
      "kind": "Refused",
      "where": "src/compiler/program.ts:2781:9",
      "reason": "overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:2789:5",
      "phase": "lowering",
      "kind": "Refused",
      "where": "src/compiler/program.ts:2781:9",
      "reason": "overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:2805:5",
      "phase": "lowering",
      "kind": "Refused",
      "where": "src/compiler/program.ts:2781:9",
      "reason": "overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:2809:5",
      "phase": "lowering",
      "kind": "Refused",
      "where": "src/compiler/program.ts:2781:9",
      "reason": "overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance"
    },
    {
      "root_kind": "",
      "depth": 0,
      "measurement": "measured on a checker-rejected program",
      "unit": "src/compiler/program.ts:2838:5",
      "phase": "lowering",
      "kind": "Refused",
      "where": "src/compiler/program.ts:2781:9",
      "reason": "overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics",
      "text": "src/compiler/program.ts:2781:9: Adamic 0.1 refuses overload 1 of getDiagnosticsHelper parameter getDiagnostics cannot be served by implementation parameter getDiagnostics; make the implementation accept every value admitted by this overload, without mutable widening or bivariance"
    }
  ]
}
```

### Gone findings

| Kind | Baseline location | Reason | Phases |
|---|---|---|---|
| Boundary | `src/compiler/core.ts:406:21` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/core.ts:409:21` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1127:9` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1128:9` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1129:9` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1130:9` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1133:9` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1517:5` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1522:5` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1532:5` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1579:5` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1587:5` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1609:9` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1654:9` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1671:9` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1726:5` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1783:9` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1816:13` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1858:9` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| Boundary | `src/compiler/program.ts:1877:5` | failed statement; state rolled back; continued at next statement or declaration | lowering |
| NotYet | `src/compiler/core.ts:406:47` | a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | lowering |
| NotYet | `src/compiler/core.ts:409:48` | a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | lowering |
| NotYet | `src/compiler/program.ts:1127:78` | an optional chain longer than one step | lowering |
| NotYet | `src/compiler/program.ts:1128:51` | reading mode | lowering |
| NotYet | `src/compiler/program.ts:1129:32` | reading key | lowering |
| NotYet | `src/compiler/program.ts:1130:14` | reading result | lowering |
| NotYet | `src/compiler/program.ts:1133:26` | reading result | lowering |
| NotYet | `src/compiler/program.ts:1517:91` | a narrowed union member whose object tag cannot be checked with typeof; keep differently held object kinds in separately typed variables | lowering |
| NotYet | `src/compiler/program.ts:1532:11` | a function value that captures the variable its own initializer declares | lowering |
| NotYet | `src/compiler/program.ts:1622:58` | a spread that adds a field the source doesn't have | lowering |
| NotYet | `src/compiler/program.ts:1656:17` | an array of T | lowering |
| NotYet | `src/compiler/program.ts:1673:17` | a readonly T[] where readonly (string \| FileReference)[] goes, with no representation Adamic can name for the stored elements (T stored, read as string \| FileReference) | lowering |
| NotYet | `src/compiler/program.ts:1733:9` | a generic function as a value | lowering |
| NotYet | `src/compiler/program.ts:1816:18` | a dictionary member narrowed to an array without an element-checked view | lowering |
| NotYet | `src/compiler/program.ts:1938:9` | a generic function as a value | lowering |
| NotYet | `src/compiler/program.ts:4632:54` | a SpreadElement | lowering |
| NotYet | `src/compiler/program.ts:4654:54` | a SpreadElement | lowering |
| NotYet | `src/compiler/program.ts:4668:93` | a SpreadElement | lowering |
| NotYet | `src/compiler/program.ts:4686:128` | a SpreadElement | lowering |
| Refused | `src/compiler/program.ts:1522:5` | a placeholder reset of a parameter | lowering |
| Refused | `src/compiler/program.ts:1579:31` | a generic instantiated at void whose body reads the value: parseConfigHostFromCompilerHostLike's T = void (read from the signature), read at src/compiler/program.ts:5106:54 as a binding's initializer | lowering |
| Refused | `src/compiler/program.ts:2796:16` | a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | refusal_scan |
| Refused | `src/compiler/program.ts:2829:16` | a value of type DiagnosticWithLocation seen as Diagnostic, which can write SourceFile \| undefined where SourceFile is read | refusal_scan |
| Refused | `src/compiler/utilities.ts:10012:105` | overload 1 of getSupportedExtensionsWithJsonIfResolveJsonModule parameter supportedExtensions cannot be served by implementation parameter supportedExtensions | lowering |

## Setup and validation

Fresh detached worktrees were created for both exact commits in an isolated scratch bare repository that borrows existing objects read-only. The shared checkout was not modified. Each worktree ran `git submodule update --init --recursive --depth 1` and `npm ci` in `stage3/api`. Both were clean when measured.

TypeScript was cloned with `--depth 1 --branch v6.0.3` and verified at `050880ce59e30b356b686bd3144efe24f875ebc8`. `ADAMIC_TYPESCRIPT_SOURCE` pointed to it. `GOTMPDIR`, `TMPDIR`, `GOCACHE`, `GOMODCACHE`, npm cache and stage3 caches were set to scratch for builds and measurements. Initial input scratch on memory-backed `/tmp` was moved to disk before measurement, with a symlink preserving its original paths.

Tools: `go version go1.27.1 linux/amd64`, Node `v24.19.0`. Both adapted trees contain 79 compiler `.ts` files. Compiler/census implementation diff between the commits: empty. Provenance and complete adapted compiler source hashes are retained in local scratch.

For each tree (using its own checkout and overlay):

```sh
source /workspace/program-full-census-scratch/env.sh
stage3/apply.sh --build-product "$SCRATCH/$NAME-adapted"
python3 stage3/census/latent/make_overlay.py "$PWD" "$SCRATCH/$NAME-overlay"
gofmt -w "$SCRATCH/$NAME-overlay/"*.go
go build -buildvcs=false -overlay="$SCRATCH/$NAME-overlay/overlay.json" \
  -o "$SCRATCH/$NAME-census" ./stage3/census/latent/tool
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
LATENT_ONLY_FILE="$SCRATCH/$NAME-adapted/src/compiler/program.ts" \
CENSUS_MODULE_FORMAT=CommonJSBundle \
LATENT_TIMES="$SCRATCH/$NAME-times.jsonl" \
LATENT_PROGRESS_FILE="$SCRATCH/$NAME-progress.json" \
  "$SCRATCH/$NAME-census" "$SCRATCH/$NAME-adapted/src/compiler" "$SCRATCH/$NAME.jsonl"
```

The sequential runner launches each command in its own process group, polls `wait4(WNOHANG)` every 0.2 seconds, sends SIGTERM at 6,000 seconds and SIGKILL after another 30 seconds if needed, records that child’s `ru_maxrss`, and stops on timeout or nonzero exit. Neither signal was needed. Exact commands and resource records are retained in local scratch. No speculative mode, mutants, source reduction, or scan-only wrapper was enabled for either real census. No IR or executable output was produced. The source-directory argument loads the full compiler project; the exact file filter limits full lowering observations to program.ts.

Both documented legacy `audit.py` runs passed. Both documented `full_audit.py` runs failed the same assertion: the synthetic `three` unit’s lowering Refused reasons were `{'var'}`, while the audit expected `{'a string as a condition', 'a number as a condition', 'var'}`. The audit and compiler were left unchanged. Audit logs and complete synthetic streams are retained in local scratch. This failure limits validation of the measurement overlay; these results are observations, not a fresh acceptance ruling.

Observed checker headers: base: `checker_rejected=false`, 0 diagnostics, change: `checker_rejected=false`, 0 diagnostics. The overlay’s static measurement label says “measured on a checker-rejected program”; the actual header flags above are the observed checker state and must not be replaced by that label.

Both streams contain exactly one selected-file observation and a final `relation_report` trailer. The comparison audit independently reproduced all new and gone signatures using zero-context `git diff` line mapping, verified byte-equal mapped lines, target removal, stream completion and nonempty phases; the independent audit script and log are retained in local scratch. Baseline unit statuses: 235 attempted, 13 deferred generic instantiation; adapted: 238 attempted, 12 deferred. Raw finding records: 1,625 baseline, 648 adapted. Isolated lowering, recovery boundaries, deferred generics and omitted backend/ownership passes retain the limitations described in the latent README.

## Local evidence and provenance

The report-only push contains this file. Full raw streams, timing streams, comparison and line maps, synthetic audit results, build/adaptation logs, source hashes and runner scripts remain in `/workspace/program-full-census-scratch`. Both actual streams have zero checker diagnostics, and no final result was inferred from a partial stream.

base measurement binary SHA-256: `00ac2bda66cd0283464e62a8f521f7a4cf31443e5ce220856aca54e906cab803`.

base adapted program.ts SHA-256: `0ed189a95d2204a4c8a0e7bd6602d09cd85e1705ea9495f224905a24f16f3688`.

base recursive submodules:

```text
5360753e19e4417675a90e92a1823bea66b13918 cohere (5360753e)
 d92d9bfee114c80be2c375d72edae966176e3a4f cohere/TypeScript (d92d9bfe)
```

change measurement binary SHA-256: `f75aaf0bc503e972a72bf08a0934d6005c33357c0628f1bddba5e6489d13801a`.

change adapted program.ts SHA-256: `73ce6c78ec4f0ea22b2f729a07bcf48797f1edca46a3dce654c6942e51ebc41f`.

change recursive submodules:

```text
5360753e19e4417675a90e92a1823bea66b13918 cohere (5360753e)
 d92d9bfee114c80be2c375d72edae966176e3a4f cohere/TypeScript (d92d9bfe)
```
