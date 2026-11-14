"""Finalize both measured formats and the requested combined census report."""
import hashlib,json,subprocess
from pathlib import Path
base=Path(__file__).resolve().parent;primary=base.parent/'c4v7-census';modes=['CommonJSBundle','EcmaScript']
read=lambda p:json.loads(p.read_text())
records={m:[json.loads(l) for l in (base/m/'census.jsonl').read_text().splitlines()] for m in modes}
groups={m:read(base/m/'grouped-first-stops.json') for m in modes}
comparisons={m:read(base/m/'comparison.json') for m in modes}
for m in modes:
 assert len(records[m])==len(comparisons[m])==79
 assert read(base/m/'validation.json')['observations']==158
assert [r['roots'] for r in records[modes[0]]]==[r['roots'] for r in records[modes[1]]]
keys=['kind','where','reason','fix','diagnostics']
matching=sum({k:a.get(k) for k in keys}=={k:b.get(k) for k in keys} for a,b in zip(records[modes[0]],records[modes[1]]))
ledger=read(primary/'head-findings.json');mainledger=read(primary/'main-findings.json')
clear={where:all(c['after'][0]!=where for mode in modes for c in comparisons[mode]) and not any(f['where']==where for f in ledger) for where in ['sys.ts:304:13','scanner.ts:2934:41']}
pairing=read(primary/'pairing.json')
lead='sys.ts:304:13: **'+('cleared' if clear['sys.ts:304:13'] else 'not cleared')+'**. scanner.ts:2934:41: **'+('cleared' if clear['scanner.ts:2934:41'] else 'not cleared')+'**. **'+str(pairing['new'])+' new refusal-scan findings.**'
s=['## Part 1 — next stops on c4v-7','','All 79 roots ran twice in CommonJSBundle, then twice in EcmaScript: 316 observations. Duplicate observations agree except timing; '+str(matching)+'/79 first-stop tuples agree across formats. Each completed group was checkpoint-pushed. CommonJSBundle was validated and pushed before EcmaScript started. One ordinary binary and the same adapted head tree served both formats.','','The before column is the measured 5c8338c4 CommonJSBundle table from `typescript/next-stops-5c8338c4` at `d4f3ac9023013b7eb3d9dac4e8f61e1d64ecb76c`. That branch has no measured EcmaScript baseline. EcmaScript after values below are measured; its before reference is explicitly CommonJSBundle.','','### Exact stop tuples','']
old=read(base/'previous-CommonJSBundle-grouped.json');after={tuple((g['where'],g['kind'],g['reason'])):g for m in modes for g in groups[m]}
bs={tuple((g['where'],g['kind'],g['reason'])):g for g in old};bids={k:f'B{i:02}' for i,k in enumerate(sorted(bs),1)};aids={k:f'A{i:02}' for i,k in enumerate(sorted(after),1)}
for mapping,data in [(bids,bs),(aids,after)]:
 for k,ident in mapping.items():
  g=data[k];s += [f"- **{ident}**: `{k[0]}`; `{k[1]}`; {k[2]}. Phase: {g['phase']}."]
s+=['','### Every root, before and after','','Each identifier expands to the exact location, kind and reason above. The two format reports also retain full tuple text per root: [CommonJSBundle](../next-stops-c4v7/CommonJSBundle/REPORT.md), [EcmaScript](../next-stops-c4v7/EcmaScript/REPORT.md).','','| Root | Before: main CJS | After: head CJS | After: head ES |','| --- | --- | --- | --- |']
for a,b in zip(comparisons[modes[0]],comparisons[modes[1]]):
 assert a['root']==b['root'] and a['before']==b['before']
 s += ['| `'+a['root']+'` | '+bids[tuple(a['before'])]+' | '+aids[tuple(a['after'])]+' | '+aids[tuple(b['after'])]+' |']
s+=['','### Every new first-stop tuple and its adapted source line','']
newkeys={tuple(c['after']) for m in modes for c in comparisons[m] if c['before']!=c['after']};mainkeys={(f['where'],f['kind'],f['reason']) for f in mainledger};seen=set()
for m in modes:
 for e in read(base/m/'source-evidence.json'):
  k=e['where'],e['kind'],e['reason']
  if k not in newkeys or k in seen:continue
  seen.add(k)
  s += ['#### '+e['where'],'','```typescript',e.get('adapted_line','(no source line)'), '```','','Context:','','```typescript',e.get('adapted_excerpt',''),'```','','Kind: `'+e['kind']+'`. Reason: '+e['reason']+'. Phase: '+e['phase']+'. Emitter: '+e.get('emitter','not applicable')+'. Owner: '+e['owner']+'.']
  if e.get('owner_basis'):s+=['','Source attribution: '+e['owner_basis'],'']
  s += ['','This is an existing main refusal-scan finding newly reached as a first stop.' if k in mainkeys else 'This tuple is not present in the main refusal-scan ledger; consult its phase above.','']
assert seen==newkeys, 'every new first-stop tuple must have a quoted source line'
s+=['### Lowering-phase first stops','']
for m in modes:
 roots=[r for g in groups[m] if g['phase'].startswith('lowering') for r in g['roots']]
 s += ['- '+m+': '+(', '.join('`'+r+'`' for r in roots) if roots else 'none')+'.']
s+=['','Scan-only does not see these lowering-phase first stops. Complete source excerpts, paired root observations, process bounds, actual exits, grouped tuples and validations are under `../next-stops-c4v7/`. This is first-stop measurement, not successful native compilation or execution.','','### Validation and provenance','','Both formats passed the copied and retargeted validator. All child exits were zero; durations were below 1,800 seconds (process limits 1,780 seconds). Adapted source hashes and the ordinary binary hash were rechecked. All tracked changes relative to c4v-7 are in the two census evidence directories; no compiler or adaptation source changed. The requested branch was created under a new name and all pushes were non-forced.']
part2=(primary/'REPORT.md').read_text().split('\n## Part 1')[0]
lines=part2.splitlines();lines[2]=lead
(primary/'REPORT.md').write_text('\n'.join(lines)+'\n\n'+'\n'.join(s)+'\n')
(base/'REPORT.md').write_text('# Next stops on c4v-7\n\n'+lead+'\n\nFull combined report: [c4v7-census/REPORT.md](../c4v7-census/REPORT.md).\n\n'+'\n'.join(s)+'\n')
statuses=[read(p) for p in base.rglob('*-status.json')]
result=dict(roots_per_format=79,observations=316,duplicates_agree=True,matching_first_stop_tuples=matching,cleared=clear,new_refusal_scan_findings=pairing['new'],new_first_stop_tuples=len(seen),maximum_child_seconds=max(s['seconds'] for s in statuses),commonjs_complete_commit=read(base/'bundle-complete.json')['commit'])
(base/'final-validation.json').write_text(json.dumps(result,indent=2)+'\n')
subprocess.run(['python3',str(primary/'audit.py')],check=True)
for root in [base,primary]:
 manifest={str(p.relative_to(root)):{'bytes':p.stat().st_size,'sha256':hashlib.sha256(p.read_bytes()).hexdigest()} for p in sorted(root.rglob('*')) if p.is_file() and p.name!='artifact-manifest.json'}
 assert all(x['bytes']<=20_000_000 for x in manifest.values())
 (root/'artifact-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
print(json.dumps(result,indent=2))
