import json
from pathlib import Path
out=Path(__file__).resolve().parent;r=json.loads((out/'verification.json').read_text());lane=r['lane'];js=r['javascript']
text=f'''# Changed-file undefined-slot amendment

Feature: `{r['feature']}` on `{r['base']}`.
Evidence branch: `typescript/changed-files-undefined-evidence`.
TypeScript: v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`.

The sole source amendment is the timer adapter's whole-owner after pin:

```ts
const changedFilesInLastPoll: (WatchedFileWithUnchangedPolls | undefined)[] = [];
```

Its upstream before pin and all other owners are unchanged. The implementation
commit contains only owners.json and its README amendment. No ledger line was
added: the owner README gives no type-only ledger convention. No Loom submission
was made.

| Proof | Result |
|---|---|
| Full sys.ts census, CommonJSBundle | Main 191 events, amendment 190; one finding gone, zero new or changed |
| Unique Refused/NotYet sites | 64 → 63; all 63 remaining sites unchanged |
| Unit status | All 136 units unchanged |
| Restoring mutant | 191 events; every normalized observation exactly matches main |
| Emitted sys.js | {js['bytes']:,} bytes in both trees; byte-identical |
| adamic types | Exit 0 on all 79 roots, main and amendment |
| Whole-owner contract | Exit 0; 13 owners, 9 files, LF/CRLF, drift/half/duplicate/moved guards and zero-write replay |
| Actual amended-owner replay | Zero edits and zero writes on measured after-tree |
| Existing source-extracting witnesses | Main/after JSON byte-identical |
| stage3/apply.sh | Main and amendment exit 0; all 747 fresh amendment source files match measured after-tree |
| Lane product preparation | Exit 0 |
| Whole stage3/lane package | Exit 0; {lane['passed_tests']} passing Go tests, all 398 manifest shards; {lane['elapsed']} s |
| Lane generated files | API/runtime hashes unchanged; all three files left unchanged |

## Every finding delta

Exactly one finding is gone:

- **Refused, src/compiler/sys.ts:290:19**, unit `sys.ts:288:5`, phase
  `refusal_scan`: “a value of type WatchedFileWithUnchangedPolls[] seen as
  (WatchedFileWithUnchangedPolls | undefined)[], which can write
  WatchedFileWithUnchangedPolls | undefined where WatchedFileWithUnchangedPolls
  is read”.

New findings: **none**. Changed findings: **none**. No boundary event changed.
The raw observation delta is also exactly that one removed finding.
`census-delta.json` retains its complete reason/text and every observation delta.
Only tree prefixes and the exact 14-byte sys.ts span relocation are normalized.
The authoritative input preserves main's 1,975 CRLF line endings.

The identifier audit finds exactly seven occurrences: the declaration,
pollQueue argument, two length reads, two identity comparisons, and push.
There is no new reader requiring narrowing. No guard or runtime edit was added.

## Restoring mutant and JavaScript

`sys-restoring-mutant.ts.txt.gz` is the named restoring mutant source.
It is constructed from the amendment by restoring only the old declaration and
is byte-identical to main sys.ts. Its independent complete census restores
290:19 and matches all main observations after path normalization.

The common emitted sys.js SHA-256 is
`{js['sha256']}`.
The full emitted files are retained. The emission options are ES2024, ESNext,
sourceMap false, using stock TypeScript 6.0.3; transpile diagnostics are empty.

## Evidence and limits

`COMMANDS.md`, the retained scripts and per-run receipts identify exact commands,
flags, source hashes, every compiler root, binaries, tools and exits. All three
census headers have empty diagnostics, checker_rejected false, full mode and
CommonJSBundle; each has a relation_report trailer. The ordinary Load/Lower
no-output assertions and the independent census audit pass. The lane uses its unchanged expectations, including the existing sanctioned
Public APIs baseline exception. The census retains
its historical measurement label in raw records. The 63 remaining existing
Refused/NotYet sites are retained in full.

The initial amended snapshot accidentally normalized CRLF to LF. That preview
was terminated (exit -15), excluded, and replaced by the exact byte-preserving
snapshot. Its partial stream and receipts are retained under excluded-preview.
The corrected types check, witnesses and JavaScript emission were rerun against
the authoritative snapshot. SHA256.json covers all retained evidence files.
'''
(out/'REPORT.md').write_text(text)
