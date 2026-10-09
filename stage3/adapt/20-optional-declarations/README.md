# 20: truthful optional declarations

This adaptation widens an owned optional declaration from `p?: T` to
`p?: (T) | undefined` when a TS2412, TS2375 or TS2379 finding supplies evidence
that the implementation deliberately stores undefined. The question mark,
assignments, object keys and every runtime expression remain. A present undefined
key has different `in`, `hasOwn`, enumeration and spread behavior from an absent
key. Removing a write or conditionally constructing a key is not this adaptation.

The resolver uses stock npm TypeScript 6.0.3's compiler API, not source regexes.
It parses current text, resolves property symbols to their declarations, checks
that every non-undefined source constituent fits the declared value type, and
edits each selected owner once per program. It rechecks the whole dependency
closure after each pass, stopping only when no eligible owner remains. This also
tolerates preceding type-import adaptations: no source line is a rewrite address.
Type insertions use parsed UTF-16 offsets and preserve existing source trivia.
Nested type additions compose. Optional methods, live methods, accessors and
overload sets are declined: unioning an optional method's callable requires a
property conversion that changes variance and public declaration shape.
Only function, constructor and conditional type annotations need parentheses.
An inherited optional view of the same proven present-undefined slot receives
the same truthful union; this resolves the slot at all its declaration owners.

Both absence and present undefined are now admitted. Consumers must therefore
prove more than key presence before using the value. Every new or changed
stock-checker diagnostic is reported with location in `newly_exposed_consumers`;
runtime consumers stay unchanged. Type-only inherited owner coherence and the
build-error dispositions below follow the user's revised scope.
The evidence-to-owner ledger is `resolved_findings`. `initially_declined` retains
seed exclusions even if another legitimate write to a shared owner removes the
finding incidentally. `declined` lists remaining optional-code findings and reasons.

The fixed checker options are strict, exactOptionalPropertyTypes,
noUncheckedIndexedAccess, noImplicitReturns, noFallthroughCasesInSwitch,
erasableSyntaxOnly and verbatimModuleSyntax; ES2024, ESNext modules, Bundler
resolution, force module detection, ES2024 library and no implicit type packages.
The stock resolver is tooling for rejected source. The census's Adamic checker
remains the measurement authority, including its sounder regex declarations.

## Running

Input is the externally cached TypeScript tree pinned by stage 3: repository
`https://github.com/microsoft/TypeScript.git`, tag `v6.0.3`, commit
`050880ce59e30b356b686bd3144efe24f875ebc8`. No upstream source is committed here.
Generate upstream's diagnostics file before adapting a prepared compiler tree.

Install the stock tool in scratch, then run the requested command directly:

```sh
npm install --prefix /tmp/adamic-optional/npm typescript@6.0.3 --no-audit --no-fund
export NODE_PATH=/tmp/adamic-optional/npm/node_modules
node stage3/adapt/20-optional-declarations/adapt.cjs /tmp/adamic-optional/typescript > /tmp/adamic-optional/adapt.json 2> /tmp/adamic-optional/adapt.stderr
node stage3/adapt/20-optional-declarations/adapt.cjs /tmp/adamic-optional/typescript > /tmp/adamic-optional/idempotence.json 2> /tmp/adamic-optional/idempotence.stderr
```

`TSC_ADAPT_TYPESCRIPT` may instead name the stock `lib/typescript.js` explicitly.
Otherwise resolution uses the tooling's normal Node module search, then the
checkout's dependencies. A version other than 6.0.3 is rejected, including
upstream's locked build compiler (5.9.3). The JSON report goes to stdout and
progress to stderr; only eligible owner declarations are written in the checkout.

## Declined contracts

The original survey's C and R cases remain separate obligations:

| Case | Original pinned site | Reason |
| --- | --- | --- |
| E022, E-index | `src/compiler/checker.ts:33923` | `childrenTypes[0]` is an unproven indexed read, even after `length === 1`; it cannot seed a declaration edit. |
| E023, E-index | `src/compiler/checker.ts:43517` | `getTypeArguments(type as GenericType)[0]` requires an arity/presence invariant. |
| E070, E-generic | `src/compiler/factory/nodeFactory.ts:5202` | `T["comment"]` may be a narrower arbitrary specialization. |
| E027, E-other | `src/compiler/emitter.ts:1014` | The expanded chain concerns `Bundle \| SourceFile` and missing required fields, not optionality. |
| E087, additional generic review | `src/compiler/resolutionCache.ts:1386` | The reset receiver is a type parameter; its specialization may narrow the field. |

A bare indexed read is never a seed. An existing explicit undefined branch can
be evidence of intentional absence. Generic/unknown/indexed-access value
contracts, required or inferred slots, external owners and live methods remain
unchanged. A structural chain must actually end in an undefined incompatibility;
a diagnostic code alone is insufficient. Shared-owner changes can discharge an
indexed diagnostic incidentally, but do not prove the indexed read invariant.

## Current revision and proof

