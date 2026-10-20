Whole program 2/79 (Oct 8: 2); own file 56/79 (56); lowering 2/79 (0).<br>
tsc entry: 235 diagnostics (Oct 8: 320).<br>
Measured main: `295df3bc6aa3d47933ccf675da1c33e745856313`.<br>
Meter elapsed: 35m 38s including recovery; nproc 5, CPU quota 4.<br>
Not measured: native output/execution, exhaustive lowering, or the complete compiler gate.

# Stage 3 meter on main 295df3bc

Main and the historical area label both measure the same main commit above. The retired area/stage3 branch was not fetched or measured. Each compiler was built independently in its own detached checkout. Both ordinary scores are 2/79 whole program, 56/79 own file, and 2/79 lowering. The lowering passes are `src/compiler/corePublic.ts` and `src/compiler/hostErrors.ts`. Lowering here is the ordinary census's successful IR lowering, without invoking a backend; it is not a native binary claim.

The tsc entry fails with 235 diagnostics on each tree, 85 fewer than Oct 8. Its first diagnostic is `src/compiler/builder.ts:2273:9 TS2375`: the object supplied as `ReusableBuilderProgramState` has an `outSignature` that may be undefined. The first diagnostic is no longer builder.ts:1246:69 TS2345. The 56-file own-checker pass set is exactly the Oct 8 pass set: no additions or losses. This comparison spans all changes on main since Oct 8 and does not isolate one commit's causal effect.

The denominator is verified against upstream: 77 checked-in compiler sources, upstream's generated `diagnosticInformationMap.generated.ts`, and adaptation 47's `hostErrors.ts`. No upstream root was removed. The 78 original/generated upstream roots measure whole program 1/78 and own file 55/78; hostErrors.ts supplies the extra pass in each. Three JSON/config inputs are listed by the ordinary census but fail the extension gate and are excluded from the denominator.

## Commands and pinning

The checked-in `twice-daily.sh` reads `STAGE3_METER_COMPILER`, `STAGE3_METER_LATENT`, and `STAGE3_METER_RUNS`. It reads neither `STAGE3_METER_MAIN_REF` nor `STAGE3_METER_AREA_REF`, and unconditionally fetches retired area/stage3. After `git fetch origin` and cutting `typescript/stage3-meter-295df3bc` from origin/main, the retained [run-local copy](validation/pinned-twice-daily.sh) resolves both supplied SHAs locally. Its [compressed diff](validation/launcher.diff.gz) also records the fixed output directory, local dependency transport, and overlay compatibility adapter. The report's area ref name is the reporter's historical label; its SHA is the main SHA above.

The bounded initial launch, captured by [launch.sh](validation/launch.sh), was:

```sh
run=$PWD/stage3/meter/runs/20261019T233755Z.9xwivl15
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
STAGE3_METER_MAIN_REF=295df3bc6aa3d47933ccf675da1c33e745856313 \
STAGE3_METER_AREA_REF=295df3bc6aa3d47933ccf675da1c33e745856313 \
STAGE3_METER_COMPILER=per-ref STAGE3_METER_LATENT=first-error \
STAGE3_METER_RUNS=$PWD/stage3/meter/runs \
timeout --kill-after=60s 14400 bash "$run/validation/pinned-twice-daily.sh" > "$run/meter-start.log" 2>&1
```

That launch stopped before measurement on overlay incompatibilities. Main's ordinary command bodies were resumed under `timeout --kill-after=60s 5400 bash -c ...`; they are retained verbatim in [main-ordinary-recovery.sh](validation/main-ordinary-recovery.sh). The remaining commands ran under these bounded controllers:

```sh
timeout --kill-after=60s 14400 bash "$run/validation/recover-run.sh" > "$run/meter-recovery.log" 2>&1
timeout --kill-after=60s 14400 bash "$run/validation/recover-area.sh" > "$run/area-recovery.log" 2>&1
timeout --kill-after=60s 14400 python3 "$run/validation/finalize-run.py" > "$run/validation/finalization.log" 2>&1
```

The first recovery controller was paused while its main latent child continued, permitting area to run independently, as in Oct 8's recovery. The child exited 0 and all 79 records passed `latent_summary` before the paused controller was retired. No compiler was interrupted. Its controller exit 143 is intentional; [main-latent-completion.json](validation/main-latent-completion.json), area-recovery.exit, and [measurement-completion.json](validation/measurement-completion.json) record the successful measurement exits. Both latent commands used `LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1`. All output went to logs, never through a pipe. Background jobs were checked throughout; the meter stayed below two hours.

## Recovery and limits

