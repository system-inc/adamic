| Restore site and quoted line | Written object allocations | Saved value source object allocations | One literal, same one |
|---|---|---|---|
| `src/compiler/program.ts:4957` — `host.compilerHost.fileExists = originalFileExists;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/program.ts:4958` — `host.compilerHost.directoryExists = originalDirectoryExists;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/program.ts:4959` — `host.compilerHost.getDirectories = originalGetDirectories;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:782` — `host.readFile = cache.originalReadFile;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:783` — `host.fileExists = cache.originalFileExists;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:784` — `host.directoryExists = cache.originalDirectoryExists;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:785` — `host.createDirectory = cache.originalCreateDirectory;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:786` — `host.writeFile = cache.originalWriteFile;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:787` — `compilerHost.getSourceFile = cache.originalGetSourceFile;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/tsbuildPublic.ts:788` — `state.readFileWithCache = cache.originalReadFileWithCache;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/watchPublic.ts:658` — `compilerHost.readFile = originalReadFile;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/watchPublic.ts:659` — `compilerHost.fileExists = originalFileExists;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/watchPublic.ts:660` — `compilerHost.directoryExists = originalDirectoryExists;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/watchPublic.ts:661` — `compilerHost.createDirectory = originalCreateDirectory;` | Unmeasured | Unmeasured | Unmeasured |
| `src/compiler/watchPublic.ts:662` — `compilerHost.writeFile = originalWriteFile!;` | Unmeasured | Unmeasured | Unmeasured |

# Stopped measurement: #fj0vqtf, Lower’s ground

**This is not the requested allocation census. No tsc restore has an allocation verdict.**
“Unmeasured” means the instrument never queried that receiver on a complete whole-program graph;
it does **not** mean the head returned `Unknown`, and it does not imply “no” or zero allocations.
The table retains all discovered restore candidates, including callable-property siblings.

## Controls

The scratch-only `probe_test.go.txt` fully lowers two small TypeScript programs using this exact
head, assigns allocation sites, builds `newAllocationFlowGraph`, and calls its unchanged
`ReachingAllocations` directly. Both the save and restore name the same module-level `host` local.
The programs are embedded in the probe; they differ only in `const host: Host = a` versus
`const host: Host = flag ? a : b`, with `flag = Date.now() > 0`, two arrow-bearing object literals,
`const original = host.read`, and `host.read = original`.

| Control | Written receiver | Saved-read receiver | Requested one-and-same predicate |
|---|---|---|---|
| One literal | `Sites:[-1], Unknown:false` | Same local, same query | **yes** |
| Choice of two literals | `Sites:[-2 -1], Unknown:false` | Same local, same query | **no, 2** |

`-1` is the `a` literal (control source line 2, `const a: Host = {read: (): string => 'a'};`);
`-2` is `b` (line 3, `const b: Host = {read: (): string => 'b'};`). These numbers are graph-local,
not source identities. `controls.log` records the passing run. These controls test a cardinality
predicate over the head’s actual reach results, not a new compiler rule.

Both restore programs **lower successfully on 949a2198**. This matters: this head’s
`hostMemberValueRestore` (`internal/lower/host_member_value.go:254`) asks the graph for both sets,
checks unknown reach and known closures, then applies `hostMemberLayout` over the combined sites.
It does not require a singleton or equality of the two sets. `hostMemberReach`
(`internal/lower/host_member_write.go:201`) likewise permits multiple plain literals. The proposed
one-and-same ruling is a separate measurement target, not this head’s existing acceptance rule.

## Method and stopping reason

The requested complete whole-program instrument was **not built**. The preserved probe only
attempts ordinary loading/lowering for the whole roots and makes direct reach queries for the
controls. It does not bypass earlier walls or reach tsc’s restore bodies. Thus the central task
remains undone; the report stops here rather than presenting a source-flow approximation as the
head’s own answer.

