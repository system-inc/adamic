# Paired kind-predicate latent census

Base: `df84b97437661778245a02862f38799f4bdf31cf`. Head: `b03fe07e48cce2b4a11f183bb3b4a8ba06bb60b9`.

Both successful runs used one adapted tree and first-error mode. The head is one commit directly on this base.

Open-tag sites: **1251 before, 1006 after**. Of the base sites, **239 now have no finding**, **6 moved to another reason**, and **1006 retain the open-tag reason**.

Regressions: **0**. Units whose lowering first stop changed: **0**.

| Unique finding kind | Base | Head |
| --- | ---: | ---: |
| NotYet | 559 | 559 |
| Refused | 4280 | 4036 |
| panic | 0 | 0 |
| error | 0 | 0 |
| SkippedDependency | 0 | 0 |

Moved open-tag sites by new reason:

- 1: a cast the runtime can't check
- 1: a value of type ComputedPropertyName seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace
- 4: a value of type Expression seen as SourceMapRange | undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace

Five of these six locations already had their other finding on the base. Only the ComputedPropertyName-to-TextRange refusal is newly revealed. All six locations were already blocked, so this is no newly failing source site.

| Source token | Remaining reason | Present on base |
| --- | --- | --- |
| `src/compiler/moduleSpecifiers.ts:906:35` | a cast the runtime can't check | yes |
| `src/compiler/transformers/classFields.ts:932:41` | a value of type Expression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | yes |
| `src/compiler/transformers/classFields.ts:935:47` | a value of type Expression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | yes |
| `src/compiler/transformers/destructuring.ts:559:214` | a value of type ComputedPropertyName seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | no |
| `src/compiler/transformers/esDecorators.ts:1597:45` | a value of type Expression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | yes |
| `src/compiler/transformers/esDecorators.ts:1600:51` | a value of type Expression seen as SourceMapRange \| undefined, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | yes |

## Units whose first stop moved

Only lowering-phase events define a unit stop here. File-wide refusal-scan owners are retained in the JSON but are not treated as attempted units.

None. Every attempted unit retains its first lowering stop or remains finding-free.

## Regressions

None observed. No previously finding-free attempted unit now stops, no previously finding-free refusal-scan site now stops, and no reason became a panic.

Changed first stops in units already blocked on the base are reported above, rather than counted as a newly failing unit. The per-site JSON retains all newly exposed findings, their phases, and their owning attempts.

## Unit lowering time on this box

One sequential sample per compiler. These are sums of the meter's unit timers, not whole-process elapsed times. File scans, checker work, and other shared setup are excluded.

| Sum | Base seconds | Head seconds | Head minus base |
| --- | ---: | ---: | ---: |
| wall_seconds | 552.475779 | 543.306559 | -9.169220 |
| cpu_seconds | 627.255324 | 621.177584 | -6.077740 |

Ten biggest movers by absolute wall-time delta. All per-unit wall and CPU deltas are in `delta.json`.

| Unit | Base wall s | Head wall s | Delta wall s | Delta CPU s |
| --- | ---: | ---: | ---: | ---: |
| `src/compiler/resolutionCache.ts:1711:1` (resolutionIsSymlink) | 1.250825 | 0.414083 | -0.836742 | -0.852667 |
| `src/compiler/utilities.ts:5522:1` (isNamedEvaluationSource) | 0.996416 | 1.658051 | +0.661634 | +0.656517 |
| `src/compiler/transformers/ts.ts:235:1` (transformTypeScript) | 0.309807 | 0.876467 | +0.566660 | +0.213991 |
| `src/compiler/utilitiesPublic.ts:870:1` (symbolName) | 1.105273 | 1.626089 | +0.520816 | +5.334119 |
| `src/compiler/utilities.ts:7296:1` (getRawJSDocModifierFlagsNoCache) | 0.967771 | 0.480757 | -0.487014 | -3.291516 |
| `src/compiler/programDiagnostics.ts:89:1` (createProgramDiagnostics) | 0.592697 | 0.148871 | -0.443826 | -0.446448 |
| `src/compiler/transformers/utilities.ts:618:1` (isStaticPropertyDeclaration) | 0.425413 | 0.777668 | +0.352255 | +2.507914 |
| `src/compiler/utilitiesPublic.ts:2143:1` (needsScopeMarker) | 1.283383 | 1.620805 | +0.337423 | +3.308114 |
| `src/compiler/factory/utilities.ts:631:1` (isOuterExpression) | 0.396765 | 0.727540 | +0.330775 | +0.313648 |
| `src/compiler/diagnosticInformationMap.generated.ts:4:1` (KindImportDeclaration) | 0.401971 | 0.072994 | -0.328977 | -0.732625 |

