# Baobab: next stops on 464e5daa

Both formats independently measured all 79 sorted compiler roots. All 79 first-stop (location, kind, exact reason) tuples agree between formats. Each root has two raw observations agreeing except timing: 316 raw observations across both modes.

CommonJSBundle was completed, validated and pushed at ef97b8ca2 before any EcmaScript root ran. The final branch is one evidence commit directly on 464e5daa23a60588f5aaa656e95d93af05d582c3. Checkpoint history is retained separately on typescript/next-stops-464e5daa-checkpoint.

## Grouped tables

### CommonJSBundle

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:142:81 | Refused | a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | refusal_scan | internal/lower/invariance.go: (*lowering).wideningRefusal (via refuseWidening) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | scanner.ts:469:12 | Refused | a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go: (*lowering).methodValueStop (via methodValueRefusal) |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |

### EcmaScript

| Roots | Location | Kind | Exact reason | Phase | Owner |
| ---: | --- | --- | --- | --- | --- |
| 73 | sys.ts:142:81 | Refused | a value of type typeof PollingInterval seen as Levels, whose readonly field Low becomes writable: a readonly field may hold something narrower than PollingInterval.Low, which a write of number would replace | refusal_scan | internal/lower/invariance.go: (*lowering).wideningRefusal (via refuseWidening) |
| 2 | - | accepted | - | accepted | lower.LowerWithOptions (accepted; no stop) |
| 1 | commandLineParser.ts:293:35 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | diagnosticInformationMap.generated.ts:13:45 | NotYet | reading an enum before its runtime initialization; move the call after the enum declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |
| 1 | scanner.ts:469:12 | Refused | a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it) | refusal_scan | internal/lower/refusals.go: (*lowering).methodValueStop (via methodValueRefusal) |
| 1 | semver.ts:71:9 | NotYet | a namespace read before runtime initialization, directly or through a reachable call; move that read or call after the namespace declaration | lowering (initialization preflight) | internal/lower/namespaces.go: (*lowering).namespaceInitialization |

## Changes from the requested baselines

Bundle versus stage3/meter/README.md, Nov 13 method-change reading: 74 changed, 5 unchanged. Of the 73 formerly at core.ts:2599:15 (process), 72 now stop at sys.ts:142:81 and sys.ts itself stops at scanner.ts:469:12. core.ts itself moves from performanceCore.ts:29:15 to sys.ts:142:81, bringing the new shared group to 73.

ES versus typescript/next-stops-f44b5c27-2: all 73 formerly at sys.ts:57:1 now stop at sys.ts:142:81; the other 6 roots are unchanged.

The sys.ts root is Refused at scanner.ts:469:12 in both modes, exactly: a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it).

Files lowering whole in both modes: corePublic.ts and hostErrors.ts, 2/79. No native-output success is claimed.

## Reproduction and evidence

The six requested scripts were copied from 3dc2a59dede1bff6eef557cb987f6ea0a3e6a462 and adapted to the new pins, paths and per-format evidence directories. The requested source README is absent; REPORT.md and tooling-adaptations.md supply its recipe. Fresh-run changes are detailed in each mode report and prior-tooling-adaptations.md. measure.py records the bounded invocation recipe. group.py runs at most ten roots, three at a time after the initial sequential bundle batch; run.py bounds each group and child to 1780 seconds and reports every child exit.

stage3/apply.sh ran once. Its tree and ordinary census binary were reused for both formats. Root status JSON and raw observations record module_format; every root invocation explicitly inherited CENSUS_MODULE_FORMAT. No ordinary compiler or census-tool overlay was used. Only this new evidence directory differs from the measured commit.

The initial census build exited 1 because the fresh worktree lacked checker submodules. Pinned submodule initialization and the retry exited 0. These actual exits remain in status files; no failed root measurement is omitted.

Each mode REPORT.md contains the full grouped table, all root memberships, owners, phases and comparisons. comparison.json retains exact before/after tuples; source-evidence.json retains adapted excerpts and emitter names. owners.json comes from the measured commit; code methods identify stops without an exact owner mapping. Retained .go.txt files show dispatch and phase ordering.

validation.json in each mode checks all roots, duplicate agreement, actual exits, duration bounds, explicit format, baseline tuples, binary/source hashes and source integrity. cross-format-validation.json records independent agreement. binaries.json and ordinary-buildinfo.txt identify the built ordinary binary. provenance.json identifies compiler, tooling, TypeScript and environment.

Completed logs are compressed losslessly; compression.json keeps raw and compressed SHA256. Artifacts above 20 MB are omitted in favor of their SHA256; none of the committed artifacts exceeds 20 MB. artifact-manifest.json identifies every final evidence file except itself.

## Checkpoints

| Checkpoint | Commit |
| --- | --- |
| baobab: CommonJSBundle checkpoint 10/79 roots | 95c09f00c4d00bdb5d895b7fd1bc42cdd420dcf5 |
| baobab: CommonJSBundle checkpoint 20/79 roots | 204471632709ec5487362e6910d9eafa92cf2559 |
| baobab: CommonJSBundle checkpoint 30/79 roots | 6a4e4913ffba59b64c3d0438c568578df7275769 |
| baobab: CommonJSBundle checkpoint 40/79 roots | f0db03478d6848e9a50e5cc70410b6d77a707bf3 |
| baobab: CommonJSBundle checkpoint 50/79 roots | 99cfa4639edd53fdabe584084312e3b4dd9b7f6f |
| baobab: CommonJSBundle checkpoint 60/79 roots | f89676a193ef98e91992f724f7511def54d8ff99 |
| baobab: CommonJSBundle checkpoint 70/79 roots | b3fecd1c2e33ed46a5a144ac40b5c3b0cd0032ef |
| baobab: CommonJSBundle checkpoint 79/79 roots | ef97b8ca2819275dc195bbc6b594c9d4aab0b1ae |
| baobab: EcmaScript checkpoint 10/79 roots | eec7139781e04575ba8be4a3a3530722e532ccc0 |
| baobab: EcmaScript checkpoint 20/79 roots | 20c3d00a7e40a300adf50496e1484f55ce7d8d05 |
| baobab: EcmaScript checkpoint 30/79 roots | 1f2fc9fe112d49fea935c28f280a8d85db0fab3a |
| baobab: EcmaScript checkpoint 40/79 roots | b0fa9b945b93a6c67323e7b4468f2d79e7d97aaa |
| baobab: EcmaScript checkpoint 50/79 roots | 454b768029360dd575f4096c81e0b72f64186db7 |
| baobab: EcmaScript checkpoint 60/79 roots | 05ab3455c5e592655c5e8acebff9aac5b2675f48 |
| baobab: EcmaScript checkpoint 70/79 roots | 4a6b2b83f7d96cbd9386801cb86c1a23724c0d09 |
| baobab: EcmaScript checkpoint 79/79 roots | 2e34469322dc393360da325c05e5d80a73c4312a |
