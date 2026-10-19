from pathlib import Path
import subprocess,json,time,hashlib
r=Path(__file__).resolve().parents[3];o=Path(__file__).resolve().parent;results=[]
def run(name,cmd,limit):
 start=time.monotonic()
 with (o/(name+'.jsonl')).open('w') as log:
  try: code=subprocess.run(cmd,cwd=r,stdout=log,stderr=subprocess.STDOUT,timeout=limit).returncode
  except subprocess.TimeoutExpired: code=124
 results.append(dict(name=name,command=cmd,exit=code,seconds=round(time.monotonic()-start,3)));(o/'commands.json').write_text(json.dumps(results,indent=2)+'\n');print(name,code,results[-1]['seconds'],flush=True);assert code==0,name
def hashes():return {str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (r/'internal/oracle/testdata/counts').rglob('*.counts')}
before=hashes();run('update-counts',['go','test','./internal/oracle','-run','TestCountsAreRecorded','-count=1','-parallel=4','-timeout','600s','-json','-args','-update-counts'],660);after=hashes();assert before==after;(o/'regeneration-proof.json').write_text(json.dumps(dict(files=len(after),byte_identical=True),indent=2)+'\n')
run('equality-proof',['python3',str(o/'equality-proof.py')],30)
run('determinism-census',['go','test','./internal/oracle','-run','^(TestCountsFilesCensus|TestCountsRegenerateDeterministically|TestPredicateCountsRegenerateDeterministically)$','-count=1','-parallel=4','-timeout','90s','-json'],120)
run('oracle',['go','test','./internal/oracle','-count=1','-parallel=4','-timeout','1200s','-json'],1250)
run('mutants',['python3',str(o/'run-mutants.py')],600)
run('ir-guard',['go','test','./internal/ir','-run','TestCallTargetExportInputs|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted','-count=1','-parallel=4','-timeout','90s','-json'],120)
run('parallel-guard',['go','test','./cmd/adamic-gate','-run','TestEveryTestIsParallelOrSaysWhy','-count=1','-timeout','90s','-json'],120)
run('vet',['go','vet','./internal/oracle'],180)
files=subprocess.check_output(['git','diff','HEAD','--name-only','--diff-filter=ACMR','--','internal/oracle/*.go'],cwd=r,text=True).splitlines();run('gofmt',['gofmt','-l',*files],30);assert not (o/'gofmt.jsonl').read_text().strip()
run('equality-final',['python3',str(o/'equality-proof.py')],30)
