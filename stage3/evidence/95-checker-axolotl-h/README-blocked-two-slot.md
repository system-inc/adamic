# 95h stopped at the services contract boundary

Status: **blocked, not an active adaptation**. The requested
`typescript/axolotl-checker-h` branch remains at 95a
`ef4d26a4a38b7890317e1ba71f8ad928fd8588f6`. It contains no 95h directory.
The rejected candidate, proof runners and observations live only on
`typescript/axolotl-checker-h-evidence`, outside stage3/adapt, so apply does
not discover or execute them. 95c stays unchanged at `7b1e14c56`.

The stop condition is the one in the brief: the widening exposes a diagnostic
outside the original object's own contract. No additional host slot is changed
to make the count pass. No public slot is needed by the attempted repair.

## The complete 54261 literal and the 97b seam

Before editing, stock TypeScript 6.0.3 parsed, bound and typed every one of the
14 property values. Every destination declaration belongs to the @internal
ModuleSpecifierResolutionHost interface in src/compiler/types.ts. The interface
is absent from the pinned public typescript.d.ts baseline. The preflight records
its source declaration locations and internal status, and the baseline's bytes
and SHA-256. The factory itself is file-private.

| Member | May store undefined as the property value? | Owner |
| --- | --- | --- |
| getCommonSourceDirectory | No | unchanged |
| getCurrentDirectory | No | unchanged |
| getSymlinkCache | Yes, maybeBind result | 95h candidate |
| getPackageJsonInfoCache | No, always an arrow; its return can be undefined | unchanged |
| useCaseSensitiveFileNames | No | unchanged |
| redirectTargetsMap | No | unchanged |
| getRedirectFromSourceFile | No, always an arrow; its return can be undefined | unchanged |
| isSourceOfProjectReferenceRedirect | No | unchanged |
| fileExists | No | unchanged |
| getFileIncludeReasons | No | unchanged |
| readFile | Yes, explicit undefined conditional arm | 95h candidate |
| getDefaultResolutionModeForFile | No | unchanged |
| getModeForResolutionAtIndex | No | unchanged |
| getGlobalTypingsCacheLocation | Yes, maybeBind result | accepted 97b I7 |

The accepted peer is **cli/axolotl-97b**, found with git ls-remote '*97b*',
commit **27162cab2bb0054b4ed7cad5a9e258f67f8030c1**, directory
`97-cli-axolotl-b`. Its exact two-file snapshot is under peer-adapters/.
54261 needs its getGlobalTypingsCacheLocation widening. H does not edit that
member, directoryExists, realpath or trace; 97b owns all four.

The final H candidate edits only readFile and getSymlinkCache, each as an
indexed method type plus undefined, preserving method parameter bivariance.
Each has its own ownerBeforeSha256. The factory function has a separate owner
pin. There is **no whole-interface or whole-file acceptance pin**. The full
literal's member order, initializer text, value types, undefined-valued subset
and destination symbols are guarded. Its three undefined-valued slots are
partitioned into two owned by H and one supplied by 97b. All plans are parsed,
JavaScript-compared in both comments modes and reread before one write.

The initial investigation before the 97b seam is archived separately under
initial-three-slot-investigation/. It is historical evidence, not the final
candidate or its measurement baseline.

## Exact diagnostic result with 97b present

All canonical measurements compare **95a + 97b** against
**95a + H candidate + 97b**, with identical dependency-free environments and
unchanged production internal/load options. The real stage3/census/tool is used.
Root and aggregate observations agree and are counted once.

| Profile | Before whole program | After whole program | checker.ts before / after | Exact removed / added |
| --- | ---: | ---: | ---: | ---: |
| src/tsc/tsc.ts | 125 | 124 | 17 / 16 | 1 / 0 |
| src/typescript/typescript.ts, including services/server | 3341 | 3340 | 17 / 16 | 2 / 1 |
| All 711 source files imported by a scratch .a root | 8488 | 8487 | 23 / 22 | 2 / 1 |

The all-source profile deliberately includes library and harness sources, unlike
the normal services/server entry. It is a supplemental stress measurement, not
a claimed normal project baseline. Every source file receives a before/after
own-file row in each profile, including zero rows. The per-file CSV and JSON
reports cover src/services too.

| Own-file attribution | tsc before / after | services/server before / after | all-source before / after |
| --- | ---: | ---: | ---: |
| src/compiler/checker.ts | 17 / 16 | 17 / 16 | 23 / 22 |
| src/compiler/types.ts | 0 / 0 | 0 / 0 | 0 / 0 |
| src/services/utilities.ts | 0 / 0, not reached by tsc entry | 153 / 153 | 153 / 153 |
| All src/services files combined | 0 / 0 | 2755 / 2755 | 2756 / 2756 |

54261's original TS2375 disappears. At services/utilities.ts:2454:5, the
existing TS2375 naming readFile is replaced by one naming
**getModuleSpecifierCache**:

```text
Types of property 'getModuleSpecifierCache' are incompatible.
  Type '(() => ModuleSpecifierCache) | undefined' is not assignable to type '() => ModuleSpecifierCache'.
    Type 'undefined' is not assignable to type '() => ModuleSpecifierCache'.
```

Thus totals fall, but the strict complete-message comparison has one added
message. check-census.py exits **1** on the services/server comparison.
getModuleSpecifierCache is **absent from the original 14-member literal**;
it is filled by the separate services createModuleSpecifierResolutionHost
literal. It is not one of 97b's four owned ModuleSpecifierResolutionHost
members either. Widening it here would violate the requested exact slot scope.
The candidate therefore stops rather than claiming a no-added-diagnostic pass.