This revision supersedes `e3535e7`'s measured 405-owner implementation. It removes
all **39** optional-method conversions, retains **366** property-owner edits,
and adds **seven** inherited optional declaration owners. The current adapter
edits **373 distinct declarations** for the initial **725** optional-code findings.
The shared apply diff is **24 files, 371 lines replaced**. The method conversions
were JavaScript-neutral but altered callable variance and public declaration
syntax; they do not satisfy the now-required API snapshot contract.

Merged latest adaptation 10, `a3ef0dc`, in merge `299738f`. Shared apply uses its
unchanged numeric ordering. `stage3/apply.sh` and `stage3/oracle/run.sh` are
unmodified. Upstream source and baseline changes stay in the external pinned
checkout, never in Adamic. Heavy apply, default oracle and census runs are
sequential.

### The nine build errors

The seven `compilerHost` errors were cascading from a rejected initializer,
not an `in` guard. This corrects the first revision's explanation. No statement,
assignment, object key, callback implementation or guard was changed.

| Original location | Error | Disposition |
| --- | --- | --- |
| `src/compiler/tsbuildPublic.ts:787:5` | TS2322, cached `WriteFileCallback` versus optional-boolean callable | Leave `SolutionBuilderHostBase.writeFile?` at `tsbuildPublic.ts:222` a method. Converting it changed method bivariance; widening the callback implementation contract is unproven. The property cache declaration still truthfully admits undefined. |
| `src/services/services.ts:1769:13` | TS2322, mutable versus readonly reference-name arrays | Leave `CompilerHost.resolveTypeReferenceDirectives?` at `types.ts:8161` a method. Its original variance accepts the forwarded host callback; proving a readonly callable implementation would require broader consumer work. |
| `src/services/services.ts:1779:39` | TS18048, `compilerHost` possibly undefined | The preceding initializer now succeeds, restoring its existing flow narrowing. No additional edit. |
| `src/services/services.ts:1782:13` | TS2345, compiler host argument | Same initializer correction; no additional edit. |
| `src/services/services.ts:1786:9` | TS18048 | Same initializer correction; no additional edit. |
| `src/services/services.ts:1788:32` | TS2345 | Same initializer correction; no additional edit. |
| `src/services/services.ts:1796:23` | TS18048 | Same initializer correction; no additional edit. |
| `src/services/services.ts:1798:20` | TS18048 | Same initializer correction; no additional edit. |
| `src/services/services.ts:1799:34` | TS18048 | Same initializer correction; no additional edit. |

Optional methods have no syntax for unioning the callable itself with undefined.
A callable-property conversion would also violate the sanctioned snapshot diff.
The adapter declines every optional method it encounters, with owner location
and reason in `declined_owners`, instead of inventing a new calling contract.

### Inherited declaration owners

A proven undefined store through a derived/interface-intersection view also
belongs to its inherited optional views. The resolver walks interface bases and
peer bases, resolves the same property's symbols, and adds the union only to
compatible, owned, typed optional declarations. Required, generic, unknown,
external and method contracts are still excluded by the original selection rules.
The original write finding is retained as evidence; inherited rows have
`resolution` and `from` in the ledger. A set deduplicates every owner.

The seven added owners, beyond the retained original property edits, are:

| Owner | Pristine location |
| --- | --- |
| `NonRelativeNameResolutionCache.isReadonly` | `moduleNameResolver.ts:891` |
| `NamedDeclaration.name` | `types.ts:1763` |
| `DeclarationStatement.name` | `types.ts:1793` |
| `ObjectLiteralElement.name` | `types.ts:1982` |
| `TypeReference.mapper` | `types.ts:6650` |
| `ResolvedTypeReferenceDirectiveWithFailedLookupLocations.failedLookupLocations` | `types.ts:8118` |
| `ResolvedTypeReferenceDirectiveWithFailedLookupLocations.affectingLocations` | `types.ts:8119` |

The new owners make the same slot truthful through all existing optional views;
there are no JavaScript edits. They eliminate the declaration inheritance errors
that previously changed APILibCheck. The focused upstream test now passes
**6 checks** against its original reference baseline.

### Mechanical public snapshot proof

`check-baselines.cjs` uses stock TypeScript 6.0.3 to parse both snapshots and
pinned source owner declarations. It verifies the pristine commit, unmodified
tracked source/reference inputs, and that upstream's pristine declaration emitter
reproduces the original API reference exactly. Namespace-qualified owner paths
keep `ts.Diagnostic` distinct from `ts.server.protocol.Diagnostic`.

From the original snapshot it constructs exactly the permitted additions at the
373-owner ledger's matching public optional property types, keeping each `?`.
Function/constructor/conditional types alone receive necessary parentheses. The
constructed full text must equal the newly emitted snapshot byte-for-byte; this
also proves every changed line belongs to an edited declaration, with no member,
signature, whitespace, name, unrelated line or unlisted owner changes. Upstream
expands some enum annotations, so the existing emitted annotation is retained
exactly, rather than reconstructed from source tokens.

Observed result: **189 edited snapshot declarations, exactly 189 changed lines**.
All **60,930 other reference baselines** are byte-identical to pristine. Only
`tests/baselines/reference/api/typescript.d.ts` is accepted in the external tree,
and only after proof passes. APILibCheck's reference is unchanged. The tool's
`--accept-api` switch authorizes only this proved snapshot; its default is read-only.

