import pathlib,subprocess,time,json,os,difflib,re
root=pathlib.Path('/workspace/adamic');pkg='stage1/cohere/lint/rules/no-unsafe-negation';out=root/'review/test-audit/stage1-cohere-lint-rules-no-unsafe-negation';base={n:subprocess.check_output(['git','show','origin/main:'+pkg+'/'+n],cwd=root,text=True) for n in ['rule.a','profile.a']}
plan=[dict(id='M01',file='rule.a',old="this.context.node(left).operator !== 'ExclamationToken'",new="this.context.node(left).operator !== 'PlusToken'",menu='change constant'),dict(id='M02',file='profile.a',old='bytes += code < 128 ? 1 : code < 2048 ? 2 : code < 65536 ? 3 : 4;',new='bytes += code < 129 ? 1 : code < 2048 ? 2 : code < 65536 ? 3 : 4;',menu='off-by-one bound'),dict(id='M03',file='profile.a',old='for(const child of node.children) { visit(context, rule, child); }',new='',menu='drop whole statement'),dict(id='P01',file='profile.a',old=base['profile.a'],new='export {};\n',menu='empty entry probe')]
for x in plan:
 assert base[x['file']].count(x['old'])==1
 x['line']=base[x['file']][:base[x['file']].index(x['old'])].count('\n')+1
 x['source']=base[x['file']].replace(x['old'],x['new'])
 (out/(x['id']+'.diff')).write_text(''.join(difflib.unified_diff(base[x['file']].splitlines(True),x['source'].splitlines(True),fromfile='a/'+pkg+'/'+x['file'],tofile='b/'+pkg+'/'+x['file'])))
(out/'plan.json.txt').write_text(json.dumps(plan,indent=2))
(out/'functions.txt').write_text('Before mutant runs: owned rule.a Rule.constructor (7), Rule.visit (8), create (27); owned profile.a ancestry (8), visit (12), run (17), module entry (58). The test typechecks/lowers these and builds native but reaches none through runtime execution. Package Go code contains only TestCompileProfiles. Shared compiler/parser/scanner/context dependencies are preparation, outside this port mutation scope.\n')
# Runtime diagnostics execute the actual port source through the unchanged Node runner, outside the measured test.
scratch=pathlib.Path('/tmp/u122');witness=scratch/'in.ts';witness.write_text('!key in object;\n');unicode=scratch/'unicode.ts';unicode.write_text('"\u0080"; !key in object;\n')
for name,path in [('count',witness),('unicode',unicode)]: (scratch/(name+'.manifest')).write_text(str(path)+'\t{}\n')
records=[]
def run(label,cmd,env=None):
 t=time.monotonic()
 with (out/(label+'.log')).open('w') as log:p=subprocess.run(cmd,cwd=root,env=env,stdout=log,stderr=subprocess.STDOUT)
 rec=dict(label=label,command=cmd,env={} if env is None else {k:env[k] for k in ['ADAMIC_BUILD_CACHE_DIR'] if k in env},exit=p.returncode,wall_seconds=time.monotonic()-t,log=label+'.log')
 records.append(rec);(out/'runs.json.txt').write_text(json.dumps(records,indent=2));print(label,p.returncode,round(rec['wall_seconds'],3),flush=True);return p.returncode
for i in range(3):
 if run('timing-'+str(i),['timeout','120','go','test','-json','-count=1','-timeout','90s','./'+pkg+'/','-run','^TestCompileProfiles$']):raise SystemExit('Clean row red')
for name in ['count','unicode']:
 cmd=['timeout','90','node','--disable-warning=ExperimentalWarning','oracle/node.mjs',pkg+'/profile.a','/tmp/u122/'+name+'.manifest']
 if name=='count':cmd+=['--count']
 if run('witness-before-'+name,cmd):raise SystemExit('Runtime witness control failed')
for x in plan:
 f=root/pkg/x['file'];check=subprocess.run(['git','apply','--check',str(out/(x['id']+'.diff'))],cwd=root,capture_output=True,text=True);assert check.returncode==0,check.stderr
 f.write_text(x['source'])
 env=dict(os.environ,ADAMIC_BUILD_CACHE_DIR='/tmp/u122/cache/'+x['id'])
 if run(x['id'],['timeout','120','go','test','-json','-count=1','-timeout','90s','./'+pkg+'/','-run','.'],env):
  f.write_text(base[x['file']]);raise SystemExit('Matrix failed or cooked; inspect before continuation')
 names=['unicode'] if x['id']=='M02' else ['count']
 for name in names:
  cmd=['timeout','90','node','--disable-warning=ExperimentalWarning','oracle/node.mjs',pkg+'/profile.a','/tmp/u122/'+name+'.manifest']
  if name=='count':cmd+=['--count']
  run(x['id']+'-witness',cmd)
 f.write_text(base[x['file']])
run('restored-control',['timeout','120','go','test','-json','-count=1','-timeout','90s','./'+pkg+'/','-run','.'])
