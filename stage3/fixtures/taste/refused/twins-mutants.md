# Item 8 stage 2: taste refused-directory twin

Base: assertions commit `16d924bade1c2619188aba59d7a304e0c92cf396`. Branch: `typescript/item8-twins-taste`. Shape follows wave 1 (`d579ff21`) and the subsequent evidence ruling.

The directory contains the one requested source fixture and its historical `.refused` record. The source was moved with `git mv`, with identical bytes; the historical refusal record remains unchanged. No Go logic or expected output changed, and no tests were added; the Go test edit changes only the source extension in its path.

## Comparison

Each outcome records exit, stdout and stderr; `\n` is an LF byte. Before is `.a`, after is `.ts`. Equal outcome identifiers mean equal exits and byte-identical stdout/stderr.

| File (now .ts) | Native before | Native after | JavaScript before | JavaScript after | Identical | Type predicate | Non-null assertion | Definite assignment |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 21_truthy_loops.ts | N1 | N1 | J1 | J1 | yes | no | no | no |

No non-twin was found. Native and JavaScript outcomes also equal each other.

| Outcomes | Exit | Stdout (JSON string) | Stderr (JSON string) |
| --- | --- | --- | --- |
| N1, J1 | 0 | `"true\ntrue\nouter\nmissing\ntrue\n"` | `""` |

## Method and readers

Setup: `export GOPROXY='https://proxy.golang.org|direct'; timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1` (exit 0). Commands source `/workspace/adamic-tools/env.sh`. The pinned upstream source is `/tmp/item8-typescript` at `050880ce59e30b356b686bd3144efe24f875ebc8`; locked API dependencies are installed.

Before the rename, `/tmp/item8-capture` loaded and lowered the `.a`, built `native.C` through `native.Build` with `native.Options{}`, executed the binary, and executed `javascript.JavaScript` through `node --disable-warning=ExperimentalWarning oracle/node.mjs`. These match `releasedUncached` and `onJavaScriptBackend`. The same capture binary ran after the rename. Records: `/tmp/item8-taste-before.json` and `/tmp/item8-taste-after.json`; logs use the same prefixes and `.log` suffix. Both backend records compare equal after normalizing only their fixture-key extension. No capture source or binary is committed.

A whole-repository `git grep -n 21_truthy_loops` is saved in `/tmp/item8-21_truthy_loops-readers.log`. The live Go path is in `checked_views_v2_migration_test.go`; it changes only `.a` to `.ts`. The status and fixture-union paths change only their extensions. The manifest, source-site sample list, and their generators (`sample.cjs`, `mutants.py`) now name the actual `refused/21_truthy_loops.ts` relative path, correcting their pre-existing omission of `refused/` as well as the extension. All other metadata, site counts, sample quotas, source spans and recorded outcomes remain unchanged. Dated scouts, verification evidence, native closure-convention records, logs, `evidence.json`, `mutants.json`, and the historical `.refused` diagnostic retain their old paths. Historical audit failures are not repaired.

The manifest correction also fixes the stale root path by naming the existing `refused/21_truthy_loops.ts` fixture.

No test reads manifest.json, so its path correction needs no mutant under the user's ruling. The fixture-union and sample list contain identities and source provenance, not runtime outcomes; their paths are updated without changing their counts or other fields.

No counts record or count registration exists for this fixture, so there is no counts rename, no count numbers to diff, and no counts mutant.

## Door mutants

Each mutation was applied alone and restored before the full suites. Each command exits 1, with zero passing and four failing test events (parents and Node leaf included). No mutant changes enter the commit.

### Status path restored to .a; fixture remains .ts

Changed only `refused/21_truthy_loops.ts` back to `.a` in `taste/status.json`.

Command: `go test ./stage3/fixtures -run '^TestFixturesTaste$/^taste$/^refused$/^21_truthy_loops[.]a$' -count=1 -v > /tmp/item8-taste-mutant-status.log 2>&1`.

Witness: `TestFixturesTaste/taste/refused/21_truthy_loops.a` (and its `/node` leaf).

```text
fixtures_test.go:382: recorded Node byte comparison failed: got exit=70 stdout="" stderr="adamic: panic: Error [ERR_MODULE_NOT_FOUND]: Cannot find module '/workspace/adamic/stage3/fixtures/taste/refused/21_truthy_loops.a' imported from /workspace/adamic/oracle/node.mjs\n"; want exit=0 stdout="true\ntrue\nouter\nmissing\ntrue\n" stderr=""
fixtures_test.go:405: unexpected stage 0 error: load: no file at /workspace/adamic/stage3/fixtures/taste/refused/21_truthy_loops.a
```

### Fixture restored to .a; status remains .ts

Moved only the fixture back with `git mv`, keeping status at `.ts`.

Command: `go test ./stage3/fixtures -run '^TestFixturesTaste$/^taste$/^refused$/^21_truthy_loops[.]ts$' -count=1 -v > /tmp/item8-taste-mutant-file.log 2>&1`.

Witness: `TestFixturesTaste/taste/refused/21_truthy_loops.ts` (and its `/node` leaf).

```text
fixtures_test.go:382: recorded Node byte comparison failed: got exit=70 stdout="" stderr="adamic: panic: Error [ERR_MODULE_NOT_FOUND]: Cannot find module '/workspace/adamic/stage3/fixtures/taste/refused/21_truthy_loops.ts' imported from /workspace/adamic/oracle/node.mjs\n"; want exit=0 stdout="true\ntrue\nouter\nmissing\ntrue\n" stderr=""
fixtures_test.go:405: unexpected stage 0 error: load: no file at /workspace/adamic/stage3/fixtures/taste/refused/21_truthy_loops.ts
```

## Validation

Full fixture package:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/item8-typescript go test ./stage3/fixtures/... -count=1 -json -timeout=60m > /tmp/item8-taste-fixtures.jsonl 2> /tmp/item8-taste-fixtures.stderr
```

PASS, exit 0: one package; 34 top-level tests and 742 subtests passed (776 passing test events), one top-level test skipped, zero failures. Package elapsed: 146.492 seconds. Stderr is empty.

Full oracle package:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/item8-typescript go test ./internal/oracle/... -count=1 -json -timeout=60m > /tmp/item8-taste-oracle.jsonl 2> /tmp/item8-taste-oracle.stderr
```

PASS, exit 0: one package; 1777 top-level tests and 6153 subtests passed (7930 passing test events). 442 top-level tests and 99 subtests skipped (541 skip events). Zero failures. Package elapsed: 910.440 seconds. Stderr is empty.

Final checks verified identical fixture bytes, identical status outcomes, unchanged sample counts/quotas/source spans, path-only metadata changes, an extension-only Go edit, an unchanged historical refusal record, and no counts record for this name. `git diff --check` passed. No transient capture source, mutant edit, dependency installation or test log is committed.
