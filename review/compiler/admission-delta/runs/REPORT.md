Recorded two real full-manifest runs with cold revision builds; no compiler or admission-tool features changed.
Tool source is 85144830; base is cf735d9f; fixes head is ced32bf9; delivery commit is reported in the final response.
Cold compiler builds passed; both admission commands exited 1 with JSON verdict fail, admitted 0, sampled 0.
No new mutants; no newly accepted programs meant no Node, JavaScript-backend, or native runtime comparisons.
Not covered: a real refusal-to-admission runtime delta; both manifests also had 16 missing-dependency errors and five compiler timeouts.

## Pinned revisions and results

Both requested comparisons use base `cf735d9fba9e38de6368575e5630e44375a86eaf`, the origin/main tip at the initial fetch. Run 1 uses that same SHA as head. Run 2 uses `ced32bf9b1f410b8be07f078adb78365a20a9f5b`, the latest fixes tip at the initial fetch, with `--budget 60`. The final fetch still showed the same fixes tip; origin/main had advanced to `76c59c81e8617cea1892a01841494895927a712c`. The completed runs retain their original pins.

| Run | Programs | JSON verdict | Newly admitted | Sampled | Omitted | Exit |
|---|---:|---|---:|---:|---:|---:|
| main | 916 | fail | 0 | 0 | 0 | 1 |
| fixes | 920 | fail | 0 | 0 | 0 | 1 |

Main corpora: witnesses 2, fixtures 801, gaps 107, review 6, fuzz 0.
Fixes corpora: witnesses 2, fixtures 801, gaps 107, review 10, fuzz 0.

Main classes: 833 accepted-by-both, 62 refused-by-both, 16 compiler-error, 5 compiler-timeout.
Fixes classes: 837 accepted-by-both, 62 refused-by-both, 16 compiler-error, 5 compiler-timeout.
Neither run had newly refused programs.

## Why run 2 did not exercise a runtime delta

Observation: all four review programs added by the fixes branch compiled successfully with both the main and fixes compilers, using their head-pinned inputs:

| Program | Main compile exit | Fixes compile exit | Class |
|---|---:|---:|---|
| review/compiler/fx5-every-refusal/initial-literal-layout.a | 0 | 0 | accepted-by-both |
| review/compiler/fx5-literal-fix/all-number.a | 0 | 0 | accepted-by-both |
| review/compiler/fx5-literal-fix/nested-number.a | 0 | 0 | accepted-by-both |
| review/compiler/fx5-literal-refusal/all-number.a | 0 | 0 | accepted-by-both |

Observation: the refusal-introducing commits `e49ae19e` and `5f466b83` are ancestors of the fixes tip, and neither is an ancestor of the pinned main tip. Their ancestry checks are in timings.json.

Inference: the fixes restore admissions lost on intermediate worker branches, but these full manifests show no programs refused by the requested main baseline and accepted by the fixes tip. Changing base to an intermediate refusal revision would test a different comparison and was not done.

There are no three-runtime disagreements to report because zero programs were selected. Node, JavaScript-backend, and native output observations are null in both result documents. This is not evidence that the fixes' output is correct.

## Phase wall times

Seconds below are measured wall times. Classification spans the first compiler invocation through the last. The admission-command total also includes the tool's isolated corpus checkout, metadata verification, sampling, JSON writing, and cleanup; it therefore exceeds the classification span. Selected-program comparison time is effectively zero because there were no admissions.

| Phase | Main vs main | Main vs fixes |
|---|---:|---:|
| Compiler checkout | 1.869279 | 3.425664 |
| Pinned cohere worktree | 0.515187 | 0.816338 |
| Pinned TypeScript worktree | 2.570064 | 2.222446 |
| Cold Go build | 163.788113 | 163.270385 |
| Manifest generation | 0.214799 | 0.214791 |
| Admission classification | 233.996838 | 236.163782 |
| Selected-program comparison | 0.000000 | 0.000000 |
| Admission command total | 235.978674 | 238.102479 |

Both Go cache directories contained zero entries before their build. Each actual compiler was built from the pinned revision in a detached checkout with the exact pinned cohere and TypeScript worktrees. No tracked cmd or internal files changed in either build checkout. Binary SHA-256 digests, cache identities, source SHAs, and all phase commands are in timings.json. Run 2 reuses the main binary cold-built for run 1 and uses a separately cold-built fixes binary.

The cold builds exceeded the admission tool's fixed 120 s compiler-build limit. They were therefore performed explicitly, then supplied using the tool's existing `--base-binary` and `--head-binary` flags. This verifies real cold revision builds and real classifications, but does not prove the tool's automatic build path can finish cold under its existing limit.

## Manifest generator provenance

