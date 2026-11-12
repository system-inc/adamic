from pathlib import Path
import json,hashlib,subprocess,shutil
repo=Path('/workspace/path-components-adapt');run=Path('/workspace/scratch/path-adapt-final');out=Path('/workspace/path-components-evidence/stage3/census/path-from-components-adapt');final=out/'final';final.mkdir(exist_ok=True)
read=lambda n:json.loads((run/n).read_text())
required=['lane-go','generate-shards','apply','contract','final-after','replay','types-before','types-after','stock-after','build','build-comparison','census-before','census-after','census-mutant']
for label in required:assert read(label+'-status.json')['child_exit_code']==0,label
passes=[];failures=[];packagePass=False
for line in (run/'lane-go.jsonl').read_text().splitlines():
 try:d=json.loads(line)
 except json.JSONDecodeError:continue
 if d.get('Action')=='pass' and d.get('Test','').startswith('TestShard_'):passes.append(d['Test'])
 if d.get('Action')=='fail':failures.append(d)
 if d.get('Action')=='pass' and not d.get('Test') and d.get('Package','').endswith('/stage3/lane'):packagePass=True
assert len(set(passes))==398 and len(passes)==398 and not failures and packagePass,(len(passes),failures,packagePass)
summary={'shardPasses':398,'failures':[],'packagePassed':True,'command':read('lane-go-status.json')['command'],'seconds':read('lane-go-status.json')['seconds']}
(run/'lane-summary.json').write_text(json.dumps(summary,indent=2)+'\n')
meta=read('provenance.json');commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip();parent=subprocess.check_output(['git','rev-parse','HEAD^'],cwd=repo,text=True).strip();assert parent==meta['main'],(parent,meta['main'])
for name,h in meta['adaptationInputs'].items():assert hashlib.sha256(subprocess.check_output(['git','show','HEAD:'+name],cwd=repo)).hexdigest()==h,name
meta['adaptationCommit']=commit;meta['adaptationParent']=parent;meta['validationInputsMatchCommit']=True;(run/'provenance.json').write_text(json.dumps(meta,indent=2)+'\n')
for p in run.iterdir():
 if p.is_file() and p.name not in ['adamic','latent']:shutil.copyfile(p,final/p.name)
