# 95h candidate: the basic node-builder host contract

lane: run as part of a Correctness wave

This is a non-shipping candidate. It is retained only on
`typescript/axolotl-checker-h-evidence` if its required no-new-diagnostic gate
fails. The 95h branch stays on 95a until that gate passes. 95c is unchanged.

Base: 95a `ef4d26a4a38b7890317e1ba71f8ad928fd8588f6`.
Seam: accepted `cli/axolotl-97b`, commit
`27162cab2bb0054b4ed7cad5a9e258f67f8030c1`, directory
`stage3/adapt/97-cli-axolotl-b`. Apply orders 95h before 97b.

## Whole literal audit before editing

`createBasicNodeBuilderModuleSpecifierResolutionHost`, checker.ts:54260,
is file-private. Its entire 14-member return literal was parsed, bound, and
typed before editing. All destination declarations belong to the @internal
ModuleSpecifierResolutionHost interface. That interface is absent from the
pinned typescript.d.ts baseline. No public declaration is edited.

| Member | Property value may be undefined? | Disposition |
| --- | --- | --- |
| getCommonSourceDirectory | No | unchanged |
| getCurrentDirectory | No | unchanged |
| getSymlinkCache | Yes, maybeBind result | 95h |
| getPackageJsonInfoCache | No; always an arrow, its result can be undefined | unchanged |
| useCaseSensitiveFileNames | No | unchanged |
| redirectTargetsMap | No | unchanged |
| getRedirectFromSourceFile | No; function result can be undefined | unchanged |
| isSourceOfProjectReferenceRedirect | No | unchanged |
| fileExists | No | unchanged |
| getFileIncludeReasons | No | unchanged |
| readFile | Yes, explicit undefined conditional arm | 95h |
| getDefaultResolutionModeForFile | No | unchanged |
| getModeForResolutionAtIndex | No | unchanged |
| getGlobalTypingsCacheLocation | Yes, maybeBind result | 97b I7 |

54261 needs the getGlobalTypingsCacheLocation widening too, supplied exclusively
by 97b. 95h does not edit that slot, directoryExists, realpath, or trace.

95h's two own slots use `?: { method(...): R; }["method"] | undefined`,
matching adaptation 32 and 97b while preserving method parameter bivariance.
No new runtime default, assertion, cast or optional chain is added. The
existing literal and every evaluation point stay unchanged.

Stock TypeScript 6.0.3 through 47-host-errors' compiler() verifies bound slot
symbols, declaration/read types, and all 14 initializer types and spellings.
The private factory function has an ownerBeforeSha256. Each of 95h's two
members has its own ownerBeforeSha256; there is no interface or file hash pin.
All output is planned, parsed, compared as JavaScript in both comments modes,
and reread before one write. CRLF and unrelated bytes are preserved without a
printer. The independent verifier reconstructs just the two member edits and
compares all 746 source files. It accepts unrelated 97b member changes in the
control tree because it pins only 95h's own members.

## Required diagnostic gate

The real production census tool measures the tsc entry, the public TypeScript
entry (which includes services and server), and an all-source .a importer of
all 711 source files. Root and aggregate observations must agree and count
once. Complete diagnostic texts, not only locations or totals, are compared.
Per-file before/after counts include every source file and every services file.
The all-source profile deliberately includes upstream library and harness
sources, so its baseline differs from the ordinary services/server entry.

All measurements for this candidate use 97b on both sides: 95a+97b versus
95a+95h+97b. Counts, the exact reader difference, the 97b seam, and any resulting
stop decision are recorded on the separate evidence branch. No logs, censuses
or proof helpers belong in this adaptation directory.

## Final disposition

Blocked with 97b present. The tsc entry is 125 to 124, but the services/server
entry is 3341 to 3340 with a replacement TS2375 at utilities.ts:2454 naming
getModuleSpecifierCache. It is absent from the original 54261 literal. The
strict no-added-message gate rejects the candidate, so no active adapter is
shipped. See the evidence-directory README for the complete counts and seam.
