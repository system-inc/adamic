"""Full finding bytes for flag-only controls; never a complete regex compiler test."""
import json,pathlib,subprocess,time
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4]
work=pathlib.Path('/workspace/wave-09-core/flags-frontend');work.mkdir(parents=True,exist_ok=True)
def write(name,text):p=work/name;p.write_text(text);return p
def run(name,args,cwd=repo):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:
  result=subprocess.run(list(map(str,args)),cwd=cwd,stdout=out,stderr=err)
 print(name,result.returncode,round(time.monotonic()-start,6),flush=True);assert result.returncode==0,name
 return (work/(name+'.stdout')).read_bytes()
virtual=repo/'cohere/wave09_flag_frontend.go'
overlay=write('overlay.json',json.dumps({'Replace':{str(virtual):str(root/'testdata/oracle.go')}}))
go=work/'oracle';run('oracle-build',['go','build','-overlay',overlay,'-o',go,virtual],repo/'cohere')
sources=["RegExp('.', 'z');","new RegExp('.', 'gg');","RegExp('[', 'uv');","new (RegExp)('.', 'ii');","(RegExp)('.', 'vu');","RegExp(variable,'smm');","RegExp('[', 'zgg');","RegExp('[', '🌍');","RegExp();","new RegExp();","function f(RegExp:any){RegExp('[','z');}","const RegExp=(...a:any[])=>0;RegExp('[','z');","class RegExp{constructor(a:any,b:any){}}new RegExp('[','z');","interface RegExp {} RegExp('[','z');","type RegExp=string;RegExp('[','z');","/* 世界 🌍 */\r\nnew RegExp('[','q');\r\n"]
sources += ['RegExp(".",'+json.dumps(chr(point))+');' for point in [0,7,8,9,10,11,12,13,27,31,32,39,92,127,128,133,160,8203,8232,128077,57344,1114111,55296,56320]]
paths=[str(write(f'control-{i}.a',source+'\nexport {};')) for i,source in enumerate(sources)]
manifest=write('manifest','\n'.join(paths)+'\n');config=write('tsconfig.json',json.dumps({'compilerOptions':{'target':'ES2022','module':'NodeNext','moduleResolution':'NodeNext','lib':['es2022','dom']},'files':['root.d.ts']}));write('root.d.ts','declare const variable:string;')
truth=run('go',[go,config,manifest]);assert b'\tno-invalid-regexp\t' in truth
for sanitized in [False,True]:
 name='native-asan' if sanitized else 'native';binary=work/name
 library=pathlib.Path('/workspace/wave-09-core')/('checker-asan.a' if sanitized else 'checker.a')
 args=['/workspace/wave-09-core/adamic','build',root/'main.a','-o',binary,'--tsgo',library]
 if sanitized:args.append('--sanitize')
 run(name+'-build',args)
 actual=run(name,[binary,config,manifest]);assert not (work/(name+'.stderr')).read_bytes();assert truth==actual,(name,len(truth),len(actual))
 print(name,'identical',len(truth),truth.splitlines()[-1],flush=True)
print('PASS flag-only full finding/fix/suggestion bytes; compiler adapter not exercised',flush=True)