comparison=read('build-comparison.json');delta=read('census-delta.json');replay=read('replay.json')
assert replay['result']['edits']==0 and replay['writes']==0 and replay['bytesUnchanged'] and replay['mtimeUnchanged']
assert len(comparison['rows'])==12 and all(r['identical'] for r in comparison['rows'])
assert not read('stock-after.json')['diagnostics'] and len(read('stock-after.json')['roots'])==79
assert len(delta['removedUniqueFindings'])==6 and not delta['addedUniqueFindings']
text=f'''# getPathFromPathComponents adaptation: final proof

**All required gates pass.** Adaptation commit **`{commit}`** is the single commit on main **`{meta['main']}`**, branch `typescript/path-from-components-adapt`. It contains only the new `stage3/adapt/99-correctness-path-from-components` directory and the upstream ledger entry. Measurement output is on `typescript/path-from-components-evidence`, outside stage3/adapt.

## Corrected replay gate and pre-existing refusal

The earlier stop was superseded by the user's correction. Adapter 30's `declined site changed: core.ts:992:21` refusal reproduces on unmodified main when 30 is replayed over the **fully adapted** tree. **That is not apply's path: apply runs each adapter on its predecessor's output, before later adaptations add `from[i]!` in addRange.** This historical refusal is not attributed to the new path owner and does not require an amendment to 30 for this change.

The isolated `30.plan(ts, "path.ts", text, true)` accepts both baseline and new owner text unchanged: zero edits and four declined reads. The valid replay proofs are the successful **complete stage-by-stage apply.sh run**, and this adaptation's **zero-edit, zero-write replay over its own output**, with bytes and mtime unchanged. No existing adapter or declined-site guard was changed. Historical observations are retained in [30-baseline.log](30-baseline.log), [30-path-only.json](30-path-only.json), and the superseded [HISTORICAL-STOP.md](HISTORICAL-STOP.md). The interrupted old Go run is historical; the final run below completed.

## Final verification on main's tip

| Gate | Result |
|---|---|
| Complete apply.sh, stage by stage | child 0, {read('apply-status.json')['seconds']:.2f} seconds |
| New adaptation's owner/drift contract | child 0; ten no-write refusals, unrelated-owner and LF/CRLF checks |
| New adaptation replayed on its output | child 0; files 0, edits 0, writes 0; bytes and mtime unchanged |
| Whole `go test` stage3/lane package | **child 0; all 398 shards pass; zero failures; package PASS** |
| `generate_shards.py --check` | child 0 |
| Adamic types over sorted 79 roots | child 0 before and after |
| Stock TypeScript 6.0.3 compiler project | 79 roots, zero diagnostics |
| Actual upstream build | child 0 |
| Ten emitted JavaScript outputs | all byte-identical to baseline |
| Both public declaration files | byte-identical to baseline |
| Full-mode, single-file path.ts census | child 0 before and after; no checker diagnostics; output guards enabled |
| Restoring mutant | census child 0; original cast refusal restored at path.ts:513:45 |

The completed whole Go command is unfiltered:

```sh
go test -json -count=1 -timeout=28m -parallel=4 ./stage3/lane
```

It completed in **{summary['seconds']:.2f} seconds**, including all fixture tests and all **398 unique TestShard_ passes**. [final/lane-summary.json](final/lane-summary.json), [final/lane-go-status.json](final/lane-go-status.json) and [final/lane-go.jsonl](final/lane-go.jsonl) retain the count, exact exit and full stream. This was run to completion; it is not the historical interrupted 28-shard observation.

[final/provenance.json](final/provenance.json) pins main, upstream TypeScript (`050880ce59e30b356b686bd3144efe24f875ebc8`), tools, binary hashes, apply product key and every adaptation/ledger input hash. Those hashes were checked against the final commit's Git objects, so the tested working-tree bytes equal the committed files. The lane's shared product contains the new overloads. The earlier baseline adapted tree is reusable because the intervening main changes only affect Adamic lowering, not stage3's source/adaptation inputs; the final main binaries reran its types and census checks.

## Adapter contents

The new adapter changes only getPathFromPathComponents: two `@internal` overloads for `readonly Path[] -> Path` and `readonly string[] -> string`, and a string implementation with no casts. The overloads precede the existing `//// Path Formatting` comment and JSDoc, preserving emitted comment bytes. `owners.json` pins the complete before/after owner region, including its comments and overload set, with exact text and SHA-256.

Planning parses the whole file and verifies all preconditions and the complete after owner before the sole write. Owner drift, partial edits, missing/moved/duplicate owners and parse errors are refused without writing. Replay makes zero edits and writes, preserving bytes and mtime. Unrelated owners and CRLF are preserved. Full path.ts transpile equality is also checked before writing. See the committed adaptation README and [final/contract.json.log](final/contract.json.log), [final/replay.json](final/replay.json).

The upstream ledger entry records this type-only change and the evidence branch. No compiler, oracle, shard, public API sanction, other adapter or forEachAncestorDirectory owner was changed. [owner-pins-grep.txt](owner-pins-grep.txt) retains the complete earlier pin search: 30's declined root read is the direct target-owner pin; other name hits are copied import blocks.

## forEachAncestorDirectory remains unchanged

Its existing overloads resolve five source calls as **string 3, Path 2**. Implementation-only P is not a caller-visible type parameter: one actual generic P argument resolves to the string overload. The stock checker evidence, complete call positions and signatures remain in [p-instantiations.json](p-instantiations.json).

The same-shape candidate retaining the Path/string overloads but using a string directory/string callback implementation and removing `as P` produces **TS2394**, “This overload signature is not compatible with its implementation signature.” Its JavaScript is identical, but the stock zero-diagnostic gate fails. The README records the count and this reason for exclusion. No callback-union workaround was introduced. See [p-candidate-stock.json](p-candidate-stock.json) and [p-candidate-path.ts.txt](p-candidate-path.ts.txt).

## Exact source and output evidence

The established getPathFromPathComponents source instantiations are **string 13, Path 2**, including five harness alias calls. The target is absent from the upstream public API baseline and sanctioned-api inventory; those complete measurements remain on [typescript/path-from-components](https://github.com/system-inc/adamic/tree/typescript/path-from-components/stage3/census/path-from-components).

The final actual build compared every one of these ten emitted JavaScript files and both public declarations byte for byte, retaining sizes and matching hashes in [final/build-comparison.json](final/build-comparison.json):

'''
for row in comparison['rows']:text+=f"- `{row['file']}`: {row['bytes']} bytes, SHA-256 `{row['beforeSha256']}`.\n"
text+='''
Source maps and internal declarations are outside that byte-identity claim. The comparison uses raw bytes, not normalized tokens or removed comments. [final/stock-after.json](final/stock-after.json) records the exact 79 project roots and empty diagnostic list; the earlier baseline stock project has zero diagnostics in [p-candidate-stock.json](p-candidate-stock.json). Adamic's explicit sorted root lists, own exits and timings are in [final/types-before-status.json](final/types-before-status.json) and [final/types-after-status.json](final/types-after-status.json).

## Exact census delta and restoring mutant

Both final main-tip censuses used `LATENT_FULL=1`, exact `LATENT_ONLY_FILE` selectors, and `LATENT_ASSERT_NO_OUTPUT=1`, loading the complete compiler directory. Both header diagnostic lists are empty (`checker_rejected: false`); the recorder's fixed historical measurement label does not override those headers.

After normalizing the two inserted lines, the unique `(kind, source location, exact reason)` comparison removes these six findings and adds **none**:

| Baseline path.ts site | Kind | Exact reason removed |
|---|---|---|
'''
for kind,where,reason in delta['removedUniqueFindings']:text+=f"| `{where.split('path.ts:')[-1]}` | {kind} | `{reason}` |\n"
text+='''
These are four distinct source positions: the two cast expressions and their two statement boundaries. No other diagnostic site changes after line-shift normalization. `forEachAncestorDirectory`'s `getDirectoryPath(directory) as P` refusal remains, shifting from original 1103:28 to after 1105:28. The overloads and new implementation have zero findings. See [final/census-before.jsonl](final/census-before.jsonl), [final/census-after.jsonl](final/census-after.jsonl), [final/census-delta.json](final/census-delta.json).

The restoring scratch mutant replaces the after owner with its original generic before owner, including `"" as T`; it restores the exact original line numbering and the refusal **path.ts:513:45**, reason `a cast the runtime can't check`. The final main binary remeasured this in [final/census-mutant.jsonl](final/census-mutant.jsonl) and [final/mutant-result.json](final/mutant-result.json). The historical expected-red detector also exits 1 in [restoring-mutant-check-status.json](restoring-mutant-check-status.json). No real adaptation source was mutated.

## Bounds and reproduction

Every final process is supervised by [final/run.py](final/run.py): it reports its child's own exit, bounds it at 29 minutes, and reaps its process group. The whole Go test also has a 28-minute internal bound. No process ran for 30 minutes and no final bound was reached. Generated physical Go evidence files use `.go.txt`; [final/build-overlay.py](final/build-overlay.py) applies that naming before generation. Binaries are excluded from evidence; their SHA-256 values are retained in provenance.

Reproduce using the pinned upstream checkout, locked stage3 parser/build dependencies and the committed adapter. Run apply.sh into a new output, then the saved final drivers (adjust their recorded absolute paths together when relocating). Keep every output outside stage3/adapt. The adapter README links this evidence branch. All final status files report child 0; the historical stop and failed ancestor candidate remain explicitly identified as history and rejection evidence.
'''
(out/'REPORT.md').write_text(text)
manifest={str(p.relative_to(out)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(out.rglob('*')) if p.is_file() and p.name!='SHA256SUMS.json'}
(out/'SHA256SUMS.json').write_text(json.dumps(manifest,indent=2)+'\n')
print('final report written; all gates and commit identity verified')
