"""Independent scope/count/evidence checks, with failing artifact mutants."""
import collections,copy,gzip,json,sys
from pathlib import Path
repo=Path(__file__).resolve().parents[4];unit=Path(__file__).resolve().parent
m=json.load(open(sys.argv[1]));c=json.load(open(sys.argv[2]));ledger=json.loads((unit/'latent-ledger.json').read_text());reference=json.loads((repo/'stage3/drivers/parser/cases-reference.json').read_text());raw={}
with gzip.open(repo/'stage3/meter/runs/20261008T035244Z.latent-full/compiler/full.jsonl.gz','rt') as f:
 for line in f:
  for x in json.loads(line).get('findings',[]):raw[tuple(x.get(k) for k in ['kind','where','reason','text'])]=x

def audit(m,c,l):
 assert m['timing_instrumented'] and not c['timing_instrumented']
 assert m['rows']==c['rows'] and m['options']==c['options'] and m['typeTrees']==c['typeTrees']
 rows=[x for x in m['rows'] if x['group'] not in ['compiler','directed']]
 assert len(rows)==10406 and sum(x['nodes'] for x in rows)==1323111
 expected={r['input_path']:r for r in reference['cases']};assert set(expected)=={x['name'] for x in rows}
 for x in rows:assert x['nodes']==expected[x['name']]['node']['node_count'] and x['raw_sha256']==expected[x['name']]['source_sha256']
 assert sum(x['jsdoc_roots'] for x in rows)==1758 and sum(x['jsdoc_descendants'] for x in rows)==12534
 assert sum(x['jsdoc_diagnostics'] for x in rows)==34
 compiler=[x for x in m['rows'] if x['group']=='compiler'];assert len(compiler)==82 and sum(x['jsdoc_roots'] for x in compiler)==4157
 for r in m['options']:
  want=r['mode']=='ParseAll' or r['mode']!='ParseNone' and (r['kind'] in ['JS','JSX'] or r['mode']=='ParseForTypeErrors' and r['link'])
  assert r['jsdoc_roots']==int(want)
 for g,t in m['timings'].items():
  assert len(t['rounds'])==5
  for pair in t['rounds']:
   assert pair['ParseNone']['calls']==0 and pair['ParseNone']['jsdoc_ms']==0
   assert 0<=pair['ParseAll']['jsdoc_ms']<=pair['ParseAll']['parse_ms']
 keys=[]
 for x in l['findings']:
  k=tuple(x[k] for k in ['kind','where','reason','text']);assert k in raw;keys.append(k)
  location=x['where'][x['where'].find('src/compiler/'):].rsplit(':',2)[0];assert x['file']==location
 assert len(keys)==len(set(keys)) and len(keys)==592
 assert collections.Counter(x['kind'] for x in l['findings'])==l['counts']=={'Refused':192,'NotYet':400}
audit(m,c,ledger)
for name,change in [('missing case',lambda a,b,l:a['rows'].pop(100)),('invented JSDoc count',lambda a,b,l:a['rows'][0].update(jsdoc_roots=999)),('changed latent reason',lambda a,b,l:l['findings'][0].update(reason='MUTANT_WRONG_REASON'))]:
 a,b,l=copy.deepcopy((m,c,ledger));change(a,b,l)
 try:audit(a,b,l)
 except AssertionError:print('CAUGHT:',name)
 else:raise AssertionError('surviving artifact mutant: '+name)
print('PASS: source timing control, 10406 cases, modes, pinned latent sites, 3 artifact mutants')
