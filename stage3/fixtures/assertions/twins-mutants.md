# Item 8 stage 2: assertions twins

Base: `c85076c84741e66d7fe01c6596c6354e85570d83` (`typescript/stage3-fixtures-accept-ts`). Shape follows `d579ff21e10791f78ab269628849513f7db5a39c` and its v2/TWINS.md, with the user's subsequent evidence ruling.

The directory matched the three requested fixtures; the other seventeen remain unchanged. All three were moved with `git mv`, with unchanged source bytes. No Go logic or expected output changed, and no tests were added; Go test edits change path extensions only.

## Comparison

Each outcome records exit, stdout and stderr; `\n` is an LF byte. Before is `.a`, after is `.ts`. N and J mean native and JavaScript. Equal identifiers mean equal exits and byte-identical stdout/stderr.

| File (now .ts) | Native before | Native after | JavaScript before | JavaScript after | Identical | Type predicate | Non-null assertion | Definite assignment |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 11_parenthesized_kind.ts | N1 | N1 | J1 | J1 | yes | no | no | no |
| 13_flag_downcast.ts | N2 | N2 | J2 | J2 | yes | no | no | no |
| 20_type_flag_mask.ts | N3 | N3 | J3 | J3 | yes | no | no | no |

No non-twin was found. Every native outcome also equals its JavaScript outcome.

| Outcomes | Exit | Stdout (JSON string) | Stderr (JSON string) |
| --- | --- | --- | --- |
| N1, J1 | 0 | `"110\n110\n"` | `""` |
| N2, J2 | 0 | `"8\n0\n"` | `""` |
| N3, J3 | 0 | `"2\n0\n"` | `""` |

## Method and readers

Setup: `export GOPROXY='https://proxy.golang.org|direct'; timeout --kill-after=60s 1800 bash cloud/setup.sh > /tmp/setup.log 2>&1` (exit 0). Commands source `/workspace/adamic-tools/env.sh`. The pinned upstream source is `/tmp/item8-typescript` at `050880ce59e30b356b686bd3144efe24f875ebc8`; locked API dependencies are installed.

Before any rename, `/tmp/item8-capture` loaded and lowered each `.a`, built `native.C` through `native.Build` with `native.Options{}`, executed the binary, and executed `javascript.JavaScript` through `node --disable-warning=ExperimentalWarning oracle/node.mjs`. These are the APIs and invocation used by `releasedUncached` and `onJavaScriptBackend`. The same capture binary repeated the runs after the renames. Before/after JSON records are `/tmp/item8-assertions-before.json` and `/tmp/item8-assertions-after.json`; logs have the same prefixes and `.log` suffix. All six backend records compare equal after normalizing only the fixture extension in their keys. No capture source or binary is committed.

Whole-repository `git grep -n` searches for each basename are saved as `/tmp/item8-<basename>-readers.log`. Live Go paths are in `checked_views_flag_downcast_test.go` and `checked_views_v2_migration_test.go`. Status rows, manifest rows, extraction specifications in `make-fixtures.cjs`, and fixture-union identities change only `.a` to `.ts`. Dated scouts, verification records, source-provenance comments, reports and logs retain their historical paths. The existing Python audit is already failing on baseline: it compares three historical logs with `native: null` against current Compiles statuses; this is not repaired here.

No test reads manifest.json, so its path corrections need no mutant under the user's ruling. The fixture-union JSON contains identities, not runtime outcomes, and has no test reader; its names are updated to match the migrated paths.

No counts records or count registrations exist for these three fixtures, so there are no counts renames, count numbers to diff, or counts mutant.

## Door mutants

Each mutation was applied alone and restored before the full suites. Both commands exit 1, with zero passing and four failing test events (including parent tests and the Node leaf). No mutant changes enter the commit.

### Status path restored to .a; fixture remains .ts

Only `11_parenthesized_kind.ts` in `status.json` was changed back to `.a`.

Command: `go test ./stage3/fixtures -run '^TestFixturesAssertions$/^assertions$/^11_parenthesized_kind[.]a$' -count=1 -v > /tmp/item8-assertions-mutant-status.log 2>&1`.