Four real-input mutants are rejected by this named check: remove an edited
`?`; append undefined to unedited protocol `Diagnostic.source`; omit
`CompilerOptions.traceResolution` from the owner ledger; append a newline to
an unrelated reference JS baseline. Each exits 1 and every input is restored.
The unrelated reference mutant passes the API reconstruction first and fails
only the all-other-reference comparison.

### Default oracle

The unchanged default command used all suites, no test regex, four workers and
the default 2,400-second test limit. Install exited 0 (3.600s), checked build
exited 0 (2.808s), tests exited 0 (477.923s): **106,367 passing, 0 failing,
0 pending**, with **zero baseline diffs**. Total wall time **484.400s**.
The preparation also ran upstream's checked build before accepting the API
snapshot. No `--no-typecheck` option or weakened compiler option was used.

The default oracle compares against the mechanically proved API reference.
Every other reference, including APILibCheck, is original. The after-oracle
snapshot/reference proof reruns independently, preventing acceptance of any
unrelated reference change.

### Complete whole-program census

The unchanged full census driver ran every input entry, followed by all compiler
roots, sequentially. Each final pass attempted **81 entries** and checked **78
TypeScript roots** together. The whole program remains checker-rejected and
never reaches lowering; the three JSON entries give extension-loading errors.
Prepared latest-10 source has zero TS1484 on both sides.

| Reason | After latest 10 | After 10 + 20 |
| --- | ---: | ---: |
| TS1294 | 180 | 180 |
| TS1484 | 0 | 0 |
| TS2304 | 6 | 6 |
| TS2320 | 0 | 0 |
| TS2322 | 120 | 118 |
| TS2339 | 28 | 28 |
| TS2345 | 733 | 736 |
| TS2366 | 1 | 0 |
| TS2375 | 77 | 19 |
| TS2379 | 31 | 14 |
| TS2412 | 617 | 27 |
| TS2420 | 1 | 1 |
| TS2430 | 0 | 0 |
| TS2488 | 11 | 11 |
| TS2532 | 225 | 225 |
| TS2538 | 7 | 7 |
| TS2556 | 2 | 2 |
| TS2591 | 54 | 54 |
| TS2684 | 2 | 2 |
| TS2722 | 2 | 2 |
| TS2740 | 1 | 1 |
| TS2769 | 10 | 10 |
| TS7006 | 1 | 1 |
| TS7029 | 83 | 83 |
| TS7030 | 252 | 252 |
| TS7031 | 1 | 1 |
| TS18046 | 8 | 8 |
| TS18048 | 380 | 380 |

Whole-program total **2,833 → 2,168**. TS2412/2375/2379 total **725 → 60**,
net **665** removed. TS2345 rises **733 → 736** as widened fields reveal stricter
consumer argument contracts, listed below. TS2322 falls **120 → 118** and
TS2366 falls **1 → 0**; all other counts are unchanged. Inherited declaration
coherence leaves TS2320 and TS2430 at zero, unlike the superseded revision.

The generated diagnostics source contributes one TS2375 beyond the historical
77-root survey, explaining **77** rather than **76**. During fresh apply the
stock checker reports **2,729 → 2,064**: one unregenerated TS1484 and one
not-yet-installed `source-map-support` TS2307 are present on both sides. After
upstream generation/dependency installation the stock counts are **2,727 →
2,062**. The meter uses its own sounder library; its complete table above is
the authority. No checker options were changed.

The first standalone before-generator invocation accidentally omitted its input
filename and exited 0 with usage. That exploratory count was 2,834. The final
before pass was rerun in full with the correct input; every diagnostic chain
except that stale generated TS1484 matches the discarded pass exactly.

### JavaScript identity and idempotence

Upstream's ordinary checked build passed on the latest-10 control and on the
adapted tree. All **10 emitted JavaScript files, 28,599,479 bytes**, including
the actual compiler bundled in `run.js`, are byte-identical. The comparison
checks full byte arrays and file sets, not only hashes. A copied artifact with
an appended statement is rejected by the same comparator. Declaration output
is the proved exception; maps and build metadata are not claimed identical.

| JavaScript artifact | SHA-256 on both sides |
| --- | --- |
| `_tsc.js` | `3e526e28c780ada5f46e8e7c06f075fa51910580c2dc185fbc66a55ebd60fc34` |
| `_tsserver.js` | `1992fc3518a4aa5443638a25794281995ec3ce84820692713296262997d5443b` |
| `_typingsInstaller.js` | `71515c1c1aa29c8616309f7dea4d3a32eed662df3a4dbcf4345e7522ece56ac1` |
| `run.js` | `0520cd6de4d4eae7cac4fcaebd6057baebbb059f7401334bbae53b2b5fee6fc1` |
| `tsc.js` | `2cffde0b8c6760dfb0b5b0382bbb7e00ba6a8b2d981b9205b256a700a481d983` |
| `tsserver.js` | `e3ccfeec65ec5c470b8ffc5611878a31182650c7e8a062c38a719f83b523edcb` |
| `tsserverlibrary.js` | `6a0b343fa3a3f53cf198c83a18035c104b6990b2a260272654f9a6b2f2380fdf` |
| `typescript.js` | `8fe41e154da327616c99568f551fe13d48764b73f08abf1fd773d5ce74870905` |
| `typingsInstaller.js` | `2d48a038523ea7c626fa2796a30fa5fdc07612ea5fab83544423009d7683a213` |
| `watchGuard.js` | `008f2ceb504ee9b8c5d47054d43033e1ec3f2fef03f18d8aa1b2dd70cba6fc88` |


