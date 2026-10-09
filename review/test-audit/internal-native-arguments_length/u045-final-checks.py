import pathlib,json,difflib,subprocess,os,time
p=pathlib.Path('review/test-audit/internal-native-arguments_length');probes=json.loads((p/'probes.json').read_text());extra=dict(id='P_REUSE',file='internal/native/reuse.go',entry='func planReuse(program *ir.Program, lending map[int]bool) *reusePlan {',change='return &reusePlan{consumed: map[int]bool{}, spreads: map[*ir.Statement]map[int]bool{}, moves: map[*ir.Statement]map[int]bool{}, arrays: map[*ir.Statement]map[int]bool{}, lending: lending}')
probes.append(extra)
for m in probes:
 f=pathlib.Path(m['file']);base=f.read_text();changed=base.replace(m['entry'],m['entry']+'\n\tif len([]int{}) == 0 { '+m['change']+' }')
 (p/'diffs'/ (m['id']+'.diff')).write_text(''.join(difflib.unified_diff(base.splitlines(True),changed.splitlines(True),fromfile='a/'+str(f),tofile='b/'+str(f))))
 try:
  f.write_text(changed)
  with open(p/(m['id']+'-vet.log'),'w') as out:r=subprocess.run(['timeout','120','go','vet','./internal/native/'],stdout=out,stderr=subprocess.STDOUT)
  assert r.returncode==0,(m['id'],r.returncode)
  if m['id']=='P_REUSE':
   env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u045/cache/P_REUSE';pattern='^('+ '|'.join(json.loads((p/'names.json').read_text()))+')$';cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',pattern];start=time.monotonic()
   with open(p/'P_REUSE.log','w') as out:r=subprocess.run(cmd,stdout=out,stderr=subprocess.STDOUT,env=env)
   commands=json.loads((p/'commands.json').read_text());commands.append(dict(id='P_REUSE',seconds=time.monotonic()-start,exit=r.returncode,command=' '.join(cmd)));(p/'commands.json').write_text(json.dumps(commands,indent=2));print('P_REUSE',r.returncode,flush=True)
 finally:f.write_text(base)
(p/'probes.json').write_text(json.dumps(probes,indent=2))
# Validate witness weakening standalone.
patch=(p/'diffs/W_BUILD.diff');subprocess.run(['git','apply',str(patch)],check=True)
try:
 with open(p/'W_BUILD-vet.log','w') as out:r=subprocess.run(['go','vet','./internal/native/'],stdout=out,stderr=subprocess.STDOUT)
 assert r.returncode==0
finally:subprocess.run(['git','apply','-R',str(patch)],check=True)
# Every patch applies cleanly to starting sources.
for patch in (p/'diffs').glob('*.diff'):subprocess.run(['git','apply','--check',str(patch)],check=True)
print('all diffs apply; all Go diffs vetted',flush=True)
