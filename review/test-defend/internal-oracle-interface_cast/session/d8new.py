import pathlib,time,subprocess,os,json
p=pathlib.Path('/workspace/adamic/review/test-defend/internal-oracle-interface_cast/session')
while 'completed' not in (p/'follow-driver.log').read_text():time.sleep(1)
file=pathlib.Path('internal/lower/interface_cast.go');orig=file.read_text()
try:
 subprocess.run(['git','apply',str(p/'D8.diff')],check=True)
 with (p/'D8-vet.log').open('w') as f:v=subprocess.run(['go','vet','./internal/lower/'],stdout=f,stderr=subprocess.STDOUT)
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^(TestFractionalPowersReachRuntime|TestReviewPrograms.*)$'];start=time.monotonic()
 with (p/'D8-new.log').open('w') as f:r=subprocess.run(cmd,stdout=f,stderr=subprocess.STDOUT,env={**os.environ,'ADAMIC_GATE_UNCACHED':'1','ADAMIC_BUILD_CACHE_DIR':'/tmp/interface-defense/cache/D8'})
 ev=[]
 for l in (p/'D8-new.log').read_text().splitlines():
  try:ev.append(json.loads(l))
  except:pass
 (p/'D8-new-matrix.json').write_text(json.dumps({'mutant':'D8','command':cmd,'exit':r.returncode,'vet_exit':v.returncode,'wall':time.monotonic()-start,'statuses':{e['Test']:e['Action'] for e in ev if e.get('Test') and '/' not in e['Test'] and e['Action'] in ['pass','fail','skip']}},indent=2))
finally:file.write_text(orig)
print('completed',flush=True)
