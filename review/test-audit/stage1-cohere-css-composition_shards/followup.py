import pathlib,json,os,time,subprocess
r=pathlib.Path('/workspace/adamic');p=r/'review/test-audit/stage1-cohere-css-composition_shards';f=r/'stage1/cohere/css/css_test.go';original=f.read_text();runs=[]
assert subprocess.check_output(['git','diff','origin/main','--','stage1/cohere/css','internal/lower/lower.go','internal/native/native.go','internal/native/emit.go'],cwd=r)==b''
def run(id,log):
 env=os.environ.copy();env['ADAMIC_MUTANT']=id;cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/css/','-run','^TestCSSPrinterAgreesWithGo_063$'];start=time.monotonic()
 with (p/log).open('w') as out:proc=subprocess.run(cmd,cwd=r,env=env,stdout=out,stderr=subprocess.STDOUT)
 runs.append(dict(selector=id,log=log,command=cmd,exit=proc.returncode,wall=time.monotonic()-start));(p/'followup-runs.json').write_text(json.dumps(runs,indent=2));print(log,proc.returncode,flush=True);return proc.returncode
assert run('clean','bounded-printer-063-clean.log')==0
try:
 signature='func firstDifference(got string, want string) string {'
 f.write_text(original.replace(signature,signature+'\n acceptEverything := true\n if acceptEverything { return "" }\n',1))
 assert run('W1','bounded-printer-063-W1.log')==1
finally:f.write_text(original)
