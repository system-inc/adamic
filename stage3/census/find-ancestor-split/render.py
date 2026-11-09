import json,hashlib,gzip,shutil
from pathlib import Path
out=Path(__file__).resolve().parent
prefix='/workspace/scratch/fa-adapted/src/'
def read(side):
 path=out/((('baseline-'+side) if side in ['parent','head'] else side)+'-entry-run.jsonl')
 with path.open() as stream:
  header=json.loads(next(stream));rows=set();units=set();first={}
  for line in stream:
   row=json.loads(line)
   for unit in row.get('units') or []:units.add(unit['where'])
   for f in row.get('findings') or []:
    rows.add(' | '.join([f['unit'].removeprefix(prefix),f['phase'],f['kind'],f['where'].removeprefix(prefix),f['reason'][:110]]))
    if f['phase']=='lowering':first.setdefault(f['unit'],(f['where'],f['kind'],f['reason']))
 return header,rows,units,first
all_data={side:read(side) for side in ['parent','head','a','b']}
header=all_data['parent'][0]
assert all(d[0]==header for d in all_data.values())
assert not header['checker_rejected'] and not header['diagnostics'] and header['latent_mode']=='first-error'
assert all(d[2]==all_data['parent'][2] for d in all_data.values())
sets={side:d[1] for side,d in all_data.items()}
for side in ['parent','head']:
 assert sets[side]==set((out/('baseline-findings-by-unit-'+side+'.txt')).read_text().splitlines())
for side,s in sets.items():(out/('findings-by-unit-'+side+'.txt')).write_text(''.join(r+'\n' for r in sorted(s)))
parent_only=set((out/'baseline-findings-only-parent.txt').read_text().splitlines());head_only=set((out/'baseline-findings-only-head.txt').read_text().splitlines())
assert parent_only==sets['parent']-sets['head'];assert head_only==sets['head']-sets['parent']
cleared={r for r in parent_only if ' | refusal_scan | ' in r};added={r for r in head_only if ' | refusal_scan | ' in r}
assert len(cleared)==28 and len(added)==5
novel=(sets['a']|sets['b'])-(sets['parent']|sets['head'])
summary={};comparisons={}
for side in ['a','b']:
 for ref in ['parent','head']:
  for direction,rows in [('only-'+ref,sets[ref]-sets[side]),('only-'+side,sets[side]-sets[ref])]:
   (out/('findings-'+side+'-vs-'+ref+'-'+direction+'.txt')).write_text(''.join(r+'\n' for r in sorted(rows)))
  comparisons[side+'-vs-'+ref]=dict(only_reference=len(sets[ref]-sets[side]),only_variant=len(sets[side]-sets[ref]))
 kept_cleared=cleared-sets[side];kept_added=added&sets[side]
 summary[side]=dict(kept_clearings=len(kept_cleared),kept_additions=len(kept_added),kept_clearing_rows=sorted(kept_cleared),kept_addition_rows=sorted(kept_added),novel=sorted(sets[side]-(sets['parent']|sets['head'])))
scan_units={r.split(' | ',1)[0] for r in cleared|added}
scan_first_stop_changes={}
for side in ['head','a','b']:
 scan_first_stop_changes[side]=[dict(unit=u,parent=all_data['parent'][3].get(prefix+u),other=all_data[side][3].get(prefix+u)) for u in sorted(scan_units) if all_data['parent'][3].get(prefix+u)!=all_data[side][3].get(prefix+u)]
(out/'scan-unit-first-stops.json').write_text(json.dumps(scan_first_stop_changes,indent=2)+'\n')
(out/'comparison.json').write_text(json.dumps(dict(headers_equal=True,header=header,units_equal=True,comparisons=comparisons,variants=summary),indent=2)+'\n')
def esc(s):return s.replace('|','&#124;')
def table(rows):
 result=['| Finding: unit &#124; phase &#124; kind &#124; site &#124; reason[:110] | Parent | Head | Variant a | Variant b |','| --- | --- | --- | --- | --- |']
 for row in sorted(rows):result.append('| '+esc(row)+' | '+' | '.join('present' if row in sets[s] else 'absent' for s in ['parent','head','a','b'])+' |')
 return result
