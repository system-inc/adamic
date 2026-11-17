# Proof

Measured on the adapted tree stage3/apply.sh builds at main 433726c2 with road A (99-method-value-properties)
installed. Every run was a sealed Loom proof job. The codemod (codemod-b.cjs), the owner generator (gen-owners-b.cjs),
the scan and the census scripts are on `runtime-performance/method-values-landed` under measure/rv75m9k/, and the
landed-form check on `runtime-performance/method-values-b-evidence`.

## Sites

On road A's tree the stock checker finds 49 method value reads under the meter's 71 names. This adaptation takes 43, 6
stay refused and named in README.md, and stock tsc reads 0 before and after (J-1fhp1ww, J-sx6cxf7).

Every view's receiver, root and each link of its chain, is a const or a parameter or variable the program never writes:
the codemod collects every assignment, compound assignment, ++ and -- target in the program and refuses a site whose
receiver touches one (J-ectam69 prints the evidence per site). Writes by element access or Object.assign aren't seen
that way; the census and the stock checker are the second line.

## The landed form, stacked on road A (J-sx6cxf7, at this head)

- apply builds with both directories: patch-set rows 99-method-value-properties (2 files) and 99-method-value-views
  (10 files);
- the scan reads 6 with stock tsc 0; road B's mutant brings back the 49, tsc 0; re-adapting gives 6; a replay writes
  nothing;
- the engine reproduces the codemod's output in all 79 files of a fresh tree, and its mutant restores every file
  (J-pkn4h1b);
- TestProduct_Stage3AdaptedTree passes (hereby's whole build, harness and testRunner included, 223 s, a cache miss);
- the whole lane passes (J-a8pyf3p, below);
- TestGetterReadThroughView passes. Settling the rest of internal/oracle's products on that box ran out of disk quota
  in /run/loom-scratch, the box's; the pin built what it reads.

## Full single-file censuses, before and after road B

CommonJSBundle, LATENT_FULL=1, LATENT_ASSERT_NO_OUTPUT=1, LATENT_ONLY_FILE, all 79 roots loaded, paired as a multiset
by (kind, phase, reason, text) with positions and the tree's path stripped.

| File | Paired | Removed | Added | Job |
|---|---:|---:|---:|---|
| checker.ts | 8,536 | 0 | 0 | J-fwbz4h7, J-b9w7r5e (split) |
| parser.ts | 3,148 | 14 | 0 | J-pjmh9ve |
| nodeFactory.ts | 3,606 | 0 | 0 | J-em2bd4g |
| program.ts | 1,496 | 4 | 12 | J-6zt3yfz |
| emitter.ts | 1,110 | 2 | 0 | J-em2bd4g |
| classFields.ts | 633 | 0 | 0 | J-em2bd4g |
| executeCommandLine.ts | 388 | 2 | 0 | J-6zt3yfz |
| transformers/utilities.ts | 146 | 0 | 0 | J-sx6cxf7 |
| watch.ts | 79 | 0 | 0 | J-6zt3yfz |
| tsbuildPublic.ts | 56 | 0 | 0 | J-6zt3yfz |
| watchPublic.ts | 34 | 0 | 0 | J-6zt3yfz |

`adamic types` over the 79 roots: 0 diagnostics before and after (J-6zt3yfz).

**Removed:** the target refusals for isEmitNotificationEnabled (emitter.ts 868, 944), onWatchStatusChange and
afterProgramCreate (executeCommandLine.ts 865, 999), hasInvalidatedResolutions and hasInvalidatedLibResolutions
(program.ts 1604, 1685), and program.ts's earlier stop at 480:54 ("a literal method through a view that erases its
receiver", System.createDirectory) with its Boundary.

**parser.ts's 14** are seven NotYets "a method sharing its name with a getter" (3812 to 3914, the JSDoc type parsers)
and their Boundaries, as second copies recorded under a caller unit. Lower's deciding run (J-g96pcx2, a trace of
eepPropertyPresence on the evidence branch only) shows each read is reached with the name an accessor name and the
symbol a method, so the stop fires at road B exactly as before, and each is recorded in its own unit in both runs. It's
order, ruled by Lower on #rv75m9k.

