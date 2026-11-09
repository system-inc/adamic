# findAncestor compiler bar census

Task 448a51k. Report-only comparison on one adapted TypeScript 6.0.3 tree.

| Side | Compiler pin | First marked catalogue unit | Attempted unit | Location |
| --- | --- | ---: | ---: | --- |
| base | `c1e9379ef3ceffd00588fb40050e29ea1681a739` | 3842 | 3444 | `compiler/transformers/es2020.ts:1:1` (None) |
| head | `5cc5cee86331f52b085826205a425dcebc7f4654` | 3818 | 3420 | `compiler/transformers/typeSerializer.ts:132:1` (createRuntimeTypeSerializer) |

The compiler bar passes before the earlier onset: 0 changed refusal-scan rows there. Changes at or after onset are classified as relation-limit noise under the requested rule. This classification does not establish that every later change was caused by the limit.

## Method

Base `c1e9379ef3ceffd00588fb40050e29ea1681a739` was fetched explicitly by full SHA after the server rejected its abbreviation. Both named reference branches and main were fetched. Head `5cc5cee86331f52b085826205a425dcebc7f4654` has base as its sole parent. Its only changed paths are `internal/lower/overload_predicate_result.go` and its test; no `stage3/adapt` path changes. See [pins](pins.json) and fetch logs.

Read c6f6688c, `stage3/census/find-ancestor-63c11488/REPORT.md`, and copied its guarded own-pin entry-overlay, shared-input, parallel-run and separate checkpoint-ref method. No adaptations were carried. `source /workspace/adamic-tools/env.sh`, `bash cloud/setup.sh`, and `npm ci --prefix stage3/api` completed successfully. `ADAMIC_TYPESCRIPT_SOURCE` names the checkout at `050880ce59e30b356b686bd3144efe24f875ebc8`. Base `stage3/apply.sh` built one tree at `/workspace/scratch/bar-adapted`. Both sides use exactly those bytes; their hashes were checked again after the runs. See [apply log](apply.log), [patch set](patch-set.md), [source-hashes.json](source-hashes.json), and [binary provenance](provenance.json).

Each compiler was built from its own pin with the pinned shared checker and its own guarded latent plus entry overlay. Both ran first-error mode on `src/tsc/tsc.ts`, simultaneously, with `LATENT_ASSERT_NO_OUTPUT=1`, `LATENT_RELATION_PER_UNIT=1`, `GOMEMLIMIT=5GiB`, and `GOGC=70`. Whole stdout/stderr logs and streams are retained. The process wrapper reports the child exit code and cleans its process group; a 30-minute lack of progress terminates a stalled child. Unit and relation-query timings are separate evidence. No backend or native output is run.

Catalogue ordinals are one-based in the original stream file/unit order, including deferred declarations. Attempted ordinals count the per-unit relation probes only. Onset includes the unit whose post-lowering check first reports overflow. Every moved source row is classified by its recorded unit against the earlier catalogue onset. Full reasons drive the exact comparison; the requested lists use `reason[:110]`, sorted and unique. Truncation can hide differences, so exact rows and the literal list diff are both retained.

The pinned `stage3/census/pair/pair.py` expects source coordinates for every finding and does not accept the relation trailer at `where: program`. Its explicitly preserved pair-input copies omit only the trailer and each unit's observational `relation_overflow` field; all source findings, eligibility, units, and headers remain. The whole original streams drive onset classification and the separate run-end relation comparison. Pairing used the pinned stock TypeScript package. This projection changes no production or pair-tool source.

## Counts

| Moved exact finding rows | Removed from base | Added on head | Total |
| --- | ---: | ---: | ---: |
| head | 10 | 4 | 14 |
| relation-limit noise | 35 | 34 | 69 |
| relation-limit finding itself | 0 | 0 | 0 |

| Moved refusal-scan rows | Removed from base | Added on head |
| --- | ---: | ---: |
| head | 0 | 0 |
| relation-limit noise | 30 | 5 |
| relation-limit finding itself | 0 | 0 |

| Side | Files | Catalogue units | Probed units | Exact finding rows | Truncated list rows | Checker diagnostics |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| base | 81 | 4563 | 4059 | 6444 | 6444 | 0 |
| head | 81 | 4563 | 4059 | 6437 | 6437 | 0 |

The run-end relation finding is identical on both sides, once per stream, so zero of those findings move. The report metadata retains the different first affected unit of each side. Counts are unique `(unit, phase, kind, site, full reason)` rows. Removed and added rows are separate observations, not paired causal transitions. All finding kinds are retained. Equal rows are excluded from moved counts.

## Every head row in full

Format: `side | catalogue unit number | unit | phase | kind | site | full reason`.