Witness: `TestFixturesAssertions/assertions/11_parenthesized_kind.a` (and its `/node` leaf).

```text
fixtures_test.go:382: recorded Node byte comparison failed: got exit=70 stdout="" stderr="adamic: panic: Error [ERR_MODULE_NOT_FOUND]: Cannot find module '/workspace/adamic/stage3/fixtures/assertions/11_parenthesized_kind.a' imported from /workspace/adamic/oracle/node.mjs\n"; want exit=0 stdout="110\n110\n" stderr=""
fixtures_test.go:405: unexpected stage 0 error: load: no file at /workspace/adamic/stage3/fixtures/assertions/11_parenthesized_kind.a
```

### Fixture restored to .a; status remains .ts

Only the fixture was moved back with `git mv`; status still named `.ts`.

Command: `go test ./stage3/fixtures -run '^TestFixturesAssertions$/^assertions$/^11_parenthesized_kind[.]ts$' -count=1 -v > /tmp/item8-assertions-mutant-file.log 2>&1`.

Witness: `TestFixturesAssertions/assertions/11_parenthesized_kind.ts` (and its `/node` leaf).

```text
fixtures_test.go:382: recorded Node byte comparison failed: got exit=70 stdout="" stderr="adamic: panic: Error [ERR_MODULE_NOT_FOUND]: Cannot find module '/workspace/adamic/stage3/fixtures/assertions/11_parenthesized_kind.ts' imported from /workspace/adamic/oracle/node.mjs\n"; want exit=0 stdout="110\n110\n" stderr=""
fixtures_test.go:405: unexpected stage 0 error: load: no file at /workspace/adamic/stage3/fixtures/assertions/11_parenthesized_kind.ts
```

## Validation

Full fixture package:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/item8-typescript go test ./stage3/fixtures/... -count=1 -json -timeout=60m > /tmp/item8-assertions-fixtures.jsonl 2> /tmp/item8-assertions-fixtures.stderr
```

PASS, exit 0: one package; 34 top-level tests and 742 subtests passed (776 passing test events), one top-level test skipped, zero failures. Package elapsed: 115.923 seconds. Stderr is empty.

Full oracle package:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/item8-typescript go test ./internal/oracle/... -count=1 -json -timeout=60m > /tmp/item8-assertions-oracle.jsonl 2> /tmp/item8-assertions-oracle.stderr
```

PASS, exit 0: one package; 1777 top-level tests and 6153 subtests passed (7930 passing test events). 442 top-level tests and 99 subtests skipped (541 skip events). Zero failures. Package elapsed: 1179.245 seconds. Stderr is empty.

Final checks verified unchanged source bytes for all three fixtures, unchanged status outcomes and all other manifest/union fields, path-only Go edits, no counts records for these names, and no changes to the other seventeen fixture sources. `git diff --check` passed.

## Existing audit finding

On the unmodified base, `python3 stage3/fixtures/assertions/audit.py stage3/census/data /tmp/item8-typescript` exits 1 at `audit.py:47` with `AssertionError: native differs from Node` (log `/tmp/item8-assertions-audit-baseline.log`). The first failing row is `11_parenthesized_kind.a`; evaluating the same comparison for every Compiles row identifies these three stale records:

| Historical log fixture | Saved native | Saved Node stdout = current status stdout | Node stderr | Node exit |
| --- | --- | --- | --- | --- |
| 11_parenthesized_kind.a | `null` | `"110\n110\n"` | `""` | 0 |
| 13_flag_downcast.a | `null` | `"8\n0\n"` | `""` | 0 |
| 20_type_flag_mask.a | `null` | `"2\n0\n"` | `""` | 0 |

The old builds exited 1: the first reported `Adamic 0.1 refuses a cast the runtime can't check`; the other two reported `Adamic 0.1 refuses a number as a condition`. No native run was saved. This is a stale-log mismatch, not a current native/JavaScript disagreement; current before/after outcomes are recorded above. Historical logs and the audit implementation remain unchanged.