**program.ts's 12 added** are the next stops of one statement road B reaches further into: the `compilerHost` object
literal in createCompilerHostWorker. Before road B it stopped at 480:54. Now it reaches "a function value that captures
the variable its own initializer declares" (the literal's own `fileName => compilerHost.readFile(fileName)`), memoize's
body in core.ts (1895 to 1901: "assigning to an Identifier", "reading value"), and "a presence test of a method whose
presence the receiver's construction doesn't prove" at `system.getEnvironmentVariable ?` (493). All are in code road B
doesn't touch, and each existed behind the stop road B removes.

## The getter pin (Lower's ruling)

internal/oracle's TestGetterReadThroughView: a factory whose interface declares a method and whose object implements it
with a getter returning a closure, read through `{ readonly make: (typeof factory)["make"] }`, captured and called. Node
prints 12, 13, factory: the getter runs at every read. Native and the JavaScript backend agree (J-rb661pc). The mutant
that makes finishAccessors skip the Property read (class_accessors.go, `if false && names[expression.Name] && ...`)
turns the native row red; one that leaves accessorNames empty survives, as it should, since accessorNames only feeds
refusals.

## Emitted JavaScript

JavaScript-changing by construction: each view adds one `const` before its statement, each arrow one `const` at its
scope, and the reads name the view or the arrow. transpileModule's diff of every changed file is in J-1fhp1ww's log (which still carried toKey's arrow, since left):
those lines and nothing else.

## The five arrows: Adamic's two counts (#trxxkge's conditions)

An arrow reads its member when it's called, where tsc read it when the value was taken. Counted with the stock checker
over all of upstream's src (601 files), by `arrow-counts.cjs` on the evidence branch:

1. **Writes to a wrapped member after the arrow is made: 0.**
   - Nothing in src writes `directoryExists` or `createDirectory` on a value typed System. The writes to those names
     land on CompilerHostLikeForCache (program.ts:605, :615, changeCompilerHostLikeToUseCache's overrides), on a
     compiler host (program.ts:4907, :4956; watchPublic.ts:661, :662) and on a SolutionBuilderHost (tsbuildPublic.ts
     :785, :786). changeCompilerHostLikeToUseCache never receives `system`: its five callers hand it a CompilerHost
     (watchPublic.ts:128, executeCommandLine.ts:919, services.ts:1781), a `CompilerHost & ResolutionCacheHost`
     (watchPublic.ts:637) and a SolutionBuilderHost (tsbuildPublic.ts:759). The System arrows also read at the same
     moment tsc reads: `(compilerHost.x || system.x)(path)` takes and calls the member in one expression.
   - tsbuildPublic's `compilerHost` is the object literal createCompilerHostFromProgramHost makes (tsbuildPublic.ts:435;
     its `getCanonicalFileName` is `createGetCanonicalFileName(...)`, watch.ts), and nothing writes it: the writes to
     `getCanonicalFileName` in src land on a server Project (server/project.ts:586) and on the test TestServerHost
     (virtualFileSystemWithWatch.ts:451, watchEnvironment.ts:586).
2. **Identity, `length` or `name` reads where tsc's function value flowed: 0.** The value passed to
   createModuleResolutionCache and createTypeReferenceDirectiveResolutionCache, followed through every callee it's
   handed to (createModuleOrTypeReferenceResolutionCache, createPerDirectoryResolutionCache,
   createNonRelativeNameResolutionCache, createPackageJsonInfoCache, toPath), is only ever passed on or called (path.ts:766):
   13 uses, none compared, stored or read for a property. The two `||` sites call the value where they take it.

## The whole lane

J-a8pyf3p at this head: `PASS stage3 landing lane`, 106,366 passing, 1 failing (the sanctioned Public APIs
acknowledgement, baseline diff api/typescript.d.ts), 0 pending, no timeout, in 390 s. It's the Node witness of the
value-identical claim, Correctness's condition.

## Not run

- A count of the arrows' calls per lane run. The five arrows sit off the per-node paths (a directory created or tested,
  a resolution cache made once per build state).
