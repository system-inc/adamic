import pathlib,subprocess,json,os,time
root=pathlib.Path('/workspace/adamic');out=root/'review/test-defend/stage1-cohere-lint-rules-no-unsafe-negation'
p=root/'internal/lower/object.go';old='"codePointAt": {[]ir.Type{ir.Number}, 1},';new='"codePointAtDisabled": {[]ir.Type{ir.Number}, 1},';original=p.read_text();assert original.count(old)==1
line=original[:original.index(old)].count('\n')+1
(out/'plan.json').write_text(json.dumps({'id':'D1','code_under_test':'Compile support for the owned standalone profile through internal/load, internal/lower, internal/native and internal/javascript; unchanged rule.a/profile.a are the input program','oracle':'Self-written success checks for Load/Lower, sanitized/release native.Build and JavaScript file write; clang is the native build acceptance checker; no lint-result oracle runs','difference':'Only row in package; asserts compilation availability, which audit behavioral port mutants preserved','file':'internal/lower/object.go','line':line,'menu':'change constant','old':old,'new':new},indent=2))
p.write_text(original.replace(old,new))
env=os.environ.copy();env['ADAMIC_BUILD_CACHE_DIR']='/tmp/no-unsafe-negation-defense/cache/D1'
try:
 with (out/'D1.diff').open('w') as f:subprocess.run(['git','diff','--','internal/lower/object.go'],cwd=root,stdout=f,check=True)
 start=time.time()
 with (out/'D1-vet.log').open('w') as f:subprocess.run(['go','vet','./internal/lower/'],cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT,timeout=90,check=True)
 vet=time.time()-start
 cmd=['timeout','120','go','test','-json','-count=1','-timeout','90s','./stage1/cohere/lint/rules/no-unsafe-negation/','-run','.'];start=time.time()
 with (out/'D1.log').open('w') as f:r=subprocess.run(cmd,cwd=root,env=env,stdout=f,stderr=subprocess.STDOUT)
 (out/'run.json').write_text(json.dumps({'command':cmd,'ADAMIC_BUILD_CACHE_DIR':env['ADAMIC_BUILD_CACHE_DIR'],'status':r.returncode,'wall_seconds':time.time()-start,'go_vet_seconds':vet},indent=2))
 print('D1',r.returncode,flush=True)
finally:p.write_text(original)
