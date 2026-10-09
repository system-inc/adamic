Built: an exact 320-stop inventory, code-pattern groups, isolated topic rebuild measurements, and 14 Node-held .a witnesses.
Commits: main 45487a809f89885a3fc651cd590e7dabf31362dc; scout evidence 7e6ed35b; stricter-options-next 8f32e51e8fc41b8f1177453213ca5453ce764486; delivery SHA is in the handoff.
Results: main reproduces 320 identities; paired original-project entry reports 319 -> 67, with 252 disappearances and zero new diagnostic identities; the .a driver remains loader-blocked.
Mutants: all 14 semantic source mutants differ from fixed Node observations; inventory deletion, duplication, coordinate, code, clearance-total, false-driver-green and site-status mutants are rejected.
Limits: no native parser binary, no emitted-guard proof for these 320 sites, and no full gate; direct .a-driver clearance counts are unmeasured.

# Compiler findings

The exact requested native driver rebuild on main plus stricter-options-next exits 1 before diagnostic collection:

```
adamic: load: project .ts imported into an Adamic-option program needs separate checker ownership, not implemented yet: /workspace/cache/checker-stops-adapted/src/compiler/corePublic.ts
```

This is the **new first stop for the driver**, not 320 cleared errors. The topic distinguishes `.a` checker ownership from project `.ts` ownership. Do not waive this refusal or manufacture a `.ts` Adamic driver to obtain a green build.

A separate authorized measurement uses tsc's **existing** `src/compiler/parser.ts` as the build entry. It selects the existing `src/compiler/tsconfig.json`, which extends `src/tsconfig-base.json`, retains its es2020 library and Node declarations, and audits additional Adamic soundness obligations. The topic build reaches **67 checker diagnostics** and exits 1, first **builder.ts:2273:9 TS2375**. All 67 are baseline identities; no new identities appear. This is a compiler-source project build, **not the native dump driver's acceptance build**.

Supplying the project's requested Node declarations and installed source-map-support changes the control: main on this same original entry reports **319**, all baseline identities except `sys.ts:1598:69 TS2307`. Topic reports **67**, so **252** control identities disappear. Against the pinned 320 there are **253** disappearances: **one dependency effect plus 252 topic/project-options effects**. Library selection, host declarations, iterator-completion typing and option scheduling all differ in this comparison. Counts of absent diagnostics are observations, **not counts of emitted runtime checks**.

The remaining groups are: 51 optional-field/interface stops, 7 definite-argument stops, and 9 result/assignment relation stops. Sixty remaining baseline messages explicitly mention exactOptionalPropertyTypes. The other seven include the implements diagnostic and six nullable uses in moduleNameResolver/resolutionCache; inspect their full messages and obligation attribution rather than calling all seven optional-field errors from their codes alone. The topic's own whole-program report names missing own-presence representation prerequisites, 5bb775ca and alias/presence fixes through 7e7464e6. This scout does not implement or import those prerequisites.

Compiler priorities supported by this evidence:

1. Resolve separate checker ownership for the `.a` dump driver importing the compiler's project files. Until then the direct driver cannot benefit from the project-mode admission path.
2. Finish present-undefined versus absent own-field representation and the rejected optional contracts. The fixture observes `Object.hasOwn({read: undefined}, 'read') === true`; deleting that write changes behavior.
3. Attribute the remaining argument/result relations individually. Indexed access, iterator completion and optional property relations can share diagnostic codes; code alone is insufficient.
4. Preserve the project's selected lib and Node host declarations. The 55 host-declaration, 7 host-member, 2 host-inference and 1 Set-shape baseline stops are not all missing native language operations. Native host implementation remains separate work.

# Groups and exact data

[TABLE.md](TABLE.md) ranks 14 semantic groups and gives up to three examples. [groups.csv](groups.csv) further splits every group by diagnostic code and provides three examples **per code-pattern row**. Where fewer than three sites exist, all sites are provided; none are duplicated to fill a quota.