lines=['# Split findAncestor refusal-scan census','','Task 448a51k. Report only. Base: `1060360b7b0d15dd6f0c0e8ec7bf0254f9fff000`. Baseline evidence: `c6f6688cc161d695bb79ee740ae153724f964d8d`, directory `stage3/census/find-ancestor-63c11488/`.','','Variant a: `7aabf3739`, without `predicateOverloadUndefinedServes`. Variant b: `52a75455d`, without the tested-result arm in `predicateOverloadReturn`. Full own-pin compiler and checker provenance is in [provenance.json](provenance.json). `7216e359` differs from the baseline head `63c11488` only in a test fixture; no production compiler source differs.','','A scratch checkout of 1060360b carried both complete adaptation directories from eeb87728, verified against the baseline input hashes, and ran `bash stage3/apply.sh /workspace/scratch/fa-adapted` once using TypeScript 6.0.3 source at `050880ce59e30b356b686bd3144efe24f875ebc8`. The complete [source hash comparison](source-hash-check.json) matches the baseline exactly. Both variants used that one adapted tree.','','Each guarded entry overlay was generated from its own variant checkout with `stage3/census/latent/make_overlay.py` followed by `stage3/meter/entry_overlay.py`. Each binary was built with `go build -p=2 -buildvcs=false -overlay=... ./stage3/census/latent/tool`. Overlay evidence uses `.go.txt`. Runs were parallel, with `LATENT_ASSERT_NO_OUTPUT=1`, `GOMEMLIMIT=5GiB`, `GOGC=70`, and a separate `LATENT_TIMES` file; no other LATENT flags were inherited. The entry was `/workspace/scratch/fa-adapted/src/tsc/tsc.ts`. Both no-output guards were exercised. No native output was produced.','','The parent was reused, not rerun. Parent, baseline head, and both variant stream headers are exactly equal: first-error mode, 81 sources, zero checker diagnostics. Unit identities also match. The finding serialization reproduces both baseline lists exactly: `unit | phase | kind | site | reason[:110]`, sorted and unique. See [comparison.json](comparison.json) and the `findings-*-vs-*.txt` lists.','','The main table contains all 28 baseline scan clearings and all 5 baseline scan additions. The supplied only-parent and only-head lists also contain lowering differences; those appear in the second table to retain every supplied row. Any variant finding absent from both baseline sides is appended to the main table.','']
lines+=table(cleared|added|novel)
lines+=['',f'Variant findings present in neither parent nor head: {len(novel)}.', '', 'Among units recording the 33 requested scan rows, first-stop changes against parent: '+', '.join(side+'='+str(len(scan_first_stop_changes[side])) for side in ['head','a','b'])+'. See [scan-unit-first-stops.json](scan-unit-first-stops.json).','','## Other rows from the supplied difference lists','']+table((parent_only|head_only)-(cleared|added))
lines+=['','## Runs','','| Variant | Child exit | Wall seconds | Peak RSS MiB |','| --- | ---: | ---: | ---: |']
for side in ['a','b']:
 status=json.loads((out/(side+'-entry-run.status.json')).read_text());timing=json.loads((out/(side+'-entry-run.time.json')).read_text());assert status['exit']==0 and status['stopped'] is None
 lines.append('| '+side+' | '+str(status['exit'])+' | '+str(status['wall_seconds'])+' | '+str(round(timing['peak_rss_kib']/1024,1))+' |')
lines+=['','Raw streams are `a-entry-run.jsonl.gz` and `b-entry-run.jsonl.gz`. Reused baseline streams are `baseline-parent-entry-run.jsonl.gz` and `baseline-head-entry-run.jsonl.gz`. Exact commands, own child exits, unit progress, overlay bytes, source hashes, and checkpoint receipts are retained. The runner stops a process group after 30 minutes without output or unit progress and cleans up remaining group members after child exit.','','## Per-variant result','']
for side in ['a','b']:
 d=summary[side]
 sentence=f'Variant {side} keeps {d["kept_clearings"]} of the 28 array-widening clearings and {d["kept_additions"]} of the 5 method-bivariance additions.'
 lines.append(sentence);print(sentence,flush=True)
text='\n'.join(lines)+'\n';assert '\u2013' not in text and '\u2014' not in text
(out/'REPORT.md').write_text(text)
