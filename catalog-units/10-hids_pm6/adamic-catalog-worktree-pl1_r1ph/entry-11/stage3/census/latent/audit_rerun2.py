"""Independently recount checker ratios, variance owners and deltas; kill artifact mutants."""
import collections, copy, functools, gzip, hashlib, json, pathlib, re, sys
ROOT=pathlib.Path(__file__).resolve().parent;adapted=pathlib.Path(sys.argv[1]).resolve()
stock=json.loads(pathlib.Path(sys.argv[2]).read_text());report=json.loads((ROOT/'REPORT.json').read_text())
LABEL='measured on a checker-rejected program';KINDS=['NotYet','Refused','SkippedDependency','error','panic']
raw={}
for run in report['runs']:
 with gzip.open(ROOT/'data/rerun2'/(run['name']+'.jsonl.gz'),'rt') as f:raw[run['name']]=[json.loads(l) for l in f]
with gzip.open(ROOT/'data/cumulative.jsonl.gz','rt') as f:raw['cumulative']=[json.loads(l) for l in f]

def keys(rows):return {(f['kind'],f['where'],f['reason'],f['text']) for r in rows[1:] for f in r['findings']}
def kind_counts(s):
 counter=collections.Counter(k[0] for k in s);return {k:counter[k] for k in KINDS}
def family(f):
 if f['kind']!='Refused':return None
 if '(adamic/invariant-mutable)' in f['text']:return 'mutable-invariance'
 if '(method-signature-style)' in f['text']:return 'method-parameter-bivariance'
 if '(adamic/contravariant-override)' in f['text']:return 'contravariant-override'
 return {'a readonly inherited field redeclared mutable':'readonly-to-mutable-override','an override that widens its return type':'covariant-return-override'}.get(f['reason'])
@functools.lru_cache(maxsize=None)
def owner(where):
 file,line,col=where.rsplit(':',2);text=(adapted/file).read_bytes().decode()
 lines=text.splitlines(keepends=True)
 prefix=''.join(lines[:int(line)-1])
 line_prefix=lines[int(line)-1].encode('utf-16-le')[:2*(int(col)-1)].decode('utf-16-le')
 position=len((prefix+line_prefix).encode())
 candidates=[d for d in stock[file] if d['start']<=position<d['end']]
 if not candidates:return dict(where=file+':1:1',name='<file scope>',kind='KindSourceFile',start=0,end=len(text.encode()),qualified_name='<file scope>')
 declaration=min(candidates,key=lambda d:(d['end']-d['start'],-d['start']))
 ancestors=sorted([d for d in candidates if d['start']<=declaration['start'] and declaration['end']<=d['end']],key=lambda d:-(d['end']-d['start']))
 result=dict(declaration,name=declaration['name'].strip(),qualified_name=' / '.join(d['name'].strip() for d in ancestors))
 return result

