# 95h: truthful internal module-specifier host slots

lane: run as part of a Correctness wave

Base: 95a `ef4d26a4a38b7890317e1ba71f8ad928fd8588f6`.
Only this adaptation directory belongs to H. 95c stays at `7b1e14c56`.
The earlier blocked two-slot attempt is historical evidence at
`a6ce2a25d`; the extended contract below supersedes that attempt.

## Scope, internal audit and 97b ownership

H widens exactly four optional members of the **@internal**
`ModuleSpecifierResolutionHost` interface in src/compiler/types.ts:
`getSymlinkCache`, `readFile`, `getModuleSpecifierCache`, and
`getNearestAncestorDirectoryWithPackageJson`. Each receives an indexed method
type plus undefined, `x?: { x(...): R; }["x"] | undefined`, preserving method
parameter bivariance. All four declarations were resolved to that internal
interface before editing. The interface is absent from the pinned public
`typescript.d.ts` baseline. No public API line is edited.

Accepted 97b is **cli/axolotl-97b**, commit
`27162cab2bb0054b4ed7cad5a9e258f67f8030c1`, directory
`97-cli-axolotl-b`. It owns `directoryExists`, `realpath`,
`getGlobalTypingsCacheLocation`, and `trace`; H never edits them.
54261 requires its global-cache widening too. Apply orders H before 97b.
All canonical census comparisons have 97b on both sides.

Five contextually typed or derived object literals were inventoried across
src/compiler and src/services. The whole return literal at checker.ts:54261
has 14 members; exactly getSymlinkCache, readFile and
getGlobalTypingsCacheLocation may hold undefined as property values.
getPackageJsonInfoCache is always an arrow whose *return* may be undefined.

The entire services createModuleSpecifierResolutionHost literal at
utilities.ts:2454 has 16 members:

| Member | Property value may be undefined? | Owner |
| --- | --- | --- |
| fileExists | No | unchanged |
| getCurrentDirectory | No | unchanged |
| readFile | Yes | H |
| useCaseSensitiveFileNames | No, guarded fallback | unchanged |
| getSymlinkCache | Yes | H |
| getModuleSpecifierCache | Yes | H |
| getPackageJsonInfoCache | No, always an arrow | unchanged |
| getGlobalTypingsCacheLocation | Yes | 97b |
| redirectTargetsMap | No | unchanged |
| getRedirectFromSourceFile | No, always an arrow | unchanged |
| isSourceOfProjectReferenceRedirect | No | unchanged |
| getNearestAncestorDirectoryWithPackageJson | Yes | H |
| getFileIncludeReasons | No | unchanged |
| getCommonSourceDirectory | No | unchanged |
| getDefaultResolutionModeForFile | No | unchanged |
| getModeForResolutionAtIndex | No | unchanged |

The other literals are createProgram, its getEmitHost helper, and the services
getModuleSpecifierResolverHost forwarding spread. Their ModuleSpecifierResolutionHost
undefined stores are owned by 97b, or forward the audited services producer.
No further host member requires widening. Other existing Program/EmitHost
members are inventoried as different destinations and are not changed by H.

## Adapter and independent verifier

Stock TypeScript 6.0.3 through 47-host-errors' compiler() parses and binds the
compiler plus services. The adapter inventories every relevant contextual
literal, all its members, its undefined-valued subset and bound destinations,
and the known forwarding spread. All five writer functions have individual
owner text pins. Each H member has its own ownerBeforeSha256. There is **no
whole-interface or whole-file acceptance pin**, so 97b can modify its own
members independently. Own slot symbols, optionality and unchanged read types
are checked; unrecognized writer, member or owner shapes are refused.

All edits are source slices, planned and parsed before one types.ts write.
Inputs are reread and complete stock JavaScript is compared in both comments
modes before writing. CRLF is preserved without a printer. No assertion, cast,
optional chain, default or runtime change is added. --check performs no writes.
A second run reports files 0, declarations 0; 97b also reruns with zero edits.

verify.cjs is independent of the planner. It reconstructs only H's four parsed
member edits, compares all 746 src files, checks CRLF and complete JavaScript,
and leaves unrelated 97b edits in the control intact. Exactly one file changes;
CRLF stays 10,681. types.ts JavaScript is byte-identical:

| Comments | Bytes | SHA-256 before and after |
| --- | ---: | --- |
| Retained | 139659 | `4839bedc439cad06969bad2a3d805a7c0b480f06df73fcff62d2cd0957009fc3` |
| Removed | 121398 | `f041d720cfda4ccb137ffc32198ec3d2fce46dbf13404df737dc637e21290639` |

