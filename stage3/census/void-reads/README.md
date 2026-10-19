Measured 79 adapted compiler roots from main 141fec7702d4aed21429724b2afd7efcd980f277.<br>
Found 491 void-typed reads: 383 returns, 98 concise arrow bodies, 7 conditional branches, 3 truthiness operands.<br>
Proved 351 annotation-only adaptations with byte-identical JavaScript; 42 sites remain unclassified.<br>
Reserved all 98 concise arrow bodies and all 16 union-with-void sites for the language lead.<br>
Node corpus: 106366 passed, one independently reproduced API baseline failure; 369/393 sites observed; mutant adds exactly one site.

# Void reads, October 19, 2026

This is a measurement, with a partial classification. No compiler source repair is committed. The table is [void-reads-2026-10-19.json](void-reads-2026-10-19.json). Observations (locations, types, forms, resolved declarations, runtime counters) and judgments are separate fields. Lines and UTF-16 columns are one-based. Root hashes identify the exact adapted inputs.

| Form | Count |
|---|---:|
| concise arrow body | 98 |
| condition or truthiness | 3 |
| conditional-expression branch | 7 |
| return statement | 383 |
| All other forms in the brief | 0 |
| **Total** | **491** |

| Judgment | Count |
|---|---:|
| adaptation | 351 |
| real read, with impossibility proved | 0 |
| concise arrow body, unjudged | 98 |
| unclassified | 42 |

## Construction and boundaries

Stock TypeScript **6.0.3**, installed from stage3/api's lockfile, runs the checker. Program construction is reused from checker-census.cjs at **21f6268d6dad5d39c985558b8469473d4f68a303**, using its loader library profile and compiler options. This includes its loader-compatible RegExp/string declaration edits. The prelude is loaded with exactly the MapIterator and SetIterator next declarations removed, preserving line positions. Its original and filtered hashes are in the table. These are measurement-only overlays outside the checkout, not edits under internal/.

The census visits 480940 runtime expression nodes in the 79 roots. It uses TypeScript's context-aware expression predicate, skips erased type syntax, and does not mistake declaration names for reads. It discards expression-statement results, void operands, comma left operands, for initializer/incrementor expressions, and comma results or parentheses carrying a discarded value. Nested expressions remain separate sites. There is no deduplication by source line or statement. Unions containing void are excluded from the 491.

The measurement program has **234 pre-existing checker diagnostics** under this construction. This is not a claim that the adapted tree is checker-clean. The annotation overlay also has 234 diagnostics and introduces zero diagnostics, comparing file, code, message and multiplicity before exploratory type queries.

## Adaptations

For 350 of the proved sites, the resolved implementation has no value-returning statement, excluding nested functions. Replace its explicit void return annotation with undefined, or add an undefined return annotation when void was inferred. For example, binder.ts's bind changes from:

```typescript
function bind(node: Node | undefined): void {
```

to:

```typescript
function bind(node: Node | undefined): undefined {
```

Its body is untouched. Every site records the exact signature replacement, offsets in the measured file, declaration line, and the checker result **undefined** (or **boolean** for the marker handler below). The combined overlay changes 310 annotations across eight files; edits are listed once under adaptationProof.files. This is a batch proof: apply the complete listed overlay, rather than assuming that any arbitrary subset was separately checked.

Full-file ts.transpileModule output, using the measurement compiler options, is compared byte for byte before and after. Each pair of SHA-256 hashes is recorded. verify.cjs independently reconstructs all eight outputs and rejects eight mutants, each appending a runtime 0; statement to one rewritten file. No adaptation is credited solely because it looks erasable. The overlay is never written to tsc's source.

One additional adaptation refines outofbandVarianceMarkerHandler from a void-returning callback to a boolean-returning callback. Its assignments either return true from assignments to unreliable/unmeasurable or forward to the saved handler. Runtime coverage observes true at checker.ts:23401, so treating void as necessarily undefined would be wrong here. The table records the implementations and the checked annotation rewrite. This refinement is included in the same byte-identical overlay proof.

## Runtime observations and limits

