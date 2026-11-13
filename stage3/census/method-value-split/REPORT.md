# Method-value split

**346 non-callee method-access sites: 92 presence (26.6%), 34 optional-call (9.8%), 220 escapes (63.6%).** These are a **static approximation of stage 0’s unbound-method rule**, not measured refusal counts. Every site has exactly one class in [sites.json](sites.json).

The literal non-callee access predicate includes **54 write-only assignment targets**, recorded as `escapes` with `accessRole: "write"`. Excluding these leaves **292 actual read sites: 92 presence, 34 optional-call, 166 escapes**. The broad inventory has 161 distinct method names and 198 owner/name groups.

**Two presence-classified sites forward a method value through `||` into a variable** (`program.ts:1603` and `program.ts:1684`). They satisfy the brief’s explicit “left of `||`” shape, but not its stricter “only a truth test” wording. They carry `logicalOrMayForwardMethodValue: true`. Treating these conservatively as escapes gives **90 presence, 34 optional-call, 222 escapes**, or **90 / 34 / 168** after excluding write targets. Thus **90** is the conservative presence-only read count.

## Top 20 methods by sites

Grouped by declaring owner and method name; ties sorted by owner/name. Includes write targets under the broad access predicate.

| Owner.method | Sites | Presence | Optional-call | Escapes |
| --- | ---: | ---: | ---: | ---: |
| `NodeFactory.liftToBlock` | 20 | 0 | 0 | 20 |
| `ParenthesizerRules.parenthesizeExpressionForDisallowedComma` | 12 | 0 | 0 | 12 |
| `CompilerHost.getSourceFile` | 7 | 0 | 0 | 7 |
| `ModuleResolutionHost.getCurrentDirectory` | 7 | 0 | 7 | 0 |
| `ParenthesizerRules.parenthesizeLeftSideOfAccess` | 7 | 0 | 0 | 7 |
| `CompilerHost.getCanonicalFileName` | 6 | 0 | 0 | 6 |
| `ParenthesizerRules.parenthesizeOperandOfPrefixUnary` | 6 | 0 | 0 | 6 |
| `ModuleResolutionHost.fileExists` | 5 | 0 | 0 | 5 |
| `CompilerHost.getParsedCommandLine` | 4 | 2 | 0 | 2 |
| `ProgramHost.resolveLibrary` | 4 | 2 | 0 | 2 |
| `ProgramHost.resolveModuleNameLiterals` | 4 | 2 | 0 | 2 |
| `ProgramHost.resolveModuleNames` | 4 | 2 | 0 | 2 |
| `SymbolTracker.trackSymbol` | 4 | 4 | 0 | 0 |
| `System.getEnvironmentVariable` | 4 | 3 | 0 | 1 |
| `WatchCompilerHost.afterProgramCreate` | 4 | 1 | 0 | 3 |
| `WatchHost.onWatchStatusChange` | 4 | 1 | 1 | 2 |
| `BuilderProgram.hasChangedEmitSignature` | 3 | 0 | 2 | 1 |
| `EmitHost.getCanonicalFileName` | 3 | 0 | 0 | 3 |
| `ModuleResolutionHost.getGlobalTypingsCacheLocation` | 3 | 1 | 1 | 1 |
| `ModuleSpecifierResolutionHost.getPackageJsonInfoCache` | 3 | 0 | 3 | 0 |

Of the 92 presence sites, **56 have at least one direct call of the same resolved method in the recorded guarded regions; all 56 use text-equal receivers for every such call**. The other 36 have no such call. Their universal same-receiver flag is vacuously true; `hasGuardedCall` and `guardedCallCount` distinguish them. No presence site has a detected different-receiver call.

46 owner/name groups occur only as presence in this syntactic inventory. This is not a count of adaptations that can be removed: it includes the logical-OR caveat and does not establish stage 0 coverage.

## SourceFileLike.getPositionOfLineAndCharacter

**The only two resolved readers of this particular interface member in all of `src/` are:**

| File:line:column | Class | Evidence |
| --- | --- | --- |
| `src/compiler/scanner.ts:469:12` | presence | Tests the member; the guarded call at 470 uses the same `sourceFile` receiver text. |
| `src/compiler/scanner.ts:470:9` | direct-call | `sourceFile.getPositionOfLineAndCharacter(line, character, allowEdits)`; excluded from the non-callee census. |

One other member access has the same spelling but resolves to **SourceFile**, not SourceFileLike:

| File:line:column | Class | Declaring owner |
| --- | --- | --- |
| `src/services/formatting/smartIndenter.ts:594:27` | direct-call | `SourceFile` in `src/services/types.ts:172` |

All 16 `src/**/tsconfig*.json` files were inspected, creating programs for those with roots, including harness, services, server and testRunner. An independent lexical AST inventory over all `src/` `.ts`, `.tsx`, `.js` and `.jsx` files found no additional property/literal-key readers missing from the project inventory. There are no harness/testRunner readers of the member. [readers.json](readers.json) records all three same-spelling accesses and resolved owners.

A textual inventory also finds definitions, imports, free-function calls in `program.ts`/`sourcemap.ts`, and the provider object property in `server/editorServices.ts:3939`. These are not reads of this method member; no same-name destructuring reader was found.

**SourceFileLike is in `tests/baselines/reference/api/typescript.d.ts`**, in merged declarations at lines 5912 and 5915. **Its optional `getPositionOfLineAndCharacter` member is not.** The member is marked `@internal` at `src/compiler/types.ts:4296–4297`. The same-named method at baseline line 5958 belongs to `SourceFile`. The baseline also exposes the free function taking `SourceFileLike` at line 8502.

