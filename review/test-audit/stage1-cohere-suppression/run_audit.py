import pathlib,subprocess,time,json,os,difflib,re
out=pathlib.Path('review/test-audit/stage1-cohere-suppression');plan=json.loads((out/'frozen-plan.json').read_text());rows=['TestEachGapStandsWhereGapsMdSaysItDoes','TestThePortParsesAsGoCohereDoes'];records=[]
def run(name,cmd,env=None):
 start=time.monotonic()
 with (out/(name+'.log')).open('w') as f:r=subprocess.run(cmd,env=env,stdout=f,stderr=subprocess.STDOUT)
 es=[]
 for line in (out/(name+'.log')).read_text().splitlines():
  try:es.append(json.loads(line))
  except:pass
 rec={'name':name,'command':' '.join(cmd),'wall':time.monotonic()-start,'exit':r.returncode,'fails':[e['Test'] for e in es if e.get('Action')=='fail' and 'Test'in e],'passes':[e['Test'] for e in es if e.get('Action')=='pass' and 'Test'in e],'elapsed':[e.get('Elapsed') for e in es if e.get('Action')in ['pass','fail'] and 'Test'not in e],'cooked':'test timed out' in (out/(name+'.log')).read_text() or r.returncode==124,'panic':'panic:' in (out/(name+'.log')).read_text()};records.append(rec);(out/'runs.json').write_text(json.dumps(records,indent=2)+'\n');print(name,rec['exit'],rec['wall'],rec['fails'],flush=True);return rec
def test(name,pattern='.',env=None):return run(name,['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/suppression/','-run',pattern],env)
def diff(ident,file,original,modified):
 (out/(ident+'.patch')).write_text(''.join(difflib.unified_diff(original.splitlines(True),modified.splitlines(True),fromfile='a/'+file,tofile='b/'+file)))
for row in rows:
 for i in range(3):
  r=test('timing-'+row+'-'+str(i+1),'^'+row+'$')
  if r['exit']:raise SystemExit('red baseline timing')
for m in plan['mutants']:
 f=pathlib.Path(m['file']);original=f.read_text();assert original.count(m['from'])==1;modified=original.replace(m['from'],m['to']);diff(m['id'],str(f),original,modified)
 env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u138/cache/'+m['id'])
 try:
  f.write_text(modified)
  if m['id']=='M4':run(m['id']+'-vet',['timeout','90','go','vet','./internal/native/'],env)
  target='stage1/cohere/suppression/gaps/2_return_undefined_array.ts' if m['id']=='M4' else 'stage1/cohere/suppression/main.ts'
  run(m['id']+'-build',['timeout','90','go','run','./cmd/adamic','build',target,'-o','/tmp/u138/'+m['id'],'--sanitize'],env)
  r=test(m['id'],env=env)
  if r['panic'] or r['cooked']:
   for row in rows:test(m['id']+'-isolated-'+row,'^'+row+'$',env)
 finally:f.write_text(original)
# Valid empty-index probe, without unreachable code interfering with type narrowing.
f=pathlib.Path('stage1/cohere/suppression/suppression.ts');original=f.read_text();start=original.index('export function build(');modified=original[:start]+"export function build(sourceText: string): Index {\n\treturn new Index([], [], buildLineIndex(sourceText));\n}\n";diff('P1',str(f),original,modified)
try:
 f.write_text(modified);env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u138/cache/P1');run('P1-build',['timeout','90','go','run','./cmd/adamic','build','stage1/cohere/suppression/main.ts','-o','/tmp/u138/P1','--sanitize'],env);test('P1',env=env)
finally:f.write_text(original)
# Empty IR keeps the emitter runnable while eliminating every lowered statement/output.
f=pathlib.Path('internal/lower/lower.go');original=f.read_text();start=original.index('func Lower(');end=original.index('\ntype lowering struct');modified=original[:start]+'func Lower(ctx context.Context, program *load.Program) (*ir.Program, error) {\n\treturn &ir.Program{}, nil\n}\n'+original[end:];modified=modified.replace('\t"fmt"\n','').replace('\t"path/filepath"\n','');diff('P2',str(f),original,modified)
try:
 f.write_text(modified);env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u138/cache/P2');run('P2-vet',['timeout','90','go','vet','./internal/lower/'],env);test('P2',env=env)
finally:f.write_text(original)
# Mixed positive/witness row: disable comparison, observe each subcase rather than treating all as witnesses.
f=pathlib.Path('stage1/cohere/suppression/suppression_test.go');original=f.read_text();start=original.index('func firstDifference(');end=original.index('// lowered checks',start);modified=original[:start]+'func firstDifference(got string, want string) string { return "" }\n\n'+original[end:];diff('W1',str(f),original,modified)
try:
 f.write_text(modified);run('W1-vet',['timeout','90','go','vet','./stage1/cohere/suppression/']);test('W1')
finally:f.write_text(original)
index='/tmp/u138-index';env=dict(os.environ,GIT_INDEX_FILE=index)
with (out/'apply-check.log').open('w') as log:
 for patch in sorted(out.glob('*.patch')):
  subprocess.run(['git','read-tree',plan['commit']],env=env,stdout=log,stderr=log,check=True);log.write(str(patch)+'\n');log.flush();subprocess.run(['git','apply','--cached','--check',str(patch)],env=env,stdout=log,stderr=log,check=True)
test('final-clean')
