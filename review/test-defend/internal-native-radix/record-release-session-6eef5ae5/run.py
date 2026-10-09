import pathlib,json,os,subprocess,time,difflib,re
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/internal-native-radix';(p/'diffs').mkdir(exist_ok=True)
plan=[('B1','heap.c','\teach->free = NULL;\n','','Drop fresh/spare-chunk free-list reset, targeting bulk allocation and reuse'),('B2','map.c','\tfree(map->entries);\n\tfree(map->buckets);','\tfree(map->entries);','Drop bucket storage destruction, targeting record workload cleanup'),('B3','map.c','for (size_t bucket = hash_key(map, key) & mask;; bucket = (bucket + 1) & mask)','for (size_t bucket = hash_key(map, key) & mask;; bucket = (bucket + 2) & mask)','Change collision lookup stride, targeting bulk hit/miss workloads'),('R1','count.h','#define ADAMIC_COUNT_FREE() (adamic_counted.frees++, adamic_counted.live--)','#define ADAMIC_COUNT_FREE() (adamic_counted.frees++)','Drop immediate live-count decrement, contrasting direct live assertions with exit counts'),('R2','class_inheritance.c','for (size_t index = 0; index < object->shape->count; index++)','for (size_t index = 1; index < object->shape->count; index++)','Off-by-one first child, targeting two-reference chain destruction'),('R3','heap.c','\tdraining = false;\n','','Drop drain reset, targeting a second release after earlier drain completion')]
items=[]
for id,file,old,new,aim in plan:
 file='internal/native/runtime/'+file;s=(root/file).read_text();assert s.count(old)==1;(p/'diffs'/(id+'.diff')).write_text(''.join(difflib.unified_diff(s.splitlines(True),s.replace(old,new).splitlines(True),fromfile='a/'+file,tofile='b/'+file)));items.append(dict(id=id,file=file,line=s[:s.index(old)].count('\n')+1,old=old,new=new,aim=aim))
(p/'plan.json').write_text(json.dumps(items,indent=2))
# Include every current row in direct runtime callers, plus original audited regexp/radix competitors.
files=['record_test.go','runtime_profile_test.go','heap_test.go','string_views_test.go','map_hash_test.go','regexp_test.go','radix_test.go','library_test.go']
rows=[]
for file in files:
 rows+=['TestRuntimeKeyKeepsBoundaries'] if file=='library_test.go' else re.findall(r'^func (Test\w+)\(', (root/'internal/native'/file).read_text(),re.M)
(p/'matrix-rows.json').write_text(json.dumps(rows,indent=2));env=os.environ.copy();env['ADAMIC_RECORD_BENCH']='1';env['TMPDIR']='/tmp/defend-radix';runs=json.loads((p/'runs.json').read_text()) if (p/'runs.json').exists() else []
def run(id,label,pattern):
 env['ADAMIC_BUILD_CACHE_DIR']='/workspace/scratch/defend-radix/cache/'+id;cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./internal/native/','-run',pattern];start=time.monotonic();log=id+'-'+label+('-retry' if id=='clean' else '')+'.log'
 with (p/log).open('w') as f:r=subprocess.run(cmd,env=env,cwd=root,stdout=f,stderr=subprocess.STDOUT)
 events=[]
 for l in (p/log).read_text().splitlines():
  try:events.append(json.loads(l))
  except:pass
 runs.append(dict(id=id,label=label,command=cmd,log=log,exit=r.returncode,wall=time.monotonic()-start,failed=[e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test')],passed=[e['Test'] for e in events if e.get('Action')=='pass' and e.get('Test')],errors=[e.get('Output') for e in events if e.get('OutputType')=='error'],timeout=any('panic: test timed out' in e.get('Output','') for e in events)));(p/'runs.json').write_text(json.dumps(runs,indent=2));print(id,label,r.returncode,round(time.monotonic()-start,2),flush=True);return runs[-1]
r=run('clean','matrix','^('+'|'.join(rows)+')$');assert not r['failed'],'red bounded baseline'
for m in items:
 diff=p/'diffs'/(m['id']+'.diff');subprocess.run(['git','apply',str(diff)],cwd=root,check=True)
 try:
  # Explicit compile validation with the runtime's C11 warning flags, both standard modes.
  validation=[]
  for sanitize in [False,True]:
   flags=['-std=c11','-Wall','-Wextra','-Werror','-Wcast-function-type-strict','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-DADAMIC_COUNT','-I',str(root/'internal/native/runtime')]+(['-O1','-g','-fsanitize=address,undefined','-fno-sanitize-recover=all'] if sanitize else ['-O2'])
   file=m['file'] if not m['file'].endswith('.h') else 'internal/native/runtime/heap.c';cmd=['clang']+flags+['-c',file,'-o','/workspace/scratch/defend-radix/'+m['id']+('-san' if sanitize else '')+'.o']
   with (p/(m['id']+('-san' if sanitize else '')+'-compile.log')).open('w') as f:c=subprocess.run(cmd,cwd=root,stdout=f,stderr=subprocess.STDOUT)
   assert c.returncode==0,(m['id'],'compile');validation.append(dict(command=cmd,exit=c.returncode))
  (p/(m['id']+'-validation.json')).write_text(json.dumps(validation,indent=2))
  r=run(m['id'],'matrix','^('+'|'.join(rows)+')$');completed=set(r['failed']+r['passed'])
  for name in ['TestRecordBenchmark','TestRuntimeReleasePaths']:
   if name not in completed:run(m['id'],name,'^'+name+'$')
 finally:subprocess.run(['git','apply','-R',str(diff)],cwd=root,check=True)