The original measured source, without annotation adaptations, was copied to a separate scratch tree. Every non-arrow site was wrapped in a value-preserving observer. The standard stage3 oracle built the instrumented compiler and ran its default upstream test corpus on Node with four workers. The observer records completed evaluations: an expression that throws or exits before returning produces no value hit. Thus these are completed-read counts, not attempted-call counts.

Oracle status: **fail**. Reported test totals: **{"passing":106366,"failing":1,"pending":0}**. Baseline differences: **1**. Across 24 process snapshots, **30080592 completed reads** cover **369 sites**; **24 instrumented sites have zero hits**. Each site's runtime field contains its count and every distinct observed value. Undefined is represented explicitly. Object, function and symbol identities, if present, are process-local; numbers retain NaN, infinities and negative zero.

Numeric values 0, 1, 2 and 4 were also observed at sys.exit call sites. Test host overrides use an expression-bodied assignment that returns the assigned exit code; production and throwing host implementations are recorded separately as implementation evidence. The observed numeric results must not be replaced by undefined.

A finite corpus cannot prove that an arbitrary host callback always returns undefined. Zero hits do not prove unreachability. This unit does not claim that the remaining 42 reads have no JavaScript-identical adaptation: the conservative no-value-return analysis did not prove one, and no impossibility proof was established. They include callbacks and host hooks, interface/overload signatures without an implementation body, and forwarding functions with return expressions. The table includes resolved declarations, including implementation bodies where the checker resolves directly to one. An exhaustive implementation inventory for external callbacks was not established. No site is labeled real read by default.

The sole corpus failure is the public API declaration baseline. An uninstrumented adapted-tree control ran the Public APIs tests: 30 passed and one failed. Its baseline diff is byte-identical to the instrumented run (SHA-256 3da14ced9f2c7eea0ae8b4dabbe6067a91510b0ab0d950ddb83903a9eb50fa28). This establishes the failure independently of the counter instrumentation. No baseline was accepted or changed.

The counter control passes eight values with six distinct value kinds, preserves return identity, and catches removal of the hit-count increment. This is recorded in counter-proof.json. Raw process snapshots and oracle logs are retained in evidence/.

## Unclassified sites

