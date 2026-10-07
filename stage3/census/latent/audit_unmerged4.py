"""Independently recount unique lowering sites and top-ten rankings for untouched tips."""
import collections, copy, gzip, hashlib, json, pathlib, subprocess
ROOT=pathlib.Path(__file__).resolve().parent;report=json.loads((ROOT/'REPORT.json').read_text());DATA=ROOT/'data/unmerged4'
LABEL='measured on a checker-rejected program';KINDS=['NotYet','Refused','SkippedDependency','error','panic']
raw={}
for r in report['runs']:
 with gzip.open(DATA/(r['name']+'.jsonl.gz'),'rt') as f:raw[r['name']]=[json.loads(l) for l in f]

def check(value):
 assert value['measurement']==LABEL,'label'
 for r in value['runs']:
  assert r['unmerged'] and not r['features'] and r['head']==r['main'],'untouched pin'
  tree=pathlib.Path(r['tree']);adapted=pathlib.Path(r['adapted'])
  assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=tree,text=True).strip()==r['head'],'untouched pin'
  for area in ['internal','stage3/adapt','stage3/apply.sh','stage3/apply.py']:
   assert not subprocess.check_output(['git','diff','HEAD','--',area],cwd=tree),'untouched compiler/adaptations'
  assert subprocess.check_output(['git','rev-parse','HEAD:cohere'],cwd=tree,text=True).strip()==subprocess.check_output(['git','rev-parse','HEAD'],cwd=tree/'cohere',text=True).strip(),'dependency pin'
  rows=raw[r['name']];h=rows[0];assert h['checker_rejected'] and r['measurement']==LABEL,'label'
  manifest={m['file']:m['sha256'] for m in r['source']['files']}
  actual={str(p.relative_to(adapted)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (adapted/'src/compiler').rglob('*.ts')}
  assert manifest==actual and len(actual)==(78 if r['name']=='main-unmerged' else 79),'source'
  assert len(rows[1:])==len(actual) and {row['file'] for row in rows[1:]}==set(manifest),'coverage'
  keys={(f['kind'],f['where'],f['reason'],f['text']) for row in rows[1:] for f in row['findings']}
  counts=collections.Counter(k[0] for k in keys)
  assert r['lowering_counts']=={k:counts[k] for k in KINDS},'unique totals'
  reasons=collections.Counter(k[0]+': '+k[2] for k in keys)
  assert r['per_reason']==dict(reasons),'reason counts'
  expected=sorted([(k.split(': ',1)[0],k.split(': ',1)[1],v) for k,v in reasons.items() if k.split(': ',1)[0] in ['NotYet','Refused']],key=lambda x:(-x[2],x[0]+': '+x[1]))[:10]
  assert [(x['kind'],x['reason'],x['count']) for x in r['top_10']]==expected,'top ten'
  assert all(x['measurement']==LABEL for x in r['top_10']),'label'
  assert r['checker_total']==len(h['diagnostics'])==len(h['diagnostic_sites']),'checker total'
  clean=sorted(f for f in actual if not any(d['file']==f for d in h['diagnostic_sites']))
  ratio=r['checker_clean_ratio']
  assert r['checker_clean_files']==clean and ratio['numerator']==len(clean) and ratio['denominator']==len(actual) and ratio['fraction']==f'{len(clean)}/{len(actual)}' and ratio['percent']==100*len(clean)/len(actual),'source denominator'
  for row in rows[1:]:
   file=row['file'];own={k for k in keys if k[1].rsplit(':',2)[0]==file}
   assert {k:r['per_file'][file][k] for k in KINDS}=={k:sum(s[0]==k for s in own) for k in KINDS},'file attribution'
   assert r['per_file'][file]['checker']==sum(d['file']==file for d in h['diagnostic_sites']),'checker file'
   for unit in row['units']:
    if 'body_start' not in unit:continue
    ds=[d['text'] for d in h['diagnostic_sites'] if d['file']==file and d['start']<unit['body_end'] and (d['end']>unit['body_start'] or d['start']>=unit['body_start'])]
    assert unit['checker_diagnostics']==ds and (unit['status']=='skipped_checker_body')==bool(ds),'body scope'
    if ds:assert not any(f['unit']==unit['where'] for f in row['findings']),'body scope'
check(report);print('both untouched tips, dependency pins, source hashes, exact per-tree file coverage, unique counts, exact reasons, top ten and body eligibility passed')
mutants=[('untouched pin',lambda r:r['runs'][0].__setitem__('head','wrong')),('source',lambda r:r['runs'][0]['source']['files'][0].__setitem__('sha256','wrong')),('unique totals',lambda r:r['runs'][0]['lowering_counts'].__setitem__('NotYet',-1)),('unique totals',lambda r:r['runs'][1]['lowering_counts'].__setitem__('Refused',-1)),('reason counts',lambda r:r['runs'][0]['per_reason'].__setitem__(next(iter(r['runs'][0]['per_reason'])),-1)),('top ten',lambda r:r['runs'][0]['top_10'][0].__setitem__('count',-1)),('top ten',lambda r:r['runs'][1]['top_10'].reverse()),('top ten',lambda r:r['runs'][1]['top_10'].pop()),('file attribution',lambda r:r['runs'][0]['per_file'][next(iter(r['runs'][0]['per_file']))].__setitem__('NotYet',-1)),('label',lambda r:r.__setitem__('measurement','wrong')),('source denominator',lambda r:r['runs'][1]['checker_clean_ratio'].__setitem__('denominator',78))]
for name,mutate in mutants:
 changed=copy.deepcopy(report);mutate(changed)
 try:check(changed)
 except AssertionError as e:assert str(e)==name,(name,e);print(name+' mutant caught')
 else:raise AssertionError(name+' mutant survived')
unit=next(u for row in raw[report['runs'][0]['name']][1:] for u in row['units'] if u.get('checker_diagnostics'))
old=unit['checker_diagnostics'];unit['checker_diagnostics']=[]
try:check(report)
except AssertionError as e:assert str(e)=='body scope';print('body-scope evidence mutant caught')
else:raise AssertionError('body evidence mutant survived')
finally:unit['checker_diagnostics']=old
print('all 12 dedicated artifact/evidence mutants caught')
