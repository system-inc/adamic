# Wave 19 checker audit

Audit baseline: origin/area/stage1-lint `1f9e223d8ef6d0a914407d4b37d0e56568521f4d`.
Cohere pin: `7945d102a6c18dd36adf9114a758ce646e8b2359`.
Legacy branch: codex/typeaware-wave-19, checkpoint `2037685271385ee8e0bb83d599697e22dc45e95d`.
Scope: the nine outstanding wave-19 rules; superseded React reservations are excluded. This is an audit only. No port, checker change, parity run or mutant pass is claimed.

Paths below are relative to `cohere/internal/lint/rules/`. Symbols are spelled as upstream calls them. A missing result facet does not mean the harness has no implementation of the underlying checker operation.

## Missing facts and shared predicates

| Exact upstream call or helper | Rules blocked | Upstream location and missing contract |
| --- | --- | --- |
| `ctx.Program.Options().IsolatedDeclarations.IsTrue()` | `@typescript-eslint/consistent-generic-constructors` | `typescript/consistent_generic_constructors.go:240`. The area's `options` question returns strict-null checking only; no isolated-declarations option. |
| `ctx.Program.Options().NoPropertyAccessFromIndexSignature.IsTrue()` | `@typescript-eslint/dot-notation` | `typescript/dot_notation.go:88`. Compiler-option fact absent. |
| `ctx.TypeChecker.GetIndexInfosOfType(objectType)` | `@typescript-eslint/dot-notation` | `typescript/dot_notation.go:117`. Needs index signatures and their key types, with property-presence distinction. |
| `dotNotationFirstModifier(property)`; `property.Declarations[0].Modifiers()` | `@typescript-eslint/dot-notation` | `typescript/dot_notation.go:133`, `:137`. Ordered modifiers of the first property declaration, skipping decorators; property-info does not expose these. |
| `resolvesToAGlobal(ctx, callee)` | `@typescript-eslint/no-array-constructor`, `symbol-description` | Calls: `core/no_array_constructor.go:205`, `core/symbol_description.go:95`; helper: `core/no_new_native_nonconstructor.go:99`. Shared RuleContext predicate missing. It examines the symbol's first declaration and declaration-file status, not just its name. Existing type-origin exposes declaration origins for type symbols; symbol-origin exposes only a value-declaration filename. An arbitrary node-symbol declaration-origin fact would make the shared predicate exact. |
| `identifierIsShadowed(ctx, operand)` | `valid-typeof` | Call: `core/valid_typeof.go:147`; helper: `core/no_undef_init.go:159`. Shared RuleContext predicate missing. Any source declaration means shadowed; unresolved or zero-declaration symbols do not. It is not the negation of resolvesToAGlobal. Needs arbitrary node-symbol declaration origins as above. |
| `ast.IsGlobalScopeAugmentation(container.Parent)`; `ast.IsExternalModule(container.AsSourceFile())` | `nexus/correctness-no-uncleared-race-timeout`, `nexus/correctness-no-process-exit-after-output` | `nexus/correctness_no_uncleared_race_timeout.go:261`, `:263`; `nexus/correctness_no_process_exit_after_output.go:458`, `:461`. Declaration ancestry and global-augmentation/external-module flags are absent. The class/interface-only declarations question supplies spans, not this ancestry. |
| `writers.ctx.TypeChecker.GetResolvedSignature(call)`; `signature.Declaration()`; `ast.GetFunctionFlags(declaration)` | `nexus/correctness-no-process-exit-after-output` | `nexus/correctness_no_process_exit_after_output.go:314`, `:318`, `:335`. Existing signature-shape does resolve calls, but does not return the declaration, source/body ownership or function flags needed for following local writers. |
| `program.SourceFiles()`; `program.ResolveModule(importer, specifier)`; `program.GetSourceFileForResolvedModule(resolved.ResolvedFileName)` | `nexus/correctness-require-blocking-standard-streams` | `nexus/correctness_require_blocking_standard_streams.go:271`, `:279`, `:283`. Program source/module graph and dependency closure absent. This is the first stopping point; declaration ownership for Process/namespace symbols also needs richer declaration metadata later. |
| `checker.Checker_getResolvedSignature(ctx.TypeChecker, call, nil, checker.CheckModeNormal)`; `resolved.Target()`; `declared.TypeParameters()` | `require-await` | `core/require_await.go:988`, `:992`, `:993`, `:1031`. Declared generic target, type parameters and correspondence with instantiated arguments are absent; instantiated signature-shape alone is insufficient. |
| `checker.Checker_getIndexTypeOfType(ctx.TypeChecker, part, checker.Checker_stringType(ctx.TypeChecker))`; corresponding `checker.Checker_numberType(ctx.TypeChecker)` call | `require-await` | `core/require_await.go:1081`, `:1092` (rest parameter also `:1019`). General string/number index-type lookup by type identity absent; the existing rest-parameter number-index extraction is narrower. |
| `type_checking.IsThenableType(ctx.TypeChecker, node, returnPart)`; `type_checking.GetCallSignatures(...)` | `require-await` | `core/require_await.go:751`; GetCallSignatures calls at `:742`, `:1096`. Shared thenability and union-aware call-signature helpers absent on RuleContext. Existing call-returns/call-count facts do not themselves supply these complete shared judgments. |