- The original snapshot generator rejected `impliedTargetQuery func(*ast.Node) *checker.Type`. This is a test-only query override, assigned only in compiler tests; production leaves it nil. The [run-local adapter](validation/statecopy-adapter.diff.gz) recognizes only this exact field/type and preserves nil, rejecting a populated hook. Unknown function fields still fail. First-error statement lowering does not use full-mode snapshots. Nil preservation, rejection of populated/unknown hooks, and a permissive-hook mutant were exercised.
- The refusal generator rejected new outer `return err` preflights. Its [adapter](validation/refusal-adapter.diff.gz) accepts only the exact `referenceRefusal(module)` and `parallelPreflight(module)` conditionals and records their errors. Unknown preflights and missing child walks still fail before output. Production refusal code is retained by the original AST clone mechanism.
- A cold apply had no adapted-product manifest and no `ADAMIC_TYPESCRIPT_SOURCE`. TypeScript v6.0.3 was cloned at `050880ce59e30b356b686bd3144efe24f875ebc8`; `ADAMIC_TYPESCRIPT_SOURCE=/tmp/badrnrf-inputs/TypeScript` supplied it. The stage3/api lockfile was installed with `npm ci --ignore-scripts --no-audit --no-fund` in scratch storage and linked into each detached checkout. Unmodified apply verified its pinned package digest. Main built the adapted product locally; area restored that verified keyed product into a fresh directory. Initial failure logs are retained under validation/initial-overlay-failure and validation/second-overlay-failure.
- Dependency clones used the verified local checkouts as transport, then checked out each recorded pin: cohere `0183ccf8c57a775f874ca73be1adfc06e9ad1562`, checker TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. The inherited Go flags are `-buildvcs=false -trimpath`, so the binaries do not carry a VCS stamp. Source pin/diff checks, executable hashes, build information, and input manifests are retained in [provenance.json](validation/provenance.json). Both compiler checkouts have unchanged tracked sources. Ordinary compiler options and prelude were not changed. Go build-information logs have trailing whitespace stripped in their readable views; exact raw `.log.gz` copies and [SHA-256 records](validation/buildinfo-log-formatting.json) are retained. Census JSONL and diagnostic bytes are unmodified. The initial literal reason-key equality assertion failed because TypeScript renders absolute adapted-tree paths inside some type names; comparison of the complete finding sites showed zero differences after replacing the known tree prefix. The corrected provenance check normalizes that prefix in reason keys as well. Raw reports and diagnostic streams are unchanged, and the initial assertion log is retained.
- The guarded census is partial: diagnosed bodies and deferred generics may not be examined, and fresh isolated lowering state can differ from ordinary rooted lowering. No backend, native driver, native-output proof, whole compiler gate, or milestone update was run. This version of the meter did not write stage3/progress.json.

## Validation

The current checkout discovers 36 meter tests, whereas Oct 8's retained README reports 46. The prior progress/root/symlink test additions and writer are absent from this main revision; they were not recreated outside the unit's territory.

```sh
timeout 1800 python3 -m unittest discover -s stage3/meter -p '*_test.py' > "$run/validation/tests-default.log" 2>&1
CENSUS_BINARY="$binaries/census" LATENT_CENSUS_BINARY="$binaries/latent-census" ENTRY_CENSUS_BINARY="$binaries/entry-census" \
  timeout 1800 python3 -m unittest discover -s stage3/meter -p '*_test.py' > "$run/validation/tests-real-binaries.log" 2>&1
```

Here `$binaries=/tmp/adamic-gate/stage3-meter.RXmuND/main-binaries`. The default invocation passes with four optional probes skipped. With all three actual binaries enabled, 34/36 tests pass with no skips. The two failures are original entry fixtures that expect `debugger` or string/number conditions to remain lowering blockers: observed NotYet is 0 instead of positive, and Refused is 1 instead of 3. The real checker imported-error attribution and planted latent-finding probes pass unchanged.

Run-local copies of those two entry probes use stable `var` witnesses; no production test was edited. Both pass, including clean imported reach, unloaded-file exclusion, imported-error blocking, nested rejected-entry measurement, and the first-error-only mutant. Their [exact fixture diff](validation/entry-probes.diff.gz) is retained. The command was:

```sh
PYTHONPATH=$PWD/stage3/meter CENSUS_BINARY="$binaries/census" LATENT_CENSUS_BINARY="$binaries/latent-census" ENTRY_CENSUS_BINARY="$binaries/entry-census" \
  timeout 300 python3 -m unittest discover -s "$run/validation/current-probes" -p entry_test.py -k CensusTests > "$run/validation/current-entry-probes.log" 2>&1
```

Additional checks, each with retained commands and witness logs:

- `report-mutants.py`: treating whole-program status as own-file status fails the planted imported-diagnostic assertion; omitting Refused reasons fails the combined-reason assertion.
- `METER_SCRIPT_UNDER_TEST=.../compiler-selection-mutant.sh timeout 300 python3 -m unittest discover -s stage3/meter -p compiler_test.py -k test_per_ref_builds_and_runs_each_pinned_compiler`: fails on the actual `compiled with checkout` witness despite claimed main metadata.
- `check-statecopy-adapter.py` and `check-refusal-adapter.py`: the narrow compatibility guards and their malformed/permissive mutants behave as described above.
- `timeout 1200 python3 stage3/census/latent/audit_output_guards.py /tmp/adamic-gate/stage3-meter.RXmuND/main-source /tmp/adamic-gate/stage3-meter.RXmuND/main-binaries/latent-overlay /tmp/badrnrf-output-guard-mutants`: both usable-IR and permissive-loader mutants are caught by their actual output guards.
- `lower-ledger.py` checks the 79-file coverage, own-file flags/diagnostic counts and unique latent totals against the meter report. `ledger-mutants.py` corrupts real own-file status, removes a coverage record while preserving every finding, and drops Refused findings. All three fail their intended assertions. `provenance.py` verifies pinned clean tracked sources, input hashes and paired observations.

## Setup and timing

`export GOPROXY='https://proxy.golang.org|direct'; timeout 5400 bash cloud/setup.sh > /tmp/badrnrf-setup.log 2>&1` succeeded, then `/workspace/adamic-tools/env.sh` was sourced. Node 24.19.0, Go 1.27.1, clang 20.1.8; nproc 5; cgroup cpu.max `400000 100000`. Setup prints cumulative elapsed times:

```text
setup: node ready (0.065s)
setup: go ready (0.094s)
setup: clang ready (0.482s)
setup: markdown dependencies ready (1.042s)
setup: submodules ready (374.654s)
setup: go build ready (523.938s)
setup: build cache warm (524.062s)
setup: done on 5 processors (524.093s)
```

The complete timing lines and host flags are in [setup.log](validation/setup.log). Meter elapsed above starts with the first launch after setup and includes failed preparation/recovery and overlapping paired work. It is an elapsed run duration, not a compiler benchmark.

## Top lowering reasons and owners

These are unique latent sites on a checker-rejected program, in historical first-error mode. Main records NotYet 766, Refused 4149, SkippedDependency 2, errors 0, panics 2. Area records NotYet 766, Refused 4149, SkippedDependency 2, errors 0, panics 2. OWNER BLANK rows go to @system_adamic; ownership is the pinned meter owners.json mapping.

