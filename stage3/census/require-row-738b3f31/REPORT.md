# Require row paired census

Task 00qn5az. Report only.

Base: `71b126b06aa5840838330965f77b089fbffee3e4`. Head: `738b3f315311eff497d8703e62f879d3b2b3ee2e`. The CLI tip was checked before setup and still matched the requested head. TypeScript 6.0.3 source pin: `050880ce59e30b356b686bd3144efe24f875ebc8`.

Box: `6f9f78360b0e`; 5 visible CPUs; CPU quota `400000 100000`; MemTotal:       18440100 kB. See provenance.json for affinity, memory limit and exact tree.

## Paired first stops

All 79 roots, sorted with Python Path ordering, ran individually in groups of 20, 20, 20 and 19 on each side. Each invocation produced two observations. All non-timing fields agreed; only the first counts. The ordinary census uses no overlay.

| Roots | Base location | Base kind | Base reason | Head location | Head kind | Head reason | Result | Root names |
| ---: | --- | --- | --- | --- | --- | --- | --- | --- |
| 73 | core.ts:2599:15 | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | core.ts:2599:15 | NotYet | the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column, | identical | _namespaces/ts.moduleSpecifiers.ts, _namespaces/ts.performance.ts, _namespaces/ts.ts, binder.ts, builder.ts, builderPublic.ts, builderState.ts, builderStatePublic.ts, checker.ts, commandLineParser.ts, emitter.ts, executeCommandLine.ts, expressionToTypeNode.ts, factory/baseNodeFactory.ts, factory/emitHelpers.ts, factory/emitNode.ts, factory/nodeChildren.ts, factory/nodeConverters.ts, factory/nodeFactory.ts, factory/nodeTests.ts, factory/parenthesizerRules.ts, factory/utilities.ts, factory/utilitiesPublic.ts, moduleNameResolver.ts, moduleSpecifiers.ts, parser.ts, path.ts, performance.ts, performanceCore.ts, program.ts, programDiagnostics.ts, resolutionCache.ts, scanner.ts, semver.ts, sourcemap.ts, symbolWalker.ts, sys.ts, tracing.ts, transformer.ts, transformers/classFields.ts, transformers/classThis.ts, transformers/declarations/diagnostics.ts, transformers/declarations.ts, transformers/destructuring.ts, transformers/es2015.ts, transformers/es2016.ts, transformers/es2017.ts, transformers/es2018.ts, transformers/es2019.ts, transformers/es2020.ts, transformers/es2021.ts, transformers/esDecorators.ts, transformers/esnext.ts, transformers/generators.ts, transformers/jsx.ts, transformers/legacyDecorators.ts, transformers/module/esnextAnd2015.ts, transformers/module/impliedNodeFormatDependent.ts, transformers/module/module.ts, transformers/module/system.ts, transformers/namedEvaluation.ts, transformers/taggedTemplate.ts, transformers/ts.ts, transformers/typeSerializer.ts, transformers/utilities.ts, tsbuild.ts, tsbuildPublic.ts, utilities.ts, utilitiesPublic.ts, visitorPublic.ts, watch.ts, watchPublic.ts, watchUtilities.ts |
| 2 | - | accepted |  | - | accepted |  | identical | corePublic.ts, hostErrors.ts |
| 1 | debug.ts:525:37 | Refused | a cast the runtime can't check | debug.ts:525:37 | Refused | a cast the runtime can't check | identical | core.ts |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | identical | debug.ts |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | identical | diagnosticInformationMap.generated.ts |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | identical | types.ts |

Exact location/kind/reason comparison: 79/79 identical. Every root and both observations are retained. Accepted means the entire resolved root program lowers; no native output is measured.

## Full core.ts site list

Every distinct location/phase/kind/reason finding is listed in source order, including lowering findings. Boundary markers are retained only in raw streams. Sites at line 2599 and later are marked `*`. Repeated recording events are deduplicated here and retained in the raw stream.

| Mark | Location | Phase | Kind | Exact reason | Base | Head |
| --- | --- | --- | --- | --- | --- | --- |
|  | core.ts:738:9 | lowering | NotYet | `a value of type NonNullable<T>` | present | present |
|  | core.ts:739:31 | lowering | NotYet | `an array of NonNullable<T>` | present | present |
|  | core.ts:741:15 | lowering | NotYet | `a value of type NonNullable<T>` | present | present |
|  | core.ts:742:26 | lowering | NotYet | `reading next` | present | present |
|  | core.ts:756:9 | lowering | NotYet | `an array of T` | present | present |
|  | core.ts:759:12 | lowering | NotYet | `reading deduplicated` | present | present |
|  | core.ts:1049:63 | lowering | NotYet | `an array of T` | present | present |
| * | core.ts:2599:15 | refusal_scan | NotYet | `the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column,` | present | present |
| * | core.ts:2599:80 | refusal_scan | Refused | `an ambient value declaration` | present | absent |
| * | core.ts:2604:12 | lowering | NotYet | `the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column,` | present | present |
| * | core.ts:2604:19 | refusal_scan | NotYet | `the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column,` | present | present |
| * | core.ts:2605:14 | refusal_scan | NotYet | `the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column,` | present | present |
| * | core.ts:2606:13 | refusal_scan | NotYet | `the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column,` | present | present |