Repeated adapter runs on the same final tree each report **zero new declarations**.
All **745 source files** retain their SHA-256 map. An actual source newline
mutant fails that same map comparison and is restored byte-for-byte. The fixture
passes **six property owners**, inherited/peer views, method deferral, callable
precedence, the new `in`-guard diagnostic location, indexed/generic/wrong/live
exclusions, identical emitted JavaScript, Node key fingerprints and second-run
source byte stability.

### Source diff and evidence

There are **373** owners, **737** evidence rows including **13** inherited-owner
rows, and **69** initially declined optional-code chains. Remaining findings and
new consumers follow below. Idempotence emits zero new owner edits. Line-based
diff sizes differ from declaration counts because some lines hold multiple owners.

### Consumers exposed by the stock checker

The prepared dependency/generation state reports every newly exposed consumer
without runtime fixes. The 13 new or changed chains in the apply ledger follow;
a changed chain may be an existing diagnostic with a different structural cause.

```text
src/compiler/checker.ts:1683:9: TS2322: Type '(signature: Signature, kind: SignatureDeclaration["kind"], enclosingDeclaration?: Node, flags?: NodeBuilderFlags, internalFlags?: InternalNodeBuilderFlags, tracker?: SymbolTracker, maximumLength?: number, verbosityLevel?: number, out?: WriterContextOut) => SignatureDeclaration | undefined' is not assignable to type '{ (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined): (SignatureDeclaration & { ...; }) | undefined; (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined, internalFlags?: InternalNodeBui...'.
  Type 'SignatureDeclaration | undefined' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
    Type 'FunctionExpression' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
      Type 'FunctionExpression' is not assignable to type 'FunctionExpression & { typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
        Type 'FunctionExpression' is not assignable to type '{ typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
          Types of property 'typeArguments' are incompatible.
            Type 'NodeArray<TypeNode> | undefined' is not assignable to type 'NodeArray<TypeNode>'.
              Type 'undefined' is not assignable to type 'NodeArray<TypeNode>'.
src/compiler/checker.ts:21434:98: TS2379: Argument of type '{ errors?: Diagnostic[] | undefined; }' is not assignable to parameter of type '{ errors?: Diagnostic[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'errors' are incompatible.
    Type 'Diagnostic[] | undefined' is not assignable to type 'Diagnostic[]'.
      Type 'undefined' is not assignable to type 'Diagnostic[]'.
src/compiler/checker.ts:46159:176: TS2345: Argument of type 'ErrorOutputContainer | undefined' is not assignable to parameter of type '{ errors?: Diagnostic[]; } | undefined'.
  Type 'ErrorOutputContainer' is not assignable to type '{ errors?: Diagnostic[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'errors' are incompatible.
      Type 'Diagnostic[] | undefined' is not assignable to type 'Diagnostic[]'.
        Type 'undefined' is not assignable to type 'Diagnostic[]'.
src/compiler/commandLineParser.ts:2677:9: TS2322: Type '{ path: string; originalPath: undefined; prepend?: boolean | undefined; circular?: boolean | undefined; }[] | undefined' is not assignable to type 'readonly ProjectReference[] | undefined'.
  Type '{ path: string; originalPath: undefined; prepend?: boolean | undefined; circular?: boolean | undefined; }[]' is not assignable to type 'readonly ProjectReference[]'.
    Type '{ path: string; originalPath: undefined; prepend?: boolean | undefined; circular?: boolean | undefined; }' is not assignable to type 'ProjectReference' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
      Types of property 'originalPath' are incompatible.
        Type 'undefined' is not assignable to type 'string'.
src/compiler/emitter.ts:1011:33: TS2345: Argument of type 'SourceFile | Bundle' is not assignable to parameter of type 'Bundle'.
  Property 'sourceFiles' is missing in type 'SourceFile' but required in type 'Bundle'.
src/compiler/emitter.ts:1014:31: TS2345: Argument of type 'SourceFile | Bundle' is not assignable to parameter of type 'SourceFile'.
  Type 'Bundle' is missing the following properties from type 'SourceFile': statements, endOfFileToken, fileName, path, and 27 more.
src/compiler/moduleNameResolver.ts:635:5: TS2375: Type '{ resolvedTypeReferenceDirective: ResolvedTypeReferenceDirective | undefined; failedLookupLocations: string[] | undefined; affectingLocations: string[] | undefined; resolutionDiagnostics: Diagnostic[] | undefined; }' is not assignable to type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'resolutionDiagnostics' are incompatible.
    Type 'Diagnostic[] | undefined' is not assignable to type 'Diagnostic[]'.
      Type 'undefined' is not assignable to type 'Diagnostic[]'.
src/compiler/tracing.ts:175:17: TS2339: Property 'phase' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:24: TS2339: Property 'name' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:30: TS2339: Property 'args' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:36: TS2339: Property 'time' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:42: TS2339: Property 'separateBeginAndEnd' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/watchPublic.ts:383:52: TS2379: Argument of type '{ configFileName: string; optionsToExtend: CompilerOptions | undefined; watchOptionsToExtend: WatchOptions; extraFileExtensions: readonly FileExtensionInfo[]; system: System; createProgram: CreateProgram<...> | undefined; reportDiagnostic: DiagnosticReporter | undefined; reportWatchStatus: WatchStatusReporter | unde...' is not assignable to parameter of type 'CreateWatchCompilerHostOfConfigFileInput<T>' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'createProgram' are incompatible.
    Type 'CreateProgram<T> | undefined' is not assignable to type 'CreateProgram<T>'.
      Type 'undefined' is not assignable to type 'CreateProgram<T>'.
```

