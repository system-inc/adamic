import json,pathlib
s=pathlib.Path('/workspace/scratch/path-mint-618/evidence')
r={}
for label,af,bf in [('compiler','base-stock/javascript.json','after-stock/javascript.json'),('allSource','base-all-js.json','after-all-js.json')]:
 a=json.loads((s/af).read_text());b=json.loads((s/bf).read_text())
 if label=='allSource':a=a['hashes'];b=b['hashes']
 r[label]={'baseFiles':len(a),'afterFiles':len(b),'new':sorted(b.keys()-a.keys()),'gone':sorted(a.keys()-b.keys()),'changed':[k for k in sorted(a.keys()&b.keys()) if a[k]!=b[k]]}
 assert a==b,r[label]
(s/'javascript-comparison.json').write_text(json.dumps(r,indent=2)+'\n');print(r)
a=json.loads((s/'base-stock/stock.json').read_text());b=json.loads((s/'after-stock/stock.json').read_text())
assert a['roots']==b['roots'] and len(a['roots'])==79
assert a['diagnostics']==b['diagnostics']==[]
(s/'stock-comparison.json').write_text(json.dumps({'version':a['version'],'roots':79,'baseDiagnostics':0,'afterDiagnostics':0,'new':[],'gone':[],'changed':[]},indent=2)+'\n')
