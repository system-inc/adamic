"""Check supplemental whole-file records without merging duplicate measurements."""
import gzip,hashlib,json,pathlib,re
O=pathlib.Path(__file__).resolve().parent;E=O/'evidence'
checks={};records={}
for label in ['head','log']:
 d=E/'latent'/label/'r00';m=json.loads((d/'metrics.json').read_text())
 assert m['returncode']==0 and not m.get('validation_error') and m['wall_seconds']<1760
 for k,v in m.items():
  if k.endswith('_sha256'):
   with gzip.open(d/(k[:-7]+'.gz'),'rb') as f:assert hashlib.sha256(f.read()).hexdigest()==v
 with gzip.open(d/'records.jsonl.gz','rb') as f:records[label]=f.read()
 rows=[json.loads(line) for line in records[label].splitlines()]
 assert rows[0]['latent_mode']=='first-error' and len(rows)==2 and len(rows[1]['units'])==251
 checks[label]=m
assert records['head']==records['log']
maxima=dict(judged=0,past=0,kept=0);walks=0
with gzip.open(E/'latent/log/r00/pairs.log.gz','rt') as f:
 for line in f:
  m=re.match(r'walk judged=(\d+) past=(\d+) kept=(\d+) stop=',line)
  if m:
   walks+=1
   for key,val in zip(maxima,map(int,m.groups())):maxima[key]=max(maxima[key],val)
value=dict(equal=True,bytes=len(records['head']),raw_records_sha256=hashlib.sha256(records['head']).hexdigest(),units=251,head='62ddd47969e5168e5a0e74a23498b85d144b2361',log='29de73556749a0374da691bb9599e0e024d1214a',metrics=checks,supplemental_walks=walks,supplemental_maxima=maxima,excluded_from_primary_totals=True)
(E/'whole-public-byte-identity.json').write_text(json.dumps(value,indent=2)+'\n')
print(json.dumps({k:v for k,v in value.items() if k!='metrics'},indent=2))
