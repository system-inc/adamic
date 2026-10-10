import concurrent.futures,json,os,pathlib,re,subprocess,time
ROOT=pathlib.Path('/workspace/adamic');OUT=ROOT/'review/compiler/views-v4/sweep'
def run(job):
 name,args=job;start=time.monotonic()
 with (OUT/(name+'.log')).open('w') as log:
  try: status=subprocess.run(['timeout','90',*args],cwd=ROOT,stdout=log,stderr=subprocess.STDOUT).returncode
  except Exception as error: log.write(str(error));status=125
 return dict(name=name,command=args,status=status,seconds=round(time.monotonic()-start,3))
mode=os.sys.argv[1]
if mode=='tests':
 names=[x for x in (OUT/'oracle-list.log').read_text().splitlines() if re.fullmatch(r'Test\w+',x)]
 names += sorted({n for p in (ROOT/'internal/oracle').glob('*view*test.go') if not p.name.startswith('review') for n in re.findall(r'^func (Test\w+)\(',p.read_text(),re.M)} - set(names))
 records=[x for x in (OUT/'records-list.log').read_text().splitlines() if re.fullmatch(r'Test\w+',x)]
 jobs=[('oracle-'+n,['go','test','./internal/oracle','-run','^'+n+'$','-count=1','-timeout','90s','-v']) for n in names]
 jobs += [('records-'+n,['go','test','./internal/native','-run','^'+n+'$','-count=1','-timeout','90s','-v']) for n in records if n!='TestRecordBenchmark']
 packages=set()
 for path in (ROOT/'stage1').rglob('GAPS.md'):
  if re.search(r'\bcallables?\b|\bviews?\b',path.read_text(),re.I): packages.add(str(path.parent.relative_to(ROOT)))
 (OUT/'gaps-packages.json').write_text(json.dumps(sorted(packages),indent=2))
 for package in sorted(packages):
  listing=subprocess.run(['timeout','90','go','test','./'+package,'-list','Gap'],cwd=ROOT,capture_output=True,text=True)
  (OUT/('list-'+package.replace('/','-')+'.log')).write_text(listing.stdout+listing.stderr)
  for n in listing.stdout.splitlines():
   if re.fullmatch(r'Test\w+',n):jobs.append(('gaps-'+package.replace('/','-')+'-'+n,['go','test','./'+package,'-run','^'+n+'$','-count=1','-timeout','90s','-v']))
 jobs.append(('counts',['go','test','./internal/oracle','-run','^TestCountsAreRecorded$','-count=1','-timeout','90s','-v']))
 (OUT/'test-plan.json').write_text(json.dumps(jobs,indent=2))
 completed={json.loads(line)['name'] for line in (OUT/'tests.jsonl').read_text().splitlines()} if (OUT/'tests.jsonl').exists() else set()
 jobs=[job for job in jobs if job[0] not in completed]
 with (OUT/'tests.jsonl').open('a') as results,concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
  for result in pool.map(run,jobs):results.write(json.dumps(result)+'\n');results.flush();print(result['name'],result['status'],result['seconds'],flush=True)
else:
 paths=sorted([p for directory in ['internal/oracle','stage3'] for p in (ROOT/directory).rglob('*.a')])
 (OUT/'fixtures.json').write_text(json.dumps([str(p.relative_to(ROOT)) for p in paths],indent=2))
 def check(path):
  row={'path':str(path.relative_to(ROOT))}
  for label,binary in [('main','/tmp/adamic-main-sweep-check'),('v4','/tmp/adamic-v4-sweep-check')]:
   start=time.monotonic()
   try:
    got=subprocess.run(['timeout','60',binary,str(path)],cwd=ROOT,capture_output=True,text=True,timeout=65)
    row[label]=json.loads(got.stdout) if got.returncode==0 else {'admitted':False,'stage':'probe','error':got.stderr,'exit':got.returncode}
   except Exception as error:row[label]={'admitted':False,'stage':'probe','error':str(error)}
   row[label]['seconds']=round(time.monotonic()-start,3)
  return row
 with (OUT/'admission.jsonl').open('w') as results,concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:
  for i,result in enumerate(pool.map(check,paths),1):results.write(json.dumps(result)+'\n');results.flush()
