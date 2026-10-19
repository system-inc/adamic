from pathlib import Path
import subprocess,json,time,hashlib
r=Path(__file__).resolve().parents[3];o=Path(__file__).resolve().parent;rows=[]
def run(name,cmd,limit=180):
 start=time.monotonic()
 with (o/(name+'.jsonl')).open('w') as log:
  try:exit=subprocess.run(cmd,cwd=r,stdout=log,stderr=subprocess.STDOUT,timeout=limit).returncode
  except subprocess.TimeoutExpired:exit=124
 rows.append(dict(name=name,command=cmd,exit=exit,seconds=round(time.monotonic()-start,3)));(o/'followup-commands.json').write_text(json.dumps(rows,indent=2)+'\n');print(name,exit,flush=True)
def hashes():return {str(p.relative_to(r)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (r/'internal/oracle/testdata/counts').rglob('*.counts')}
# Run only after the complete oracle package passes.
before=hashes();run('update-counts',['go','test','./internal/oracle','-run','TestCountsAreRecorded','-count=1','-parallel=4','-timeout','600s','-json','-args','-update-counts'],660);after=hashes();assert before==after;print('FULL UPDATE BYTE-IDENTICAL',len(after),flush=True);(o/'regeneration-proof.json').write_text(json.dumps(dict(files=len(after),byte_identical=before==after),indent=2)+'\n')
run('mutants',['python3',str(o/'run-mutants.py')],600)
run('equality-after',['python3',str(o/'equality-proof.py')])
run('ir-guard',['go','test','./internal/ir','-run','TestCallTargetExportInputs|TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted','-count=1','-parallel=4','-timeout','90s','-json'],120)
run('parallel-guard',['go','test','./cmd/adamic-gate','-run','TestEveryTestIsParallelOrSaysWhy','-count=1','-timeout','90s','-json'],120)
run('vet',['go','vet','./internal/oracle'],180)
files=subprocess.check_output(['git','diff','HEAD','--name-only','--diff-filter=ACMR','--','internal/oracle/*.go'],cwd=r,text=True).splitlines()+['internal/oracle/count_files_test.go'];run('gofmt',['gofmt','-l',*files])
