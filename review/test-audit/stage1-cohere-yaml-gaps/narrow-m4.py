import shlex
import pathlib,os,subprocess,time,json,signal
p=pathlib.Path('/tmp/u152');root=pathlib.Path('/workspace/adamic');f=root/'stage1/cohere/yaml/lexer.ts';old=f.read_text();records=[]
env=os.environ.copy();env['ADAMIC_YAML_LIBRARY']=str(p/'library');env['ADAMIC_BUILD_CACHE_DIR']=str(p/'cache'/'M4-narrow')
def run(id,cmd):
 start=time.monotonic()
 with (p/(id+'.log')).open('w') as out:r=subprocess.run(cmd,env=env,cwd=root,stdout=out,stderr=subprocess.STDOUT)
 rec={'id':id,'command':shlex.join(cmd),'wall':round(time.monotonic()-start,3),'exit':r.returncode};records.append(rec);(p/'narrow-runs.json').write_text(json.dumps(records,indent=2));print(rec,flush=True)
def orphan_ids():
 ss=subprocess.check_output(['ps','-eo','pid=,ppid=,args='],text=True);return {int(s.split()[0]) for s in ss.splitlines() if len(s.split())>=3 and s.split()[1]=='1' and s.split()[2].startswith('/tmp/adamic-gate/Test')}
(p/'lexer-behavior.ts').write_text("import { Lexer } from '/workspace/adamic/stage1/cohere/yaml/lexer.ts';\nconsole.log(JSON.stringify(new Lexer().lex('[a, -x]')));\n")
cmd=['timeout','2','node','--disable-warning=ExperimentalWarning',str(root/'oracle/node.mjs'),str(p/'lexer-behavior.ts')]
run('M4-behavior-before',cmd)
try:
 f.write_text(old.replace('return unit === 44 || unit === 91','return unit === 45 || unit === 91',1))
 run('M4-behavior-after',cmd)
 for test in ['TestLexerMatchesGo','TestPropsMatchGo']:
  before=orphan_ids()
  run('M4-narrow-'+test,['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run','^'+test+'$'])
  for pid in orphan_ids()-before:
   try:os.killpg(pid,signal.SIGKILL)
   except ProcessLookupError:pass
   with (p/'orphan-cleanup.txt').open('a') as out:out.write(f'Killed new orphaned audit process group {pid} after {test} timed out.\n')
finally:f.write_text(old)
run('restored-baseline',['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/yaml/','-run','^(TestLexerGaps|TestStructuralPositionRefusal|TestClosedStringPresenceGap|TestClosedLexerGaps|TestClosedValuePresenceGap|TestLexerMatchesGo|TestLexerMutants|TestPropsMatchGo|TestPropsMutants|TestSharedSliceAppendMatchesNode|TestScalarMutants)$'])
