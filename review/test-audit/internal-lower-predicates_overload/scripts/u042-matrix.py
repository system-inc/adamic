import pathlib,json,time,subprocess,os,re
R=pathlib.Path('/workspace/adamic');E=R/'review/test-audit/internal-lower-predicates_overload';P=R/'internal/lower';base=json.load(open(E/'base.json'));meta=json.load(open(E/'function-offsets.json'));plans=json.load(open(E/'plan.json'));rows=json.load(open(E/'scope.json'))
while not (E/'timing.done').exists():time.sleep(1)
for f in {p['file'].split('/')[-1] for p in plans}:
 s=base[f];edits=[];copies=[]
 for fn in meta[f]:
  dispatch=''
  for p in plans:
   if not p['file'].endswith('/'+f) or not fn['start']<=p['pos']<fn['end']:continue
   segment=s[fn['start']:fn['end']];pos=p['pos']-fn['start'];changed=segment[:pos]+p['new']+segment[pos+len(p['old']):];newname=fn['name']+'_audit_'+p['id'];namepos=fn['namepos']-fn['start'];changed=changed[:namepos]+newname+changed[namepos+len(fn['name']):];copies.append(changed);call=fn['receiver']+newname+'('+fn['args']+')';dispatch+='\nif auditSelector()=="'+p['id']+'" { '+('return '+call if fn['results'] else call+'; return')+' }\n'
  if dispatch:edits.append((fn['body']+1,dispatch))
 for pos,text in sorted(edits,reverse=True):s=s[:pos]+text+s[pos:]
 (P/f).write_text(s+'\n'+'\n'.join(copies)+'\n')
(P/'audit_selector.go').write_text('package lower\nimport "os"\nfunc auditSelector() string {return os.Getenv("ADAMIC_MUTANT")}\n')
subprocess.run(['gofmt','-w']+[str(P/f) for f in ['predicates.go','predicates_proof.go','expression.go','audit_selector.go']])
with (E/'switch-vet.log').open('w') as log:rc=subprocess.run(['go','vet','./internal/lower/'],cwd=R,stdout=log,stderr=subprocess.STDOUT).returncode
assert rc==0,'switch vet failed'
(E/'switched-source').mkdir(exist_ok=True)
for f in ['predicates.go','predicates_proof.go','expression.go','audit_selector.go']:(E/'switched-source'/(f+'.fixture')).write_text((P/f).read_text())
records=[]
for p in plans:
 id=p['id'];env=dict(os.environ,ADAMIC_MUTANT=id,ADAMIC_BUILD_CACHE_DIR='/tmp/u042/cache/'+id);start=time.monotonic()
 with (E/(id+'.log')).open('w') as log:rc=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','.'],cwd=R,env=env,stdout=log,stderr=subprocess.STDOUT).returncode
 text=(E/(id+'.log')).read_text();panic='panic:' in text or rc==124
 if panic:
  for row in rows:
   with (E/(id+'-'+row+'.log')).open('w') as log:subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/lower/','-run','^'+row+'$'],cwd=R,env=env,stdout=log,stderr=subprocess.STDOUT)
 records.append(dict(id=id,seconds=time.monotonic()-start,exit=rc,panic_or_timeout=panic));(E/'matrix-time.json').write_text(json.dumps(records,indent=2))
for f in ['predicates.go','predicates_proof.go','expression.go']:(P/f).write_text(base[f])
(P/'audit_selector.go').unlink()
for p in plans:
 f=p['file'].split('/')[-1];s=base[f];(P/f).write_text(s[:p['pos']]+p['new']+s[p['pos']+len(p['old']):]);start=time.monotonic()
 with (E/(p['id']+'-vet.log')).open('w') as log:rc=subprocess.run(['go','vet','./internal/lower/'],cwd=R,stdout=log,stderr=subprocess.STDOUT).returncode
 assert rc==0,p['id'];(P/f).write_text(s)
 with (E/'vet-times.jsonl').open('a') as out:out.write(json.dumps(dict(id=p['id'],seconds=time.monotonic()-start,exit=rc))+'\n')
(E/'matrix.done').write_text('done')
