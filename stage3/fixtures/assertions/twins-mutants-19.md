# Item 8 stage 2: four remaining checked-downcast twins

Base: main at `4ef2f6b4afd71ca0c33b4349fe598feb47192ee1`. Read commits `10a46c6b52205d827250ec4b50371920254b82ba`, `8344852add763f2c55c6201d3a877e67ca36b068`, and `ddc0e7d65ec9ce11c8b7371ffe391d16a9098b8b` in full before migration. This change follows their source, companion, comparison, and mutant shape.

All four requested fixtures moved with `git mv`; their source bytes, including provenance comments, are identical to the base. `lane1/optional-read-control.a` stays unchanged: the lane1 precedent excluded both optional fixtures rather than requiring them to move together, and this control has no checked downcast. Every other source keeps its extension.

## Comparison

Before is `.a`, after is `.ts`. Each repeated identifier means equal exits and byte-identical stdout and stderr. N and J denote native and JavaScript.

| File (now .ts) | Native before | Native after | JavaScript before | JavaScript after | Identical |
| --- | --- | --- | --- | --- | --- |
| assertions/19_identifier_kind.ts | N1 | N1 | J1 | J1 | yes |
| lane1/optional-read.ts | N2 | N2 | J2 | J2 | yes |
| lane4b/fixtures/optional-absent.ts | N3 | N3 | J3 | J3 | yes |
| lane4b/fixtures/optional-receiver.ts | N4 | N4 | J4 | J4 | yes |

Every native result also equals its JavaScript result. None uses a type predicate, non-null assertion, or definite assignment.

| Outcomes | Exit | Stdout (JSON string) | Stderr (JSON string) |
| --- | --- | --- | --- |
| N1, J1 | 0 | `"name\nname\nmissing\n"` | `""` |
| N2, J2 | 0 | `"name\nmissing\n"` | `""` |
| N3, J3 | 0 | `"absent\n"` | `""` |
| N4, J4 | 0 | `"ok\nabsent\n"` | `""` |

## Method and readers

Required setup: `export GOPROXY='https://proxy.golang.org|direct'`, then `timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1`. PASS, exit 0, 331.354 seconds. Build and test commands source `/workspace/adamic-tools/env.sh`. Pinned upstream TypeScript is `/workspace/scratch/typescript`, commit `050880ce59e30b356b686bd3144efe24f875ebc8`; API dependencies were installed with `npm ci --prefix stage3/api`.

A temporary capture program calls `load.Load`, `lower.Lower`, `native.C`, `native.Build` with `native.Options{}`, and `javascript.JavaScript`. It executes native binaries directly and JavaScript with `node --disable-warning=ExperimentalWarning oracle/node.mjs <generated program.mjs>`, matching the oracle helpers. The same capture binary runs before and after migration. All eight backend records compare equal with `cmp`: `/workspace/scratch/m8dv-before.json` and `/workspace/scratch/m8dv-after.json`. Capture logs use the same prefixes with `.log`. No capture utility enters a commit.

Whole-repository basename searches are saved in `/workspace/scratch/m8dv-readers.log`. The assertion's manifest, status, extraction specification, and fixture-union identities change only `.a` to `.ts`. The status outcome and Node result are unchanged. Manifest and fixture-union rows contain provenance and identities, not additional runtime outcomes. Historical logs, shard evidence, reports, and a-check records retain their observed paths, as in the precedents.

No counts records or count registrations exist for any of these four fixtures. The optional control's counts remain untouched. Interface holding tests already use `interfaceFixturePath`, directly or through `interfaceFixture`, so no Go test needs editing. No path under internal/lower, internal/native, or internal/javascript changes.

## Door mutant

Only `19_identifier_kind.ts` in `status.json` was temporarily changed back to `19_identifier_kind.a`; the fixture remained `.ts`. All other fields stayed identical. Command:

```sh
go test ./stage3/fixtures -run '^TestFixturesAssertions$/^assertions$/^19_identifier_kind[.]a$' -count=1 -v > /workspace/scratch/m8dv-mutant-status.log 2>&1
```

RED, exit 1. Witness: `TestFixturesAssertions/assertions/19_identifier_kind.a` and its Node leaf. Zero passing and four failing test events.

```text
fixtures_test.go:382: recorded Node byte comparison failed: got exit=70 stdout="" stderr="adamic: panic: Error [ERR_MODULE_NOT_FOUND]: Cannot find module '/workspace/adamic/stage3/fixtures/assertions/19_identifier_kind.a' imported from /workspace/adamic/oracle/node.mjs\n"; want exit=0 stdout="name\nname\nmissing\n" stderr=""
fixtures_test.go:405: unexpected stage 0 error: load: no file at /workspace/adamic/stage3/fixtures/assertions/19_identifier_kind.a
```

Restored the status filename to `.ts`, then ran:

```sh
go test ./stage3/fixtures -run '^TestFixturesAssertions$/^assertions$/^19_identifier_kind[.]ts$' -count=1 -v > /workspace/scratch/m8dv-restored-status.log 2>&1
```

GREEN, exit 0. All six test events passed, including Node, native, and stage0. No mutant edit enters the commit. This status row is the only recorded runtime outcome moved.

## Validation

