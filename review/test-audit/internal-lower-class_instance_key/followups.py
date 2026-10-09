import pathlib,subprocess,json,time,os
p=pathlib.Path(__file__).parent;times={}
def run(args,name):
 start=time.monotonic()
 with (p/(name+'.log')).open('w') as f:r=subprocess.run(args,stdout=f,stderr=subprocess.STDOUT)
 times[name]={'exit':r.returncode,'wall':round(time.monotonic()-start,3),'command':args};(p/'followup-timings.json').write_text(json.dumps(times,indent=2)+'\n');print(name,times[name],flush=True)
 return r.returncode
assert subprocess.run(['git','diff','--exit-code','--','internal/lower'],capture_output=True).returncode==0,'Production sources must be restored first'
mat=json.loads((p/'matrix.json').read_text());survivors=[m for m,states in mat.items() if not any(s=='fail' for s in states.values())]
w=pathlib.Path('internal/lower/audit_u030_witness_test.go');w.write_text((p/'survivor-witness.go.txt').read_text())
try:
 run(['timeout','120','go','test','-v','-count=1','-timeout','90s','./internal/lower/','-run','^TestAuditU030SurvivorWitness$'],'survivor-original')
 for mid in survivors:
  files=[m['file'] for m in json.loads((p/'menu.json').read_text()) if m['id']==mid];original={f:pathlib.Path(f).read_text() for f in files}
  try:
   subprocess.run(['git','apply',str(p/(mid+'.diff'))],check=True);os.environ['ADAMIC_BUILD_CACHE_DIR']='/tmp/u030/cache/witness-'+mid
   run(['timeout','120','go','test','-v','-count=1','-timeout','90s','./internal/lower/','-run','^TestAuditU030SurvivorWitness$'],'survivor-'+mid)
  finally:
   for f,s in original.items():pathlib.Path(f).write_text(s)
finally:w.unlink()
extra=json.loads((p/'extra-timings-needed.json').read_text())
for r in extra:
 assert not r.endswith(' family'),'Family timing requires all members'
 for i in range(1,4):run(['timeout','120','go','test','-count=1','-timeout','90s','./internal/lower/','-run','^'+r+'$'],r+'-'+str(i))