| Location | Form | Completed hits | Values seen | Expression |
|---|---|---:|---|---|
| src/compiler/binder.ts:3038:24 | return statement | 2900 | [{"type":"undefined"}] | bindEnumDeclaration(node as EnumDeclaration) |
| src/compiler/binder.ts:3643:16 | return statement | 2900 | [{"type":"undefined"}] | isEnumConst(node)             ? bindBlockScopedDeclaration(node, SymbolFlags.ConstEnum, SymbolFlags.ConstEnumExcludes)             : bindBlockScopedDeclaration(node, SymbolFlags.RegularEnum, SymbolFlags.RegularEnumExcludes) |
| src/compiler/checker.ts:8231:60 | return statement | 2666 | [{"type":"undefined"}] | bindElement(e) |
| src/compiler/checker.ts:26996:32 | return statement | 29470 | [{"type":"undefined"}] | inferFromTypes(apparentSource, target) |
| src/compiler/checker.ts:49109:24 | return statement | 617649 | [{"type":"undefined"}] | checkPropertySignature(node as PropertySignature) |
| src/compiler/checker.ts:49248:24 | return statement | 10735 | [{"type":"undefined"}] | checkBindingElement(node as BindingElement) |
| src/compiler/checker.ts:51425:32 | condition or truthiness | 1674693 | [{"type":"undefined"}] | markLinkedReferences(node, ReferenceHint.Unspecified) |
| src/compiler/debug.ts:865:16 | return statement | 0 | [] | console.log(formatControlFlowGraph(flowNode)) |
| src/compiler/emitter.ts:1747:28 | return statement | 4046 | [{"type":"undefined"}] | emitModuleDeclaration(node as ModuleDeclaration) |
| src/compiler/emitter.ts:1976:28 | return statement | 175639 | [{"type":"undefined"}] | emitBinaryExpression(node as BinaryExpression) |
| src/compiler/executeCommandLine.ts:579:16 | return statement | 1 | [{"type":"number","value":"1"}] | sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped) |
| src/compiler/executeCommandLine.ts:584:16 | return statement | 0 | [] | sys.exit(ExitStatus.Success) |
| src/compiler/executeCommandLine.ts:589:16 | return statement | 0 | [] | sys.exit(ExitStatus.Success) |
| src/compiler/executeCommandLine.ts:594:16 | return statement | 2 | [{"type":"number","value":"0"}] | sys.exit(ExitStatus.Success) |
| src/compiler/executeCommandLine.ts:599:16 | return statement | 1 | [{"type":"number","value":"1"}] | sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped) |
| src/compiler/executeCommandLine.ts:605:20 | return statement | 4 | [{"type":"number","value":"1"}] | sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped) |
| src/compiler/executeCommandLine.ts:613:24 | return statement | 2 | [{"type":"number","value":"1"}] | sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped) |
| src/compiler/executeCommandLine.ts:620:24 | return statement | 0 | [] | sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped) |
| src/compiler/executeCommandLine.ts:632:24 | return statement | 1 | [{"type":"number","value":"1"}] | sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped) |
| src/compiler/executeCommandLine.ts:643:20 | return statement | 7 | [{"type":"number","value":"1"}] | sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped) |
| src/compiler/executeCommandLine.ts:663:24 | return statement | 0 | [] | sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped) |
| src/compiler/executeCommandLine.ts:667:20 | return statement | 1 | [{"type":"number","value":"0"}] | sys.exit(ExitStatus.Success) |
| src/compiler/executeCommandLine.ts:707:20 | return statement | 0 | [] | sys.exit(ExitStatus.Success) |
| src/compiler/executeCommandLine.ts:833:16 | return statement | 0 | [] | sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped) |
| src/compiler/executeCommandLine.ts:839:16 | return statement | 1 | [{"type":"number","value":"0"}] | sys.exit(ExitStatus.Success) |
| src/compiler/executeCommandLine.ts:845:16 | return statement | 0 | [] | sys.exit(ExitStatus.Success) |
| src/compiler/executeCommandLine.ts:850:16 | return statement | 0 | [] | sys.exit(ExitStatus.DiagnosticsPresent_OutputsSkipped) |
| src/compiler/executeCommandLine.ts:899:12 | return statement | 2053 | [{"type":"number","value":"1"},{"type":"number","value":"2"},{"type":"number","value":"0"},{"type":"number","value":"4"}] | sys.exit(exitStatus) |
| src/compiler/executeCommandLine.ts:938:12 | return statement | 478 | [{"type":"number","value":"0"},{"type":"number","value":"2"},{"type":"number","value":"1"}] | sys.exit(exitStatus) |
| src/compiler/executeCommandLine.ts:965:12 | return statement | 1239 | [{"type":"number","value":"0"},{"type":"number","value":"2"},{"type":"number","value":"1"}] | sys.exit(exitStatus) |
| src/compiler/moduleNameResolver.ts:1139:20 | return statement | 31885 | [{"type":"undefined"}] | underlying.forEach((elem, key) => {                 const [specifier, mode] = memoizedReverseKeys.get(key)!;                 return cb(elem, specifier, mode);             }) |
| src/compiler/moduleNameResolver.ts:1141:24 | return statement | 21855 | [{"type":"undefined"}] | cb(elem, specifier, mode) |
| src/compiler/parser.ts:1361:20 | return statement | 8005 | [{"type":"undefined"}] | (overrideSetExternalModuleIndicator \|\| setExternalModuleIndicator)(file) |
| src/compiler/programDiagnostics.ts:162:32 | return statement | 647 | [{"type":"undefined"}] | computedDiagnostics!.add(                             createDiagnosticExplainingFile(                                 program,                                 diagnostic.file && program.getSourceFileByPath(diagnostic.file),                                 diagnostic.fileProcessingReason,                                 diagnostic.diagnostic,                                 diagnostic.args \|\| emptyArray,                             ),                         ) |
| src/compiler/programDiagnostics.ts:172:32 | return statement | 9 | [{"type":"undefined"}] | computedDiagnostics!.add(filePreprocessingLibreferenceDiagnostic(program, diagnostic)) |
| src/compiler/programDiagnostics.ts:174:32 | return statement | 344 | [{"type":"undefined"}] | diagnostic.diagnostics.forEach(d => computedDiagnostics!.add(d)) |
| src/compiler/sys.ts:569:12 | return statement | 87 | [{"type":"undefined"}] | curSysLog(s) |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts:68:20 | return statement | 0 | [] | previousOnEmitNode(hint, node, emitCallback) |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts:71:20 | return statement | 52511 | [{"type":"undefined"}] | esmOnEmitNode(hint, node, emitCallback) |
| src/compiler/transformers/module/impliedNodeFormatDependent.ts:73:16 | return statement | 24677 | [{"type":"undefined"}] | cjsOnEmitNode(hint, node, emitCallback) |
| src/compiler/watch.ts:994:28 | condition or truthiness | 176 | [{"type":"undefined"}] | host.trace(s) |
| src/compiler/watchUtilities.ts:221:16 | return statement | 0 | [] | host.writeFile!(fileName, data, writeByteOrderMark) |