Before migration, `go test ./stage3/fixtures ./internal/oracle -run '^(TestFixturesAssertions|TestCheckedViewOptionalReadAgreesWithNode|TestCheckedViewObjectPrimitiveSource)$' -count=1 -json -timeout=20m` passed with exit 0. Logs: `/workspace/scratch/m8dv-before-holding.jsonl` and `.stderr`. The object primitive test holds both lane4b files, so these three top-level tests hold all four fixtures.

Source and companion audits passed: all four source files equal the base byte for byte, and all four companion files equal the base with only the requested extension substitutions. `git diff --check` passed.

## Stage 2 scratch merge

The detached scratch checkout started at `b94fae07ac6d63543c000ef543a91fa693d7c3a2` and merged migration commit `ab1d86903` cleanly, producing local merge `05a9902641644a8d4a67cdb9006ebda613cdb933`. The merge is not pushed. Dependency installations are reused from the main checkout. Merge log: `/workspace/scratch/m8dv-stage2-merge.log`.

Holding command:

```sh
go test ./stage3/fixtures ./internal/oracle -run '^(TestFixturesAssertions|TestCheckedViewOptionalReadAgreesWithNode|TestCheckedViewObjectPrimitiveSource)$' -count=1 -json -timeout=20m > /workspace/scratch/m8dv-stage2-holding.jsonl 2> /workspace/scratch/m8dv-stage2-holding.stderr
```

Exit 1 overall: oracle PASS (38.376 seconds), fixtures FAIL (31.682 seconds). All four migrated fixture subtests pass, including Node, native, and stage0 for `19_identifier_kind.ts`, both lane4b optional subtests, and lane1's checked read and unchanged control. The aggregate `TestFixturesAssertions` fails only on untouched `09_interface_kind.a`: stage 2 changes its recorded `NotYet` into `Refused` before reaching its array-element gap.

```text
fixtures_test.go:472: gap changed: update status.json and check the native output; got {Outcome:Refused What:stage3/fixtures/assertions/09_interface_kind.a:15:10: Adamic 0.1 refuses a downcast from Node to ObjectLiteralExpression in .a whose checks no proof erases; declare Node as a union with ObjectLiteralExpression as a member and narrow it, as .a writes !: if (expression.kind !== 211) panic('why'), or move the file to .ts, where the cast stays a checked cast Stdout:}; recorded {Outcome:NotYet What:stage3/fixtures/assertions/09_interface_kind.a:15:9: stage 0 can't lower checked view union arm awaits views-v3: array element kind yet Stdout:}
```

An isolated rerun of the migrated assertion passed, exit 0:

```sh
go test ./stage3/fixtures -run '^TestFixturesAssertions$/^assertions$/^19_identifier_kind[.]ts$' -count=1 -json -timeout=20m > /workspace/scratch/m8dv-stage2-target-assertion.jsonl 2> /workspace/scratch/m8dv-stage2-target-assertion.stderr
```

The unrelated `.a` fixture and its recorded outcome remain unchanged. This prevents claiming that every aggregate holding test passes on stage 2; no additional fixture migration or compiler change is included.

## Full main-tip validation

Ran the full packages used by all three earlier waves:

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript go test ./stage3/fixtures/... ./internal/oracle/... -count=1 -json -timeout=60m > /workspace/scratch/m8dv-full.jsonl 2> /workspace/scratch/m8dv-full.stderr
```

All three holding tests PASS: `TestFixturesAssertions`, `TestCheckedViewOptionalReadAgreesWithNode`, and `TestCheckedViewObjectPrimitiveSource`. Every migrated fixture and the unchanged optional control pass.

The full fixtures package PASS, 233.834 seconds: 35 top-level and 763 subtests passed (798 passing events), one top-level skip, zero failures.

The full oracle package finished in 1177.773 seconds: 2066 top-level and 6578 subtests passed (8644 passing events), 665 top-level and 107 subtests skipped (772 skip events), one top-level failure. Overall command exit 1. Stderr is empty. The sole failure was an environment prerequisite, unrelated to the fixture migration:

```text
TestStdoutColumnsRefusesWASI
library_node_stdout_columns_test.go:246: want typed, named WASI refusal, got native: wasm32-wasi requires WASI_SYSROOT
```

The required default setup does not install its optional WASI SDK. Installed the repository's pinned SDK with `timeout --kill-after=60s 1800 bash cloud/setup.sh --wasi-sdk > /workspace/scratch/setup-wasi.log 2>&1` (PASS, exit 0, 24.401 seconds), then sourced `/workspace/adamic-tools/env.sh` again and reran the sole failed test:

```sh
go test ./internal/oracle -run '^TestStdoutColumnsRefusesWASI$' -count=1 -json -timeout=20m > /workspace/scratch/m8dv-wasi-rerun.jsonl 2> /workspace/scratch/m8dv-wasi-rerun.stderr
```

PASS, exit 0, 0.453 seconds. The original full-run failure remains recorded; the full suite was not rerun. No tests, expected outcomes, or compiler code were changed to address either this setup prerequisite or the unrelated stage 2 refusal.

Final scope and whitespace checks passed. Both commits touch only paths retained in the final diff. Logs, temporary capture utilities, dependency installations, and scratch merge changes are not included.