def check(value):
 assert value['measurement']==LABEL,'label'
 manifest={f['file']:f['sha256'] for f in value['source']['files']}
 actual={str(f.relative_to(adapted)):hashlib.sha256(f.read_bytes()).hexdigest() for f in (adapted/'src/compiler').rglob('*.ts')}
 assert manifest==actual and len(actual)==78,'source manifest'
 for run in [value['baseline'],*value['runs']]:
  rows=raw[run['name']];header=rows[0];sites=keys(rows)
  assert run['measurement']==LABEL and header['checker_rejected'],'label'
  assert run['checker_total']==len(header['diagnostics']),'checker total'
  assert run['checker_per_reason']==dict(sorted(collections.Counter(d.split('error ')[1].split(':')[0] for d in header['diagnostics']).items())),'checker reasons'
  assert set(run['per_file'])==set(manifest),'file coverage'
  assert run['lowering_counts']==kind_counts(sites),'finding totals'
  expected_reasons=dict(sorted(collections.Counter(k[0]+': '+k[2] for k in sites).items()))
  assert run['per_reason']==expected_reasons,'finding reasons'
  assert {(f['kind'],f['where'],f['reason'],f['text']) for f in run['findings']}==sites,'finding ledger'
  for row in rows[1:]:
   file=row['file'];entry=run['per_file'][file]
   assert entry['measurement']==LABEL,'label'
   own={k for k in sites if k[1].rsplit(':',2)[0]==file}
   assert {k:entry[k] for k in KINDS}==kind_counts(own),'per-file finding totals'
   assert entry['checker']==sum(s['file']==file for s in header['diagnostic_sites']),'per-file checker'
   skipped=set()
   for unit in row['units']:
    if 'body_start' not in unit:continue
    diagnostics=[d['text'] for d in header['diagnostic_sites'] if d['file']==file and d['start']<unit['body_end'] and (d['end']>unit['body_start'] or d['start']>=unit['body_start'])]
    assert unit['checker_diagnostics']==diagnostics and (unit['status']=='skipped_checker_body')==bool(diagnostics),'body eligibility'
    if diagnostics:skipped.add(unit['where'])
   assert not any(f['unit'] in skipped for f in row['findings']),'body eligibility'
   if run is not value['baseline']:
    assert [{k:v for k,v in d.items() if k!='measurement'} for d in row['declarations']]==stock[file],'stock declaration spans'
  units=[u for row in rows[1:] for u in row['units']]
  assert run['bodies_skipped']==sum(u['status']=='skipped_checker_body' for u in units),'unit count'
  assert run['top_level_units']==len(units) and run['functions_attempted']==sum(u['kind']=='KindFunctionDeclaration' and u['status']!='skipped_checker_body' for u in units),'unit count'
  clean=sorted(f for f,entry in run['per_file'].items() if entry['checker']==0)
  assert run['checker_clean_files']==clean,'zero diagnostic files'
  ratio=run['checker_clean_ratio']
  assert ratio['measurement']==LABEL,'label'
  assert ratio['denominator']==78,'ratio denominator'
  assert ratio['numerator']==len(clean) and ratio['fraction']==f'{len(clean)}/78' and abs(ratio['percent']-100*len(clean)/78)<1e-10,'checker clean ratio'
  if run is value['baseline']:
   assert run['variance']['total']==sum(family(f) is not None for f in run['findings']),'variance recount'
   continue
  variance=[f for f in run['findings'] if family(f)]
  v=run['variance'];assert v['measurement']==LABEL,'label'
  assert v['total']==len(variance) and v['families']==dict(sorted(collections.Counter(family(f) for f in variance).items())),'variance recount'
  expected=collections.defaultdict(list)
  for finding in variance:expected[owner(finding['where'])['where']].append(finding)
  assert set(v['per_declaration'])==set(expected) and v['owning_declaration_count']==len(expected),'variance owner coverage'
  for location,findings in expected.items():
   group=v['per_declaration'][location]
   assert {k:x for k,x in group['declaration'].items() if k!='measurement'}==owner(findings[0]['where']),'variance owner'
   assert group['count']==len(findings) and group['families']==dict(collections.Counter(family(f) for f in findings)),'variance owner counts'
   assert {(f['kind'],f['where'],f['reason'],f['text']) for f in group['findings']}=={(f['kind'],f['where'],f['reason'],f['text']) for f in findings},'variance owner sites'
   for f in group['findings']:
    assert f['variance_family']==family(f),'variance recount'
    assert {k:x for k,x in f['owning_declaration'].items() if k!='measurement'}==owner(f['where']),'variance owner'
  for delta_name,prior in [('delta_vs_previous_cumulative',value['baseline']),*([('delta_vs_cumulative_2',value['runs'][0])] if run['name'].endswith('nested') else [])]:
   d=run[delta_name];assert d['measurement']==LABEL,'label'
   if 'variance' in d:assert d['variance']==run['variance']['total']-prior['variance']['total'],'delta recount'
   assert d['checker']==run['checker_total']-prior['checker_total'] and d['checker_clean_files']==len(clean)-prior['checker_clean_ratio']['numerator'],'delta recount'
   assert d['counts']=={k:run['lowering_counts'][k]-prior['lowering_counts'][k] for k in KINDS},'delta recount'
   assert d['new_checker_clean_files']==sorted(set(clean)-set(prior['checker_clean_files'])) and d['no_longer_checker_clean_files']==sorted(set(prior['checker_clean_files'])-set(clean)),'delta recount'
   assert d['per_reason']=={k:run['per_reason'].get(k,0)-prior['per_reason'].get(k,0) for k in sorted(run['per_reason'].keys()|prior['per_reason'].keys())},'delta recount'
   for file in manifest:
    assert d['per_file'][file]['checker']==run['per_file'][file]['checker']-prior['per_file'][file]['checker'],'delta recount'
    assert {k:d['per_file'][file][k] for k in KINDS}=={k:run['per_file'][file][k]-prior['per_file'][file][k] for k in KINDS},'delta recount'
check(report)
print('stock TypeScript spans, checker files/ratios, variance owners, eligibility and deltas passed')
mutants=[
 ('source manifest',lambda r:r['source']['files'][0].__setitem__('sha256','0'*64)),
 ('checker total',lambda r:r['runs'][0].__setitem__('checker_total',1425)),
 ('per-file checker',lambda r:r['runs'][0]['per_file'][r['runs'][0]['checker_clean_files'][0]].__setitem__('checker',1)),
 ('zero diagnostic files',lambda r:r['runs'][0]['checker_clean_files'].pop()),
 ('ratio denominator',lambda r:r['runs'][0]['checker_clean_ratio'].__setitem__('denominator',77)),
 ('checker clean ratio',lambda r:r['runs'][0]['checker_clean_ratio'].__setitem__('numerator',27)),
 ('finding totals',lambda r:r['runs'][0]['lowering_counts'].__setitem__('NotYet',1149)),
 ('finding reasons',lambda r:r['runs'][0]['per_reason'].__setitem__(next(iter(r['runs'][0]['per_reason'])),-1)),
 ('unit count',lambda r:r['runs'][0].__setitem__('bodies_skipped',-1)),
 ('variance recount',lambda r:r['runs'][0]['variance'].__setitem__('total',495)),
 ('variance owner',lambda r:next(iter(r['runs'][0]['variance']['per_declaration'].values()))['declaration'].__setitem__('name','wrong owner')),
 ('variance owner counts',lambda r:next(iter(r['runs'][0]['variance']['per_declaration'].values())).__setitem__('count',-1)),
 ('delta recount',lambda r:r['runs'][0]['delta_vs_previous_cumulative'].__setitem__('checker',0)),
 ('label',lambda r:r.__setitem__('measurement','unqualified')),
]
for name,mutate in mutants:
 changed=copy.deepcopy(report);mutate(changed)
 try:check(changed)
 except AssertionError as failure:
  assert str(failure)==name,(name,failure);print(name+' mutant caught')
 else:raise AssertionError(name+' mutant survived')


entry=next(row['declarations'][0] for row in raw['cumulative-2'][1:] if row['declarations'])
old=entry['end'];entry['end']+=1
try:check(report)
except AssertionError as failure:
 assert str(failure)=='stock declaration spans';print('raw declaration span mutant caught')
else:raise AssertionError('declaration span mutant survived')
finally:entry['end']=old

print('rerun evidence audit passed')