[diagnostics.json](diagnostics.json) retains all 320 ordered diagnostics: code, complete multiline reason, file, one-based line/column, source line, smallest TypeScript AST token, enclosing expression excerpt (limited to 2,000 characters), and directly enclosing indexed expression when present. Coordinates are in **main-adapted source**, not stock-source coordinates. TypeScript's own 6.0.3 parser supplies AST syntax. There are 23 diagnosed files and 39 directly enclosing element-access expressions; 39 is a syntactic lower bound, **not the count of all index-derived values**. Callback parameters and values carried through aliases are not classified as direct element-access syntax.

[comparison.json](comparison.json) retains status for every identity in both project builds and explicitly marks direct-driver counts unmeasured. [source-hashes.json](source-hashes.json) hashes the diagnosed files; [input-hashes.json](input-hashes.json) hashes every compiler .ts input and both staged .a driver files. Raw compressed logs are under [evidence](evidence). The fresh main log reproduces all 320 pinned identities in the same order.

# Provenance and reproduction

Only this directory changes on branch codex/step24-checker-stops. Scratch checkout `/workspace/cache/checker-stops-scratch` was detached at main; `merge --no-commit --no-ff origin/codex/stricter-options-next` succeeded without conflicts. Its index tree is **6be232f53831a0d1faf8def593ebb8b382c73862**. It was never committed or pushed. Its pinned cohere module is reused from main through a cache symlink; no production files in the delivery checkout were edited.

TypeScript cache pin: **6.0.3**, **050880ce59e30b356b686bd3144efe24f875ebc8**. `bash stage3/apply.sh` built the main-adapted input outside the repository. The existing parser main.a/kinds.a were staged verbatim. Project Node dependencies were supplied by a symlink to `/workspace/cache/tsc-census/npm/node_modules`: TypeScript 6.0.3, @types/node 25.3.3, source-map-support 0.5.10. No new Adamic .ts files were written.

Setup: GOPROXY='https://proxy.golang.org|direct'; `bash cloud/setup.sh`; source `/workspace/adamic-tools/env.sh`. Node ready 0.040s; Go ready 0.065s; markdown ready 0.128s; submodules ready 0.128s; clang ready 0.278s; Go build skipped with validated warming stamp 1.253s; cache warm 1.264s; done **1.316s**. nproc **5**, CPU quota **4**. Node v24.19.0, Go 1.27.1, clang 20.1.8.

Commands below had output redirected to named logs; expected compiler rejections exit 1. No package-wide confirmation or full gate ran.

```
git -c fetch.recurseSubmodules=false fetch origin
git checkout -b codex/step24-checker-stops origin/main
git -c fetch.recurseSubmodules=false fetch origin main:refs/remotes/origin/main codex/step24-scout:refs/remotes/origin/codex/step24-scout codex/stricter-options-next:refs/remotes/origin/codex/stricter-options-next
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step24-checker-stops-setup.log 2>&1
source /workspace/adamic-tools/env.sh
bash stage3/apply.sh /workspace/cache/checker-stops-adapted > /tmp/checker-stops-apply.log 2>&1
git worktree add --detach /workspace/cache/checker-stops-scratch origin/main
# In scratch, with the main cohere module linked:
git merge --no-commit --no-ff origin/codex/stricter-options-next > /tmp/checker-stops-merge.log 2>&1
go build -buildvcs=false -o /workspace/cache/checker-stops-topic ./cmd/adamic > /tmp/checker-stops-topic-build.log 2>&1
# From main checkout, before adding project dependencies:
/workspace/cache/step24-adamic build /workspace/cache/checker-stops-adapted/parser-proof-main.a -o /workspace/cache/checker-stops-main-parser > /tmp/checker-stops-main-parser.log 2>&1
# From scratch:
/workspace/cache/checker-stops-topic build /workspace/cache/checker-stops-adapted/parser-proof-main.a -o /workspace/cache/checker-stops-topic-parser > /tmp/checker-stops-topic-parser.log 2>&1
# After providing project dependencies, original project entry on each compiler:
/workspace/cache/step24-adamic build /workspace/cache/checker-stops-adapted/src/compiler/parser.ts -o /workspace/cache/checker-stops-project-main-parser > /tmp/checker-stops-project-main-parser.log 2>&1
/workspace/cache/checker-stops-topic build /workspace/cache/checker-stops-adapted/src/compiler/parser.ts -o /workspace/cache/checker-stops-project-parser > /tmp/checker-stops-project-parser-with-types.log 2>&1
TYPESCRIPT_API=/workspace/cache/tsc-census/npm/node_modules/typescript/lib/typescript.js node stage3/scouts/step24/checker-stops/inventory.cjs /workspace/cache/checker-stops-adapted > /tmp/checker-stops-inventory-final.log 2>&1
MAIN_ADAMIC=/workspace/cache/step24-adamic TOPIC_ADAMIC=/workspace/cache/checker-stops-topic python3 stage3/scouts/step24/checker-stops/check-fixtures.py > /tmp/checker-stops-fixtures-final.log 2>&1
python3 stage3/scouts/step24/checker-stops/audit.py > /tmp/checker-stops-audit.log 2>&1
python3 stage3/scouts/step24/checker-stops/report.py > /tmp/checker-stops-report.log 2>&1
```