### Consumers exposed by the census

The final meter reports **13 new or changed whole-program chains**.
Locations are in the final adapted source. A changed chain can be a persisting
error whose structural cause changed. No runtime consumer was edited.

```text
src/compiler/checker.ts:1683:9: error TS2322: Type '(signature: Signature, kind: SignatureDeclaration["kind"], enclosingDeclaration?: Node, flags?: NodeBuilderFlags, internalFlags?: InternalNodeBuilderFlags, tracker?: SymbolTracker, maximumLength?: number, verbosityLevel?: number, out?: WriterContextOut) => SignatureDeclaration | undefined' is not assignable to type '{ (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined): (SignatureDeclaration & ...) | undefined; (signature: Signature, kind: SyntaxKind, enclosingDeclaration: Node | undefined, flags: NodeBuilderFlags | undefined, internalFlags?: InternalNodeBuilderF...'.
  Type 'SignatureDeclaration | undefined' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
    Type 'ArrowFunction' is not assignable to type '(SignatureDeclaration & { typeArguments?: NodeArray<TypeNode>; }) | undefined'.
      Type 'ArrowFunction' is not assignable to type 'ArrowFunction & { typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
        Type 'ArrowFunction' is not assignable to type '{ typeArguments?: NodeArray<TypeNode>; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
          Types of property 'typeArguments' are incompatible.
            Type 'NodeArray<TypeNode> | undefined' is not assignable to type 'NodeArray<TypeNode>'.
              Type 'undefined' is not assignable to type 'NodeArray<TypeNode>'.
src/compiler/checker.ts:21434:98: error TS2379: Argument of type '{ errors?: Diagnostic[] | undefined; }' is not assignable to parameter of type '{ errors?: Diagnostic[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'errors' are incompatible.
    Type 'Diagnostic[] | undefined' is not assignable to type 'Diagnostic[]'.
      Type 'undefined' is not assignable to type 'Diagnostic[]'.
src/compiler/checker.ts:46159:176: error TS2345: Argument of type 'ErrorOutputContainer | undefined' is not assignable to parameter of type '{ errors?: Diagnostic[]; } | undefined'.
  Type 'ErrorOutputContainer' is not assignable to type '{ errors?: Diagnostic[]; }' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
    Types of property 'errors' are incompatible.
      Type 'Diagnostic[] | undefined' is not assignable to type 'Diagnostic[]'.
        Type 'undefined' is not assignable to type 'Diagnostic[]'.
src/compiler/commandLineParser.ts:2677:9: error TS2322: Type '{ prepend?: boolean | undefined; circular?: boolean | undefined; path: string; originalPath: undefined; }[] | undefined' is not assignable to type 'readonly ProjectReference[] | undefined'.
  Type '{ prepend?: boolean | undefined; circular?: boolean | undefined; path: string; originalPath: undefined; }[]' is not assignable to type 'readonly ProjectReference[]'.
    Type '{ prepend?: boolean | undefined; circular?: boolean | undefined; path: string; originalPath: undefined; }' is not assignable to type 'ProjectReference' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
      Types of property 'originalPath' are incompatible.
        Type 'undefined' is not assignable to type 'string'.
src/compiler/emitter.ts:1011:33: error TS2345: Argument of type 'Bundle | SourceFile' is not assignable to parameter of type 'Bundle'.
  Property 'sourceFiles' is missing in type 'SourceFile' but required in type 'Bundle'.
src/compiler/emitter.ts:1014:31: error TS2345: Argument of type 'Bundle | SourceFile' is not assignable to parameter of type 'SourceFile'.
  Type 'Bundle' is missing the following properties from type 'SourceFile': statements, endOfFileToken, fileName, path, and 27 more.
src/compiler/moduleNameResolver.ts:635:5: error TS2375: Type '{ resolvedTypeReferenceDirective: ResolvedTypeReferenceDirective | undefined; failedLookupLocations: string[] | undefined; affectingLocations: ... | undefined; resolutionDiagnostics: ... | undefined; }' is not assignable to type 'ResolvedTypeReferenceDirectiveWithFailedLookupLocations' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'resolutionDiagnostics' are incompatible.
    Type 'Diagnostic[] | undefined' is not assignable to type 'Diagnostic[]'.
      Type 'undefined' is not assignable to type 'Diagnostic[]'.
src/compiler/tracing.ts:175:17: error TS2339: Property 'phase' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:24: error TS2339: Property 'name' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:30: error TS2339: Property 'args' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:36: error TS2339: Property 'time' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/tracing.ts:175:42: error TS2339: Property 'separateBeginAndEnd' does not exist on type '{ phase: Phase; name: string; args?: Args | undefined; time: number; separateBeginAndEnd: boolean; } | undefined'.
src/compiler/watchPublic.ts:383:52: error TS2379: Argument of type '{ configFileName: string; optionsToExtend: CompilerOptions | undefined; watchOptionsToExtend: WatchOptions; extraFileExtensions: readonly FileExtensionInfo[]; system: System; createProgram: CreateProgram<...> | undefined; reportDiagnostic: DiagnosticReporter | undefined; reportWatchStatus: WatchStatusReporter | unde...' is not assignable to parameter of type 'CreateWatchCompilerHostOfConfigFileInput<T>' with 'exactOptionalPropertyTypes: true'. Consider adding 'undefined' to the types of the target's properties.
  Types of property 'createProgram' are incompatible.
    Type 'CreateProgram<T> | undefined' is not assignable to type 'CreateProgram<T>'.
      Type 'undefined' is not assignable to type 'CreateProgram<T>'.
```

