import pathlib,subprocess,json,time
r=pathlib.Path('/workspace/adamic');o=r/'review/compiler/loom-r-slice-5-ac2-m2';rows=[]
def run(n,c,limit=180):
 with (o/'logs'/(n+'.jsonl')).open('w') as f:
  try:x=subprocess.run(c,cwd=r,stdout=f,stderr=subprocess.STDOUT,timeout=limit).returncode
  except subprocess.TimeoutExpired:x=124
 rows.append(dict(name=n,command=c,exit=x));(o/'additional-commands.json').write_text(json.dumps(rows,indent=2)+'\n')
 print(n,x,flush=True)
run('tsgo-product-prepare',['go','test','./internal/native','-run','^TestProduct_(TSGoArchive|SplitTSGoSource|SplitTSGoBaseline)$','-count=1','-parallel=4','-timeout','150s','-json'],180)
run('tsgo-behavior-retry',['go','test','./internal/native','-run','^TestSplitTSGoAgreesUnit0[01]$','-count=1','-parallel=4','-timeout','90s','-json'])
for j in json.loads((o/'mutant-jobs.json').read_text()):run(j['name'],['go','test',j['package'],'-overlay',str(o/(j['name']+'-overlay.json')),'-run',j['test'],'-count=1','-timeout','90s','-json'])
pin=r/'internal/oracle/testdata/review/agree/non_null.expected.json';original=pin.read_bytes()
try:
 data=json.loads(original);data['exit']=71;pin.write_text(json.dumps(data)+'\n')
 run('wrong-outcome-pin',['go','test','./internal/oracle','-run','^TestReviewProgramsAgreeWithNode_001$/^non_null[.]a$','-count=1','-timeout','90s','-json'])
finally:pin.write_bytes(original)
run('parallel-guard',['go','test','./cmd/adamic-gate','-run','TestEveryTestIsParallelOrSaysWhy|Test.*Product','-count=1','-timeout','90s','-json'])
run('ir-readers',['go','test','./internal/ir','-run','^TestCallTargetReaders','-count=1','-parallel=4','-timeout','90s','-json'])
files=subprocess.check_output(['git','diff','HEAD','--name-only','--diff-filter=ACMR','--','*.go'],cwd=r,text=True).splitlines();pkgs=sorted({'./'+str(pathlib.Path(x).parent)+'/...' for x in files});run('vet',['go','vet',*pkgs],300);run('gofmt',['gofmt','-l',*files])
run('consumers',['python3',str(o/'consumers.py')],3600)