## Source identity, reruns, undo and composition

The independent verifier compares **746 src files**. Only types.ts differs
between the 97b control and candidate, by exactly H's two declarations.
CRLF stays **10,681**. Complete types.ts JavaScript is byte-identical:

| Comments | Bytes | SHA-256 before and after |
| --- | ---: | --- |
| Retained | 139659 | `4839bedc439cad06969bad2a3d805a7c0b480f06df73fcff62d2cd0957009fc3` |
| Removed | 121398 | `f041d720cfda4ccb137ffc32198ec3d2fce46dbf13404df737dc637e21290639` |

Both adapters rerun with **0 files, 0 edits** on their combined output.

The undo mutant restores only readFile, leaving getSymlinkCache and all 97b
changes intact. It restores the exact original tsc **125 / 17** and normal
services/server **3341 / 17** diagnostic sets. H's independent verifier and
adapter --check both exit **1**. No mutant enters an oracle or apply tree.

Fresh **apply.sh exits 0**, with H before 97b. Incremental table rows are:

- `95-checker-axolotl-h | 1 | 2 | 2`
- `97-cli-axolotl-b | 3 | 12 | 12`

Both run without refusing the other's edit. The fresh output's complete
746-source map equals the separately measured candidate's map; its fresh
census is again **124 whole-program / 16 checker own-file**, and both adapters'
second runs edit nothing. Mechanical composition does not override the failed
services diagnostic gate.

Every comparison below produces identical bytes across all 746 source files:

- H then 97b versus 97b then H.
- H/B/97b versus B/H/97b, B `408f8cb16c4e97ca51997d5a340a0b1826b66ba9`.
- H/C/97b versus C/H/97b, C `7b1e14c56e54683180aa0cf7819e50ad086486eb`.
- H/D/97b versus D/H/97b, D `c6c2c01e6b0ebfe7fd8b767119c489a1843cdbc5`.
- H/E/97b versus E/H/97b, E `51d4ec4aaf82a8d8e694bc57bf6c41539dd5d630`.
- B/C/D/E/H/97b versus 97b/H/E/D/C/B.

The full stack's tsc-entry census is **110 whole-program / 2 checker own-file**.
The two remaining checker messages are 42849 and 50153. This is not a claim
that its services gate passes; the exposed services member is still outside H.

## Environment and reproduction

Setup passed: Node 0.026s, Go 0.030s, submodules 0.072s, clang 0.186s,
tsc 0.234s, Markdown 0.283s, Go build 15.013s, cache warm 15.120s,
done **15.150s**. `nproc=5`, CPU quota 4. Every build/test shell sources
`/workspace/adamic-tools/env.sh`. All test output is logged. No Go, lane,
fixture, checked-in patch table, 95c or other adaptation source is changed.

Canonical raw observations are the tsc/typescript/all before/after jsonl.gz
files. Each profile's census.json and per-file.csv contain exact differences
and every own-file count. preflight.json holds the complete before-edit audit;
source.json, mutant.json, apply.json and composition.json record the other
proofs. Logs are compressed beside them. candidate/ contains the final two-slot
non-shipping adapter and independent verifier; peer-adapters/ contains accepted
97b verbatim. Its evidence-layout helper import points to the unchanged 47
helper; copying the candidate back into stage3/adapt requires restoring that
relative import to `../47-host-errors/adapt.cjs`.

Principal commands used:

```sh
source /workspace/adamic-tools/env.sh
export NODE_PATH=/workspace/adamic/stage3/api/node_modules
node CANDIDATE/adapt.cjs AFTER97B > /tmp/95h97-after-adapt.log 2>&1
node CANDIDATE/verify.cjs BEFORE97B AFTER97B > /tmp/95h97-source.json 2>&1
/workspace/scratch/95-census BEFORE97B/src/tsc/tsc.ts /workspace/scratch/95h97-tsc-before.jsonl > /tmp/95h97-tsc-before.log 2>&1
/workspace/scratch/95-census AFTER97B/src/tsc/tsc.ts /workspace/scratch/95h97-tsc-after.jsonl > /tmp/95h97-tsc-after.log 2>&1
/workspace/scratch/95-census BEFORE97B/src/typescript/typescript.ts /workspace/scratch/95h97-typescript-before.jsonl > /tmp/95h97-typescript-before.log 2>&1
/workspace/scratch/95-census AFTER97B/src/typescript/typescript.ts /workspace/scratch/95h97-typescript-after.jsonl > /tmp/95h97-typescript-after.log 2>&1
python3 check-census.py /workspace/scratch/95h97-typescript-before.jsonl /workspace/scratch/95h97-typescript-after.jsonl --before-tree BEFORE97B --after-tree AFTER97B > /tmp/95h97-strict-gate.json 2>&1
# Expected strict gate exit: 1, added getModuleSpecifierCache message.
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/95-upstream bash stage3/apply.sh /workspace/scratch/95h97-apply > /tmp/95h97-apply.log 2>&1
```

The all-source .a importer contains one side-effect import per file in
all-source-roots.json, with absolute paths to the respective control or
candidate. It contributes no diagnostic of its own. Root and aggregate census
observations must be normalized and equal.

**Not run after the mandatory stop:** the upstream oracle, native tsc build,
full repository Go gate, or lane acceptance. No 106366/standing-API oracle
claim is made for this blocked H+97b candidate. No services slot outside the
54261 literal is repaired, and no public API sanction is inferred.
