import pathlib,json,re,subprocess,time
root=pathlib.Path.cwd();evidence=root/'review/compiler/loom-r-hidden-boundaries-ac';results=[]
def run(label,cmd):
 start=time.monotonic()
 with (evidence/(label+'.log')).open('w') as log:p=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=180)
 tests=[]
 for line in (evidence/(label+'.log')).read_text().splitlines():
  try:e=json.loads(line)
  except ValueError:continue
  if isinstance(e,dict) and e.get('Test') and e.get('Action') in ('pass','fail','skip'):tests.append(dict(name=e['Test'],action=e['Action'],seconds=e.get('Elapsed',0)))
 result=dict(label=label,command=cmd,exit=p.returncode,seconds=round(time.monotonic()-start,3),tests=tests);results.append(result)
 (evidence/'tests.json').write_text(json.dumps(results,indent=2)+'\n');print(label,p.returncode,result['seconds'],flush=True)
 if p.returncode:raise SystemExit(p.returncode)
 assert all(t['seconds']<60 for t in tests),result
files=json.loads((evidence/'reader-files.json').read_text());groups={}
for name in files:groups.setdefault(str(pathlib.Path(name).parent),[]).extend(re.findall(r'^func (Test\w+)\(t \*testing.T\)',pathlib.Path(name).read_text(),re.M))
for package,names in sorted(groups.items()):run(package.replace('/','-'),['timeout','150','go','test','./'+package,'-run','^('+'|'.join(sorted(set(names)))+')$','-count=1','-timeout','90s','-parallel','4','-json'])
run('ir-readers',['timeout','150','go','test','./internal/ir','-run','^TestCallTargetReaders$','-count=1','-timeout','90s','-json'])
run('vet',['timeout','150','go','vet','./cmd/adamic','./internal/flow','./internal/ir','./internal/lower','./internal/oracle','./internal/native/runtimeproducts'])
run('diff-check',['timeout','30','git','diff','--check'])