| Reason | Owner | NotYet | Refused | Total |
| --- | --- | ---: | ---: | ---: |
| an object refinement using an open numeric enum as a literal tag | OWNER BLANK | 0 | 1182 | 1182 |
| a cast the runtime can&#x27;t check | 01a11410 | 0 | 581 | 581 |
| a type predicate whose return is not proven (return expression is not a trusted check on node) | 01a1143b-d691 | 0 | 388 | 388 |
| an unproven predicate argument for parameter test (argument &quot;isExpression&quot;) | OWNER BLANK | 0 | 169 | 169 |
| an unproven predicate argument for parameter test (argument &quot;isTypeNode&quot;) | OWNER BLANK | 0 | 75 | 75 |
| a union or optional field read in a program with record storage | OWNER BLANK | 70 | 0 | 70 |
| a record value of type TsConfigSourceFile &#124; CompilerOptionsValue | OWNER BLANK | 54 | 0 | 54 |
| an indirect call of a checked predicate overload | OWNER BLANK | 38 | 0 | 38 |
| an unproven predicate argument for parameter test (argument &quot;isModifierLike&quot;) | OWNER BLANK | 0 | 37 | 37 |
| a type predicate whose return is not proven (return paths through KindSwitchStatement are not verified) | 01a1143b-d691 | 0 | 36 | 36 |

Both entry lowering censuses are blocked by the entry's checker diagnostics, as required in first-error mode. The directory latent census does not establish successful lowering. Its two retained main panic sites are moduleSpecifiers.ts:335:1 (nil pointer dereference) and transformers/declarations/diagnostics.ts:1:1 (Node.Text on ComputedPropertyName). Skipped dependency bodies are watch.ts:843:1 and sys.ts:1463:27. Full records are in [latent-diagnostics.json](latent-diagnostics.json) and the compressed JSONL streams.

In the table below, hostErrors.ts has an observed latent NotYet despite passing ordinary rooted lowering. Its directory-wide rejected-program context and isolated unit state differ from its ordinary root program. The ordinary score remains 2/79. Six other own-file passes have no observed latent NotYet/Refused; their imported checker failures still prevent an ordinary lowering attempt.

## For Lower: every source file

First primary diagnostic or NotYet/Refused by source line, UTF-16 column, then kind/reason. Findings are attributed to their actual file across all census attempts; both refusal_scan and lowering phases count.

Own-file pass means no primary diagnostic located in that file across the ordinary compiler-root census. Imported diagnostics can still block ordinary lowering. The latent record reaches every file, but diagnosed bodies and deferred generics may remain unexamined. A recorded finding is an observed blocker, not a successful native compilation. First finding here is source order, not event order or a claim that production lowering reached the file. Full raw diagnostic elaborations remain in census.jsonl.gz and lower-ledger.json.