### Declined optional method owners

All 64 encountered owners are left with their original method syntax and variance.
Their common reason is the inability to add a callable undefined union while
preserving public method shape. They are not proof of a non-undefined runtime
value; they remain obligations for a later adaptation with a broader API contract.

| Pristine owner location | Name |
| --- | --- |
| `src/compiler/types.ts:9991` | `readFile` |
| `src/compiler/types.ts:9993` | `getSymlinkCache` |
| `src/compiler/types.ts:9996` | `getGlobalTypingsCacheLocation` |
| `src/compiler/types.ts:8196` | `createHash` |
| `src/compiler/types.ts:7994` | `realpath` |
| `src/compiler/types.ts:9990` | `directoryExists` |
| `src/compiler/types.ts:9992` | `realpath` |
| `src/compiler/types.ts:8620` | `createHash` |
| `src/compiler/types.ts:10008` | `trace` |
| `src/compiler/types.ts:7989` | `directoryExists` |
| `src/compiler/types.ts:7996` | `getDirectories` |
| `src/compiler/types.ts:4549` | `trace` |
| `src/compiler/tsbuildPublic.ts:238` | `now` |
| `src/compiler/types.ts:8162` | `resolveModuleNameLiterals` |
| `src/compiler/types.ts:8170` | `resolveTypeReferenceDirectiveReferences` |
| `src/compiler/types.ts:8179` | `resolveLibrary` |
| `src/compiler/types.ts:8151` | `resolveModuleNames` |
| `src/compiler/types.ts:8161` | `resolveTypeReferenceDirectives` |
| `src/compiler/types.ts:8155` | `getModuleResolutionCache` |
| `src/compiler/watchPublic.ts:200` | `directoryExists` |
| `src/compiler/tsbuildPublic.ts:217` | `createDirectory` |
| `src/compiler/tsbuildPublic.ts:222` | `writeFile` |
| `src/compiler/types.ts:8135` | `getDefaultLibLocation` |
| `src/compiler/types.ts:8141` | `readDirectory` |
| `src/compiler/types.ts:7988` | `trace` |
| `src/compiler/watchPublic.ts:186` | `createHash` |
| `src/compiler/watchPublic.ts:207` | `realpath` |
| `src/compiler/watchPublic.ts:211` | `getEnvironmentVariable` |
| `src/compiler/watchPublic.ts:274` | `now` |
| `src/compiler/types.ts:8198` | `useSourceOfProjectReferenceRedirect` |
| `src/compiler/types.ts:8201` | `createDirectory` |
| `src/compiler/types.ts:9991` | `readFile` |
| `src/compiler/types.ts:9993` | `getSymlinkCache` |
| `src/compiler/types.ts:9996` | `getGlobalTypingsCacheLocation` |
| `src/compiler/types.ts:8196` | `createHash` |
| `src/compiler/types.ts:7994` | `realpath` |
| `src/compiler/types.ts:9990` | `directoryExists` |
| `src/compiler/types.ts:9992` | `realpath` |
| `src/compiler/types.ts:8620` | `createHash` |
| `src/compiler/types.ts:10008` | `trace` |
| `src/compiler/types.ts:7989` | `directoryExists` |
| `src/compiler/types.ts:7996` | `getDirectories` |
| `src/compiler/types.ts:4549` | `trace` |
| `src/compiler/types.ts:8162` | `resolveModuleNameLiterals` |
| `src/compiler/types.ts:8170` | `resolveTypeReferenceDirectiveReferences` |
| `src/compiler/types.ts:8179` | `resolveLibrary` |
| `src/compiler/types.ts:8151` | `resolveModuleNames` |
| `src/compiler/types.ts:8161` | `resolveTypeReferenceDirectives` |
| `src/compiler/types.ts:8155` | `getModuleResolutionCache` |
| `src/compiler/watchPublic.ts:200` | `directoryExists` |
| `src/compiler/types.ts:8135` | `getDefaultLibLocation` |
| `src/compiler/types.ts:8141` | `readDirectory` |
| `src/compiler/types.ts:7988` | `trace` |
| `src/compiler/watchPublic.ts:186` | `createHash` |
| `src/compiler/watchPublic.ts:207` | `realpath` |
| `src/compiler/watchPublic.ts:211` | `getEnvironmentVariable` |
| `src/compiler/watchPublic.ts:274` | `now` |
| `src/compiler/types.ts:8198` | `useSourceOfProjectReferenceRedirect` |
| `src/compiler/types.ts:8201` | `createDirectory` |
| `src/compiler/watchPublic.ts:200` | `directoryExists` |
| `src/compiler/watchPublic.ts:186` | `createHash` |
| `src/compiler/watchPublic.ts:207` | `realpath` |
| `src/compiler/watchPublic.ts:211` | `getEnvironmentVariable` |
| `src/compiler/watchPublic.ts:274` | `now` |

