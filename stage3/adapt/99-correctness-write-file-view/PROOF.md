# Proof — #b8ffarm

Source base: 8e3a784aa5a601f57342c07b68f145fe56252c21.
TypeScript: v6.0.3, 050880ce59e30b356b686bd3144efe24f875ebc8.
All measurements use private scratch copies; no shared checkout is adapted.

| Check | Observation |
|---|---|
| Whole-owner contract | LF/CRLF; drift, half states, missing, duplicate, overload, moved, parse-error and mixed-ending refusal; no-write replay and restoring replay; bytes/mtimes preserved |
| Node witness | Real before/after owner bodies; receiver-sensitive calls, getter, later reassignment, captured identity and own .call agree |
| JavaScript | 604 source files, 1,208 comparison pairs / 2,416 emits, comments retained and removed; only patchWriteFileEnsuringDirectory differs, exactly by the alias and redirected read |
| Main / candidate full sys.ts census | 59 → 58 unique findings; exactly one gone, zero new and zero changed |
| Gone finding | Refused, refusal_scan, sys.ts:1385:31: a method read as a value (writeFile would lose its object, and this with it) |
| Restoring mutant | 59 findings, exactly main again; the target returns in refusal_scan |
| Stock tsc 6.0.3 | Actual CLI, all 79 roots, exit 0 for main, candidate and mutant |
| adamic types | All sorted 79 roots, exit 0 for main, candidate and mutant |
| Upstream declarations | Separate npm ci then node_modules/.bin/hereby dts on main/candidate; both files byte-identical, exact diffs empty |
| Carried lane count | 2,200 captures and 4,627 callback calls on each side; source hashes exactly match the counted main/candidate in typescript/price-sys-public at 3ecc182ad4c79a8783e4b673c01d977d046e515e |

The exact declaration SHA-256 values are:

- typescript.d.ts: f03adcf915df559d616fcafd75181d3423eaa0f272237482cedbd025525c9516
- tsserverlibrary.d.ts: cee9e9bf9a8f67388ee878afd5e172189421bae6625076b774d68563d036fbfd

Every census is full single-file sys.ts with LATENT_FULL=1,
LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE naming the whole file, and
CENSUS_MODULE_FORMAT=CommonJSBundle. Each header has zero diagnostics and
checker_rejected:false; the meter's historical generic label does not override
the header. Both no-output guards remain enabled. This is not successful native
compilation of the remaining frontier.

The landing commit is published only after unmodified stage3/apply.sh succeeds
through the complete adaptation sequence and the whole, unfiltered
stage3/lane/run.sh returns PASS: 106,366 passing, zero pending, and only the
sanctioned public-API baseline row. No test filter or worker/count change is
allowed. The final source must match the measured candidate byte for byte;
replay of this stage must still write zero files.

Full streams, every delta with phase/attempt context, source and emit hashes,
the restoring-mutant observation, empty declaration diffs, type logs, the
carried counter traces, whole-apply/lane receipts and the final commit identity
belong on [the evidence branch](https://github.com/system-inc/adamic/tree/typescript/write-file-view-1385-evidence/stage3/census/write-file-view-1385).
The adaptation directory contains no measured output products or public rows.