Removed sites: 1. Added sites: 0. The require declaration at core.ts:2599:80 is gone at head.

The first refusal-scan site after process at head is `core.ts:2604:19`, NotYet: `the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column,`. This is the next currently recorded refusal-scan site. Its reason still names the pending process row, so it is not an independent forecast after Runtime clears process. These measurements do not establish where the 73 go after that clearance; no process-cleared branch exists in this comparison.

Head sites from line 2599 onward, in source order:

- `core.ts:2599:15` refusal_scan NotYet: the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column,

- `core.ts:2604:12` lowering NotYet: the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column,

- `core.ts:2604:19` refusal_scan NotYet: the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column,

- `core.ts:2605:14` refusal_scan NotYet: the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column,

- `core.ts:2606:13` refusal_scan NotYet: the host binding process, pending Runtime's clearance of its native column and Runtime and Platforms' clearance of its wasm32 column,

All four refusal-scan sites remaining in head core.ts name the pending process row. The later scan sites arise from hostBindingRefusal at internal/lower/host_bindings.go:268-269, which checks the row stop before allowing typeof at lines 275-276. Clearing the row can therefore clear these sites as well. The seven earlier lowering findings are independent unit-probe observations, including generic and rollback-dependent findings; their source order does not establish an ordinary next stop. Runtime clearance followed by another paired ordinary census is needed to satisfy a forecast beyond process.

## Run labels

Peak memory is Linux wait4 ru_maxrss in KiB for each measured child, including inherited child usage under Linux accounting. Times are UTC. Each child exit and process-group cleanup is recorded.

