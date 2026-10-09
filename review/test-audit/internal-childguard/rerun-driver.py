import pathlib,time,json,subprocess,os,concurrent.futures
p=pathlib.Path('/workspace/adamic/review/test-audit/internal-childguard')
while True:
 f=p/'runs.json'
 if f.exists() and any(x['id']=='M09' for x in json.loads(f.read_text())): break
 time.sleep(1)
rows=['TestChild','TestProgress','TestStalled','TestCeiling','TestExitIsNotGuardError','TestKillsProcessGroup','TestNoFirstOutput']
env=os.environ.copy(); env['ADAMIC_MUTANT']='M09'
def run(row):
 with (p/('M09.'+row+'.log')).open('w') as log:
  r=subprocess.run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/childguard/','-run','^'+row+'$'],cwd='/workspace/adamic',env=env,stdout=log,stderr=subprocess.STDOUT)
 return dict(test=row,exit=r.returncode)
with concurrent.futures.ThreadPoolExecutor(max_workers=7) as pool: results=list(pool.map(run,rows))
(p/'M09.reruns.json').write_text(json.dumps(results,indent=2))