The reach API consumes `ir.Expression` and an `ir.Program`, not checked AST nodes:
`shape_flow.go:22` delegates to the shared allocation graph, whose sources, calls, returns,
closure parameters, liveness, stores and projections are collected from IR (`view_flow.go:114`).
The normal pipeline can stop before complete bodies or receiver expressions exist. Querying a
partial graph, dropping unsupported statements, or inventing AST-to-IR edges would not establish
what this head would judge on a complete program. I did not implement or validate a faithful
scratch construction of that complete IR. This is an implementation limit of this attempt,
not a claim that the requested scratch instrument is impossible, and not a request for permission.

The final whole-root attempt is retained in `whole-program.log`. No per-site allocation query was
made on that input. The final adapted program loads, but `Lower` returns
`lower: stage 0 compiles a program from one entry file, got 79` and no IR. This is only the public
entry guard, not an exploration of the later walls or a direct reach census. It is evidence of where
this attempt stopped, not a substitute for the direct
whole-program census the task explicitly requested.

## Site discovery

`host-sites.txt` was not found in the checkout; the `ahra` CLI was unavailable. I used the allowed
AST derivation route with the pinned stock TypeScript API in `stage3/api`. `inventory.cjs` traverses
every assignment in the compiler roots, follows direct variable initializers and aliases to member
reads, and selects interface/type-literal method targets. It additionally keeps callable targets
whose RHS is named `original…` or `cache.original…`, to expose restores of values returned by a
helper or stored in a cache. This latter selection is a naming heuristic, not a proof of saved-value
provenance or an exhaustive interprocedural derivation. No reach inference occurs in this script.

I manually checked the cache provenance: `changeCompilerHostLikeToUseCache` in `program.ts` saves
`host.readFile`, `host.fileExists`, `host.directoryExists`, `host.createDirectory`, and `host.writeFile`,
returns them as `original…` fields, and `watchPublic.ts` destructures that result before restoring.
`tsbuildPublic.ts` stores the helper’s returned originals in `state.cache`, along with
`compilerHost.getSourceFile`, and restores them in `disableCache`. `program.ts`’s three direct restores
read their originals from `host.compilerHost`. These are syntactic provenance observations only;
they do not identify allocation sites of those host objects.

`sites.json` contains the exact roots, quoted assignment lines and raw inventory metadata.
Its `saved` field can be an intermediate cache-field read, not the original host-method read;
it must not be interpreted as the object whose allocation reach the task asks to measure.
Rows marked `declaredMethod:false` are callable-property siblings in the **adapted** tree, retained
for visibility; existing adaptations have converted some optional methods to properties. The
`state.readFileWithCache` row is an auxiliary state-property restore, not a host-interface method.
This discovery has not been proven exhaustive for arbitrarily renamed values returned through
other functions or fields. There is no claim that this is a complete census.

## Setup, provenance and limits

Base: `949a2198c5a869d0a1598949add6708cb7949880`, fetched from
`compiler/host-member-value-3`; its complete change was reviewed (the `.a` and `.ts` fixture twins
were also checked byte-identical). A fresh detached worktree was created at
`/workspace/scratch/restore-reach`, separate from `/workspace/adamic`.
`git submodule update --init --recursive --depth 1` completed, and `npm ci` ran in `stage3/api`.
`ADAMIC_TYPESCRIPT_SOURCE` named a separate checkout of TypeScript v6.0.3 at
`050880ce59e30b356b686bd3144efe24f875ebc8`. `stage3/apply.sh` generated the adapted input;
no adaptation was added or edited. TMPDIR, GOTMPDIR and GOCACHE were under
`/workspace/scratch/restore-reach-runtime`. Go was 1.27.1.

The only instrument installed outside this report directory was a temporary test in the private
worktree’s `internal/lower`, copied from `probe_test.go.txt` and removed after the run. No compiler
implementation was changed. All preserved Go files under stage3 end in `.go.txt`.
`reproduce.sh` installs that test temporarily and requires an already prepared private worktree
and completed adapted tree. The report branch adds only `stage3/census/restore-reach` artifacts.

A useful continuation must construct and validate the whole-program IR used by this exact graph,
retain both receiver expressions and literal-to-source locations, and query every site directly.
Until then, allocation locations, literal quotations, counts and tsc yes/no decisions remain absent.


Inventory: 79 compiler roots; 15 restore candidates, of which 6 have an interface/type-literal method declaration in the stock API.