### Remaining optional-code chains

There are 60 remaining chains in the final stock recheck. Some share an owner;
all source writes and literal keys remain. Survey C/R evidence above is declined
for indexed, generic or discriminator reasons. Remaining chains are listed with
the resolver's reason, including missing compatible declaration-owner evidence.

| Location | Code | Reason |
| --- | --- | --- |
| `src/compiler/checker.ts:5964:13` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:6035:9` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:7837:133` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:8082:62` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:10941:33` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:17890:21` | TS2412 | unproven indexed read, not optionality evidence |
| `src/compiler/checker.ts:21434:98` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:43517:20` | TS2412 | unproven indexed read, not optionality evidence |
| `src/compiler/checker.ts:43740:86` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/checker.ts:54261:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/commandLineParser.ts:2663:9` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/emitter.ts:862:55` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/emitter.ts:938:70` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/factory/nodeFactory.ts:1216:9` | TS2412 | arbitrary generic specialization |
| `src/compiler/factory/nodeFactory.ts:1315:41` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/factory/nodeFactory.ts:1403:41` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/factory/nodeFactory.ts:5496:9` | TS2412 | arbitrary generic specialization |
| `src/compiler/moduleNameResolver.ts:132:13` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/moduleNameResolver.ts:627:9` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/moduleNameResolver.ts:635:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/program.ts:472:11` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/program.ts:1876:11` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/program.ts:2599:9` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/program.ts:4956:9` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/program.ts:4957:9` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/program.ts:5107:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:990:43` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:1019:43` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:1332:58` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:1332:58` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/resolutionCache.ts:1386:9` | TS2412 | generic or unresolved receiver contract |
| `src/compiler/transformers/classFields.ts:2745:16` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/transformers/utilities.ts:689:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/transformers/utilities.ts:749:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/transformers/utilities.ts:773:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/tsbuildPublic.ts:302:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/tsbuildPublic.ts:438:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/tsbuildPublic.ts:439:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/tsbuildPublic.ts:440:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/tsbuildPublic.ts:441:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/tsbuildPublic.ts:442:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/tsbuildPublic.ts:443:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/tsbuildPublic.ts:785:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/tsbuildPublic.ts:786:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/tsbuildPublic.ts:787:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watch.ts:756:11` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watch.ts:845:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:125:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:371:65` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:383:52` | TS2379 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:506:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:529:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:530:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:534:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:535:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:542:5` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:568:17` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:661:9` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchPublic.ts:662:9` | TS2412 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |
| `src/compiler/watchUtilities.ts:125:5` | TS2375 | expanded chain has no eligible owned optional declaration with compatible present-undefined evidence |

### Mutants, commands and limits

| Mutant | Named check and observed kill |
| --- | --- |
| Drop `?` from edited `CompilerOptions.traceResolution` | Census missing-property check: total **2,168 → 2,200**; **31** missing-property chains, TS2741 **0 → 24**; all **18** directly checked omitted-key literals caught. |
| Drop only that owner union | Census optional-count check: TS2375 **19 → 20**, optional total **60 → 61**, whole total **2,168 → 2,169**. |
| Remove an edited `?` in the emitted API snapshot | Parsed snapshot reconstruction exits 1. |
| Add undefined to unedited protocol `Diagnostic.source` | Namespace-qualified owner proof exits 1; no compiler-owner false match. |
| Drop `traceResolution` from the snapshot owner ledger | Parsed reconstruction exits 1 on the unlisted addition. |
| Append a newline to unrelated `assignmentCompat1.js` reference | All-other-reference comparison exits 1 after the API proof succeeds. |
| Append a statement to one emitted JS artifact copy | The same byte-array comparator rejects it. |
| Append a newline to actual adapted `src/compiler/types.ts` | The source-idempotence hash map comparison rejects it; original bytes restored. |

