# Stage3 S3C2 meter report

The requested full-mode meter measurement **timed out (`exit=124`) after 7,200 seconds**. It did not produce a completed findings ledger or generated meter report. Consequently, the requested observation rows and quotation cannot be supplied from this run; their absence from the available artifacts is not evidence that the blockers are gone.

## Revision and invocation

- Repository: `https://github.com/system-inc/adamic.git`.
- Fetched with `git fetch origin main`.
- Recorded `origin/main`: `e42d56d2bf61ea8bc5a253597e89687494c97e8d`.
- `git merge-base --is-ancestor 87028b0889b4431a08c31ca70bc294fa0fd746f5 e42d56d2bf61ea8bc5a253597e89687494c97e8d` returned **0**.
- Report branch: `proof/meter-s3c2-report`, created directly from that measured SHA.
- Loaded `/workspace/adamic-tools/env.sh` before execution. Exported `NODE_EXTRA_CA_CERTS=/etc/ssl/certs/ca-certificates.crt`, `npm_config_cafile=/etc/ssl/certs/ca-certificates.crt`, and `npm_config_strict_ssl=true`.
- Read `stage3/meter/README.md`. `STAGE3_METER_LATENT` was unset throughout; full latent mode was selected by default.

```sh
STAGE3_METER_COMPILER=per-ref STAGE3_METER_MAIN_REF=e42d56d2bf61ea8bc5a253597e89687494c97e8d timeout --kill-after=60s 7200 bash stage3/meter/twice-daily.sh > /tmp/stage3-meter.log 2>&1; echo exit=$?
```

Observed terminal result:

```text
exit=124
```

The measurement invocation started at approximately 2026-11-12 11:45:05 UTC and ended at 13:45:04 UTC. Run directory:

```text
/workspace/adamic/stage3/meter/runs/20261112T114505Z.0CMg4R
```

`compiler-mode.json` records `mode: per-ref`, `latent_mode: full`, both compiler commits equal to the measured SHA, and `measurement_sources: {main: main, area: main}`. All three measurement binaries built, and `main/compiler-status.txt` is empty. The box had 5 available processors and a CPU quota of 400000/100000 microseconds.

### Preparation failures before measurement

There were three script invocations in total, including two preparation failures. Neither failed invocation reached a census; only the final invocation measured the program. No additional measurement was launched after its timeout.

1. The first invocation exited 1 because `env.sh` sets `TMPDIR=/tmp/adamic-gate`, which did not exist: `mktemp: failed to create directory via template ‘/tmp/adamic-gate/stage3-meter.XXXXXX’: No such file or directory`. Created the configured temporary directory. Its log is preserved at `/tmp/stage3-meter-preflight.log`; its empty run directory is `stage3/meter/runs/20261112T113710Z.60obnZ`.
2. Run `stage3/meter/runs/20261112T113732Z.jD9XAE` exited 1 during overlay preparation: `go: cannot load module ../../../cohere/TypeScript/tsc listed in go.work file: open ../../../cohere/TypeScript/tsc/go.mod: no such file or directory`. The overlay helper runs from the launcher checkout, whose submodules were uninitialized. Initialized its recorded submodule pins without editing compiler sources. Logs remain in that run's `main/latent-overlay.log`, `/tmp/stage3-meter-overlay-preparation.log`, and `/tmp/stage3-meter-launcher-submodules.log`.

## Available measurement evidence

The README specifies compiler latent observations in `main/latent.jsonl.gz` and tsc-entry observations in `main/tsc/latent.jsonl.gz`, with the entry summary in `report.json` under `trees.main.tsc_entry.lowering_census`. The `area` label copies the main measurement on successful completion. Compression and report generation happen after all censuses finish.

At timeout:

| Artifact relative to the final run directory | State |
| --- | --- |
| `main/tsc/census.jsonl` | Present: 490 bytes, 2 ordinary census records |
| `main/tsc/latent.jsonl` | Present: 7,205 bytes, exactly 1 measurement-header record, no findings records |
| `main/tsc/latent-times.jsonl` | Present: 5,404 completed-unit timing records |
| `main/census.jsonl`, `main/latent.jsonl` | Not produced; the compiler-directory censuses had not started |
| `README.md`, `report.md`, `report.json` | Not produced |
| `area/` and compressed census streams | Not produced |

The entry latent header says `latent_mode: full`, `checker_rejected: false`, `diagnostics: []`, and `measurement: measured on a checker-clean entry-root program`. This establishes a checker-clean entry, not successful lowering. The ordinary entry census records report `NotYet` at `src/compiler/sys.ts:1654:13`, reason `node:inspector.Session.post`; that is outside the requested selection.

The last completed timing record is for `src/compiler/checker.ts:16421:5` (0.35716762 wall seconds; 0.3650699999980134 CPU seconds). Timing records identify attempts, not NotYet/Refused/error/panic observations, and cannot replace the missing findings ledger.

SHA-256 of retained artifacts:

```text
main/tsc/latent.jsonl        f3b7c258d6b75eeae8b6ed704bea602c91c2f604d358070c191039ce596a1dc5
main/tsc/census.jsonl        6c55b75062bcbd1e08506eb116f575cd4916eaf776f3a6739bde3f6f4bc11c95
main/tsc/latent-times.jsonl  6c1f219429ecbeca808166b3b7b95cd13dd4990205e005e2e9a2f108c710d3a4
```

## Requested observations

**Unavailable, not zero findings.** There are no persisted latent finding records from which to extract location, measured function body, whole reason text, and kind for either requested selection:

- Every NotYet, Refused, error, or panic observation located in `src/compiler/checker.ts:21600–21950` inclusive.
- Every observation naming `elaborateElementwise`, `elaborateArrayLiteral`, `elaborateObjectLiteral`, `elaborateJsxComponents`, `generateJsxChildren`, `generateJsxAttributes`, `generateLimitedTupleElements`, or `generateObjectLiteralElements` (the request lists `generateJsxAttributes` twice).

Whether any observation in that range still reads **“a view that generators of more than one function feed” is undetermined**. The incomplete, header-only stream cannot establish either presence or absence of that reason.

## Requested first five lines of the generated report

**Unavailable:** the final run did not create its `report.md`. There are no first five lines to quote. No older run's report or reconstructed report has been substituted.

Only this report is committed. Generated run artifacts and temporary logs remain uncommitted; no compiler or meter implementation changes were made.
