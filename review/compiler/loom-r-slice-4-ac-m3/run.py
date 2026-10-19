import pathlib,subprocess,json,time,shutil,os
r=pathlib.Path.cwd();o=r/'review/compiler/loom-r-slice-4-ac-m3';(o/'logs').mkdir(exist_ok=True);results=[]
def run(n,c,limit=300):
 start=time.monotonic()
 with (o/'logs'/(n+'.jsonl')).open('w') as f:
  try:code=subprocess.run(c,cwd=r,stdout=f,stderr=subprocess.STDOUT,timeout=limit).returncode
  except subprocess.TimeoutExpired:code=124
 results.append(dict(name=n,command=c,exit=code,seconds=round(time.monotonic()-start,3)));(o/'commands.json').write_text(json.dumps(results,indent=2)+'\n');print(n,code,results[-1]['seconds'],flush=True)
run('counts',['go','test','./internal/oracle','-run','TestCountsAreRecorded','-count=1','-parallel=4','-timeout','600s','-json','-args','-update-counts'],660)
run('lower',['go','test','./internal/lower/...','-count=1','-parallel=4','-timeout','300s','-json'],330)
run('oracle',['go','test','./internal/oracle/...','-count=1','-parallel=4','-timeout','1200s','-json'],1250)
relocations=[]
for group in ['Native','Lower','Fresh','Flow','Oracle']:
 run('ir-'+group.lower(),['go','test','./internal/ir','-run','^(TestCallTargetExportInputs'+group+'|TestCallTargetReaders'+group+'|TestProduct_CallTargetReaderArchives'+group+')$','-count=1','-parallel=4','-timeout','90s','-json'],120)
 for p in pathlib.Path(os.environ['ADAMIC_BUILD_OUTPUT_DIR']).glob('local/*/exports'):
  size=sum(x.stat().st_size for x in p.rglob('*') if x.is_file());dest=pathlib.Path(os.environ['ADAMIC_BUILD_CACHE_DIR'])/'local'/p.parent.name
  if shutil.disk_usage(dest.parent).free>size+1300000000 and not dest.exists():
   shutil.move(str(p.parent),str(dest));relocations.append(str(dest));(o/'cache-relocations.json').write_text(json.dumps(relocations,indent=2)+'\n')
run('ir-guards',['go','test','./internal/ir','-run','TestCallTargetReaders|TestEveryMapAProgramHoldsIsSorted','-count=1','-parallel=4','-timeout','90s','-json'],120)
run('gate-guards',['go','test','./cmd/adamic-gate','-run','TestEveryTestIsParallelOrSaysWhy|Test.*Product','-count=1','-timeout','90s','-json'],120)
run('dependency-guards',['go','test','./internal/compilercensus','-run','TestDependencies','-count=1','-timeout','90s','-json'],120)
run('presence-mutant',['go','test','./internal/lower','-overlay',str(o/'omit-presence-overlay.json'),'-run','^TestEEPPresence[0-6]$','-count=1','-timeout','90s','-json'],120)
files=subprocess.check_output(['git','diff','HEAD','--name-only','--diff-filter=ACMR','--','*.go'],text=True).splitlines();pkgs=sorted({'./'+str(pathlib.Path(f).parent)+'/...' for f in files});run('vet',['go','vet',*pkgs],300);run('gofmt',['gofmt','-l',*files],90)
