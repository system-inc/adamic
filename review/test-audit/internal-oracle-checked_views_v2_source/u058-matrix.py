exec(open('/tmp/u058-run.py').read().split("if __name__")[0]);import difflib
menu=json.loads((out/'menu.json').read_text()); files={m['file'] for m in menu};original={f:(root/f).read_text() for f in files}
(out/'scratch').mkdir(exist_ok=True)
for f,s in original.items():
 p=out/'scratch'/f;p.parent.mkdir(parents=True,exist_ok=True);p.write_text(s)
for m in menu:
 f=m['file'];changed=original[f].replace(m['old'],m['new'],1)
 for text in m.get('extra_remove',[]):changed=changed.replace(text,'')
 (root/f).write_text(changed)
 try:
  assert run('apply-'+m['id'],['git','apply','--reverse','--check',str(out/'diffs'/(m['id']+'.diff'))])==0
  assert run('vet-'+m['id'],['timeout','120','go','vet','./'+str(pathlib.Path(f).parent)+'/'])==0,m['id']
 finally:(root/f).write_text(original[f])
for f,s in original.items():
 for m in menu:
  if m['file']==f and m['kind']!='witness':s=s.replace(m['old'],m['switch'],1)
 (root/f).write_text(s)
helper='package %s\nimport "os"\nfunc u058Mutant(id string) bool { return os.Getenv("ADAMIC_MUTANT")==id }\nfunc u058Number(id string, normal, changed int) int { if u058Mutant(id) { return changed }; return normal }\n'
for pkg in ['lower','native','javascript']:(root/f'internal/{pkg}/u058_switch.go').write_text(helper%pkg)
assert run('switch-vet',['go','vet','./internal/lower/','./internal/native/','./internal/javascript/'])==0
assert run('switch-build',['go','test','-c','./internal/oracle/','-o','/tmp/u058-oracle.test'])==0
for m in menu:
 if m['kind']=='witness':continue
 env=os.environ.copy();env.update(ADAMIC_MUTANT=m['id'],ADAMIC_BUILD_CACHE_DIR='/tmp/u058/cache/'+m['id'],ADAMIC_GATE_UNCACHED='1')
 run(m['id'],['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern],env)
 log=(out/(m['id']+'.log')).read_text()
 if 'panic:' in log and ('runtime error:' in log or 'panic: test timed out' in log):
  ev=[]
  for line in log.splitlines():
   try:ev.append(json.loads(line))
   except:pass
  finished={e.get('Test') for e in ev if e['Action'] in ['pass','fail','skip']}
  for row in rows:
   if row not in finished:run(m['id']+'-alone-'+row,['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^'+row+'$'],env)
for f,s in original.items():(root/f).write_text(s)
for pkg in ['lower','native','javascript']:(root/f'internal/{pkg}/u058_switch.go').unlink()
m=next(m for m in menu if m['id']=='W1');f=m['file'];(root/f).write_text(original[f].replace(m['old'],m['new'],1))
try:run('W1',['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^TestClockGenericReturnsT01Mutant$'])
finally:(root/f).write_text(original[f])
run('restored-baseline',['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run',pattern])
