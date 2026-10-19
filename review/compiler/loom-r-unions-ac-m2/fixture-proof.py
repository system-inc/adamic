import json,os,subprocess,time
from pathlib import Path
out=Path(__file__).resolve().parent
results=[]
def run(label,cmd,want=0):
 start=time.monotonic();r=subprocess.run(cmd,capture_output=True,timeout=90,env=dict(os.environ,ASAN_OPTIONS='detect_leaks=1:halt_on_error=1',UBSAN_OPTIONS='halt_on_error=1'));(out/(label+'.stdout.txt')).write_bytes(r.stdout);(out/(label+'.stderr.log')).write_bytes(r.stderr);results.append(dict(label=label,command=cmd,exit=r.returncode,seconds=round(time.monotonic()-start,3)));assert r.returncode==want,(label,r.stderr);return r
positive='cmd/adamic-meter/testdata/compiled/main.a';negative='cmd/adamic-meter/testdata/corpus/not_yet.a'
node=run('meter-positive-node',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(Path(positive).resolve())]);assert node.stdout==b'true\n' and not node.stderr
run('meter-positive-build',['/workspace/adamic-binaries/adamic-acheck','build',positive,'-o','/workspace/scratch/unions-m2-meter-positive','--sanitize'])
native=run('meter-positive-native',['/workspace/scratch/unions-m2-meter-positive']);assert (native.stdout,native.stderr)==(node.stdout,node.stderr)
node=run('meter-notyet-node',['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(Path(negative).resolve())]);assert node.stdout==b'true\n' and not node.stderr
stop=run('meter-notyet-status',['/workspace/adamic-binaries/adamic-acheck','c',negative],1);assert b"can't lower a nullable object lookup without separate null/undefined tags yet" in stop.stderr
(out/'fixture-proof.json').write_text(json.dumps(results,indent=2));print('PASS: exact positive witness agrees with Node under ASan/UBSan/leaks; negative witness is Node-valid and NotYet')
