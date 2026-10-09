import pathlib,time,subprocess,os,json
root=pathlib.Path('/tmp/u059')
while not (root/'M8-counts.meta').exists():time.sleep(1)
env=os.environ.copy();env['ADAMIC_MUTANT']='P2';env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u059/cache/P2';env['ADAMIC_GATE_UNCACHED']='1'
cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run','^TestCountsAreRecorded$']
start=time.monotonic()
with (root/'P2-TestCountsAreRecorded.log').open('w') as log:p=subprocess.run(cmd,env=env,stdout=log,stderr=subprocess.STDOUT)
(root/'P2-TestCountsAreRecorded.meta').write_text(json.dumps({'exit':p.returncode,'wall':time.monotonic()-start,'command':' '.join(cmd)}))
