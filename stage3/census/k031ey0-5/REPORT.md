# k031ey0 -5 paired census: partial checkpoint

Evidence only. The first 60 pairs were pushed at the requested checkpoint. Continuation groups are pushed after each group of at most 10 roots.

Parent: `f44b5c272a39bcbd7cd2fb278939ac1c1832ce2e`. Head: `e64bc123c51fc333133684f0abb3bee8e522acf3`. TypeScript v6.0.3 upstream: `050880ce59e30b356b686bd3144efe24f875ebc8`. One adapted tree prepared from parent; stage3 is identical at head. Head production sources are unmodified. The diagnostic seam is scratch-only, saved in [seam.diff](seam.diff), and was compiled but has not been run.

## Measured so far

| Side | Completed roots | First observations (s) | Duplicate observations (s) | All measured child wall (s) |
|---|---:|---:|---:|---:|
| parent | 79/79 | 1020.133749 | 988.566959 | 2015.806160 |
| head | 79/79 | 1015.804640 | 1005.022321 | 2030.314151 |

Changed first stops among completed pairs: **0**. First stop compares kind, location, exact reason and diagnostics. All duplicate observations agree in every non-timing field.

parent: **0** roots stop first on `a view of a type that reaches ...`.
head: **0** roots stop first on `a view of a type that reaches ...`.

Entry latent census: **not run on either side**. No latent unit counts, whole-pass wall times, collectViewedElements prepass timings, or stderr prepass diagnostic have been measured. Absence of a diagnostic here is not a negative latent result.

## Timing and load

The first-observation seconds are the census tool’s load+Lower elapsed time. Child wall includes both observations and process overhead; totals exclude builds and adaptation. This is a single alternating measurement, with no statistical performance claim.

parent: box `d312b35f425b`; 5 available CPUs; cpu.max `400000 100000` (4 CPU quota). Start load [2.01220703125, 4.4755859375, 3.11669921875]; end load [1.07763671875, 1.23974609375, 1.24853515625]; sampled 1-minute range 0.25–2.01. Load triples are 1/5/15-minute averages. Every child’s before/after loads and UTC timestamps are in its status JSON.

head: box `d312b35f425b`; 5 available CPUs; cpu.max `400000 100000` (4 CPU quota). Start load [1.93701171875, 4.259765625, 3.080078125]; end load [1.36669921875, 1.2822265625, 1.26171875]; sampled 1-minute range 0.51–1.97. Load triples are 1/5/15-minute averages. Every child’s before/after loads and UTC timestamps are in its status JSON.

Roots run sequentially, parent then head, in groups of at most 20; no measured children overlap. Each child has a 1780-second bound. All completed children exited 0. The initial checkpoint completed roots 1–60 in groups of 20. Continuation uses groups 61–70 and 71–79, pushed separately before proceeding. Both latent passes are not yet started.

## Unrun

parent: 0 ordinary roots:
head: 0 ordinary roots:

Both parent and head entry latent censuses remain unrun, including LATENT_TIMES, latentPrepassTimer output and stderr seam observation.

## Per-root stops and seconds