Neither requested head contains cloud/admission-corpus/manifest.py. The unchanged generator shipped at tool revision `85144830a99aa1353011562703da4f4a68da2730`, blob `220f6496b4c491bbfe23cec5ed3b9b3a6d378544`, generated each complete manifest directly from the requested SHA.

The original outputs are main.generated-manifest.json and fixes.generated-manifest.json. For execution, only the generator header is cleared in main.execution-manifest.json and fixes.execution-manifest.json, avoiding the existing tool's attempt to resolve an absent-at-head generator. Every corpus, program path, and blob is unchanged. Each result contains every input path and blob in the same order; that equality was checked for all 916 and 920 entries. The tool's raw result generator_blob is empty, and the external generator's true revision and blob are recorded in timings.json rather than falsely attributed to either head. No tool code was changed for this workaround.

## Exact admission commands

The concrete commands with absolute binary wrapper and manifest paths are recorded in timings.json. Their form was:

```sh
/tmp/admission-real-tool --base cf735d9fba9e38de6368575e5630e44375a86eaf \
  --head cf735d9fba9e38de6368575e5630e44375a86eaf \
  --base-binary <cold-main-observer> --head-binary <cold-main-observer> \
  --manifest review/compiler/admission-delta/runs/main.execution-manifest.json --json

/tmp/admission-real-tool --base cf735d9fba9e38de6368575e5630e44375a86eaf \
  --head ced32bf9b1f410b8be07f078adb78365a20a9f5b \
  --base-binary <cold-main-observer> --head-binary <cold-fixes-observer> \
  --manifest review/compiler/admission-delta/runs/fixes.execution-manifest.json \
  --budget 60 --json
```

The observers only record compiler command wall times and forward stdout, stderr, and exit codes. The compiler and tool executables are unmodified. Each compilation uses the tool's default 10 s timeout. Timed-out wrappers cannot append their completed timing records, so commands.tsv has fewer completion rows than attempted compiler invocations; timeouts are still present in the result JSON and within the classification wall span.

## Errors and preparation failures

Both admission runs failed on the same environment and timeout records. These are retained as errors, never refusals. No compiler fixes were made.

### main compiler errors

- `internal/oracle/testdata/closure_convention_host24.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_bom.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_crypto.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_digest_twice.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_encodings.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_finalized.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_input.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_random.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_utf16.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_utf8.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_writes.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_fs_file_close.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_fs_file_mkdir.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_fs_file_write_buffer.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_fs_file_write_file.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/unknown_narrowing_host.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `stage1/cohere/css/gaps/printer_boundaries.ts`: compiler-timeout; base and head both timeout
- `stage1/cohere/estree/gaps/deepBinary.ts`: compiler-timeout; base and head both timeout
- `stage1/cohere/estree/gaps/emptyGenerics.ts`: compiler-timeout; base and head both timeout
- `stage1/cohere/estree/gaps/portParserRecovery.ts`: compiler-timeout; base and head both timeout
- `stage1/cohere/estree/gaps/typeMemberInitializer.ts`: compiler-timeout; base and head both timeout

### fixes compiler errors

- `internal/oracle/testdata/closure_convention_host24.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_bom.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_crypto.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_digest_twice.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_encodings.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_finalized.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_input.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_random.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_utf16.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_utf8.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_buffer_writes.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_fs_file_close.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_fs_file_mkdir.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_fs_file_write_buffer.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/node_fs_file_write_file.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `internal/oracle/testdata/unknown_narrowing_host.a`: compiler-error; base and head both adamic: load: node:* imports require @types/node 25.3.3 installed in stage3/api/node_modules/@types/node
- `stage1/cohere/css/gaps/printer_boundaries.ts`: compiler-timeout; base and head both timeout
- `stage1/cohere/estree/gaps/deepBinary.ts`: compiler-timeout; base and head both timeout
- `stage1/cohere/estree/gaps/emptyGenerics.ts`: compiler-timeout; base and head both timeout
- `stage1/cohere/estree/gaps/portParserRecovery.ts`: compiler-timeout; base and head both timeout
- `stage1/cohere/estree/gaps/typeMemberInitializer.ts`: compiler-timeout; base and head both timeout


The initial local TypeScript clone started repacking upstream history and was stopped. preparation-attempt.json and preparation-attempt.log retain that attempt. The first cold Go build on /tmp failed after 57.655 s with no space left on device; disk-full-attempt.json, disk-full-attempt.log, and disk-full-build.log retain the failure. Both successful cold builds started afresh on /workspace with empty caches. The comparison results were not taken from those failed attempts.

No new test leaves, oracle fixtures, mutants, or counts.md rows were added. No whole package tests or full gate were run. These runs are evidence only, and all results, manifests, traces, and logs live under review/compiler/admission-delta/runs/.