## Union-with-void table

There are **16** sites, recorded in full under unionWithVoid and not judged.

| Location | Form | Type | Expression |
|---|---|---|---|
| src/compiler/checker.ts:51425:24 | return statement | void \| undefined | node && markLinkedReferences(node, ReferenceHint.Unspecified) |
| src/compiler/checker.ts:54428:16 | return statement | void \| undefined | this.inner?.pushErrorFallbackNode?.(node) |
| src/compiler/checker.ts:54432:16 | return statement | void \| undefined | this.inner?.popErrorFallbackNode?.() |
| src/compiler/executeCommandLine.ts:765:17 | concise arrow body | void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | performBuild(                     system,                     cb,                     buildOptions,                     watchOptions,                     projects,                     errors,                 ) |
| src/compiler/executeCommandLine.ts:775:20 | return statement | void \| SolutionBuilder<EmitAndSemanticDiagnosticsBuilderProgram> | performBuild(                 system,                 cb,                 buildOptions,                 watchOptions,                 projects,                 errors,             ) |
| src/compiler/executeCommandLine.ts:789:13 | concise arrow body | void \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | executeCommandLineWorker(                 system,                 cb,                 commandLine,             ) |
| src/compiler/executeCommandLine.ts:796:16 | return statement | void \| WatchOfConfigFile<EmitAndSemanticDiagnosticsBuilderProgram> | executeCommandLineWorker(system, cb, commandLine) |
| src/compiler/factory/utilities.ts:1298:26 | variable initializer | void \| BinaryExpression | machine.onLeft(nodeStack[stackIndex]!.left, userStateStack[stackIndex]!, nodeStack[stackIndex]!) |
| src/compiler/factory/utilities.ts:1299:13 | condition or truthiness | void \| BinaryExpression | nextNode |
| src/compiler/factory/utilities.ts:1330:26 | variable initializer | void \| BinaryExpression | machine.onRight(nodeStack[stackIndex]!.right, userStateStack[stackIndex]!, nodeStack[stackIndex]!) |
| src/compiler/factory/utilities.ts:1331:13 | condition or truthiness | void \| BinaryExpression | nextNode |
| src/compiler/sys.ts:517:66 | concise arrow body | void \| undefined | cache.get(path)?.callbacks.slice().forEach(cb => cb(param1, param2, param3)) |
| src/compiler/tsbuildPublic.ts:2137:23 | concise arrow body | void \| undefined | state.allWatchedExtendedConfigFiles.get(extendedConfigFilePath)?.projects.forEach(projectConfigFilePath => invalidateProjectAndScheduleBuilds(state, projectConfigFilePath, ProgramUpdateLevel.Full)) |
| src/compiler/utilities.ts:10747:16 | return statement | void \| "skip" | bindParentToChildIgnoringJSDoc(child, parent) \|\| bindJSDoc(child) |
| src/compiler/utilities.ts:10747:16 | condition or truthiness | void \| "skip" | bindParentToChildIgnoringJSDoc(child, parent) |
| src/compiler/watch.ts:994:14 | concise arrow body | void \| undefined | host.trace && host.trace(s) |

## Commands and environment

All long-running commands had a timeout and were checked while active. Test output went directly to log files. The final commands were:

```bash
git fetch origin
git switch -c typescript/stage3-void-reads origin/main
export GOPROXY='https://proxy.golang.org|direct'
timeout 840 bash cloud/setup.sh > /tmp/void-reads-setup.log 2>&1
source /workspace/adamic-tools/env.sh
nproc
timeout 120 git fetch origin typescript/stage3-onediag-evidence
timeout 180 npm ci --prefix stage3/api > /tmp/void-reads-api-install.log 2>&1
timeout 180 git clone --depth 1 --branch v6.0.3 https://github.com/microsoft/TypeScript.git /tmp/void-reads-upstream > /tmp/void-reads-clone.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/tmp/void-reads-upstream timeout 600 bash stage3/apply.sh /tmp/void-reads-tree > /tmp/void-reads-apply-retry.log 2>&1
timeout 300 node stage3/census/void-reads/census.cjs /workspace/adamic /tmp/void-reads-tree /tmp/void-reads-census stage3/census/void-reads/void-reads-2026-10-19.json --adaptations > /tmp/void-reads-census.log 2>&1
timeout 240 node stage3/census/void-reads/mutant.cjs /workspace/adamic /tmp/void-reads-tree /tmp/void-reads-mutant-final stage3/census/void-reads/void-reads-2026-10-19.json stage3/census/void-reads/mutant-proof.json > /tmp/void-reads-mutant.log 2>&1
timeout 120 cp -a --reflink=auto /tmp/void-reads-tree /tmp/void-reads-runtime
timeout 60 node stage3/census/void-reads/instrument.cjs /tmp/void-reads-runtime stage3/census/void-reads/void-reads-2026-10-19.json > /tmp/void-reads-instrument.log 2>&1
VOID_READS_COVERAGE=/tmp/void-reads-coverage NODE_OPTIONS='--require=/workspace/adamic/stage3/census/void-reads/observe.cjs' timeout 720 bash stage3/oracle/run.sh /tmp/void-reads-runtime /tmp/void-reads-oracle --workers 4 --limit-seconds 540 > /tmp/void-reads-oracle.log 2>&1
timeout 30 node stage3/census/void-reads/counter-control.cjs /tmp/void-reads-counter-control stage3/census/void-reads/counter-proof.json > /tmp/void-reads-counter-control.log 2>&1
timeout 90 node stage3/census/void-reads/verify.cjs /workspace/adamic /tmp/void-reads-tree stage3/census/void-reads/void-reads-2026-10-19.json > /tmp/void-reads-verify.log 2>&1
PYTHONDONTWRITEBYTECODE=1 timeout 240 bash stage3/oracle/run.sh /tmp/void-reads-tree /tmp/void-reads-api-control --tests 'Public APIs' --limit-seconds 60 > /tmp/void-reads-api-control.log 2>&1
timeout 30 node stage3/census/void-reads/merge-coverage.cjs stage3/census/void-reads/void-reads-2026-10-19.json /tmp/void-reads-coverage /tmp/void-reads-oracle/report.json /tmp/void-reads-api-control/report.json > /tmp/void-reads-merge.log 2>&1
node stage3/census/void-reads/report.cjs
```

The first apply attempt failed because the product cache was cold and ADAMIC_TYPESCRIPT_SOURCE was unset. The pinned checkout above fixed that. An early census against the unfinished tree rejected 78 roots; the completed product has 79. An initial adaptation diagnostic comparison was rejected because its baseline diagnostics were collected after exploratory queries; the final run collects both diagnostic sets at the same phase and introduces none. These rejected attempts were not counted as proofs.

Setup timing lines: Node 0.074s; Go 0.097s; clang 0.621s; markdown dependencies 1.328s; submodules 231.481s; shared cache 241.970s; Go build 632.756s; cache warm 633.246s; done 633.604s. nproc=5, CPU quota=4, Node v24.19.0, Go 1.27.1, clang 20.1.8. Setup's printed environment path was /workspace/adamic-tools/env.sh and was sourced. The full Go test gate was not run for this measurement-only unit; setup built all Go packages.

The scratch mutant appends a void-returning function and a binding of its call result to the first root. mutant-proof.json records **491 -> 492**, with the sole added read at **src/compiler/_namespaces/ts.moduleSpecifiers.ts:6:25**, no lost sites, and the union count unchanged. The scratch file is restored in a finally block and its hash is rechecked by verify.cjs.