| Root | Parent first stop | Head first stop | Parent seconds | Head seconds | Parent child wall | Head child wall |
|---|---|---|---:|---:|---:|---:|
| 1: `_namespaces/ts.moduleSpecifiers.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 15.332043 | 14.381307 | 28.821974 | 27.696084 |
| 2: `_namespaces/ts.performance.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 14.508507 | 13.178272 | 27.331545 | 26.975726 |
| 3: `_namespaces/ts.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 14.165130 | 13.705832 | 28.163785 | 26.596885 |
| 4: `binder.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.378378 | 13.673639 | 27.239263 | 26.454935 |
| 5: `builder.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.049549 | 13.968751 | 26.397452 | 27.124146 |
| 6: `builderPublic.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.955286 | 13.984143 | 27.729880 | 28.122706 |
| 7: `builderState.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.690254 | 13.828538 | 26.833451 | 26.930444 |
| 8: `builderStatePublic.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.755724 | 13.613995 | 26.908937 | 27.009309 |
| 9: `checker.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.188924 | 13.419509 | 26.439512 | 27.043345 |
| 10: `commandLineParser.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.809866 | 14.773836 | 27.832873 | 28.897695 |
| 11: `core.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.346171 | 13.975127 | 26.677522 | 27.467068 |
| 12: `corePublic.ts` | accepted | accepted | 0.043321 | 0.042908 | 0.117734 | 0.114888 |
| 13: `debug.ts` | NotYet semver.ts:71:9 a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | NotYet semver.ts:71:9 a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | 3.818508 | 3.430662 | 7.319936 | 6.859434 |
| 14: `diagnosticInformationMap.generated.ts` | NotYet commandLineParser.ts:293:35 reading an enum before its runtime initialization; move the call after the enum declaration | NotYet commandLineParser.ts:293:35 reading an enum before its runtime initialization; move the call after the enum declaration | 3.832435 | 3.755403 | 7.424078 | 7.101719 |
| 15: `emitter.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 15.259076 | 14.222530 | 29.632638 | 28.527300 |
| 16: `executeCommandLine.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.038079 | 13.249059 | 26.165566 | 26.836859 |
| 17: `expressionToTypeNode.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.620581 | 13.608292 | 27.615129 | 27.522744 |
| 18: `factory/baseNodeFactory.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.681336 | 13.234089 | 26.870149 | 25.972780 |
| 19: `factory/emitHelpers.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.209344 | 12.655239 | 25.949414 | 25.410981 |
| 20: `factory/emitNode.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 12.727737 | 13.126802 | 25.744810 | 25.677811 |
| 21: `factory/nodeChildren.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 14.554759 | 13.687033 | 27.753665 | 27.508191 |
| 22: `factory/nodeConverters.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.249633 | 13.171565 | 26.658236 | 26.259228 |
| 23: `factory/nodeFactory.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.638710 | 13.674907 | 28.221023 | 25.945434 |
| 24: `factory/nodeTests.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.749178 | 12.987762 | 26.522558 | 25.660428 |
| 25: `factory/parenthesizerRules.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 12.920272 | 14.114600 | 26.591741 | 28.456742 |
| 26: `factory/utilities.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 14.238400 | 12.739033 | 27.063651 | 26.186602 |
| 27: `factory/utilitiesPublic.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.201887 | 12.894423 | 27.045932 | 32.232035 |
| 28: `hostErrors.ts` | accepted | accepted | 0.064932 | 0.065047 | 0.165339 | 0.115114 |
| 29: `moduleNameResolver.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.621344 | 15.070689 | 27.456362 | 36.776760 |
| 30: `moduleSpecifiers.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 14.646407 | 14.886439 | 28.044223 | 28.887498 |
| 31: `parser.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.882557 | 13.300231 | 26.870274 | 27.386273 |
| 32: `path.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.344122 | 13.182835 | 26.429813 | 27.347508 |
| 33: `performance.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 15.477234 | 13.150449 | 28.045428 | 25.881056 |
| 34: `performanceCore.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 12.520794 | 13.029398 | 25.437635 | 26.271251 |
| 35: `program.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 14.684248 | 17.374131 | 27.722645 | 31.920971 |
| 36: `programDiagnostics.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.293512 | 13.443379 | 26.134844 | 26.314856 |
| 37: `resolutionCache.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.900825 | 13.955695 | 28.046127 | 29.131991 |
| 38: `scanner.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 14.128679 | 13.063925 | 27.702127 | 26.110667 |
| 39: `semver.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.774311 | 14.773316 | 28.102740 | 28.365698 |
| 40: `sourcemap.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 14.141069 | 13.883285 | 27.642035 | 28.106234 |
| 41: `symbolWalker.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 14.275513 | 13.300280 | 27.281342 | 26.690009 |
| 42: `sys.ts` | Refused path.ts:1105:28 a cast the runtime can't check | Refused path.ts:1105:28 a cast the runtime can't check | 13.726728 | 13.197929 | 26.562220 | 26.094679 |
| 43: `tracing.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 12.664800 | 12.961556 | 25.641291 | 25.990213 |
| 44: `transformer.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.089685 | 13.061762 | 26.247696 | 25.996043 |
| 45: `transformers/classFields.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.722419 | 13.551226 | 28.841128 | 26.949905 |
| 46: `transformers/classThis.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.854147 | 15.377913 | 28.500026 | 29.500510 |
| 47: `transformers/declarations/diagnostics.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.308668 | 13.523332 | 26.870492 | 26.789442 |
| 48: `transformers/declarations.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.655508 | 13.958276 | 27.705178 | 26.971253 |
| 49: `transformers/destructuring.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 12.488386 | 13.199763 | 25.326621 | 26.433828 |
| 50: `transformers/es2015.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.197648 | 12.995601 | 26.115222 | 25.971788 |
| 51: `transformers/es2016.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.161080 | 13.096061 | 26.050996 | 26.143620 |
| 52: `transformers/es2017.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.180575 | 13.560734 | 26.265681 | 26.846586 |
| 53: `transformers/es2018.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.897367 | 13.153736 | 27.280092 | 26.074405 |
| 54: `transformers/es2019.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.079579 | 12.965378 | 26.058095 | 26.232364 |
| 55: `transformers/es2020.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.298039 | 14.326176 | 26.348824 | 28.843020 |
| 56: `transformers/es2021.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 16.041639 | 14.017174 | 30.084609 | 27.807559 |
| 57: `transformers/esDecorators.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.921504 | 13.526231 | 27.170446 | 26.735069 |
| 58: `transformers/esnext.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.270370 | 13.395045 | 26.760531 | 26.909715 |
| 59: `transformers/generators.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 14.371806 | 13.571028 | 28.752012 | 27.571465 |
| 60: `transformers/jsx.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.436229 | 15.196168 | 27.896377 | 29.454350 |
| 61: `transformers/legacyDecorators.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 15.399753 | 12.985447 | 27.808643 | 26.005601 |
| 62: `transformers/module/esnextAnd2015.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.105748 | 13.741891 | 26.547658 | 27.039620 |
| 63: `transformers/module/impliedNodeFormatDependent.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.192489 | 13.189469 | 26.367447 | 26.336536 |
| 64: `transformers/module/module.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.550148 | 14.090598 | 26.740911 | 27.561982 |
| 65: `transformers/module/system.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.684140 | 13.339230 | 27.459253 | 25.805785 |
| 66: `transformers/namedEvaluation.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 12.875507 | 12.433685 | 25.281678 | 24.966018 |
| 67: `transformers/taggedTemplate.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 12.763547 | 12.850084 | 25.767393 | 25.677162 |
| 68: `transformers/ts.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.094377 | 13.424622 | 25.664326 | 25.953942 |
| 69: `transformers/typeSerializer.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.970567 | 13.155709 | 26.668251 | 26.578187 |
| 70: `transformers/utilities.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.201269 | 13.016134 | 26.416910 | 26.752903 |
| 71: `tsbuild.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 15.356068 | 12.610989 | 28.334091 | 25.002737 |
| 72: `tsbuildPublic.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 12.660396 | 13.242846 | 25.210579 | 26.087534 |
| 73: `types.ts` | NotYet diagnosticInformationMap.generated.ts:13:45 reading an enum before its runtime initialization; move the call after the enum declaration | NotYet diagnosticInformationMap.generated.ts:13:45 reading an enum before its runtime initialization; move the call after the enum declaration | 3.298155 | 3.267093 | 6.201956 | 6.373004 |
| 74: `utilities.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.482000 | 13.012620 | 26.323946 | 26.537722 |
| 75: `utilitiesPublic.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.003786 | 13.426927 | 25.963678 | 27.923376 |
| 76: `visitorPublic.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 13.026230 | 12.952924 | 25.447627 | 26.198430 |
| 77: `watch.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 12.849701 | 12.588529 | 25.497573 | 25.821669 |
| 78: `watchPublic.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 12.860484 | 13.198048 | 25.771202 | 25.995420 |
| 79: `watchUtilities.ts` | Refused sys.ts:57:1 an ambient function declaration | Refused sys.ts:57:1 an ambient function declaration | 12.674575 | 15.066349 | 25.702111 | 28.483301 |

Raw JSONL observations, stderr logs and command/exit/load status JSON are adjacent to this report. Setup logs are under `preparation/`. See [COMMANDS.md](COMMANDS.md) and executable [measure.py](measure.py). `analyze.py` is saved for the later full run; this checkpoint report is produced by `partial-report.py`.