| Run | Start UTC | End UTC | Seconds | Peak RSS KiB | Child exit |
| --- | --- | --- | ---: | ---: | ---: |
| apply | 2026-11-08T20:41:07.457751+00:00 | 2026-11-08T20:49:22.419780+00:00 | 494.962 | 2099284 | 0 |
| base-build | 2026-11-08T20:49:37.626326+00:00 | 2026-11-08T20:49:46.642648+00:00 | 9.016 | 1467584 | 0 |
| base-buildinfo | 2026-11-08T20:49:46.710557+00:00 | 2026-11-08T20:49:47.717767+00:00 | 1.007 | 47104 | 0 |
| base-checkpoint-core | 2026-11-08T20:50:58.280623+00:00 | 2026-11-08T20:51:04.295317+00:00 | 6.015 | 47104 | 0 |
| base-checkpoint-group-1 | 2026-11-08T21:01:19.825674+00:00 | 2026-11-08T21:01:22.831677+00:00 | 3.006 | 47136 | 0 |
| base-checkpoint-group-2 | 2026-11-08T21:18:50.093209+00:00 | 2026-11-08T21:18:53.100340+00:00 | 3.007 | 47136 | 0 |
| base-checkpoint-group-3 | 2026-11-08T21:37:28.245977+00:00 | 2026-11-08T21:37:31.249717+00:00 | 3.004 | 47136 | 0 |
| base-checkpoint-group-4 | 2026-11-08T21:55:40.553246+00:00 | 2026-11-08T21:55:43.560811+00:00 | 3.008 | 47136 | 0 |
| base-compiler-source-clean | 2026-11-08T20:49:36.619435+00:00 | 2026-11-08T20:49:37.625216+00:00 | 1.006 | 16348 | 0 |
| base-core | 2026-11-08T20:50:23.811152+00:00 | 2026-11-08T20:50:58.026506+00:00 | 34.215 | 890752 | 0 |
| base-latent-build | 2026-11-08T20:49:50.742626+00:00 | 2026-11-08T20:50:23.804194+00:00 | 33.062 | 1471464 | 0 |
| base-overlay-format | 2026-11-08T20:49:49.730380+00:00 | 2026-11-08T20:49:50.736287+00:00 | 1.006 | 47104 | 0 |
| base-overlay | 2026-11-08T20:49:47.724119+00:00 | 2026-11-08T20:49:49.727910+00:00 | 2.004 | 65432 | 0 |
| base-root-01 | 2026-11-08T20:52:56.844092+00:00 | 2026-11-08T20:53:24.874802+00:00 | 28.031 | 882768 | 0 |
| base-root-02 | 2026-11-08T20:53:24.876394+00:00 | 2026-11-08T20:53:53.931471+00:00 | 29.055 | 888956 | 0 |
| base-root-03 | 2026-11-08T20:53:53.932950+00:00 | 2026-11-08T20:54:24.035759+00:00 | 30.103 | 861176 | 0 |
| base-root-04 | 2026-11-08T20:54:24.042688+00:00 | 2026-11-08T20:54:54.206355+00:00 | 30.164 | 861268 | 0 |
| base-root-05 | 2026-11-08T20:54:54.211365+00:00 | 2026-11-08T20:55:24.243756+00:00 | 30.032 | 901916 | 0 |
| base-root-06 | 2026-11-08T20:55:24.245676+00:00 | 2026-11-08T20:55:50.280312+00:00 | 26.035 | 870048 | 0 |
| base-root-07 | 2026-11-08T20:55:50.282171+00:00 | 2026-11-08T20:56:17.301361+00:00 | 27.019 | 859464 | 0 |
| base-root-08 | 2026-11-08T20:56:17.304768+00:00 | 2026-11-08T20:56:46.336698+00:00 | 29.032 | 866244 | 0 |
| base-root-09 | 2026-11-08T20:56:46.339028+00:00 | 2026-11-08T20:57:13.360037+00:00 | 27.021 | 861632 | 0 |
| base-root-10 | 2026-11-08T20:57:13.361827+00:00 | 2026-11-08T20:57:41.391802+00:00 | 28.03 | 868268 | 0 |
| base-root-11 | 2026-11-08T20:57:41.393074+00:00 | 2026-11-08T20:58:11.513640+00:00 | 30.121 | 872524 | 0 |
| base-root-12 | 2026-11-08T20:58:11.517632+00:00 | 2026-11-08T20:58:12.519542+00:00 | 1.002 | 47136 | 0 |
| base-root-13 | 2026-11-08T20:58:12.520080+00:00 | 2026-11-08T20:58:19.531862+00:00 | 7.012 | 434700 | 0 |
| base-root-14 | 2026-11-08T20:58:19.532608+00:00 | 2026-11-08T20:58:27.541456+00:00 | 8.009 | 468396 | 0 |
| base-root-15 | 2026-11-08T20:58:27.545832+00:00 | 2026-11-08T20:58:56.600714+00:00 | 29.055 | 869124 | 0 |
| base-root-16 | 2026-11-08T20:58:56.603360+00:00 | 2026-11-08T20:59:24.698367+00:00 | 28.095 | 874192 | 0 |
| base-root-17 | 2026-11-08T20:59:24.699050+00:00 | 2026-11-08T20:59:53.723400+00:00 | 29.024 | 891696 | 0 |
| base-root-18 | 2026-11-08T20:59:53.724291+00:00 | 2026-11-08T21:00:19.753628+00:00 | 26.029 | 870196 | 0 |
| base-root-19 | 2026-11-08T21:00:19.758809+00:00 | 2026-11-08T21:00:51.786212+00:00 | 32.027 | 882436 | 0 |
| base-root-20 | 2026-11-08T21:00:51.788068+00:00 | 2026-11-08T21:01:19.819841+00:00 | 28.032 | 864132 | 0 |
| base-root-21 | 2026-11-08T21:09:44.741544+00:00 | 2026-11-08T21:10:12.769182+00:00 | 28.028 | 879804 | 0 |
| base-root-22 | 2026-11-08T21:10:12.770995+00:00 | 2026-11-08T21:10:41.825436+00:00 | 29.054 | 866536 | 0 |
| base-root-23 | 2026-11-08T21:10:41.826682+00:00 | 2026-11-08T21:11:07.850382+00:00 | 26.024 | 851744 | 0 |
| base-root-24 | 2026-11-08T21:11:07.851968+00:00 | 2026-11-08T21:11:38.943443+00:00 | 31.092 | 872480 | 0 |
| base-root-25 | 2026-11-08T21:11:38.945442+00:00 | 2026-11-08T21:12:05.969132+00:00 | 27.024 | 873324 | 0 |
| base-root-26 | 2026-11-08T21:12:05.970267+00:00 | 2026-11-08T21:12:33.015013+00:00 | 27.045 | 880728 | 0 |
| base-root-27 | 2026-11-08T21:12:33.022249+00:00 | 2026-11-08T21:13:00.050978+00:00 | 27.029 | 872364 | 0 |
| base-root-28 | 2026-11-08T21:13:00.052035+00:00 | 2026-11-08T21:13:01.054790+00:00 | 1.003 | 47136 | 0 |
| base-root-29 | 2026-11-08T21:13:01.057968+00:00 | 2026-11-08T21:13:28.083232+00:00 | 27.025 | 866200 | 0 |
| base-root-30 | 2026-11-08T21:13:28.087923+00:00 | 2026-11-08T21:13:58.140111+00:00 | 30.052 | 871332 | 0 |
| base-root-31 | 2026-11-08T21:13:58.143615+00:00 | 2026-11-08T21:14:28.171198+00:00 | 30.028 | 871604 | 0 |
| base-root-32 | 2026-11-08T21:14:28.172057+00:00 | 2026-11-08T21:14:55.198191+00:00 | 27.026 | 886284 | 0 |
| base-root-33 | 2026-11-08T21:14:55.199325+00:00 | 2026-11-08T21:15:27.255153+00:00 | 32.056 | 872324 | 0 |
| base-root-34 | 2026-11-08T21:15:27.256430+00:00 | 2026-11-08T21:15:57.299597+00:00 | 30.043 | 870868 | 0 |
| base-root-35 | 2026-11-08T21:15:57.301522+00:00 | 2026-11-08T21:16:26.339773+00:00 | 29.038 | 870240 | 0 |
| base-root-36 | 2026-11-08T21:16:26.340894+00:00 | 2026-11-08T21:16:56.543048+00:00 | 30.202 | 915784 | 0 |
| base-root-37 | 2026-11-08T21:16:56.545119+00:00 | 2026-11-08T21:17:24.771020+00:00 | 28.226 | 878776 | 0 |
| base-root-38 | 2026-11-08T21:17:24.778453+00:00 | 2026-11-08T21:17:55.021465+00:00 | 30.243 | 853788 | 0 |
| base-root-39 | 2026-11-08T21:17:55.023538+00:00 | 2026-11-08T21:18:23.058763+00:00 | 28.035 | 911160 | 0 |
| base-root-40 | 2026-11-08T21:18:23.060581+00:00 | 2026-11-08T21:18:50.090921+00:00 | 27.03 | 878204 | 0 |
| base-root-41 | 2026-11-08T21:27:57.072806+00:00 | 2026-11-08T21:28:26.101963+00:00 | 29.029 | 875532 | 0 |
| base-root-42 | 2026-11-08T21:28:26.104720+00:00 | 2026-11-08T21:28:53.143651+00:00 | 27.039 | 883900 | 0 |
| base-root-43 | 2026-11-08T21:28:53.145538+00:00 | 2026-11-08T21:29:20.178707+00:00 | 27.033 | 860056 | 0 |
| base-root-44 | 2026-11-08T21:29:20.179465+00:00 | 2026-11-08T21:29:48.203278+00:00 | 28.024 | 887468 | 0 |
| base-root-45 | 2026-11-08T21:29:48.204391+00:00 | 2026-11-08T21:30:18.362021+00:00 | 30.158 | 901976 | 0 |
| base-root-46 | 2026-11-08T21:30:18.366965+00:00 | 2026-11-08T21:30:47.545929+00:00 | 29.179 | 886456 | 0 |
| base-root-47 | 2026-11-08T21:30:47.550950+00:00 | 2026-11-08T21:31:15.590033+00:00 | 28.039 | 882956 | 0 |
| base-root-48 | 2026-11-08T21:31:15.595579+00:00 | 2026-11-08T21:31:44.624226+00:00 | 29.029 | 881904 | 0 |
| base-root-49 | 2026-11-08T21:31:44.625232+00:00 | 2026-11-08T21:32:13.659448+00:00 | 29.034 | 892564 | 0 |
| base-root-50 | 2026-11-08T21:32:13.666015+00:00 | 2026-11-08T21:32:40.697673+00:00 | 27.032 | 866476 | 0 |
| base-root-51 | 2026-11-08T21:32:40.698894+00:00 | 2026-11-08T21:33:07.731949+00:00 | 27.033 | 870192 | 0 |
| base-root-52 | 2026-11-08T21:33:07.732715+00:00 | 2026-11-08T21:33:36.789582+00:00 | 29.057 | 870068 | 0 |
| base-root-53 | 2026-11-08T21:33:36.790914+00:00 | 2026-11-08T21:34:07.856611+00:00 | 31.066 | 919544 | 0 |
| base-root-54 | 2026-11-08T21:34:07.857925+00:00 | 2026-11-08T21:34:35.894470+00:00 | 28.037 | 888416 | 0 |
| base-root-55 | 2026-11-08T21:34:35.896710+00:00 | 2026-11-08T21:35:05.999783+00:00 | 30.103 | 886480 | 0 |
| base-root-56 | 2026-11-08T21:35:06.003323+00:00 | 2026-11-08T21:35:36.042445+00:00 | 30.039 | 887268 | 0 |
| base-root-57 | 2026-11-08T21:35:36.043570+00:00 | 2026-11-08T21:36:04.078491+00:00 | 28.035 | 872808 | 0 |
| base-root-58 | 2026-11-08T21:36:04.080453+00:00 | 2026-11-08T21:36:33.183345+00:00 | 29.103 | 871960 | 0 |
| base-root-59 | 2026-11-08T21:36:33.185747+00:00 | 2026-11-08T21:37:00.215774+00:00 | 27.03 | 899176 | 0 |
| base-root-60 | 2026-11-08T21:37:00.217114+00:00 | 2026-11-08T21:37:28.244756+00:00 | 28.028 | 861948 | 0 |
| base-root-61 | 2026-11-08T21:47:20.505579+00:00 | 2026-11-08T21:47:46.539022+00:00 | 26.034 | 892812 | 0 |
| base-root-62 | 2026-11-08T21:47:46.544241+00:00 | 2026-11-08T21:48:13.586866+00:00 | 27.043 | 866760 | 0 |
| base-root-63 | 2026-11-08T21:48:13.588569+00:00 | 2026-11-08T21:48:41.721268+00:00 | 28.133 | 871752 | 0 |
| base-root-64 | 2026-11-08T21:48:41.724801+00:00 | 2026-11-08T21:49:06.743993+00:00 | 25.019 | 878684 | 0 |
| base-root-65 | 2026-11-08T21:49:06.744935+00:00 | 2026-11-08T21:49:32.762598+00:00 | 26.018 | 854144 | 0 |
| base-root-66 | 2026-11-08T21:49:32.767317+00:00 | 2026-11-08T21:49:58.794895+00:00 | 26.028 | 884828 | 0 |
| base-root-67 | 2026-11-08T21:49:58.798328+00:00 | 2026-11-08T21:50:25.927786+00:00 | 27.129 | 898156 | 0 |
| base-root-68 | 2026-11-08T21:50:25.929190+00:00 | 2026-11-08T21:50:53.005428+00:00 | 27.076 | 895672 | 0 |
| base-root-69 | 2026-11-08T21:50:53.006760+00:00 | 2026-11-08T21:51:24.173635+00:00 | 31.167 | 861904 | 0 |
| base-root-70 | 2026-11-08T21:51:24.183465+00:00 | 2026-11-08T21:51:52.216777+00:00 | 28.033 | 878008 | 0 |
| base-root-71 | 2026-11-08T21:51:52.217989+00:00 | 2026-11-08T21:52:18.240070+00:00 | 26.022 | 859884 | 0 |
| base-root-72 | 2026-11-08T21:52:18.241086+00:00 | 2026-11-08T21:52:45.269082+00:00 | 27.028 | 874128 | 0 |
| base-root-73 | 2026-11-08T21:52:45.270409+00:00 | 2026-11-08T21:52:53.284731+00:00 | 8.014 | 426616 | 0 |
| base-root-74 | 2026-11-08T21:52:53.289202+00:00 | 2026-11-08T21:53:22.333557+00:00 | 29.044 | 868612 | 0 |
| base-root-75 | 2026-11-08T21:53:22.334547+00:00 | 2026-11-08T21:53:48.356035+00:00 | 26.021 | 872720 | 0 |
| base-root-76 | 2026-11-08T21:53:48.357606+00:00 | 2026-11-08T21:54:15.411927+00:00 | 27.054 | 870412 | 0 |
| base-root-77 | 2026-11-08T21:54:15.413243+00:00 | 2026-11-08T21:54:44.455167+00:00 | 29.042 | 877292 | 0 |
| base-root-78 | 2026-11-08T21:54:44.457041+00:00 | 2026-11-08T21:55:13.521497+00:00 | 29.064 | 861028 | 0 |
| base-root-79 | 2026-11-08T21:55:13.523065+00:00 | 2026-11-08T21:55:40.551410+00:00 | 27.028 | 851128 | 0 |
| checkpoint-apply | 2026-11-08T20:49:22.572341+00:00 | 2026-11-08T20:49:25.576915+00:00 | 3.005 | 22360 | 0 |
| head-build | 2026-11-08T20:51:21.367844+00:00 | 2026-11-08T20:51:53.498046+00:00 | 32.13 | 1468608 | 0 |
| head-buildinfo | 2026-11-08T20:51:53.565085+00:00 | 2026-11-08T20:51:54.567413+00:00 | 1.002 | 47136 | 0 |
| head-checkpoint-core | 2026-11-08T20:52:53.840061+00:00 | 2026-11-08T20:52:56.843245+00:00 | 3.003 | 47136 | 0 |
| head-checkpoint-group-1 | 2026-11-08T21:09:41.735661+00:00 | 2026-11-08T21:09:44.740135+00:00 | 3.005 | 47136 | 0 |
| head-checkpoint-group-2 | 2026-11-08T21:27:54.066322+00:00 | 2026-11-08T21:27:57.070872+00:00 | 3.005 | 47136 | 0 |
| head-checkpoint-group-3 | 2026-11-08T21:47:17.490661+00:00 | 2026-11-08T21:47:20.504183+00:00 | 3.014 | 47136 | 0 |
| head-checkpoint-group-4 | 2026-11-08T22:04:16.478023+00:00 | 2026-11-08T22:04:19.486801+00:00 | 3.009 | 47136 | 0 |
| head-compiler-source-clean | 2026-11-08T20:51:20.365187+00:00 | 2026-11-08T20:51:21.367311+00:00 | 1.002 | 47104 | 0 |
| head-core | 2026-11-08T20:52:24.665739+00:00 | 2026-11-08T20:52:53.715921+00:00 | 29.05 | 899564 | 0 |
| head-latent-build | 2026-11-08T20:51:56.578757+00:00 | 2026-11-08T20:52:24.664859+00:00 | 28.086 | 1472836 | 0 |
| head-overlay-format | 2026-11-08T20:51:55.576418+00:00 | 2026-11-08T20:51:56.578202+00:00 | 1.002 | 47136 | 0 |
| head-overlay | 2026-11-08T20:51:54.568014+00:00 | 2026-11-08T20:51:55.571320+00:00 | 1.003 | 47136 | 0 |
| head-root-01 | 2026-11-08T21:01:22.833845+00:00 | 2026-11-08T21:01:51.892641+00:00 | 29.059 | 866264 | 0 |
| head-root-02 | 2026-11-08T21:01:51.893933+00:00 | 2026-11-08T21:02:21.007205+00:00 | 29.113 | 886156 | 0 |
| head-root-03 | 2026-11-08T21:02:21.011614+00:00 | 2026-11-08T21:02:51.063379+00:00 | 30.052 | 907748 | 0 |
| head-root-04 | 2026-11-08T21:02:51.065118+00:00 | 2026-11-08T21:03:19.091621+00:00 | 28.027 | 880868 | 0 |
| head-root-05 | 2026-11-08T21:03:19.096165+00:00 | 2026-11-08T21:03:48.135708+00:00 | 29.04 | 866872 | 0 |
| head-root-06 | 2026-11-08T21:03:48.142051+00:00 | 2026-11-08T21:04:15.181193+00:00 | 27.039 | 874432 | 0 |
| head-root-07 | 2026-11-08T21:04:15.183151+00:00 | 2026-11-08T21:04:42.202105+00:00 | 27.019 | 860716 | 0 |
| head-root-08 | 2026-11-08T21:04:42.205421+00:00 | 2026-11-08T21:05:08.235354+00:00 | 26.03 | 870352 | 0 |
| head-root-09 | 2026-11-08T21:05:08.236871+00:00 | 2026-11-08T21:05:35.279191+00:00 | 27.042 | 873996 | 0 |
| head-root-10 | 2026-11-08T21:05:35.282169+00:00 | 2026-11-08T21:06:02.316394+00:00 | 27.034 | 869736 | 0 |
| head-root-11 | 2026-11-08T21:06:02.317255+00:00 | 2026-11-08T21:06:34.356118+00:00 | 32.039 | 875460 | 0 |
| head-root-12 | 2026-11-08T21:06:34.357788+00:00 | 2026-11-08T21:06:35.363486+00:00 | 1.006 | 47136 | 0 |
| head-root-13 | 2026-11-08T21:06:35.368198+00:00 | 2026-11-08T21:06:43.380900+00:00 | 8.013 | 462580 | 0 |
| head-root-14 | 2026-11-08T21:06:43.384156+00:00 | 2026-11-08T21:06:51.391232+00:00 | 8.007 | 506728 | 0 |
| head-root-15 | 2026-11-08T21:06:51.392593+00:00 | 2026-11-08T21:07:20.515807+00:00 | 29.123 | 870524 | 0 |
| head-root-16 | 2026-11-08T21:07:20.517409+00:00 | 2026-11-08T21:07:51.606308+00:00 | 31.089 | 879404 | 0 |
| head-root-17 | 2026-11-08T21:07:51.608554+00:00 | 2026-11-08T21:08:17.651154+00:00 | 26.043 | 875544 | 0 |
| head-root-18 | 2026-11-08T21:08:17.655441+00:00 | 2026-11-08T21:08:46.685812+00:00 | 29.03 | 878956 | 0 |
| head-root-19 | 2026-11-08T21:08:46.689167+00:00 | 2026-11-08T21:09:14.715406+00:00 | 28.026 | 890064 | 0 |
| head-root-20 | 2026-11-08T21:09:14.716059+00:00 | 2026-11-08T21:09:41.734485+00:00 | 27.018 | 869280 | 0 |
| head-root-21 | 2026-11-08T21:18:53.102256+00:00 | 2026-11-08T21:19:22.160825+00:00 | 29.059 | 867052 | 0 |
| head-root-22 | 2026-11-08T21:19:22.162971+00:00 | 2026-11-08T21:19:49.191729+00:00 | 27.029 | 878764 | 0 |
| head-root-23 | 2026-11-08T21:19:49.193646+00:00 | 2026-11-08T21:20:13.219437+00:00 | 24.026 | 888164 | 0 |
| head-root-24 | 2026-11-08T21:20:13.224391+00:00 | 2026-11-08T21:20:41.249574+00:00 | 28.025 | 895184 | 0 |
| head-root-25 | 2026-11-08T21:20:41.252357+00:00 | 2026-11-08T21:21:09.286384+00:00 | 28.034 | 903456 | 0 |
| head-root-26 | 2026-11-08T21:21:09.288338+00:00 | 2026-11-08T21:21:37.330388+00:00 | 28.042 | 866348 | 0 |
| head-root-27 | 2026-11-08T21:21:37.340363+00:00 | 2026-11-08T21:22:06.436212+00:00 | 29.096 | 881632 | 0 |
| head-root-28 | 2026-11-08T21:22:06.437458+00:00 | 2026-11-08T21:22:07.443092+00:00 | 1.006 | 47136 | 0 |
| head-root-29 | 2026-11-08T21:22:07.447384+00:00 | 2026-11-08T21:22:37.521305+00:00 | 30.074 | 925492 | 0 |
| head-root-30 | 2026-11-08T21:22:37.525194+00:00 | 2026-11-08T21:23:05.561847+00:00 | 28.037 | 890144 | 0 |
| head-root-31 | 2026-11-08T21:23:05.564111+00:00 | 2026-11-08T21:23:32.593580+00:00 | 27.029 | 912516 | 0 |
| head-root-32 | 2026-11-08T21:23:32.594739+00:00 | 2026-11-08T21:23:59.627091+00:00 | 27.032 | 869476 | 0 |
| head-root-33 | 2026-11-08T21:23:59.628317+00:00 | 2026-11-08T21:24:29.730440+00:00 | 30.102 | 865044 | 0 |
| head-root-34 | 2026-11-08T21:24:29.733336+00:00 | 2026-11-08T21:25:00.757766+00:00 | 31.024 | 882844 | 0 |
| head-root-35 | 2026-11-08T21:25:00.758892+00:00 | 2026-11-08T21:25:29.783492+00:00 | 29.025 | 864916 | 0 |
| head-root-36 | 2026-11-08T21:25:29.784076+00:00 | 2026-11-08T21:25:57.821789+00:00 | 28.038 | 871060 | 0 |
| head-root-37 | 2026-11-08T21:25:57.827379+00:00 | 2026-11-08T21:26:25.865170+00:00 | 28.038 | 853556 | 0 |
| head-root-38 | 2026-11-08T21:26:25.867171+00:00 | 2026-11-08T21:26:52.898896+00:00 | 27.032 | 876152 | 0 |
| head-root-39 | 2026-11-08T21:26:52.901410+00:00 | 2026-11-08T21:27:25.034924+00:00 | 32.134 | 871092 | 0 |
| head-root-40 | 2026-11-08T21:27:25.036142+00:00 | 2026-11-08T21:27:54.064343+00:00 | 29.028 | 873192 | 0 |
| head-root-41 | 2026-11-08T21:37:31.251825+00:00 | 2026-11-08T21:37:59.283840+00:00 | 28.032 | 881368 | 0 |
| head-root-42 | 2026-11-08T21:37:59.284396+00:00 | 2026-11-08T21:38:26.316679+00:00 | 27.032 | 917484 | 0 |
| head-root-43 | 2026-11-08T21:38:26.317818+00:00 | 2026-11-08T21:38:54.351664+00:00 | 28.034 | 884916 | 0 |
| head-root-44 | 2026-11-08T21:38:54.353233+00:00 | 2026-11-08T21:39:21.381760+00:00 | 27.029 | 881952 | 0 |
| head-root-45 | 2026-11-08T21:39:21.383770+00:00 | 2026-11-08T21:39:49.412594+00:00 | 28.029 | 865456 | 0 |
| head-root-46 | 2026-11-08T21:39:49.416053+00:00 | 2026-11-08T21:40:20.546576+00:00 | 31.132 | 877076 | 0 |
| head-root-47 | 2026-11-08T21:40:20.562119+00:00 | 2026-11-08T21:40:50.621971+00:00 | 30.06 | 870448 | 0 |
| head-root-48 | 2026-11-08T21:40:50.623316+00:00 | 2026-11-08T21:41:17.657137+00:00 | 27.034 | 889536 | 0 |
| head-root-49 | 2026-11-08T21:41:17.658272+00:00 | 2026-11-08T21:42:00.464401+00:00 | 42.806 | 872252 | 0 |
| head-root-50 | 2026-11-08T21:42:00.467003+00:00 | 2026-11-08T21:42:49.075966+00:00 | 48.609 | 917360 | 0 |
| head-root-51 | 2026-11-08T21:42:49.077004+00:00 | 2026-11-08T21:43:14.096043+00:00 | 25.019 | 867492 | 0 |
| head-root-52 | 2026-11-08T21:43:14.106117+00:00 | 2026-11-08T21:43:45.196534+00:00 | 31.09 | 863088 | 0 |
| head-root-53 | 2026-11-08T21:43:45.197644+00:00 | 2026-11-08T21:44:12.225304+00:00 | 27.028 | 870960 | 0 |
| head-root-54 | 2026-11-08T21:44:12.231671+00:00 | 2026-11-08T21:44:39.270398+00:00 | 27.039 | 865140 | 0 |
| head-root-55 | 2026-11-08T21:44:39.274046+00:00 | 2026-11-08T21:45:06.300390+00:00 | 27.026 | 907512 | 0 |
| head-root-56 | 2026-11-08T21:45:06.305057+00:00 | 2026-11-08T21:45:33.332630+00:00 | 27.028 | 901984 | 0 |
| head-root-57 | 2026-11-08T21:45:33.334315+00:00 | 2026-11-08T21:46:01.372919+00:00 | 28.039 | 874100 | 0 |
| head-root-58 | 2026-11-08T21:46:01.374190+00:00 | 2026-11-08T21:46:27.429160+00:00 | 26.055 | 880484 | 0 |
| head-root-59 | 2026-11-08T21:46:27.434186+00:00 | 2026-11-08T21:46:53.471434+00:00 | 26.037 | 883316 | 0 |
| head-root-60 | 2026-11-08T21:46:53.472654+00:00 | 2026-11-08T21:47:17.489184+00:00 | 24.017 | 870920 | 0 |
| head-root-61 | 2026-11-08T21:55:43.564365+00:00 | 2026-11-08T21:56:11.591232+00:00 | 28.027 | 875240 | 0 |
| head-root-62 | 2026-11-08T21:56:11.594002+00:00 | 2026-11-08T21:56:38.622475+00:00 | 27.028 | 872828 | 0 |
| head-root-63 | 2026-11-08T21:56:38.624403+00:00 | 2026-11-08T21:57:05.660765+00:00 | 27.036 | 882392 | 0 |
| head-root-64 | 2026-11-08T21:57:05.661447+00:00 | 2026-11-08T21:57:32.688964+00:00 | 27.028 | 864936 | 0 |
| head-root-65 | 2026-11-08T21:57:32.698731+00:00 | 2026-11-08T21:58:04.737256+00:00 | 32.039 | 858964 | 0 |
| head-root-66 | 2026-11-08T21:58:04.738333+00:00 | 2026-11-08T21:58:31.780894+00:00 | 27.043 | 873028 | 0 |
| head-root-67 | 2026-11-08T21:58:31.782637+00:00 | 2026-11-08T21:58:59.912775+00:00 | 28.13 | 875596 | 0 |
| head-root-68 | 2026-11-08T21:58:59.915781+00:00 | 2026-11-08T21:59:26.947115+00:00 | 27.031 | 876296 | 0 |
| head-root-69 | 2026-11-08T21:59:26.948385+00:00 | 2026-11-08T21:59:55.974257+00:00 | 29.026 | 863736 | 0 |
| head-root-70 | 2026-11-08T21:59:55.979820+00:00 | 2026-11-08T22:00:26.060196+00:00 | 30.08 | 865288 | 0 |
| head-root-71 | 2026-11-08T22:00:26.061045+00:00 | 2026-11-08T22:00:54.089407+00:00 | 28.028 | 867292 | 0 |
| head-root-72 | 2026-11-08T22:00:54.090812+00:00 | 2026-11-08T22:01:22.129399+00:00 | 28.039 | 882872 | 0 |
| head-root-73 | 2026-11-08T22:01:22.131156+00:00 | 2026-11-08T22:01:30.141541+00:00 | 8.01 | 507492 | 0 |
| head-root-74 | 2026-11-08T22:01:30.145909+00:00 | 2026-11-08T22:01:58.175368+00:00 | 28.029 | 885108 | 0 |
| head-root-75 | 2026-11-08T22:01:58.176420+00:00 | 2026-11-08T22:02:24.203836+00:00 | 26.027 | 867928 | 0 |
| head-root-76 | 2026-11-08T22:02:24.205677+00:00 | 2026-11-08T22:02:51.228842+00:00 | 27.023 | 874524 | 0 |
| head-root-77 | 2026-11-08T22:02:51.231300+00:00 | 2026-11-08T22:03:18.311095+00:00 | 27.08 | 870848 | 0 |
| head-root-78 | 2026-11-08T22:03:18.313188+00:00 | 2026-11-08T22:03:46.408458+00:00 | 28.095 | 910492 | 0 |
| head-root-79 | 2026-11-08T22:03:46.409735+00:00 | 2026-11-08T22:04:16.470033+00:00 | 30.06 | 867316 | 0 |
| head-worktree | 2026-11-08T20:51:04.298041+00:00 | 2026-11-08T20:51:20.360553+00:00 | 16.063 | 1008260 | 0 |
| npm-ci | 2026-11-08T20:41:04.442492+00:00 | 2026-11-08T20:41:07.457085+00:00 | 3.015 | 117364 | 0 |
| render-report | 2026-11-08T22:04:32.658109+00:00 | 2026-11-08T22:04:33.663567+00:00 | 1.005 | 47036 | 0 |
| runner-check | 2026-11-08T20:32:54.679799+00:00 | 2026-11-08T20:32:55.682997+00:00 | 1.003 | 10624 | 0 |
| source-checkout | 2026-11-08T20:35:20.817522+00:00 | 2026-11-08T20:41:04.429585+00:00 | 343.612 | 176548 | 0 |
| source-clean | 2026-11-08T22:04:30.435232+00:00 | 2026-11-08T22:04:31.437601+00:00 | 1.002 | 28236 | 0 |

