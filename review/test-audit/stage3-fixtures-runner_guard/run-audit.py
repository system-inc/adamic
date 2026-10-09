import pathlib,subprocess,time,json,os,difflib,shlex
p=pathlib.Path('/tmp/u164');root=pathlib.Path('/workspace/adamic');f=root/'stage3/fixtures/fixtures_test.go';original=f.read_text();records=[]
guard='if erasable+transformedCalls != 1 || strings.Count(text, "stripTypeScriptTypes(source") != 1 || strings.Count(text, "new URL(\'./adamic.mjs\', import.meta.url)") != 1 {'
assert original.count(guard)==1
mutants=[('W1',guard,'if erasable+transformedCalls < 0 {'),('W2','strings.Count(text, "new URL(\'./adamic.mjs\', import.meta.url)") != 1','strings.Count(text, "new URL(\'./adamic.mjs\', import.meta.url)") > 1'),('W3','strings.Count(text, "stripTypeScriptTypes(source") != 1','strings.Count(text, "stripTypeScriptTypes(source") > 2'),('S1','func TestFixturesAssertions(t *testing.T)      { testFixtureDirectory(t, "assertions") }','func TestFixturesAssertions(t *testing.T)      {}'),('S2','func TestFixturesCycles(t *testing.T)          { testFixtureDirectory(t, "cycles") }','func TestFixturesCycles(t *testing.T)          { testFixtureDirectory(t, "enums") }'),('S3','func TestFixturesTaste(t *testing.T)           { testFixtureDirectory(t, "taste") }','func TestFixturesTaste(t *testing.T)           { return }')]
def run(id,cmd,env=None):
 start=time.monotonic()
 with (p/(id+'.log')).open('w') as out:r=subprocess.run(cmd,cwd=root,env=env,stdout=out,stderr=subprocess.STDOUT)
 rec={'id':id,'command':shlex.join(cmd),'wall':round(time.monotonic()-start,4),'exit':r.returncode};
 if env and env.get('ADAMIC_RUNNER_GUARD_REPOSITORY'):
  rec['environment']={'ADAMIC_RUNNER_GUARD_REPOSITORY':env['ADAMIC_RUNNER_GUARD_REPOSITORY']}
  rec['command']='ADAMIC_RUNNER_GUARD_REPOSITORY='+shlex.quote(env['ADAMIC_RUNNER_GUARD_REPOSITORY'])+' '+rec['command']
 records.append(rec);(p/'runs.json').write_text(json.dumps(records,indent=2));print(rec,flush=True);return r.returncode
def apply(id,new):
 diff=''.join(difflib.unified_diff(original.splitlines(True),new.splitlines(True),fromfile='a/stage3/fixtures/fixtures_test.go',tofile='b/stage3/fixtures/fixtures_test.go'));(p/(id+'.diff')).write_text(diff)
 assert run(id+'-apply-check',['git','apply','--check',str(p/(id+'.diff'))])==0
 f.write_text(new);assert run(id+'-vet',['timeout','90','go','vet','./stage3/fixtures/'])==0
repo=p/'hybrid-runner';(repo/'oracle').mkdir(parents=True,exist_ok=True)
(repo/'oracle/node.mjs').write_text("stripTypeScriptTypes(source, { mode: 'transform' });stripTypeScriptTypes(source, { mode: 'unknown' });new URL('./adamic.mjs', import.meta.url)")
env=os.environ.copy();env['ADAMIC_RUNNER_GUARD_REPOSITORY']=str(repo)
hybrid_cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage3/fixtures/','-run','^TestTransformedNodeRunnerGuardHook$']
run('W3-hybrid-before',hybrid_cmd,env)
try:
 for id,a,b in mutants:
  assert original.count(a)==1,(id,original.count(a))
  try:
   apply(id,original.replace(a,b,1))
   run(id,['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage3/fixtures/','-run','.'])
   if id=='W3':run('W3-hybrid-after',hybrid_cmd,env)
  finally:f.write_text(original)
 def empty_runner(s):
  start=s.index('\n',s.index('func transformedNodeRunner('));end=s.index('\n}\n',start)
  return (s[:start]+'\n\treturn ""'+s[end:]).replace('\t"net/url"\n','')
 try:
  apply('P1',empty_runner(original));active=os.environ.copy();active['ADAMIC_RUNNER_GUARD_REPOSITORY']=str(root)
  run('P1-hook',hybrid_cmd,active)
  run('P1-parent',['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage3/fixtures/','-run','^TestTransformedNodeRunnerGuard$'])
 finally:f.write_text(original)
 try:
  new=''.join(line for line in original.splitlines(True) if not line.startswith('func TestFixtures'))
  apply('P2',new)
  run('P2',['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage3/fixtures/','-run','^TestFixtureDirectoriesHaveTopLevelTests$'])
 finally:f.write_text(original)
 run('restored-baseline',['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage3/fixtures/','-run','.'])
finally:f.write_text(original)