## Resources

| Run | Peak RSS KiB | Peak RSS GiB | Process wall s | Exit |
| --- | ---: | ---: | ---: | ---: |
| base | 6620928 | 6.314 | 1388.706753 | 0 |
| head | 6618736 | 6.312 | 1350.613170 | 0 |

Peak RSS is Linux `wait4` child `ru_maxrss`, in KiB, collected by `measure-rss.py`. Neither run was killed. Cgroup OOM counters stayed unchanged. The cgroup memory limit was 16 GiB; CPU quota was 4 cores. Host metadata is in `evidence/box.json`.

## Method and evidence

Read `stage3/meter/twice-daily.sh` and `stage3/meter/README.md` in full. Ran the requested setup with `GOPROXY=https://proxy.golang.org|direct` and `timeout --kill-after=60s 1800 bash cloud/setup.sh > /workspace/scratch/setup.log 2>&1`; setup exited successfully. Setup warmed the initial checkout, but both measured binaries were separately built from their exact compiler pins.

The shared adapted tree was built once with the base's `stage3/apply.sh` and lockfile-pinned API packages, using upstream TypeScript v6.0.3 at `050880ce59e30b356b686bd3144efe24f875ebc8`. `git diff df84b974 b03fe07e -- stage3` is empty. No adaptation change confounds this compiler comparison. Source hashes are in `evidence/source-manifest.json`.

Both compiler worktrees were clean, with cohere `06af2e399c6c1fb93fd570b774027942747400b4` and its TypeScript checker `d92d9bfee114c80be2c375d72edae966176e3a4f`. Each used the meter's `make_overlay.py`, `gofmt`, and `go build -buildvcs=false -overlay=... ./stage3/census/latent/tool`. The overlay blocks ordinary loading and usable IR output. No backend was invoked.

For each binary, exactly as the meter's compiler-directory latent invocation:

```sh
LATENT_TIMES=times.jsonl LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 latent-census /workspace/scratch/kp-adapted/src/compiler latent.jsonl
```

The base build and run finished before the head build and run. The two runs cover all 79 compiler files, have identical checker headers and unit eligibility, and each records 4,056 timed attempts. All 1,251 baseline open-tag sites originate in the refusal scan.

Counts follow the meter's unique `(kind, where, reason, text)` definition. Pairing uses source SHA-256 and exact UTF-16 token offset, preserving column identity and all findings at a token. `delta.json` includes every changed site, all 1,251 open-tag transitions, first-stop changes, regression accounting, and all 4,056 unit timing deltas. The raw normalized streams and timers are retained under `evidence/`, with hashes in `artifacts.sha256.json`.

"No finding" means that the source token has no finding in this guarded first-error observation. It does not establish complete unit lowering or native compilation. Later stops in already blocked units are reported as exposure, with the old and new stops preserved.

The first launch did not start a census because `/usr/bin/time` is absent in this image. That launcher failure is retained as `evidence/kp-launcher-failure.log.gz`; the built base binary was reused with the kernel-backed RSS recorder. Only the two successful census runs supply these results.

## Validation

`validate.py` independently checks the saved streams, counts, source identities, timing sums, resource records, and artifact hashes. It also rechecks all shared adapted source hashes after both runs. All checks passed; see `validation.txt`.