The census mutants use identical dependencies and stock compiler API AST edits.
For only their repeated whole-program probes, a scratch Go scheduling overlay
skips the per-entry loop (`CENSUS_WHOLE_ONLY=1`); root discovery, `inspect`,
checker options and the final whole-program call are unchanged. Both primary
before/after censuses use the full unmodified driver. The required-literal
coverage excludes assertions, which bypass construction checking, and keys
supplied by spreads. Mutation sites and complete chains are in
`revision-required-sites.log`; `revision-required-literals.log` checks every
one of the 18 directly checked literal sites.

Representative replay commands, run sequentially from Adamic's root:

```sh
stage3/apply.sh /tmp/adamic-optional/revision-approved > /tmp/adamic-optional/revision-approved-apply.log 2>&1
# Extract the adapter JSON object (starts with "typescript") from apply's log
# using JSONDecoder.raw_decode, into revision-approved-adapt.json.
npm ci --prefix /tmp/adamic-optional/revision-approved --no-audit --no-fund > /tmp/adamic-optional/revision-approved-install.log 2>&1
npm run build --prefix /tmp/adamic-optional/revision-approved > /tmp/adamic-optional/revision-approved-build.log 2>&1
export NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules
node stage3/adapt/20-optional-declarations/check-baselines.cjs /tmp/adamic-optional/pristine /tmp/adamic-optional/revision-approved /tmp/adamic-optional/revision-approved-adapt.json --accept-api > /tmp/adamic-optional/revision-approved-api-proof.log 2>&1
stage3/oracle/run.sh /tmp/adamic-optional/revision-approved /tmp/adamic-optional/revision-approved-oracle > /tmp/adamic-optional/revision-approved-oracle.log 2>&1
source /workspace/adamic-tools/env.sh
go build -o /tmp/adamic-optional/census ./stage3/census/tool > /tmp/adamic-optional/census-build.log 2>&1
# before is pinned pristine plus latest 10, with identical locked dependencies;
# regenerate using scripts/processDiagnosticMessages.mjs src/compiler/diagnosticMessages.json.
/tmp/adamic-optional/census /tmp/adamic-optional/revision-census-before/src/compiler /tmp/adamic-optional/revision-census-before.jsonl > /tmp/adamic-optional/revision-census-before.log 2>&1
/tmp/adamic-optional/census /tmp/adamic-optional/revision-approved/src/compiler /tmp/adamic-optional/revision-census-after.jsonl > /tmp/adamic-optional/revision-census-after.log 2>&1
node stage3/adapt/20-optional-declarations/check-baselines.cjs /tmp/adamic-optional/pristine /tmp/adamic-optional/revision-approved /tmp/adamic-optional/revision-approved-adapt.json > /tmp/adamic-optional/revision-api-proof-after-oracle.log 2>&1
```

The API acceptance phase is explicit and restricted to the proved reference;
shared apply/oracle code is unchanged. Apply refuses an existing output path,
so choose fresh paths when replaying. Stock 6.0.3 comes from the shared API cache;
upstream's own checked build uses its locked dependencies.

Evidence logs: `revision-approved-oracle/report.json`,
`revision-approved-api-proof.log`, `revision-api-proof-after-oracle.log`,
`revision-snapshot-mutants.log`, `revision-final-tests.log`,
`revision-js-manifest.json`, `revision-idempotence-final.log`,
`revision-idempotence-source-hashes.json`, and the full census JSONL files.
The source-level declaration ledger is `revision-approved-adapt.json`.
All evidence and upstream trees are under `/tmp/adamic-optional`, outside Adamic.
The generated scoreboard is preserved as `revision-approved-patch-set.md`;
only adaptation 20's source tools and README are committed by this revision.

Toolchain setup from the first revision succeeded: Go ready 1s, clang 1s,
Node 1s, submodules 1s, cache warm 237s, total 237s; `nproc` 5, CPU quota 4.
Go 1.27.1, clang 20.1.8, Node 24.19.0. The final preparation hit scratch-disk
`ENOSPC`; obsolete self-created trees were removed, preserving proof logs, and
the checked build and snapshot proof were rerun successfully. Package checks
from the unchanged census/meter/loader tooling passed (9.711s and 3.646s);
this unit changes only source-adaptation tooling. The full native Adamic gate,
future adaptations and unproven C/R invariants are outside this unit. No PR.

## Builder assignment pattern, October 9

The remaining-67 parser ledger supplies TS2375 at builder.ts:2273 and 2310.
Replaying this adaptation after all predecessors resolves three internal owners:
ReusableBuilderProgramState.outSignature, hasErrors and emitSignatures. At the
normal numeric position an earlier incompatible member masks those optional
chains. The bounded pattern resolves the two state object assignments inside
createBuilderProgramUsingIncrementalBuildInfo and passes those three fields
through the existing selector. It retains every question mark, key and value.
The source value must contain undefined, every other constituent must fit the
annotation, and indexed or changed owner contracts fail before edits. It is
independent of the checker's first reported incompatible member.

The full apply comparison changes only those three internal declarations.
A real-source mutant replaces buildInfo.outSignature with true and is rejected
by `builder optional value contract drift: outSignature`; original bytes are
restored. Full composition results and exact commands are recorded in
../../scouts/step24/remaining-67/README.md. No public declaration is widened
by this pattern, and no compiler or shared harness changes are included.
