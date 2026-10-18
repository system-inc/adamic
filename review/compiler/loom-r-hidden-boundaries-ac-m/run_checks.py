import pathlib,subprocess,json,time
root=pathlib.Path.cwd();folder=root/'review/compiler/loom-r-hidden-boundaries-ac-m';results=[]
def run(label,cmd,limit=840):
 start=time.monotonic()
 with (folder/(label+'.log')).open('w') as log:
  try:p=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=limit);code=p.returncode
  except subprocess.TimeoutExpired:code=124
 events=[]
 for line in (folder/(label+'.log')).read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if isinstance(e,dict) and e.get('Test') and e.get('Action') in ('pass','fail','skip'):events.append(dict(package=e.get('Package'),name=e['Test'],action=e['Action'],seconds=e.get('Elapsed',0)))
 names={e['name'] for e in events};leaves=[e for e in events if not any(n.startswith(e['name']+'/') for n in names)]
 row=dict(label=label,command=cmd,exit=code,seconds=round(time.monotonic()-start,3),tests=events,over_budget=[e for e in leaves if e['seconds']>=60]);results.append(row)
 (folder/'results.json').write_text(json.dumps(results,indent=2)+'\n');print(label,code,row['seconds'],'over-budget',len(row['over_budget']),flush=True)
 return row
if __name__=='__main__':
 r=run('counts-update',['timeout','840','go','test','./internal/oracle','-run','TestCountsAreRecorded','-count=1','-timeout','13m','-parallel','4','-json','-args','-update-counts'])
 if r['exit']:raise SystemExit(r['exit'])
 # Build declared command archives before tests that only find them.
 run('command-runtime-products',['timeout','150','go','test','./cmd/adamic','-run','^TestProduct_CommandRuntime','-count=1','-timeout','90s','-parallel','4','-json'],180)
 for pkg in ['cmd/adamic','internal/ir','internal/lower','internal/flow','internal/native','internal/oracle','stage3/fixtures','stage3/census/latent/refusalrewrite']:
  run(pkg.replace('/','-'),['timeout','840','go','test','./'+pkg+'/...','-count=1','-timeout','13m','-parallel','4','-p','1','-json'])
 run('vet',['timeout','180','go','vet','./cmd/adamic/...','./internal/ir/...','./internal/lower/...','./internal/flow/...','./internal/native/...','./internal/oracle/...','./stage3/fixtures/...','./stage3/census/latent/refusalrewrite/...'],210)
 changed=subprocess.check_output(['git','diff','--cached','--name-only','--','*.go'],text=True).splitlines()
 run('gofmt',['timeout','30','gofmt','-l']+changed,40)