Initial scratch builds failed because submodule files were absent, then because VCS stamping could not inspect the symlinked module. Linking the existing module and using `-buildvcs=false` resolved those setup failures. An early inventory overlapped adaptation and had unstable coordinates; it was discarded and regenerated only after apply finished. A project-entry run without Node declarations reported 129 located errors plus TS2688; that run is retained separately and is not the final comparison.

# Fixtures and mutants

Each of the 14 groups has a `.a` fixture with an actual main diagnostic a-check header, a fixed Node golden observation, and a one-edit semantic mutant in [mutants.json](mutants.json). [fixtures.json](fixtures.json) records each Node result, mutant exit/stdout/stderr, and both standalone compiler outcomes. These standalone fixtures intentionally retain stricter Adamic ownership; they are not project-admission fixtures. All remain checker-rejected by both standalone compilers. A header checks the first captured diagnostic; it does not enumerate every secondary error.

| Group | Mutant | Catcher |
|---|---|---|
| nullable-argument | index 0 -> 1 | golden stdout changes |
| nullable-result-relation | index 0 -> 1 | golden stdout changes |
| nullable-dereference | index 0 -> 1 | Node fails rather than printing ok |
| nullable-iteration | tuple index 0 -> 1 | Node fails rather than printing ok1 |
| optional-field-presence | `{ read }` -> `{}` | present -> absent |
| caught-unknown | Error message ok -> mutant | golden stdout changes |
| host-declarations | host -> mutant marker | golden stdout changes |
| host-members | stackTraceLimit 7 -> 8 | golden stdout changes |
| host-inference | callback profile ok -> mutant | golden stdout changes |
| library-set-shape | backing set [1] -> [1, 2] | size 1 -> 2 |
| nullable-index-key | key index 0 -> 1 | ok -> undefined |
| optional-call | function index 0 -> 1 | Node fails rather than printing ok |
| overload-use | argument index 0 -> 1 | ok -> undefined |
| spread-tuple | tuple index 0 -> 1 | Node fails rather than printing ok1 |

Inventory mutants drop a diagnostic, duplicate a row, move a coordinate or change its code; independent parsing of the pinned log rejects all four. Three further comparison mutants corrupt the clearance total, falsely mark the loader-blocked driver green, or change a site status; independently parsed measurement logs reject each. [counts.md](counts.md) is refreshed for these local fixtures; no internal/oracle discovery fixtures were added. No mutant establishes emitted native guards here: reaching native emission is still blocked. The next compiler work must add guard-erasure witnesses for admitted project sites before claiming runtime correctness.