The first stopping points above suffice to block these rules. The require-await rows identify further known contract gaps, not a claim of an exhaustive dependency audit after implementing them. GetSymbolAtLocation, GetTypeAtLocation and basic resolved-signature/type-shape operations already exist inside the bridge; the requests concern their missing exposed facets and shared predicates.

## Reproducers

Each source is a small trigger for the blocked path, not a substitute for upstream case parity. Nexus cases require the rule's normal Node declarations; the timeout case also needs Promise/timer declarations.

| Rule | Minimal source / configuration |
| --- | --- |
| consistent-generic-constructors | `class Box<T>{} const x: Box<string> = new Box();`, compiler option `isolatedDeclarations: true`. |
| dot-notation | `class X { private p=1 } new X()["p"];`, rule option `allowPrivateClassPropertyAccess: true`; index-signature path: `declare const x: Record<string, number>; x["p"];`. |
| no-array-constructor | `Array(1, 2);` versus a locally shadowed `Array`. |
| symbol-description | `Symbol();` versus a locally shadowed `Symbol`. |
| valid-typeof | `function f(undefined: unknown) { return typeof x === undefined; }`. |
| no-uncleared-race-timeout | `Promise.race([work(), new Promise(resolve => setTimeout(resolve, 10))]);`. |
| no-process-exit-after-output | `function write(){process.stdout.write("x");} write(); process.exit(0);`. |
| require-blocking-standard-streams | `import "./blocking"; process.stdout.write("x");` with a sibling blocking module; deciding whether the setup is reachable requires the module graph. |
| require-await | `declare function consume<T extends () => Promise<number>>(callback:T):void; consume(async () => 1);`. |

Prior live evidence is on checkpoint `203768527`, under `stage1/cohere/lint/rules/no-script-url/wave19-unpark/`: `evidence/generic-live-refusal.log` records the sanitized native rejection of `isolated-declarations`; `blocked/` contains the per-rule source notes. The proposed generic module compiled and ran, then exited 70; zero upstream matches or mutant passes were established. `evidence/lint.jsonl` records the separate old-branch compilation failure described below. These observations were not rerun for this documentation-only audit.

## Legacy bridge migration inventory

These are wire question names on the old branch, not new upstream Go symbols. The current area's Inspect switch contains none of them. Prepared but unregistered questions were never available on the area's checker and must not be described as regressions.

| Legacy wire API | Old implementation | Registration status / consumers |
| --- | --- | --- |
| `isolated-declarations` | `bridge/tsgo/checker/isolated_declarations.go` | Registered on old wave-19 only; generic constructors. |
| `node-symbol-origin` | `bridge/tsgo/checker/node_symbol_origin.go` | Registered on old wave-19 only; arbitrary-node declaration origins, used by generic/global-symbol paths. |
| `index-signature-access` | `bridge/tsgo/checker/index_signature_access.go` | Registered on old wave-19 only; dot-notation. |
| `declaration-ancestry` | `bridge/tsgo/checker/declaration_ancestry.go` | Prepared, not production registered; Nexus declaration ancestry. |
| `wave19-resolved-callee` | `bridge/tsgo/checker/wave19_resolved_callee.go` | Prepared, not production registered; process-exit writer traversal. |
| `wave19-program-modules` | `bridge/tsgo/checker/wave19_program_modules.go` | Prepared, not production registered; blocking-stream module graph. |
| `wave19-type-signatures`, `wave19-generic-call`, `wave19-type-members`, `wave19-heritage-members` | Corresponding `bridge/tsgo/checker/wave19_*.go` files and old Adamic decoders | Prepared, not production registered; require-await. |

Migrate through the harness-owned `context.checker`; do not transplant the old rule-owned checker/program lifecycle. The C inspection interface itself has not disappeared. Existing area `type-origin`, `property-info` and signature/type graph frames should be reused where their contracts fit rather than privately copied.

### Observed Go shim API incompatibilities in the old landing branch

| Legacy API usage | Current contract / observed diagnostic location |
| --- | --- |
| `(*ast.SourceFile).Path()` | No longer exists. Area uses `PathKey()` for default-library lookup. Old `declaration_facts.go:22`, `declaration_ancestry.go:35`, `wave19_program_modules.go:28`. |
| `(*ast.SourceFile).FileName()` passed directly as `string` | Returns `tspath.RootedFilePath`; area converts with `.AsString()` when serializing. Old `facts.go:290`, `declaration_facts.go:17`, `declaration_ancestry.go:33`, `wave19_program_modules.go:25`, `:77`, `wave19_resolved_callee.go:37`. |
| Native compiler `GetSourceFileForResolvedModule(module.ResolvedFileName)` | Expects `*module.ResolvedModule`, not the filename. Old `wave19_program_modules.go:76`. This native compiler API migration is distinct from cohere's ProgramView call spelled in the audit table. |

All diagnostic paths in this subsection are under `bridge/tsgo/checker/` on checkpoint `203768527`. Compiler output stopped with too many errors; this lists observed breaks, not a complete compatibility certification. No shim, helper or registration was changed in this audit.
