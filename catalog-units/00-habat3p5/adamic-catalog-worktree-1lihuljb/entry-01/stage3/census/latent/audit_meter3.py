"""Independently verify the checker work queue, source coverage and guarded lowering ledger."""
import collections, copy, gzip, hashlib, json, pathlib, re, sys
ROOT=pathlib.Path(__file__).resolve().parent;adapted=pathlib.Path(sys.argv[1]).resolve()
report=json.loads((ROOT/'REPORT.json').read_text())
with gzip.open(ROOT/'data/meter3/meter-3.jsonl.gz','rt') as f:raw=[json.loads(l) for l in f]
LABEL='measured on a checker-rejected program'
expected=[]
for site in raw[0]['diagnostic_sites']:
 match=re.search(r':(\d+):(\d+): error (TS\d+): ',site['text']);assert match
 line,column,code=match.groups();message=site['text'][match.end():]
 expected.append(dict(site,line=int(line),column=int(column),code=code,cause=message.split('\n')[0],message=message))
 text=(adapted/site['file']).read_bytes();prefix=text[:site['start']].decode()
 assert int(line)==prefix.count('\n')+1,'raw diagnostic line'
 assert int(column)==len(prefix.rsplit('\n',1)[-1].encode('utf-16-le'))//2+1,'raw diagnostic column'
 assert site['end']>=site['start'],'raw diagnostic span'
by_file=collections.defaultdict(list)
for d in expected:by_file[d['file']].append(d)
keys={(f['kind'],f['where'],f['reason'],f['text']) for row in raw[1:] for f in row['findings']}

def check(r):
 assert r['measurement']==LABEL,'label'
 manifest={m['file']:m['sha256'] for m in r['source']['files']}
 actual={str(p.relative_to(adapted)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (adapted/'src/compiler').rglob('*.ts')}
 assert manifest==actual and len(manifest)==78,'source'
 assert {row['file'] for row in raw[1:]}==set(manifest) and len(raw[1:])==78,'coverage'
 run=r['runs'][0];assert run['checker_diagnostics']==expected,'diagnostic ledger'
 assert run['checker_total']==len(expected)==len(raw[0]['diagnostics']),'checker total'
 assert raw[0]['checker_rejected'] and run['measurement']==LABEL,'label'
 assert collections.Counter(s['text'] for s in raw[0]['diagnostic_sites'])==collections.Counter(raw[0]['diagnostics']),'raw checker ledger'
 assert run['checker_totals_by_code']==dict(collections.Counter(d['code'] for d in expected)),'totals by code'
 assert run['checker_per_reason']==run['checker_totals_by_code'],'totals by code'
 for file,v in run['per_file'].items():assert v['checker']==len(by_file[file]),'per-file checker'
 clean=sorted(f for f in manifest if not by_file[f]);assert run['checker_clean_files']==clean,'clean files'
 ratio=run['checker_clean_ratio'];assert ratio['denominator']==78 and ratio['numerator']==len(clean) and ratio['fraction']==f'{len(clean)}/78' and ratio['percent']==100*len(clean)/78,'ratio'
 queue=sorted((f for f in manifest if 1<=len(by_file[f])<=10),key=lambda f:(len(by_file[f]),f))
 assert [q['file'] for q in run['cheap_wins']]==queue,'queue files'
 for q in run['cheap_wins']:
  assert q['measurement']==LABEL and q['count']==len(by_file[q['file']]),'queue count'
  ds=q['diagnostics'];truth=by_file[q['file']]
  assert len(ds)==len(truth),'queue diagnostics'
  for d,e in zip(ds,truth):
   assert d['code']==e['code'],'queue code'
   assert d['line']==e['line'] and d['column']==e['column'],'queue line'
   assert d['cause']==e['cause'],'queue cause'
   assert d==e,'queue diagnostic metadata'
 assert run['lowering_counts']=={k:sum(s[0]==k for s in keys) for k in ['NotYet','Refused','SkippedDependency','error','panic']},'lowering totals'
 assert run['per_reason']==dict(collections.Counter(s[0]+': '+s[2] for s in keys)),'lowering reasons'
 for row in raw[1:]:
  assert row['measurement']==LABEL,'label'
  for unit in row['units']:
   if 'body_start' not in unit:continue
   sites=[s['text'] for s in raw[0]['diagnostic_sites'] if s['file']==row['file'] and s['start']<unit['body_end'] and (s['end']>unit['body_start'] or s['start']>=unit['body_start'])]
   assert unit['checker_diagnostics']==sites and (unit['status']=='skipped_checker_body')==bool(sites),'body eligibility'
   if sites:assert not any(f['unit']==unit['where'] for f in row['findings']),'body eligibility'
check(report);print('source, all 78 files, checker codes/lines, zero ratio, complete 1-to-10 queue, lowering and body eligibility passed')
mutants=[
 ('source',lambda r:r['source']['files'][0].__setitem__('sha256','wrong')),
 ('checker total',lambda r:r['runs'][0].__setitem__('checker_total',-1)),
 ('totals by code',lambda r:r['runs'][0]['checker_totals_by_code'].__setitem__('TS99999',1)),
 ('clean files',lambda r:r['runs'][0]['checker_clean_files'].pop()),
 ('ratio',lambda r:r['runs'][0]['checker_clean_ratio'].__setitem__('denominator',77)),
 ('queue files',lambda r:r['runs'][0]['cheap_wins'].pop()),
 ('queue count',lambda r:r['runs'][0]['cheap_wins'][0].__setitem__('count',11)),
 ('queue diagnostics',lambda r:r['runs'][0]['cheap_wins'][0]['diagnostics'].pop()),
 ('queue code',lambda r:r['runs'][0]['cheap_wins'][0]['diagnostics'][0].__setitem__('code','TS99999')),
 ('queue line',lambda r:r['runs'][0]['cheap_wins'][0]['diagnostics'][0].__setitem__('line',-1)),
 ('queue cause',lambda r:r['runs'][0]['cheap_wins'][0]['diagnostics'][0].__setitem__('cause','wrong')),
 ('label',lambda r:r.__setitem__('measurement','wrong')),
]
for name,mutate in mutants:
 changed=copy.deepcopy(report);mutate(changed)
 try:check(changed)
 except AssertionError as e:assert str(e)==name,(name,e);print(name+' mutant caught')
 else:raise AssertionError(name+' mutant survived')
print(f'all {len(mutants)} dedicated artifact mutants caught')
