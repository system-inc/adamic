import pathlib, subprocess, os, json, re, time, difflib
root=pathlib.Path.cwd(); p=root/'review/test-defend/internal-oracle-checked_views_v2_source/closure-merge'; ref='origin/test-audit/internal-oracle-checked_views_v2_source:'; audit='review/test-audit/internal-oracle-checked_views_v2_source/'
current=[x for x in (p/'list.log').read_text().splitlines() if x.startswith('Test')]
oldlog=subprocess.check_output(['git','show',ref+audit+'list.log']).decode(); (p/'audit-list.log').write_text(oldlog); old=[x for x in oldlog.splitlines() if x.startswith('Test')]
(p/'scope.json').write_text(json.dumps({'current':current,'audit':old,'added':sorted(set(current)-set(old)),'vanished':sorted(set(old)-set(current))},indent=2)+'\n')
def cov(f):
 d={}
 for line in (p/f).read_text().splitlines()[1:]:
  block,n,count=line.split(); d[block]=(int(n),int(count))
 return d
c,a=cov('closure.cover'),cov('array.cover'); exclusive=[k for k in c if c[k][1]>0 and a.get(k,(0,0))[1]==0]; reverse=[k for k in a if a[k][1]>0 and c.get(k,(0,0))[1]==0]
(p/'exclusive-coverage.json').write_text(json.dumps({'closure_only':exclusive,'array_only':reverse},indent=2)+'\n')
(p/'CODE-AND-ORACLE.md').write_text('CODE UNDER TEST: Adamic lowering, specifically lowering.variables and its production Refused diagnostic. ORACLE: unchanged source Node execution (exit 0 and empty stderr, not stdout); self-written complete .refused diagnostic snapshots. No Node, tests, fixtures, pins or harness edit. Coverage identifies locals.go:14 return as target-only relative to the array-pending row. Scanner var syntax is refused at 69:5; the array row reaches NotYet for array-element metadata.\n')
plan=[{'id':'D1','file_line':'internal/lower/locals.go:14','menu':'change a constant','before':'What: "var"','after':'What: "let"','purpose':'Wrong forbidden construct name for non-block-scoped declarations; full diagnostic pin versus unrelated array metadata refusal.'},{'id':'D2','file_line':'internal/lower/locals.go:14','menu':'change a constant','before':'Fix: "use const or let"','after':'Fix: "use var"','purpose':'Incorrect repair for a refused var declaration.'},{'id':'D3','file_line':'internal/lower/optional_widening.go:196','menu':'change a constant','before':'on the source type','after':'on the target type','purpose':'Incorrect repair direction for absent optional source fields in imported Error views.'}]
(p/'plan.json').write_text(json.dumps(plan,indent=2)+'\n')
paths=[]
for line in (p/'baseline.log').read_text().splitlines():
 try: x=json.loads(line)
 except ValueError: continue
 if x.get('Action')=='run' and x.get('Test','').startswith('TestNativeAgreesWithNode/'):
  name=x['Test'].split('/',1)[1]; f=root/name
  if f.is_file() and re.search(r'\bvar\b',f.read_text()): paths.append(name)
(p/'var-corpus.json').write_text(json.dumps(paths,indent=2)+'\n')
rows=sorted(set(current)-{'TestNativeAgreesWithNode','TestCountsAreRecorded'})
patterns=['^('+ '|'.join(rows[i::4])+')$' for i in range(4)]
# Go's slash-separated regexp matches each test path component independently.
components=list(zip(*(x.split('/') for x in paths)))
assert len(set(map(lambda x:len(x.split('/')),paths)))==1
corpus='^TestNativeAgreesWithNode$/'+'/'.join('^('+ '|'.join(re.escape(x) for x in sorted(set(comp)))+')$' for comp in components)
patterns.append(corpus)
(p/'matrix-plan.json').write_text(json.dumps({'batch_patterns':patterns,'top_level_rows':rows+['TestNativeAgreesWithNode'],'corpus_members':paths,'omitted_top_level':['TestCountsAreRecorded'],'note':'All standalone rows plus corpus inputs containing conservative lexical var token. Other corpus inputs and count-table replay are unknown.'},indent=2)+'\n')
env=os.environ.copy(); env.pop('ADAMIC_NATIVE_SPLIT',None); env.pop('ADAMIC_NATIVE_JOBS',None); env['ADAMIC_GATE_UNCACHED']='1'; results=[]
def run(label,cmd,cache):
 e=env.copy();e['ADAMIC_BUILD_CACHE_DIR']='/tmp/defend-closure/cache/'+cache
 started=time.monotonic()
 with (p/(label+'.log')).open('w') as log: r=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,env=e)
 item={'label':label,'command':cmd,'cache':e['ADAMIC_BUILD_CACHE_DIR'],'wall_seconds':time.monotonic()-started,'exit':r.returncode}
 events=[]
 for line in (p/(label+'.log')).read_text().splitlines():
  try: events.append(json.loads(line))
  except ValueError: pass
 item['failed']=[x['Test'] for x in events if x.get('Action')=='fail' and x.get('Test')]; item['passed']=[x['Test'] for x in events if x.get('Action')=='pass' and x.get('Test')]; item['skipped']=[x['Test'] for x in events if x.get('Action')=='skip' and x.get('Test')]
 item['package_elapsed']=[x.get('Elapsed') for x in events if x.get('Action') in ('pass','fail') and not x.get('Test')]
 item['timeout']=any('panic: test timed out' in x.get('Output','') for x in events)
 results.append(item);(p/'results.json').write_text(json.dumps(results,indent=2)+'\n');print(label,r.returncode,round(item['wall_seconds'],3),'failed',item['failed'],flush=True)
 return item
base=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/oracle/','-run']
for i,pattern in enumerate(patterns):
 item=run('clean-batch-'+str(i+1),base+[pattern],'clean')
 if item['exit']!=0: raise RuntimeError('Clean narrowed matrix failed or timed out; no mutant planted')
source=root/'internal/lower/locals.go'; original=source.read_text(); changed=original.replace('What: "var", Fix: "use const or let"','What: "let", Fix: "use const or let"',1); assert changed!=original
(p/'D1.diff').write_text(''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/internal/lower/locals.go',tofile='b/internal/lower/locals.go')))
try:
 source.write_text(changed)
 check=run('D1-vet',['timeout','120','go','vet','./internal/lower/'],'D1'); assert check['exit']==0
 for i,pattern in enumerate(patterns):
  item=run('D1-batch-'+str(i+1),base+[pattern],'D1')
  if item['timeout'] or item['exit'] not in (0,1): raise RuntimeError('Matrix incomplete; narrower replay required')
finally:
 source.write_text(original)
run('restored',base+['^(TestClosureMergeRefusals|TestCheckedViewUntaggedArrayPending)$'],'restored')
