import json,subprocess,time,os,difflib,re,signal
from pathlib import Path
p=Path('review/test-audit/stage1-cohere-lint-rules-no-underscore-dangle');owned=Path('stage1/cohere/lint/rules/no-underscore-dangle');base=(p/'base.txt').read_text().strip();package='./stage1/cohere/lint/rules/no-underscore-dangle/'
menu=[('M1','rule.a',"name.startsWith('_')","name.startsWith('$')",'constant change'),('M2','rule.a',"this.context.settings.read(name, fallback) === 'true'","this.context.settings.read(name, fallback) !== 'true'",'condition flip'),('M3','profile.a','    visit(context, rule, root);','', 'drop statement'),('M4','rule.a','children[offset + 1] ?? -1','children[offset + 0] ?? -1','off-by-one bound'),('P1','profile.a','function run(row: string, countOnly: boolean): number {','function run(row: string, countOnly: boolean): number {\n    if (row.length >= 0) return 0;','empty-answer probe')]
functions={}
for name in ('rule.a','profile.a'):
 s=(owned/name).read_text();functions[name]=re.findall(r'(?:function\s+|^\s*)(\w+)\([^\n]*?\)\s*(?::[^\n{]+)?\{',s,re.M)
(p/'inventory-menu.json').write_text(json.dumps({'code_under_test':'owned .a port lowered and compiled; no port function is executed by the Go row','oracle':'self: load, lower and native.Build success; no external authority or runtime comparison','functions_lowered':functions,'menu':[{'id':m,'file':str(owned/f),'from':a,'to':b,'operator':op} for m,f,a,b,op in menu]},indent=2)+'\n')
fixture=p/'witness.ts';fixture.write_text('const _lead = 1; const tail_ = 2; const plain = 3; this._x; function _fn(_arg) {} const { plain: _bound } = obj;\n')
manifest=p/'witness-manifest.txt';manifest.write_text(str(fixture.resolve())+'\t'+json.dumps({'allowInObjectDestructuring':False,'allowFunctionParams':False})+'\n')
node=['node','--disable-warning=ExperimentalWarning','oracle/node.mjs',str(owned/'profile.a'),str(manifest)]
def run(id,cmd,env=None):
 start=time.monotonic()
 with (p/(id+'.log')).open('w') as out:
  proc=subprocess.Popen(cmd,stdout=out,stderr=subprocess.STDOUT,env=env,start_new_session=True);rc=proc.wait()
  try:os.killpg(proc.pid,signal.SIGKILL)
  except ProcessLookupError:pass
 with (p/'commands.jsonl').open('a') as out:out.write(json.dumps({'id':id,'command':cmd,'returncode':rc,'wall_seconds':time.monotonic()-start})+'\n')
 return rc
assert run('witness-baseline',node)==0
for repeat in range(3):
 rc=run('timing-'+str(repeat),['timeout','120','go','test','-json','-count=1','-timeout','90s',package,'-run','^TestCompileProfiles$'])
 if rc:raise RuntimeError('red timing baseline')
for mid,name,old,new,operator in menu:
 file=owned/name;original=subprocess.check_output(['git','show',base+':'+str(file)],text=True);assert original.count(old)==1,(mid,old)
 changed=original.replace(old,new)
 diff=''.join(difflib.unified_diff(original.splitlines(True),changed.splitlines(True),fromfile='a/'+str(file),tofile='b/'+str(file)))
 (p/(mid+'.diff')).write_text(diff)
 with (p/'mutants.jsonl').open('a') as out:out.write(json.dumps({'id':mid,'file':str(file),'line':original[:original.index(old)].count('\n')+1,'from':old,'to':new,'operator':operator})+'\n')
 subprocess.run(['git','apply',str(p/(mid+'.diff'))],check=True)
 try:
  env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/u121/cache/'+mid
  rc=run(mid,['timeout','120','go','test','-json','-count=1','-timeout','90s',package,'-run','.'],env)
  witness_rc=run('witness-'+mid,node)
  events=[]
  for line in (p/(mid+'.log')).read_text().splitlines():
   try:events.append(json.loads(line))
   except:pass
  baseline=(p/'witness-baseline.log').read_text();after=(p/('witness-'+mid+'.log')).read_text()
  before_findings=[l for l in baseline.splitlines() if l.startswith('finding ')];after_findings=[l for l in after.splitlines() if l.startswith('finding ')]
  result={'id':mid,'kind':'probe' if mid.startswith('P') else 'production','returncode':rc,'compiled':rc==0,'failures':[e['Test'] for e in events if e.get('Action')=='fail' and e.get('Test')],'test_seconds':[e['Elapsed'] for e in events if e.get('Action')=='pass' and e.get('Test')=='TestCompileProfiles'],'witness_returncode':witness_rc,'witness_changed':baseline!=after,'before_findings':len(before_findings),'after_findings':len(after_findings),'witness_diff':list(difflib.unified_diff(baseline.splitlines(),after.splitlines(),fromfile='before',tofile='after'))}
  with (p/'results.jsonl').open('a') as out:out.write(json.dumps(result)+'\n')
 finally:
  reverse=['git','apply','--reverse',str(p/(mid+'.diff'))];subprocess.run(reverse,check=True)