## Method and evidence

One complete apply from base supplies the identical adapted tree for both sides. Head changes no adapter files. Ordinary tools and guarded full latent tools were built from the respective exact commits. Full latent runs use LATENT_FULL=1, LATENT_ONLY_FILE set to the adapted core.ts, LATENT_ASSERT_NO_OUTPUT=1, and LATENT_TIMES for progress. The latent tool receives all compiler roots as load context, then selects core.ts alone. No usable IR is produced. Deferred generic bodies and independent lowering findings are not evidence of successful whole-program lowering. Refusal scan precedes lowering, so source order across phases is not the ordinary stop sequence.

The c3b22184 reference method was read directly from commit 9714ab2dfe3a5dab10c5cbc7e42fb9238f91278f. The named core-sites-af42d3dc report was not present in the fetched main or current core-sites and core-sites-checkpoint branches. The available core-sites-4186581e report and runners were inspected for the same explicitly requested full latent mode.

Completed apply and each root group were pushed append-only to typescript/require-row-census-checkpoint. The final branch has one report commit directly on 71b126b0. No main push or force push.

Raw streams, logs, overlays and status files are in raw-streams.tar.gz. Binaries are excluded; binary hashes and build information are retained. Verify with sha256sum -c raw-streams.tar.gz.sha256. runner.py records exact child exits and reaps all process groups, with a 30-minute no-output bound.
