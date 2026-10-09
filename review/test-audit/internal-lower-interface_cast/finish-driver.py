import pathlib,json,time,subprocess,os,shutil
root=pathlib.Path('/workspace/adamic');p=root/'review/test-audit/internal-lower-interface_cast'
while True:
 f=p/'runs.json'
 if f.exists():
  try:runs=json.loads(f.read_text())
  except: runs=[]
  if len(runs)==14:break
 time.sleep(1)
menu=json.loads((p/'menu.json').read_text());files=sorted({m['file'] for m in menu})
subprocess.run(['git','restore','--']+files,cwd=root,check=True)
(root/'internal/lower/u033_audit_mutant.go').unlink()
for m in menu:
 r=subprocess.run(['git','apply','--check',str(p/(m['id']+'.diff'))],cwd=root,capture_output=True,text=True)
 assert r.returncode==0,(m['id'],r.stderr)
with (p/'standalone-apply-check.log').open('w') as log:log.write('All 12 mutants and both probes apply cleanly to starting origin/main.\n')
env=os.environ.copy();env['U033_MEASURE_SUBSUMERS']='1'
with (p/'report-generation.log').open('w') as log:subprocess.run(['python3','/tmp/u033-report.py'],cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT,check=True)
for src,dst in [('/tmp/u033-audit.py','audit-driver.py'),('/tmp/u033-scope.py','scope-driver.py'),('/tmp/u033-report.py','report-driver.py'),('/tmp/u033-finish.py','finish-driver.py')]:shutil.copyfile(src,p/dst)
subprocess.run(['gofmt','-w',str(p/'witness/main.go')],check=True)
print('Restored production code, all diffs apply, report and additional medians complete.',flush=True)
