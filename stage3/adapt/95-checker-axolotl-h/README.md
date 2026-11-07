# 95h: truthful host slots and six owned maybeBind folds

Task kpp0f8s. All six-site checks pass; the final ledger is
[fold-evidence/six-sites/checks.json](fold-evidence/six-sites/checks.json).
The completed two-site evidence remains historical, at commit
[c2714b29c](https://github.com/system-inc/adamic/tree/c2714b29c5937f553d79e98e9fae80aab8ed995a/stage3/adapt/95-checker-axolotl-h).
Only this adaptation directory changes from main
7e3c0682da96cbe01f7ae9b9820f1bd8f1993671. TypeScript is the pinned 6.0.3 source,
050880ce59e30b356b686bd3144efe24f875ebc8.

## Ownership and ruling ledger

Adamic's ruling on task kpp0f8s authorizes the stage 3 JavaScript-changing
rewrite `maybeBind(X, X.m)` to `X.m?.bind(X)` when X is one identifier.
The corrected 99x owner audit assigns these six sites to 95h, because 95h pins
the whole writer owners. 99x is amended separately on typescript/maybebind-adapt-2.

| File | Pinned implementation owner | Member |
| --- | --- | --- |
| checker.ts | createBasicNodeBuilderModuleSpecifierResolutionHost | getSymlinkCache |
| checker.ts | createBasicNodeBuilderModuleSpecifierResolutionHost | getGlobalTypingsCacheLocation |
| program.ts | createProgram | getGlobalTypingsCacheLocation |
| program.ts | createProgram > getEmitHost | realpath |
| program.ts | createProgram > getEmitHost | createHash |
| program.ts | createProgram > getEmitHost | trace |

The original before-owner pins remain exactly:

- checker owner: 9171aad5d35e92052c682093a2f2b7fb0ad49a63787488f78d0aa74161d9b9b1
- createProgram implementation: 505bb77211eb4818ad15ed0701812c6c51eb06059becdd73923a5f3e73b69152
- nested getEmitHost: cdf72d661d43306e9bf07a30681c65ef26ef550ca372d7e2a5b3960123249e60

sites.json adds after hashes for each touched owner and literal, and before/after
initializer text for all six members. The createProgram after hash includes
both its direct rewrite and the three rewrites inside getEmitHost. Owner
selection requires the implementation body, excluding createProgram's overload
signatures. No pin is widened to a different owner.

The checker maybeBind named import is removed only when the two checker uses
are gone. Its exact neighboring import gap is pinned before and after. The
specifier and comma are replaced by an empty line, retaining CRLF and all
subsequent coordinates. This keeps adapter 78's replay location evidence
identical. program.ts's maybeBind import is untouched; other uses remain there.

## Adapter contract

The planner binds compiler and services with stock TypeScript 6.0.3 through
47-host-errors' compiler(). Before-form calls must have two arguments, a callee
resolving to core.ts's maybeBind symbol, an Identifier receiver and a
non-optional property access on that same identifier symbol. Only the six
inventoried initializers qualify. After-form bind chains are also checked by
AST and receiver symbol. All member mayStoreUndefined, hostMember and
destinationOwner values remain unchanged.

Each owner accepts a complete before or complete after state. The parent
createProgram pin covers four sites, so all fourteen nonempty partial subsets
of those four sites refuse. Both mixed checker states and inconsistent checker
import states refuse as well. The complete checker and program owners may
independently be before or after. Every after owner is normalized by AST slices
back to its unchanged original before pin. createProgram's own normalized pin
is checked before its literal, so complete-tree replay still refuses there,
with main's same expected and observed hashes, never at an earlier owner.
The checker literal is checked first to reject a foreign third member at its
literal hash. All input, shape and type checks finish before any source write.

The bound value types are checked before planning and after rebinding the
complete planned tree:

| Member | Value type |
| --- | --- |
| checker getSymlinkCache | (() => SymlinkCache) or undefined |
| both getGlobalTypingsCacheLocation sites | (() => string or undefined) or undefined |
| getEmitHost realpath | ((path: string) => string) or undefined |
| getEmitHost createHash | ((data: string) => string) or undefined |
| getEmitHost trace | ((s: string) => void) or undefined |

sites.json records the exact TypeScript typeToString spellings, including each
union bar. Any type mismatch refuses and reports the observed type.

95h also retains its original four optional internal
ModuleSpecifierResolutionHost declaration repairs in types.ts: getSymlinkCache,
readFile, getModuleSpecifierCache and getNearestAncestorDirectoryWithPackageJson.
Each becomes an indexed method type unioned with undefined, preserving method
parameter bivariance and its original read type. The @internal interface and
individual slot symbols are checked. 97-cli-axolotl-b, accepted at
27162cab2bb0054b4ed7cad5a9e258f67f8030c1, owns the separate directoryExists,
realpath, getGlobalTypingsCacheLocation and trace declaration widening. Its
source is unchanged. The complete five-literal contextual inventory, including
the two unchanged services producers, remains enforced.

The return value records six rewrites, one import removal and four declaration
repairs on first apply, across three files. It records javascriptChanging true;
javascriptIdentical is true for types.ts and false for checker.ts and program.ts.
A second run and --check write nothing. All source edits are slices, without a
printer. CRLF counts are preserved in every changed file.

## Independent source and JavaScript verification

verify.cjs reconstructs the six initializer changes and checker import cleanup
independently, compares every source file, and verifies complete JavaScript in
both comment modes. --prefix additionally enforces the original owner/literal
pins and their exact after hashes. Full-tree mode compares the complete main
control against the complete candidate: later adapters can already have changed
createProgram on both trees, so their existing edits must remain byte-identical.
This is not an acceptance relaxation in adapt.cjs.

The emitted comparison covers all nine normally built JavaScript files and
both public declarations. The latter must remain byte-identical. The checker
import cleanup eliminates the previous binder maybeBind2-to-maybeBind3 drift.
Three existing temporary bindings shift in each changed bundle:

| Scope | Before | After | Occurrences before / after |
| --- | --- | --- | --- |
| checker getPackageJsonInfoCache arrow | _a | _a2 | 3 / 3 |
| createProgram tracing.pop temporary | _p | _q | 3 / 3 |
| getEmitHost getBuildInfo arrow | _a2 | _a3 | 3 / 3 |

rename-proof.cjs resolves every occurrence to its unique symbol, requires the
same declaration and two corresponding uses, rejects escape or .name reads,
and scans all bundles for string/template and property-name mentions. It also
identifies all six newly introduced bind temporaries as distinct symbols, each
with exactly three occurrences and uses only inside its own rewritten
initializer. These new temporaries receive fresh canonical comparison names.
In particular, the new _p binding is never confused with the original tracing
binding, now _q. All substitutions affect only resolved binding occurrences,
so canonicalization introduces no capture.

The complete normalized output must equal the before output with exactly the
six initializer lowerings and their uninitialized var declarations added.
Each lowering is independently checked against esbuild's ordinary upstream
target, es2020 and node14.17. No other difference or rename is allowed. Mutants
cover string/property name observations, .name reads, a missed renamed use,
and a foreign new-temporary producer.

## Six-site evidence and reproduction

New evidence is under fold-evidence/six-sites. Raw replay results, including
exact stdout, stderr, writes and both outcome records, are packed into one
replay.tar.gz. Raw shard results are packed into one shards.tar.gz, updated as
each bounded oracle piece completes. Each archive has a SHA-256 list beside it
covering the archive and every uncompressed member. pack-evidence.py verifies
all member hashes after writing. Unpacked private test views are scratch-only.
The shipping directory is kept below 300 files.

Every command is logged. The whole lane and all manifest shards run from an
isolated frozen checkout so later evidence commits cannot change their broad
artifact identity. Each oracle piece is bounded below 25 minutes, committed
and pushed plainly as it finishes. The official merge compares actual test
identities with multiplicity, counts, the sanctioned failure title and API path,
and complete raw baseline bytes against the same-artifact whole lane.

```sh
source /workspace/adamic-tools/env.sh
bash cloud/setup.sh > /workspace/scratch/95h6-setup.log 2>&1
npm ci --prefix stage3/api > /workspace/scratch/95h6-npm.log 2>&1
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source
# Run --build-product applies from the main control and candidate checkouts.
NODE_PATH=stage3/api/node_modules node stage3/adapt/95-checker-axolotl-h/contract.cjs PREFIX NEW_CONTRACT
NODE_PATH=stage3/api/node_modules node stage3/adapt/95-checker-axolotl-h/verify.cjs --prefix PREFIX NEW_CONTRACT
NODE_PATH=stage3/api/node_modules node stage3/adapt/95-checker-axolotl-h/verify.cjs BEFORE AFTER
NODE_PATH=stage3/api/node_modules node stage3/adapt/95-checker-axolotl-h/compare.cjs BEFORE AFTER NEW_NORMALIZED
python3 -B stage3/adapt/95-checker-axolotl-h/replay.py MAIN_REPO CANDIDATE_REPO BEFORE AFTER NEW_REPLAY
# adamic types receives all 79 sorted compiler roots in one invocation.
python3 -B FROZEN/stage3/lane/prepare_shards.py ARTIFACT_CACHE
bash FROZEN/stage3/lane/run.sh WHOLE --artifact ARTIFACT
# oracle-piece.py runs unchanged manifest commands in bounded groups.
STAGE3_ARTIFACT=ARTIFACT python3 -B FROZEN/stage3/lane/merge_shards.py RESULTS --whole WHOLE --output MERGE_JSON
```

Historical declaration census, the earlier two-site stop, its accepted rename
proof and its completed oracle remain in fold-evidence and in the linked
historical README. They are not claimed as six-site validation.

## Completed six-site validation

| Check | Result |
| --- | --- |
| Full --build-product applies | Main control and six-site candidate both exit 0 |
| Paired replay | All 81 outcomes match except 95h; 95h refuses at createProgram's own original pin, never earlier; adapter 78 stdout is byte-identical |
| All 79 compiler roots | Adamic types exits 0 with zero diagnostics |
| JavaScript and declarations | All nine outputs differ only by six modeled lowerings after symbol-proved normalization; both public declarations are byte-identical |
| Mutants | All 22 source mutants and all five rename-proof mutants refuse; source refusals write nothing |
| Oracle | Whole lane, all 398 shards and unchanged official merge pass on one verified artifact |

The six-site oracle has exactly 106,366 passing tests, one sanctioned Public
APIs failure, and zero pending. The official merge accounts for all 106,367
actual test identities with multiplicity and matches the complete raw baseline
diff bytes to the whole run. Every oracle piece was committed and pushed
plainly as it finished. No lane, expectation, sanction or other adaptation
source changed. The original owner/literal before hashes and every member's
classification fields are unchanged, as recorded in pin-preservation.json.

The isolated oracle checkout froze the six-site implementation at
61526db0b0ebee0f30abc42089e79f05a02c9575
and used artifact
972f66fa46df8a2e07b1905b3c832309134d1620a5f2d451dae2232978bb038c.
Its source, nine compiler JavaScript files and both public declarations were
compared byte for byte with the proven six-site after tree before testing.
Later evidence commits do not change that frozen transformation.

| Oracle piece | Seconds |
| --- | ---: |
| Preparation | 286.585 |
| Whole lane | 284.232 |
| Shards 1 through 40 | 529.266 |
| Shards 41 through 80 | 524.655 |
| Shards 81 through 120 | 530.517 |
| Shards 121 through 160 | 533.813 |
| Shards 161 through 200 | 527.644 |
| Shards 201 through 240 | 529.605 |
| Shards 241 through 280 | 527.433 |
| Shards 281 through 320 | 526.420 |
| Shards 321 through 360 | 521.347 |
| Shards 361 through 398 | 508.461 |
| Official merge | 4.974 |

All pieces are below 25 minutes. No shard exceeded its unchanged 30-second
limit. At most two shard commands ran concurrently. No task process remains
running.

The complete raw shard results are in one
[six-sites/oracle/shards.tar.gz](fold-evidence/six-sites/oracle/shards.tar.gz),
with [its SHA-256 list](fold-evidence/six-sites/oracle/shards.sha256).
The complete raw replay records, including the superseded coordinate-only
attempt, are in one [replay.tar.gz](fold-evidence/six-sites/replay.tar.gz),
with [its SHA-256 list](fold-evidence/six-sites/replay.sha256). Source mutants
and the whole lane each have their own compact archive and hash list. Every
uncompressed member was hash-verified when packed.

[The oracle summary](fold-evidence/six-sites/oracle/summary.json) records each
piece commit, timing, input identity and final counts.
[The official merge verdict](fold-evidence/six-sites/oracle/merge.json) retains
the unchanged acceptance result.
[The evidence index](fold-evidence/six-sites/evidence-index.json) lists all new
archives and hashes and records the complete adaptation directory's physical
file count, below 300. Historical two-site evidence is retained separately.

MultiMap amendment (task b5w3ycg, wave row 7): 94-correctness-multimap first exports multiMapAddValue and multiMapRemoveValue. 95h owns the two createProgram calls in findSourceFileWorker and addFileIncludeReason and rewrites them to multiMapAddValue; its whole createProgram after-pin includes both calls. The pinned import gap adds multiMapAddValue. Other calls and the final plain Map representation belong to 99-correctness-core-multimap. The interim names remain because a later rename would touch this pinned owner.

verify.cjs independently reconstructs the two MultiMap calls and pinned import gap, and contract.cjs checks the three partial MultiMap states with refusal reasons, zero writes and unchanged file hashes ([mutant evidence](tooling-mutant-evidence.json)).