## What the scan cannot see

- This symbol-and-syntax inventory does not reproduce stage 0’s traversal, first-error stopping, checker-body skips, instantiations, refusals or module format. Its counts must not be subtracted from the roughly 109 methods / 198 first-error sites in the question. A direct callee is excluded even if stage 0 rejects something else about its call.
- Only symbols with MethodSignature or MethodDeclaration declarations qualify. Function-valued properties, getters returning functions, dynamic keys, destructured reads, unresolved/`any` accesses, aliases and reflection are outside this predicate. Literal string/template/numeric element keys are supported. Assignment targets are included by the literal access predicate and explicitly marked; they are not reads.
- Presence follows the listed truth consumers, including transparent parentheses/assertions, logical chains feeding a condition, `!`, null/undefined equality comparisons and `typeof member === "function"`. Value-forwarding `||` reads are flagged separately. This is not general dataflow or proof that no value escapes through an alias.
- Guarded regions are syntactic: logical right operands plus the selected enclosing conditional/if arm or while body, accounting for immediate negation/null-comparison polarity. Nested direct calls resolving to the same method declaration are recorded and deduplicated by AST node. Calls through saved aliases, `.call`/`.apply`/`.bind`, post-guard continuation and effects in callees are not followed. Compound conditions do not establish member presence on every path through a region.
- Receiver equality is exact AST source text, not object identity or stability. Equal text can hide shadowed bindings or reevaluate getters; aliases can name the same object differently. This report proposes no rule and does not establish that all presence sites are safe to admit.
- Programs use their unchanged tsconfig options. The compiler project has 79 roots / 79 compiler files. The compiler host falls back to the installed stage3/api Node types when the configured `node` package is absent upstream; no options are weakened. Options diagnostics are empty. This does not establish a checker-clean adapted program. [scan-meta.json](scan-meta.json) records scope and TypeScript 6.0.3.

## Stage 0 cross-check: blocked; scan target confirmed

The fresh scanner-only census **could not be run**. This was a dependency/authentication blocker, not a timeout or a claim that the census was too heavy.

The latent overlay requires the pinned `cohere` submodule, absent in both the shared checkout and scratch worktree. After selecting the documented `/workspace/adamic-tools/env.sh` toolchain and creating its missing temporary directory, setup failed with:

```text
go: cannot load module ../../../cohere/TypeScript/tsc listed in go.work file: open ../../../cohere/TypeScript/tsc/go.mod: no such file or directory
```

Attempted `git submodule update --init --recursive` in scratch only. Both clone attempts failed:

```text
fatal: could not read Username for 'https://github.com': No such device or address
fatal: clone of 'https://github.com/system-inc/cohere.git' into submodule path '/workspace/scratch/method-value-split/cohere' failed
```

Consequently there is no fresh `LATENT_ONLY_FILE` result or independent refusal confirmation from this box. The scan confirms **scanner.ts:469:12 = presence, owner SourceFileLike, one text-equal guarded call at 470**.

Historical checked-in evidence separately records one Refused for the following reason in `stage3/meter/runs/20261113T131710Z.tUxRKb/main/report.json`, under `latent_lowering.reason_rows`:

```text
a method read as a value (getPositionOfLineAndCharacter would lose its object, and this with it)
```

That archived run identifies compiler commit `d2fe2586e7cb8d34e1703479d3fc18b61ff577e9`; it is not a rerun of this report. `stage3/meter/README.md:337` separately documents the location/root association as the historical ES-mode sys.ts-root first stop. Neither substitutes for the requested fresh census.

With the pinned submodule available, build the guarded latent overlay as its README directs and run:

```sh
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 \
LATENT_ONLY_FILE=/workspace/scratch/method-value-adapted-3/src/compiler/scanner.ts \
/workspace/scratch/method-value-latent-census \
  /workspace/scratch/method-value-adapted-3/src/compiler \
  /workspace/scratch/method-value-scanner.jsonl
```

## Reproduction and verification

- Base fetched and confirmed: `origin/main` = `1046dcfdbbdca449996d5045dbff7e052e680a1c`. Work stayed in `/workspace/scratch/method-value-split`, created as a fresh detached worktree. The report branch changes only this census directory; no adaptation or compiler changes.
- Public TypeScript clone: `/workspace/scratch/method-value-typescript`. HEAD and `v6.0.3` both confirmed as `050880ce59e30b356b686bd3144efe24f875ebc8`.
- **Installed projects: `stage3/api` only**, via `npm ci` against its existing lockfile (exit 0; three packages installed). Apply requested no other stage3 install. No npm install ran in upstream TypeScript.
- Apply below exited **0**. This is successful adapted-tree production, not an oracle/npm build or test-suite pass.

```sh
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/method-value-typescript \
  stage3/apply.sh /workspace/scratch/method-value-adapted-3
NODE_PATH="$PWD/stage3/api/node_modules" node --max-old-space-size=4096 \
  stage3/census/method-value-split/scan.cjs \
  /workspace/scratch/method-value-adapted-3 \
  "$PWD/stage3/census/method-value-split"
```

Validation checked unique file/line/UTF-16-column locations, exactly one class per site, every saved source line against the adapted file, guarded-call counts/receiver flags, the exact scanner target and complete lexical coverage for the named member across src. Thirteen focused classification fixtures also passed, covering compound conditions, value-forwarding logical OR, ternaries, calls, optional calls, arguments, negation, typeof/null comparisons, bind and literal element access. Baseline interface/member presence was checked separately. [evidence.json](evidence.json) pins lockfile, apply-log, patch-set and compiler source hashes and records the census setup failure.
