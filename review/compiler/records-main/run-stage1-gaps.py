from pathlib import Path
import json,re,subprocess,concurrent.futures
r=Path('.');o=r/'review/compiler/records-main';groups={}
for p in (r/'stage1').rglob('*_test.go'):
 for n in re.findall(r'^func (Test\w+)\(t \*testing.T\)',p.read_text(),re.M):
  if re.search('Gap|Gaps|Probes',n):groups.setdefault('./'+str(p.parent),set()).add(n)
def run(item):
 p,ns=item;cmd=['go','test',p,'-run','^('+'|'.join(sorted(ns))+')$','-count=1','-timeout=85s','-v'];name=p.replace('/','_')
 with (o/(name+'.log')).open('w') as log:
  try:c=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=89).returncode
  except subprocess.TimeoutExpired:c=124
 print(p,c,flush=True);return {'package':p,'tests':sorted(ns),'exit':c}
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as ex:res=list(ex.map(run,sorted(groups.items())))
(o/'stage1-gaps.json').write_text(json.dumps(res,indent=2));assert all(x['exit']==0 for x in res)
