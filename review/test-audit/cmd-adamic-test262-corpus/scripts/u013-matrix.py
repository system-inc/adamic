import pathlib,subprocess,json,time,os,difflib,re
R=pathlib.Path("/workspace/adamic");E=R/"review/test-audit/cmd-adamic-test262-corpus";P=R/"cmd/adamic-test262"
meta=json.load(open(E/'function-offsets.json'))
# wait baseline timing before modifying source
while not (E/'timing.done').exists():time.sleep(1)
base=json.load(open(E/'base.json'));plans=json.load(open(E/'plan.json'))
for f,s in base.items():
 edits=[];copies=[]
 for fn in meta.get(f,[]):
  relevant=[p for p in plans if p['file'].endswith('/'+f) and fn['start']<=p['pos']<fn['end']]
  dispatch=''
  for p in relevant:
   segment=s[fn['start']:fn['end']];pos=p['pos']-fn['start'];changed=segment[:pos]+p['new']+segment[pos+len(p['old']):];newname=fn['name']+'_audit_'+p['id'];namepos=fn['namepos']-fn['start'];changed=changed[:namepos]+newname+changed[namepos+len(fn['name']):];copies.append(changed)
   call=fn['receiver']+newname+'('+fn['args']+')';dispatch+='\nif auditSelector()=="'+p['id']+'" { '+('return '+call if fn['results'] else call+'; return')+' }\n'
  if dispatch:edits.append((fn['body']+1,dispatch))
 for pos,text in sorted(edits,reverse=True):s=s[:pos]+text+s[pos:]
 (P/f).write_text(s+'\n'+'\n'.join(copies)+'\n')
(P/'audit_selector.go').write_text('package main\nimport "os"\nfunc auditSelector() string { return os.Getenv("ADAMIC_MUTANT") }\n')
subprocess.run(['gofmt','-w']+[str(P/f) for f in base]+[str(P/'audit_selector.go')])
with (E/'switch-vet.log').open('w') as log:rc=subprocess.run(['go','vet','./cmd/adamic-test262/'],cwd=R,stdout=log,stderr=subprocess.STDOUT).returncode
assert rc==0,'switch vet failed'
records=[]
for p in plans:
 env=os.environ.copy();env['ADAMIC_MUTANT']=p['id'];start=time.monotonic()
 with (E/(p['id']+'.log')).open('w') as log:rc=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./cmd/adamic-test262/','-run','.'],cwd=R,env=env,stdout=log,stderr=subprocess.STDOUT).returncode
 records.append(dict(id=p['id'],seconds=time.monotonic()-start,exit=rc));(E/'matrix-time.json').write_text(json.dumps(records,indent=2))
# restore code, validate every standalone variant through Go vet
for f,s in base.items():(P/f).write_text(s)
(P/'audit_selector.go').unlink()
for p in plans:
 f=p['file'].split('/')[-1];s=base[f];(P/f).write_text(s[:p['pos']]+p['new']+s[p['pos']+len(p['old']):])
 with (E/(p['id']+'-vet.log')).open('w') as log:rc=subprocess.run(['go','vet','./cmd/adamic-test262/'],cwd=R,stdout=log,stderr=subprocess.STDOUT).returncode
 assert rc==0,p['id'];(P/f).write_text(s)
(E/'matrix.done').write_text('done')