```text
base | 2583 | compiler/utilities.ts:12088:1 | lowering | Refused | compiler/utilitiesPublic.ts:786:1 | overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked)
base | 1783 | compiler/utilities.ts:1291:1 | lowering | Refused | compiler/utilitiesPublic.ts:786:1 | overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked)
base | 1900 | compiler/utilities.ts:3125:1 | lowering | Refused | compiler/utilitiesPublic.ts:786:1 | overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked)
base | 1901 | compiler/utilities.ts:3130:1 | lowering | Refused | compiler/utilitiesPublic.ts:786:1 | overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked)
base | 1902 | compiler/utilities.ts:3135:1 | lowering | Refused | compiler/utilitiesPublic.ts:786:1 | overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked)
base | 1904 | compiler/utilities.ts:3150:1 | lowering | Refused | compiler/utilitiesPublic.ts:786:1 | overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked)
base | 2040 | compiler/utilities.ts:4823:1 | lowering | Refused | compiler/utilitiesPublic.ts:786:1 | overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked)
base | 1607 | compiler/utilitiesPublic.ts:1566:1 | lowering | Refused | compiler/utilitiesPublic.ts:786:1 | overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked)
base | 1516 | compiler/utilitiesPublic.ts:787:1 | lowering | Refused | compiler/utilitiesPublic.ts:786:1 | overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked)
base | 1517 | compiler/utilitiesPublic.ts:788:1 | lowering | Refused | compiler/utilitiesPublic.ts:786:1 | overload 1 of findAncestor result T | undefined cannot be served by implementation result Node | undefined (return at compiler/utilitiesPublic.ts:795:13 isn't the value its predicate checked)
head | 1900 | compiler/utilities.ts:3125:1 | lowering | NotYet | compiler/binder.ts:1034:17 | writing a checked object field without its source-slot type certificate
head | 1901 | compiler/utilities.ts:3130:1 | lowering | NotYet | compiler/binder.ts:1034:17 | writing a checked object field without its source-slot type certificate
head | 1904 | compiler/utilities.ts:3150:1 | lowering | NotYet | compiler/binder.ts:1034:17 | writing a checked object field without its source-slot type certificate
head | 1607 | compiler/utilitiesPublic.ts:1566:1 | lowering | NotYet | compiler/utilitiesPublic.ts:1562:1 | checked predicate overload target ExportDeclaration & { readonly isTypeOnly: true; readonly moduleSpecifier: Expression; }
```

## Relation-limit evidence

base:

```json
{
  "after_load": "",
  "attribution": "per attempted unit, checked after lowering; includes the unit that first exposes overflow",
  "end": "error TS2859: Excessive complexity comparing types '\" \"' and '{}'.",
  "first_affected_unit": "/workspace/scratch/bar-adapted/src/compiler/transformers/es2020.ts:1:1",
  "per_unit": true
}
```

head:

```json
{
  "after_load": "",
  "attribution": "per attempted unit, checked after lowering; includes the unit that first exposes overflow",
  "end": "error TS2859: Excessive complexity comparing types '\" \"' and '{}'.",
  "first_affected_unit": "/workspace/scratch/bar-adapted/src/compiler/transformers/typeSerializer.ts:132:1",
  "per_unit": true
}
```

Every moved noise row is listed with its catalogue unit and full unit location in [relation-limit-noise.txt](relation-limit-noise.txt). All classes, full findings and ordinals are in [moved-findings.json](moved-findings.json). The run-end relation report is compared separately as the relation-limit finding itself.

## Evidence and validation

- Whole streams: [base-entry-run.jsonl](base-entry-run.jsonl.gz), [head-entry-run.jsonl](head-entry-run.jsonl.gz).
- Whole stderr/stdout logs: [base-entry-run.log](base-entry-run.log), [head-entry-run.log](head-entry-run.log).
- Requested sorted unique lists: [findings-base.txt](findings-base.txt.gz), [findings-head.txt](findings-head.txt.gz); literal [findings.diff](findings.diff).
- Source-only pair inputs: [base-pair-input.jsonl.gz](base-pair-input.jsonl.gz), [head-pair-input.jsonl.gz](head-pair-input.jsonl.gz).
- Paired output: [entry-paired.json](entry-paired.json.gz), [entry-paired.md](entry-paired.md).
- Exact onset and count analysis: [analysis.json](analysis.json).
- Independent artifact audit: [audit.json](audit.json). Complete streams, all probe answers, monotone marks, exact list rows, lossless source-only pair projections, checkpoint scope and process cleanup were checked. Its initial plain-only list reader failed after size-based gzip; it was fixed and rerun successfully, as recorded in [audit-initial.json](audit-initial.json).
- Existing pair suite: all 15 tests passed. Both own-pin overlay relation/stream tests passed; logs and child statuses are retained.

| Run | Own child exit | Wall seconds | Peak RSS MiB |
| --- | ---: | ---: | ---: |
| base | 0 | 6674.316 | 5288.1 |
| head | 0 | 6953.023 | 5606.1 |

Checkpoint push receipts and full push logs are retained for each completed stream, on separate `typescript/find-ancestor-bar-census-checkpoint-*` refs. The final branch contains one report-only commit directly on c1e9379e. No main ref or force push is used.
