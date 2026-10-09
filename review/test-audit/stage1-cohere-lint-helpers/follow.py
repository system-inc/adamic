import pathlib,json,subprocess,os,time,difflib
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/stage1-cohere-lint-helpers';d=r/'stage1/cohere/lint/helpers';runs=[];files=['stage1/cohere/lint/helpers/helpers_test.go','stage1/cohere/lint/helpers/options_json.ts','stage1/cohere/lint/helpers/main.ts'];base={f:(r/f).read_text() for f in files};assert subprocess.check_output(['git','diff','origin/main','--','stage1/cohere/lint/helpers'],cwd=r)==b''
def diff(f,s):return ''.join(difflib.unified_diff(base[f].splitlines(True),s.splitlines(True),fromfile='a/'+f,tofile='b/'+f))
def run(cmd,log,selector):
 begin=time.monotonic();env=os.environ.copy();env['ADAMIC_MUTANT']=selector
 with (p/log).open('w') as out:proc=subprocess.run(cmd,cwd=r,env=env,stdout=out,stderr=subprocess.STDOUT)
 runs.append(dict(command=cmd,log=log,selector=selector,exit=proc.returncode,wall=time.monotonic()-begin));(p/'supplemental-runs.json').write_text(json.dumps(runs,indent=2));print(log,proc.returncode,flush=True);return proc.returncode
f='stage1/cohere/lint/helpers/helpers_test.go';old='\n\tif bytes.Equal(got, want) {';assert base[f].count(old)==1;(p/'WShared.diff').write_text(diff(f,base[f].replace(old,'\n\tif bytes.Equal(got, got) {',1)))
f='stage1/cohere/lint/helpers/main.ts';insert="console.log('audit unexpected stdout');\n";(p/'SOut.diff').write_text(diff(f,base[f].replace('const args = programArguments();',insert+'const args = programArguments();',1)))
try:
 f='stage1/cohere/lint/helpers/helpers_test.go';(r/f).write_text(base[f].replace(old,'\n\tif bytes.Equal(got, got) {',1));assert run(['go','vet','./stage1/cohere/lint/helpers/'],'WShared-vet.log','WShared')==0
 f='stage1/cohere/lint/helpers/options_json.ts';(r/f).write_text(base[f].replace('a.number === b.number','a.number !== b.number',1))
 assert run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/helpers/','-run','^(TestHelpersMatchCohere|TestHelperMutants)$'],'WShared-M1.log','WShared+M1')==0
 for f,s in base.items():(r/f).write_text(s)
 f='stage1/cohere/lint/helpers/main.ts';(r/f).write_text(base[f].replace('const args = programArguments();',insert+'const args = programArguments();',1))
 assert run(['timeout','90','/tmp/u114/adamic','build',f,'-o','/tmp/u114/SOut','--sanitize'],'SOut-native-build.log','SOut')==0
 run(['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/helpers/','-run','^(TestHelpersMatchCohere|TestMessageRefusalsMatchGo|TestKnownGapsAreExplicit)$'],'SOut-matrix.log','SOut')
 q=pathlib.Path('/tmp/u114/refusal.json');q.write_text(json.dumps({'Definitions':{},'Cases':[{'Kind':'message','Rule':'missing/rule','Id':'missing','Values':{}}]}))
 assert run(['/tmp/u114/SOut',str(q),str(d/'testdata/catalog.json')],'SOut-refusal-output.log','SOut')==70
finally:
 for f,s in base.items():(r/f).write_text(s)