## Exact production census

The unchanged real stage3/census/tool uses production internal/load options.
Each root and identical aggregate observation must agree and count once,
following 95a's compare-census.py. Full diagnostic texts are compared.
Controls are 95a+97b; candidates are 95a+H+97b, before npm installs.

| Profile | Whole program before / after | checker.ts own-file before / after | Removed / added |
| --- | ---: | ---: | ---: |
| tsc entry | 125 / 124 | 17 / 16 | 1 / 0 |
| services/server entry | 3341 / 3339 | 17 / 16 | 2 / 0 |
| all 711 sources | 8488 / 8486 | 23 / 22 | 2 / 0 |

Exactly checker.ts:54261 TS2375 and services/utilities.ts:2454 TS2375 disappear
in profiles that reach both. utilities.ts goes **153 to 152**, not sideways.
No message appears or changes elsewhere. Per-file CSV/JSON tables include
every source file, including zero rows and all src/services files. The all-source
.a importer deliberately includes upstream libraries and harness sources, so
its baseline differs from the normal services/server entry.

The undo mutant restores only readFile, leaving the other three H members and
all 97b edits intact. The source verifier and adapter check reject it, and the
census restores the original checker and services readFile incompatibilities.
No mutant enters the oracle tree. Complete results, exits and raw observations
are retained on the evidence branch.

## Evidence and reproduction

Only adapt.cjs, verify.cjs, sites.json and this README live here, since every
file in this directory feeds apply's input key. Logs, raw censuses, oracle
reports, undo controls, public API checks and composition results live on
[typescript/axolotl-checker-h-evidence](https://github.com/system-inc/adamic/tree/typescript/axolotl-checker-h-evidence/stage3/evidence/95-checker-axolotl-h).
The evidence branch's current report identifies the accepted four-slot run;
Git commit a6ce2a25d preserves the superseded stop report.

Fresh apply is run in a scratch worktree with this directory and the exact
accepted 97b directory added, in numeric order. Canonical commands, with all
test output logged:

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH=/workspace/adamic/stage3/api/node_modules
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/95-upstream
bash stage3/apply.sh /workspace/scratch/95hx-apply > /tmp/95hx-apply.log 2>&1
node stage3/adapt/95-checker-axolotl-h/verify.cjs /workspace/scratch/95h97-before /workspace/scratch/95hx-apply > /tmp/95hx-apply-source.json 2>&1
/workspace/scratch/95-census /workspace/scratch/95hx-apply/src/tsc/tsc.ts /workspace/scratch/95hx-apply-tsc.jsonl > /tmp/95hx-apply-tsc.log 2>&1
timeout --kill-after=20s 880 bash stage3/oracle/run.sh /workspace/scratch/95hx-apply /workspace/scratch/95hx-oracle --limit-seconds 720 > /tmp/95hx-oracle.log 2>&1
```

The bounded all-runner oracle completes in **352.424s**, with **106,366
passing, 1 failing, 0 pending**. Install exits 0 (2.422s), build exits 0
(22.992s), tests exit 1 (326.909s), solely for the standing Public APIs
acknowledgement case. There are no timeouts or exceptions. Built/local API
bytes are exactly main's 590,996 bytes, SHA-256
`b901adcc6ec29e299ec21401b2a155b2b2c14c046ec743fed6ee45679b726d5d`;
the accepted reference remains main's 588,626 bytes, SHA-256
`f87d6c30da287293aae741711fa94594b75d49e0977cea6bfd5982f90e2bafcf`.
There is no new public API delta. No lane acceptance is inferred. Fresh apply
exits 0 with rows `95-checker-axolotl-h | 1 | 4 | 4` and
`97-cli-axolotl-b | 3 | 12 | 12`; its exact source map equals the oracle
input. The combined B/C/D/E/H/97b tsc census is 110 whole-program and
2 checker own-file diagnostics. Both orders are checked with 97b and with B, C, D and E,
including the full forward/reverse stack; source maps and exact refs are retained.

Setup passed in 12.003s: Node 0.022s, Go 0.023s, submodules 0.071s, clang
0.157s, tsc 0.162s, Markdown 0.204s, Go build 11.863s, cache warm
11.973s. nproc=5, CPU quota 4. Every build/test shell sources
/workspace/adamic-tools/env.sh. No Go, lane, fixture, meter, checked-in patch
table or other adaptation source is changed by the shipping branch.

Not covered: native tsc compilation/lowering, browser or ESLint integration,
the whole repository Go gate, or lane acceptance. None is claimed by this head.
