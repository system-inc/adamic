"""Validate the completed follow-up receipts and independent catch inventories."""
from pathlib import Path
import json,hashlib,subprocess,collections
out=Path(__file__).resolve().parent;s=Path('/workspace/loose-slot-scratch/eleven');repo=out.parents[3]
def read(p):return json.loads(p.read_text())
def rel(w):return w.split('/src/compiler/')[-1]
provenance=read(out/'evidence/provenance.json');assert provenance['head']['sha']=='8843e0ab7b4cfaa547c08e438ca39442a4774fc6';assert provenance['base']['sha']=='89614cc0f9d6aa7b7fe96b56c5c6c1c3eace9825'
assert subprocess.check_output(['git','rev-parse','8843e0ab^'],cwd=repo,text=True).strip()==provenance['base']['sha']
assert provenance['inputs_identical']and provenance['inputs_identical_to_minus7']
for which in ['base','head']:
 tree=s/('adapted-'+which)/'src/compiler';files={str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest()for p in sorted(tree.rglob('*.ts'))};assert files==provenance[which]['files']and len(files)==79
 assert hashlib.sha256((s/(which+'-census')).read_bytes()).hexdigest()==provenance[which]['binary_sha256']
 for i in range(1,80):
  p=out/'evidence'/which/f'root-{i:02}.json';row=read(p);r=read(p.with_suffix('.time.json'));assert r['exit_code']==0 and r['seconds']>0 and r['peak_rss_KiB']>0;assert row['module_format']=='CommonJSBundle'
assert hashlib.sha256((s/'type-count').read_bytes()).hexdigest()==provenance['type_binary_sha256'];assert hashlib.sha256((out/'type_count.go.txt').read_bytes()).hexdigest()==provenance['probe_source_sha256']
typed=read(out/'evidence/type-whole.json');audit=read(out/'evidence/independent-audit.json');comparison=read(out/'evidence/comparison.json')
assert comparison['pairs_complete']==79
assert len(typed['Catches'])==audit['CatchClauses'];assert sum(c['Bound']for c in typed['Catches'])==audit['BoundCatchClauses'];assert sum(c['Captured']for c in typed['Catches'])==audit['CapturedCatchBindings']
assert {rel(c['ClauseWhere'])for c in typed['Catches']}=={rel(c['ClauseWhere'])for c in audit['Catches']}
assert {rel(c['Where'])for c in typed['UnknownCatchCalls']}=={rel(c['Where'])for c in audit['UnknownCalls']}
assert all(c['ParameterType']=='unknown'and c['MakesHEvery']for c in typed['UnknownCatchCalls'])
assert sum(c['Stopped']for c in typed['Catches'])==len({e['Where']for e in typed['Events']if e['Catch']})
old=read(out.parent/'evidence/type-whole.json');a={rel(e['Where'])for e in old['Events']};b={rel(e['Where'])for e in typed['Events']};assert set(comparison['still_sites'])==a&b;assert set(comparison['gone_sites'])==a-b;assert set(comparison['new_sites'])==b-a
for name in ['c04','u00-reconstructed','u00b-reconstructed']:
 p=read(out/'evidence/controls'/(name+'-production.json'));t=read(out/'evidence/controls'/(name+'-type.json'))
 for kind in ['production','type']:
  r=read(out/'evidence/controls'/(name+'-'+kind+'.time.json'));assert r['exit_code']==0 and r['seconds']>0 and r['peak_rss_KiB']>0
 if name=='u00-reconstructed':assert p['kind']=='NotYet'and '#yeep0sx'in p['reason']and any(e['Catch']and e['Where']==p['where']for e in t['Events']);assert t['HEvery']and not t['CaughtEvery']and sum(c['Captured']for c in t['Catches'])==1
 else:assert p['kind']=='accepted'and not t['Events']
 if name=='c04':assert not t['HEvery']and sum(c['Captured']for c in t['Catches'])==1
 if name=='u00b-reconstructed':assert t['HEvery']and not any(c['Captured']for c in t['Catches'])
for name in ['type-whole','independent-audit']:
 r=read(out/'evidence'/(name+'.time.json'));assert r['exit_code']==0 and r['seconds']>0 and r['peak_rss_KiB']>0
original=subprocess.check_output(['git','show','8bf06ebed:stage3/census/loose-slot-census/REPORT.md'],cwd=repo,text=True);report=(out.parent/'REPORT.md').read_text();assert report.endswith(original);assert report.startswith('<!-- HEAD11 START -->');assert 'partial type-level count, not a lowering'in report
assert 'reconstructions'in report and 'byte-identical Language originals'in report
for e in typed['Events']:
 path,line,col=e['Where'].rsplit(':',2);assert rel(e['Where'])in(out/'SITE-PAIRS.md').read_text();assert Path(path).read_text().splitlines()[int(line)-1]in(out/'SITE-PAIRS.md').read_text()
changes=subprocess.check_output(['git','diff','--name-only','8bf06ebed'],cwd=repo,text=True).splitlines();assert all(p.startswith('stage3/census/loose-slot-census/')for p in changes);assert not any(p.endswith('.go')for p in changes)
result={'status':'passed','head':provenance['head']['sha'],'base':provenance['base']['sha'],'ordinary_observations':158,'unchanged_root_pairs':comparison['unchanged'],'different_root_indices':comparison['different_indices'],'catch_clauses':len(typed['Catches']),'bound_catch_clauses':sum(c['Bound']for c in typed['Catches']),'captured_catch_bindings':sum(c['Captured']for c in typed['Catches']),'stopped_catch_bindings':sum(c['Stopped']for c in typed['Catches']),'unknown_parameter_calls':len({c['Where']for c in typed['UnknownCatchCalls']}),'independent_inventory_matches':True,'old22_stand':len(a&b),'old22_gone':len(a-b),'new_sites':len(b-a),'controls':'committed c04 and explicitly reconstructed u00/u00b shapes pass','original_report_preserved':True,'production_source_edits':0}
(out/'evidence/validation.json').write_text(json.dumps(result,indent=2)+'\n');print(result)