| File | Own-file checker and first diagnostic | Ordinary lowering | Latent reach and first observed NotYet/Refused (for own-file passes) |
| --- | --- | --- | --- |
| src/compiler/_namespaces/ts.moduleSpecifiers.ts | pass | blocked by checker | yes; no NotYet/Refused observed; partial census, not a lowering pass; units attempted=1 |
| src/compiler/_namespaces/ts.performance.ts | pass | blocked by checker | yes; no NotYet/Refused observed; partial census, not a lowering pass; units attempted=1 |
| src/compiler/_namespaces/ts.ts | pass | blocked by checker | yes; no NotYet/Refused observed; partial census, not a lowering pass; units attempted=77 |
| src/compiler/binder.ts | pass | blocked by checker | yes; src/compiler/binder.ts:375:41 Refused: an object refinement using an open numeric enum as a literal tag (phase refusal_scan; owner OWNER BLANK); units attempted=16 |
| src/compiler/builder.ts | fail: src/compiler/builder.ts:2273:9 TS2375: Type &#x27;{ fileInfos: Map&lt;Path, BuilderState.FileInfo&gt;; compilerOptions: CompilerOptions; semanticDiagnosticsPerFile: Map&lt;Path, readonly ReusableDiagnostic[]&gt;; ... 7 more ...; checkPending: true &#124; undefined; }&#x27; is not assignable to type &#x27;ReusableBuilderProgramState&#x27; with &#x27;exactOptionalPropertyTypes: true&#x27;. Consider adding &#x27;undefined&#x27; to the types of the target&#x27;s properties.<br>  Types of property &#x27;outSignature&#x27; are incompatible.<br>    Type &#x27;EmitSignature &#124; undefined&#x27; is not assignable to type &#x27;EmitSignature&#x27;.<br>      Type &#x27;undefined&#x27; is not assignable to type &#x27;EmitSignature&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/builderPublic.ts | pass | blocked by checker | yes; no NotYet/Refused observed; partial census, not a lowering pass; units attempted=16 |
| src/compiler/builderState.ts | pass | blocked by checker | yes; src/compiler/builderState.ts:254:60 Refused: an object refinement using an open numeric enum as a literal tag (phase refusal_scan; owner OWNER BLANK); units attempted=5 |
| src/compiler/builderStatePublic.ts | pass | blocked by checker | yes; no NotYet/Refused observed; partial census, not a lowering pass; units attempted=3 |
| src/compiler/checker.ts | fail: src/compiler/checker.ts:1683:9 TS2322: Type &#x27;(signature: Signature, kind: SignatureDeclaration[&quot;kind&quot;], enclosingDeclaration?: Node, flags?: NodeBuilderFlags, internalFlags?: InternalNodeBuilderFlags, tracker?: SymbolTracker, maximumLength?: number, verbosityLevel?: number, out?: WriterContextOut) =&gt; SignatureDeclaration &#124; undefined&#x27; is not assignable to type &#x27;{ (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node &#124; undefined, flags: NodeBuilderFlags &#124; undefined): (SignatureDeclaration &amp; ...) &#124; undefined; (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node &#124; undefined, flags: NodeBuilderFlags &#124; undefined, internalFlags?: InternalNodeBuilderF...&#x27;.<br>  Type &#x27;SignatureDeclaration &#124; undefined&#x27; is not assignable to type &#x27;(SignatureDeclaration &amp; { typeArguments?: NodeArray&lt;TypeNode&gt;; }) &#124; undefined&#x27;.<br>    Type &#x27;ArrowFunction&#x27; is not assignable to type &#x27;(SignatureDeclaration &amp; { typeArguments?: NodeArray&lt;TypeNode&gt;; }) &#124; undefined&#x27;.<br>      Type &#x27;ArrowFunction&#x27; is not assignable to type &#x27;ArrowFunction &amp; { typeArguments?: NodeArray&lt;TypeNode&gt;; }&#x27; with &#x27;exactOptionalPropertyTypes: true&#x27;. Consider adding &#x27;undefined&#x27; to the types of the target&#x27;s properties.<br>        Type &#x27;ArrowFunction&#x27; is not assignable to type &#x27;{ typeArguments?: NodeArray&lt;TypeNode&gt;; }&#x27; with &#x27;exactOptionalPropertyTypes: true&#x27;. Consider adding &#x27;undefined&#x27; to the types of the target&#x27;s properties.<br>          Types of property &#x27;typeArguments&#x27; are incompatible.<br>            Type &#x27;NodeArray&lt;TypeNode&gt; &#124; undefined&#x27; is not assignable to type &#x27;NodeArray&lt;TypeNode&gt;&#x27;.<br>              Type &#x27;undefined&#x27; is not assignable to type &#x27;NodeArray&lt;TypeNode&gt;&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/commandLineParser.ts | fail: src/compiler/commandLineParser.ts:2077:79 TS2345: Argument of type &#x27;string &#124; undefined&#x27; is not assignable to parameter of type &#x27;string&#x27;.<br>  Type &#x27;undefined&#x27; is not assignable to type &#x27;string&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/core.ts | fail: src/compiler/core.ts:1637:11 TS2740: Type &#x27;{ has(element: TElement): boolean; add(element: TElement): Set&lt;TElement&gt;; delete(element: TElement): boolean; clear(): void; ... 6 more ...; [Symbol.toStringTag]: string; }&#x27; is missing the following properties from type &#x27;Set&lt;TElement&gt;&#x27;: union, intersection, difference, symmetricDifference, and 3 more. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/corePublic.ts | pass | accepted | yes; no NotYet/Refused observed; ordinary lowering accepted, no backend tested; units attempted=9 |
| src/compiler/debug.ts | fail: src/compiler/debug.ts:200:19 TS2339: Property &#x27;captureStackTrace&#x27; does not exist on type &#x27;ErrorConstructor&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/diagnosticInformationMap.generated.ts | pass | blocked by checker | yes; no NotYet/Refused observed; partial census, not a lowering pass; units attempted=3 |
| src/compiler/emitter.ts | pass | blocked by checker | yes; src/compiler/emitter.ts:454:66 Refused: an object refinement using an open numeric enum as a literal tag (phase refusal_scan; owner OWNER BLANK); units attempted=48, deferred_generic_instantiation=2 |
| src/compiler/executeCommandLine.ts | pass | blocked by checker | yes; src/compiler/executeCommandLine.ts:109:41 Refused: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields (phase refusal_scan; owner OWNER BLANK); units attempted=49, deferred_generic_instantiation=1 |
| src/compiler/expressionToTypeNode.ts | pass | blocked by checker | yes; src/compiler/expressionToTypeNode.ts:167:12 Refused: a cast the runtime can&#x27;t check (phase lowering; owner 01a11410); units attempted=9 |
| src/compiler/factory/baseNodeFactory.ts | pass | blocked by checker | yes; src/compiler/factory/baseNodeFactory.ts:42:16 NotYet: new a ParenthesizedExpression (phase lowering; owner OWNER BLANK); units attempted=3 |
| src/compiler/factory/emitHelpers.ts | pass | blocked by checker | yes; src/compiler/factory/emitHelpers.ts:156:9 Refused: a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method&#x27;s parameters both ways), so it can be handed what it can&#x27;t take (phase refusal_scan; owner adaptation 70, stage 3); units attempted=45 |
| src/compiler/factory/emitNode.ts | pass | blocked by checker | yes; src/compiler/factory/emitNode.ts:43:40 Refused: a cast the runtime can&#x27;t check (phase lowering; owner 01a11410); units attempted=18, deferred_generic_instantiation=23 |
| src/compiler/factory/nodeChildren.ts | pass | blocked by checker | yes; src/compiler/factory/nodeChildren.ts:11:34 NotYet: new an Identifier (phase lowering; owner OWNER BLANK); units attempted=6 |
| src/compiler/factory/nodeConverters.ts | pass | blocked by checker | yes; src/compiler/factory/nodeConverters.ts:55:35 Refused: an object refinement using an open numeric enum as a literal tag (phase refusal_scan; owner OWNER BLANK); units attempted=3 |
| src/compiler/factory/nodeFactory.ts | pass | blocked by checker | yes; src/compiler/factory/nodeFactory.ts:522:9 Refused: a function taking readonly T[] &#124; undefined seen as one taking readonly T[] &#124; undefined (tsc relates a method&#x27;s parameters both ways), so it can be handed what it can&#x27;t take (phase refusal_scan; owner adaptation 70, stage 3); units attempted=25, deferred_generic_instantiation=1 |
| src/compiler/factory/nodeTests.ts | pass | blocked by checker | yes; src/compiler/factory/nodeTests.ts:236:5 Refused: a type predicate whose return is not proven (return expression is not a trusted check on node) (phase refusal_scan; owner 01a1143b-d691); units attempted=228 |
| src/compiler/factory/parenthesizerRules.ts | pass | blocked by checker | yes; src/compiler/factory/parenthesizerRules.ts:63:9 Refused: a function taking BinaryOperator seen as one taking SyntaxKind (tsc relates a method&#x27;s parameters both ways), so it can be handed what it can&#x27;t take (phase refusal_scan; owner adaptation 70, stage 3); units attempted=3 |
| src/compiler/factory/utilities.ts | pass | blocked by checker | yes; src/compiler/factory/utilities.ts:190:75 Refused: an object refinement using an open numeric enum as a literal tag (phase refusal_scan; owner OWNER BLANK); units attempted=100, deferred_generic_instantiation=8 |
| src/compiler/factory/utilitiesPublic.ts | pass | blocked by checker | yes; src/compiler/factory/utilitiesPublic.ts:16:5 Refused: a type predicate whose return is not proven (return expression is not a trusted check on node) (phase refusal_scan; owner 01a1143b-d691); units attempted=3, deferred_generic_instantiation=1 |
| src/compiler/hostErrors.ts | pass | accepted | yes; src/compiler/hostErrors.ts:3:59 NotYet: in when a possibly absent spread has no property presence descriptors (phase lowering; owner OWNER BLANK); units attempted=2 |
| src/compiler/moduleNameResolver.ts | fail: src/compiler/moduleNameResolver.ts:132:13 TS2375: Type &#x27;{ name: string; subModuleName: string; version: string; peerDependencies: string &#124; undefined; }&#x27; is not assignable to type &#x27;PackageId&#x27; with &#x27;exactOptionalPropertyTypes: true&#x27;. Consider adding &#x27;undefined&#x27; to the types of the target&#x27;s properties.<br>  Types of property &#x27;peerDependencies&#x27; are incompatible.<br>    Type &#x27;string &#124; undefined&#x27; is not assignable to type &#x27;string&#x27;.<br>      Type &#x27;undefined&#x27; is not assignable to type &#x27;string&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/moduleSpecifiers.ts | pass | blocked by checker | yes; src/compiler/moduleSpecifiers.ts:138:31 NotYet: RegExp with a nonconstant pattern (phase lowering; owner OWNER BLANK); units attempted=47, deferred_generic_instantiation=1, panic=1 |
| src/compiler/parser.ts | pass | blocked by checker | yes; src/compiler/parser.ts:433:39 NotYet: new a ParenthesizedExpression (phase lowering; owner OWNER BLANK); units attempted=53, deferred_generic_instantiation=22 |
| src/compiler/path.ts | pass | blocked by checker | yes; src/compiler/path.ts:278:1 Refused: overload 1 of getDirectoryPath result Path cannot be served by implementation result string (phase lowering; owner OWNER BLANK); units attempted=77, deferred_generic_instantiation=4 |
| src/compiler/performance.ts | pass | blocked by checker | yes; src/compiler/performance.ts:15:1 Refused: an ambient value declaration (phase refusal_scan; owner OWNER BLANK); units attempted=24 |
| src/compiler/performanceCore.ts | pass | blocked by checker | yes; src/compiler/performanceCore.ts:28:1 Refused: an ambient value declaration (phase refusal_scan; owner OWNER BLANK); units attempted=12 |
| src/compiler/program.ts | fail: src/compiler/program.ts:406:25 TS18046: &#x27;e&#x27; is of type &#x27;unknown&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/programDiagnostics.ts | pass | blocked by checker | yes; src/compiler/programDiagnostics.ts:168:52 Refused: a value of type never[] seen as DiagnosticArguments, which can write string &#124; number &#124; boolean &#124; readonly string[] &#124; SourceFile &#124; undefined where never is read (phase refusal_scan; owner adaptation 70, stage 3); units attempted=5 |
| src/compiler/resolutionCache.ts | fail: src/compiler/resolutionCache.ts:990:43 TS2379: Argument of type &#x27;{ entries: readonly T[]; containingFile: string; containingSourceFile: SourceFile &#124; undefined; redirectedReference: ResolvedProjectReference &#124; undefined; ... 6 more ...; deferWatchingNonRelativeResolution: false; }&#x27; is not assignable to parameter of type &#x27;ResolveNamesWithLocalCacheInput&lt;string &#124; FileReference, SourceFile &#124; undefined, ResolvedTypeReferenceDirectiveWithFailedLookupLocations, ResolvedTypeReferenceDirective&gt;&#x27; with &#x27;exactOptionalPropertyTypes: true&#x27;. Consider adding &#x27;undefined&#x27; to the types of the target&#x27;s properties.<br>  Types of property &#x27;reusedNames&#x27; are incompatible.<br>    Type &#x27;readonly T[] &#124; undefined&#x27; is not assignable to type &#x27;readonly (string &#124; FileReference)[]&#x27;.<br>      Type &#x27;undefined&#x27; is not assignable to type &#x27;readonly (string &#124; FileReference)[]&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/scanner.ts | pass | blocked by checker | yes; src/compiler/scanner.ts:224:29 Refused: Object enumeration with getter or accessor descriptors (phase lowering; owner OWNER BLANK); units attempted=74, deferred_generic_instantiation=10 |
| src/compiler/semver.ts | pass | blocked by checker | yes; src/compiler/semver.ts:60:59 Refused: a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string &#124; number; minor: number; patch: number; prerelease: string &#124; readonly string[]; build: string &#124; readonly string[]; }, which can write string &#124; number where number is read (phase refusal_scan; owner adaptation 70, stage 3); units attempted=29 |
| src/compiler/sourcemap.ts | fail: src/compiler/sourcemap.ts:216:58 TS2345: Argument of type &#x27;string &#124; null &#124; undefined&#x27; is not assignable to parameter of type &#x27;string &#124; null&#x27;.<br>  Type &#x27;undefined&#x27; is not assignable to type &#x27;string &#124; null&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/symbolWalker.ts | pass | blocked by checker | yes; src/compiler/symbolWalker.ts:49:23 Refused: a value of type (type: Type) =&gt; { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) =&gt; { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read (phase refusal_scan; owner adaptation 70, stage 3); units attempted=2 |
| src/compiler/sys.ts | fail: src/compiler/sys.ts:51:68 TS2503: Cannot find namespace &#x27;NodeJS&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/tracing.ts | pass | blocked by checker | yes; src/compiler/tracing.ts:96:19 NotYet: a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports (phase refusal_scan; owner codex/namespaces-tsc); units attempted=10 |
| src/compiler/transformer.ts | fail: src/compiler/transformer.ts:367:9 TS2532: Object is possibly &#x27;undefined&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/transformers/classFields.ts | fail: src/compiler/transformers/classFields.ts:1835:21 TS2532: Object is possibly &#x27;undefined&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/transformers/classThis.ts | pass | blocked by checker | yes; src/compiler/transformers/classThis.ts:33:10 NotYet: a function returning ClassThisAssignmentBlock (phase lowering; owner OWNER BLANK); units attempted=5, deferred_generic_instantiation=2 |
| src/compiler/transformers/declarations.ts | fail: src/compiler/transformers/declarations.ts:1678:110 TS2345: Argument of type &#x27;ExpressionWithTypeArguments &#124; undefined&#x27; is not assignable to parameter of type &#x27;DeclarationDiagnosticProducing&#x27;.<br>  Type &#x27;undefined&#x27; is not assignable to type &#x27;DeclarationDiagnosticProducing&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/transformers/declarations/diagnostics.ts | pass | blocked by checker | yes; src/compiler/transformers/declarations/diagnostics.ts:139:5 Refused: a type predicate whose return is not proven (return expression is not a trusted check on node) (phase refusal_scan; owner 01a1143b-d691); units attempted=8 |
| src/compiler/transformers/destructuring.ts | pass | blocked by checker | yes; src/compiler/transformers/destructuring.ts:107:31 Refused: a value of type VariableDeclaration &#124; DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace (phase refusal_scan; owner adaptation 70, stage 3); units attempted=22 |
| src/compiler/transformers/es2015.ts | fail: src/compiler/transformers/es2015.ts:2803:38 TS2532: Object is possibly &#x27;undefined&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/transformers/es2016.ts | pass | blocked by checker | yes; src/compiler/transformers/es2016.ts:63:52 Refused: an unproven predicate argument for parameter test (argument &quot;isExpression&quot;) (phase refusal_scan; owner OWNER BLANK); units attempted=2 |
| src/compiler/transformers/es2017.ts | fail: src/compiler/transformers/es2017.ts:599:86 TS2532: Object is possibly &#x27;undefined&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/transformers/es2018.ts | pass | blocked by checker | yes; src/compiler/transformers/es2018.ts:399:148 Refused: an unproven predicate argument for parameter test (argument &quot;isExpression&quot;) (phase refusal_scan; owner OWNER BLANK); units attempted=4 |
| src/compiler/transformers/es2019.ts | pass | blocked by checker | yes; src/compiler/transformers/es2019.ts:46:48 Refused: an unproven predicate argument for parameter test (argument &quot;isBlock&quot;) (phase refusal_scan; owner OWNER BLANK); units attempted=2 |
| src/compiler/transformers/es2020.ts | pass | blocked by checker | yes; src/compiler/transformers/es2020.ts:65:17 NotYet: an indirect call of a checked predicate overload (phase refusal_scan; owner OWNER BLANK); units attempted=2 |
| src/compiler/transformers/es2021.ts | pass | blocked by checker | yes; src/compiler/transformers/es2021.ts:53:78 Refused: an unproven predicate argument for parameter test (argument &quot;isLeftHandSideExpression&quot;) (phase refusal_scan; owner OWNER BLANK); units attempted=2 |
| src/compiler/transformers/esDecorators.ts | fail: src/compiler/transformers/esDecorators.ts:351:9 TS2322: Type &#x27;{ kind: &quot;class&quot;; next: LexicalEnvironmentStackEntry &#124; undefined; classInfo: ClassInfo &#124; undefined; savedPendingExpressions: Expression[] &#124; undefined; }&#x27; is not assignable to type &#x27;LexicalEnvironmentStackEntry &#124; undefined&#x27;.<br>  Type &#x27;{ kind: &quot;class&quot;; next: LexicalEnvironmentStackEntry &#124; undefined; classInfo: ClassInfo &#124; undefined; savedPendingExpressions: Expression[] &#124; undefined; }&#x27; is not assignable to type &#x27;ClassLexicalEnvironmentStackEntry&#x27; with &#x27;exactOptionalPropertyTypes: true&#x27;. Consider adding &#x27;undefined&#x27; to the types of the target&#x27;s properties.<br>    Types of property &#x27;savedPendingExpressions&#x27; are incompatible.<br>      Type &#x27;Expression[] &#124; undefined&#x27; is not assignable to type &#x27;Expression[]&#x27;.<br>        Type &#x27;undefined&#x27; is not assignable to type &#x27;Expression[]&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/transformers/esnext.ts | fail: src/compiler/transformers/esnext.ts:183:34 TS2345: Argument of type &#x27;Statement &#124; undefined&#x27; is not assignable to parameter of type &#x27;Statement&#x27;.<br>  Type &#x27;undefined&#x27; is not assignable to type &#x27;Statement&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/transformers/generators.ts | fail: src/compiler/transformers/generators.ts:1283:39 TS2345: Argument of type &#x27;Statement &#124; undefined&#x27; is not assignable to parameter of type &#x27;Statement&#x27;.<br>  Type &#x27;undefined&#x27; is not assignable to type &#x27;Statement&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/transformers/jsx.ts | fail: src/compiler/transformers/jsx.ts:309:58 TS2345: Argument of type &#x27;JsxChild &#124; undefined&#x27; is not assignable to parameter of type &#x27;JsxChild&#x27;.<br>  Type &#x27;undefined&#x27; is not assignable to type &#x27;JsxChild&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/transformers/legacyDecorators.ts | pass | blocked by checker | yes; src/compiler/transformers/legacyDecorators.ts:183:73 NotYet: a value of type NodeArray&lt;ClassElement&gt; (phase lowering; owner OWNER BLANK); units attempted=2 |
| src/compiler/transformers/module/esnextAnd2015.ts | fail: src/compiler/transformers/module/esnextAnd2015.ts:100:46 TS2345: Argument of type &#x27;Expression &#124; undefined&#x27; is not assignable to parameter of type &#x27;FileReference &#124; Node&#x27;.<br>  Type &#x27;undefined&#x27; is not assignable to type &#x27;FileReference &#124; Node&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts | pass | blocked by checker | yes; src/compiler/transformers/module/impliedNodeFormatDependent.ts:89:16 Refused: an object refinement using an open numeric enum as a literal tag (phase refusal_scan; owner OWNER BLANK); units attempted=2 |
| src/compiler/transformers/module/module.ts | pass | blocked by checker | yes; src/compiler/transformers/module/module.ts:204:42 NotYet: a CompilerOptions seen as Pick&lt;CompilerOptions, &quot;module&quot; &#124; &quot;target&quot;&gt; (fixed objects and records have different storage; copy explicitly) (phase refusal_scan; owner adaptation 70, stage 3); units attempted=3 |
| src/compiler/transformers/module/system.ts | pass | blocked by checker | yes; src/compiler/transformers/module/system.ts:206:22 NotYet: an assignment value to a member (phase lowering; owner OWNER BLANK); units attempted=2 |
| src/compiler/transformers/namedEvaluation.ts | pass | blocked by checker | yes; src/compiler/transformers/namedEvaluation.ts:56:78 NotYet: a value of type WrappedExpression&lt;AnonymousFunctionDefinition&gt; (phase lowering; owner OWNER BLANK); units attempted=19, deferred_generic_instantiation=2 |
| src/compiler/transformers/taggedTemplate.ts | pass | blocked by checker | yes; src/compiler/transformers/taggedTemplate.ts:44:46 Refused: an unproven predicate argument for parameter test (argument &quot;isExpression&quot;) (phase refusal_scan; owner OWNER BLANK); units attempted=5 |
| src/compiler/transformers/ts.ts | pass | blocked by checker | yes; src/compiler/transformers/ts.ts:248:42 NotYet: a CompilerOptions seen as Pick&lt;CompilerOptions, &quot;module&quot; &#124; &quot;target&quot;&gt; (fixed objects and records have different storage; copy explicitly) (phase refusal_scan; owner adaptation 70, stage 3); units attempted=5 |
| src/compiler/transformers/typeSerializer.ts | pass | blocked by checker | yes; src/compiler/transformers/typeSerializer.ts:213:64 Refused: an object refinement using an open numeric enum as a literal tag (phase refusal_scan; owner OWNER BLANK); units attempted=6 |
| src/compiler/transformers/utilities.ts | pass | blocked by checker | yes; src/compiler/transformers/utilities.ts:127:17 NotYet: a value of type NodeArray&lt;ExportSpecifier&gt; &#124; NodeArray&lt;ImportSpecifier&gt; (phase lowering; owner OWNER BLANK); units attempted=44, deferred_generic_instantiation=6 |
| src/compiler/tsbuild.ts | pass | blocked by checker | yes; src/compiler/tsbuild.ts:177:17 NotYet: a function returning ResolvedConfigFileName (phase lowering; owner OWNER BLANK); units attempted=5 |
| src/compiler/tsbuildPublic.ts | fail: src/compiler/tsbuildPublic.ts:302:5 TS2412: Type &#x27;(() =&gt; Date) &#124; undefined&#x27; is not assignable to type &#x27;() =&gt; Date&#x27; with &#x27;exactOptionalPropertyTypes: true&#x27;. Consider adding &#x27;undefined&#x27; to the type of the target.<br>  Type &#x27;undefined&#x27; is not assignable to type &#x27;() =&gt; Date&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/types.ts | pass | blocked by checker | yes; src/compiler/types.ts:5913:37 Refused: a type predicate whose return is not proven (there is no body proving this parameter) (phase refusal_scan; owner 01a1143b-d691); units attempted=875 |
| src/compiler/utilities.ts | pass | blocked by checker | yes; src/compiler/utilities.ts:612:45 NotYet: an array of never (phase lowering; owner OWNER BLANK); units attempted=834, deferred_generic_instantiation=59 |
| src/compiler/utilitiesPublic.ts | pass | blocked by checker | yes; src/compiler/utilitiesPublic.ts:616:5 Refused: a type predicate whose return is not proven (return expression is not a trusted check on node) (phase refusal_scan; owner 01a1143b-d691); units attempted=242, deferred_generic_instantiation=9 |
| src/compiler/visitorPublic.ts | pass | blocked by checker | yes; src/compiler/visitorPublic.ts:123:5 Refused: overload 1 of visitNode parameter node cannot be served by implementation parameter node (phase lowering; owner OWNER BLANK); units attempted=28, deferred_generic_instantiation=13 |
| src/compiler/watch.ts | fail: src/compiler/watch.ts:756:11 TS2375: Type &#x27;{ getSourceFile: (fileName: string, languageVersionOrOptions: CreateSourceFileOptions &#124; ScriptTarget, onError?: ((message: string) =&gt; void) &#124; undefined, shouldCreateNewSourceFile?: boolean &#124; undefined) =&gt; SourceFile &#124; undefined; ... 17 more ...; jsDocParsingMode: JSDocParsingMode &#124; undefined; }&#x27; is not assignable to type &#x27;CompilerHost&#x27; with &#x27;exactOptionalPropertyTypes: true&#x27;. Consider adding &#x27;undefined&#x27; to the types of the target&#x27;s properties.<br>  Types of property &#x27;getDefaultLibLocation&#x27; are incompatible.<br>    Type &#x27;(() =&gt; string) &#124; undefined&#x27; is not assignable to type &#x27;() =&gt; string&#x27;.<br>      Type &#x27;undefined&#x27; is not assignable to type &#x27;() =&gt; string&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/watchPublic.ts | fail: src/compiler/watchPublic.ts:125:5 TS2412: Type &#x27;((data: string) =&gt; string) &#124; undefined&#x27; is not assignable to type &#x27;(data: string) =&gt; string&#x27; with &#x27;exactOptionalPropertyTypes: true&#x27;. Consider adding &#x27;undefined&#x27; to the type of the target.<br>  Type &#x27;undefined&#x27; is not assignable to type &#x27;(data: string) =&gt; string&#x27;. | blocked by checker | own-file checker fails; latent record retained |
| src/compiler/watchUtilities.ts | pass | blocked by checker | yes; src/compiler/watchUtilities.ts:124:34 Refused: a cast the runtime can&#x27;t check (phase refusal_scan; owner 01a11410); units attempted=21, deferred_generic_instantiation=5 |

### Count by first-stop class

Each of the 79 files contributes once: its first TS code if its own checker fails, otherwise its first observed lowering reason. No observed finding stays an explicit coverage limitation. There are 23 own-file checker stops, 49 files with a first latent NotYet/Refused, six blocked ordinary roots without a latent finding, and one ordinary lowering pass with no latent finding. The second ordinary lowering pass, hostErrors.ts, is among the 49 observed latent findings because that program context differs.

| Class | Files |
| --- | ---: |
| Refused: an object refinement using an open numeric enum as a literal tag | 7 |
| TS2345 | 7 |
| no NotYet/Refused observed (partial census) | 6 |
| Refused: a type predicate whose return is not proven (return expression is not a trusted check on node) | 4 |
| TS2532 | 4 |
| Refused: a cast the runtime can&#x27;t check | 3 |
| Refused: an unproven predicate argument for parameter test (argument &quot;isExpression&quot;) | 3 |
| TS2375 | 3 |
| NotYet: a CompilerOptions seen as Pick&lt;CompilerOptions, &quot;module&quot; &#124; &quot;target&quot;&gt; (fixed objects and records have different storage; copy explicitly) | 2 |
| NotYet: new a ParenthesizedExpression | 2 |
| Refused: an ambient value declaration | 2 |
| TS2322 | 2 |
| TS2412 | 2 |
| NotYet: RegExp with a nonconstant pattern | 1 |
| NotYet: a function returning ClassThisAssignmentBlock | 1 |
| NotYet: a function returning ResolvedConfigFileName | 1 |
| NotYet: a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports | 1 |
| NotYet: a value of type NodeArray&lt;ClassElement&gt; | 1 |
| NotYet: a value of type NodeArray&lt;ExportSpecifier&gt; &#124; NodeArray&lt;ImportSpecifier&gt; | 1 |
| NotYet: a value of type WrappedExpression&lt;AnonymousFunctionDefinition&gt; | 1 |
| NotYet: an array of never | 1 |
| NotYet: an assignment value to a member | 1 |
| NotYet: an indirect call of a checked predicate overload | 1 |
| NotYet: in when a possibly absent spread has no property presence descriptors | 1 |
| NotYet: new an Identifier | 1 |
| Refused: Object enumeration with getter or accessor descriptors | 1 |
| Refused: a function taking BinaryOperator seen as one taking SyntaxKind (tsc relates a method&#x27;s parameters both ways), so it can be handed what it can&#x27;t take | 1 |
| Refused: a function taking Expression[] seen as one taking readonly Expression[] (tsc relates a method&#x27;s parameters both ways), so it can be handed what it can&#x27;t take | 1 |
| Refused: a function taking readonly T[] &#124; undefined seen as one taking readonly T[] &#124; undefined (tsc relates a method&#x27;s parameters both ways), so it can be handed what it can&#x27;t take | 1 |
| Refused: a type predicate whose return is not proven (there is no body proving this parameter) | 1 |
| Refused: a value of type (type: Type) =&gt; { visitedTypes: Type[]; visitedSymbols: Symbol[]; } seen as (root: Type) =&gt; { visitedTypes: readonly Type[]; visitedSymbols: readonly Symbol[]; }, which can write readonly Type[] where Type[] is read | 1 |
| Refused: a value of type VariableDeclaration &#124; DestructuringAssignment seen as TextRange, whose readonly field pos becomes writable: a readonly field may hold something narrower than number, which a write of number would replace | 1 |
| Refused: a value of type never[] seen as DiagnosticArguments, which can write string &#124; number &#124; boolean &#124; readonly string[] &#124; SourceFile &#124; undefined where never is read | 1 |
| Refused: a value of type { major: number; minor: number; patch: number; prerelease: string; build: string; } seen as { major: string &#124; number; minor: number; patch: number; prerelease: string &#124; readonly string[]; build: string &#124; readonly string[]; }, which can write string &#124; number where number is read | 1 |
| Refused: an unproven predicate argument for parameter test (argument &quot;isBlock&quot;) | 1 |
| Refused: an unproven predicate argument for parameter test (argument &quot;isLeftHandSideExpression&quot;) | 1 |
| Refused: optional property getPositionOfLineAndCharacter in SourceFileLike absent from structural source SourceFile, which can hide fields | 1 |
| Refused: overload 1 of getDirectoryPath result Path cannot be served by implementation result string | 1 |
| Refused: overload 1 of visitNode parameter node cannot be served by implementation parameter node | 1 |
| TS18046 | 1 |
| TS2339 | 1 |
| TS2379 | 1 |
| TS2503 | 1 |
| TS2740 | 1 |
| accepted by ordinary lowering; no first stop | 1 |
| Total | 79 |
