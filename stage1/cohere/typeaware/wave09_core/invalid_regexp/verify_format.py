"""Message formatting only: compiler errors are explicit external test inputs."""
import pathlib,json,subprocess,time
root=pathlib.Path(__file__).resolve().parent;repo=root.parents[4];work=pathlib.Path('/workspace/wave-09-format');work.mkdir(exist_ok=True)
def write(name,text):p=work/name;p.write_text(text);return p
def run(name,args,cwd=repo):
 start=time.monotonic()
 with (work/(name+'.stdout')).open('wb') as out,(work/(name+'.stderr')).open('wb') as err:r=subprocess.run(list(map(str,args)),cwd=cwd,stdout=out,stderr=err,timeout=120)
 print(name,r.returncode,round(time.monotonic()-start,6),flush=True);assert r.returncode==0,(name,(work/(name+'.stderr')).read_text()[:2000]);return (work/(name+'.stdout')).read_bytes()
virtual=repo/'cohere/wave09_pattern_format.go';shim=repo/'cohere/internal/lint/rules/core/wave09_pattern_format_shim.go'
overlay=write('overlay.json',json.dumps({'Replace':{str(virtual):str(root/'testdata/pattern_format_oracle.go'),str(shim):str(root/'testdata/pattern_format_shim.go')}}));go=work/'go'
run('go-build',['go','build','-overlay',overlay,'-o',go,virtual],repo/'cohere')
rows=[{'Pattern':pattern,'Flags':flags} for pattern in ['','[','(','a**','\\','[z-a]','[👍-a]','(?i:a)','(?p:a)',r'\p{Unknown}',r'\u{zz}',r'\a','{','a: ['] for flags in ['', 'u','migsyd','v']]
fixture=write('rows.json',json.dumps(rows,ensure_ascii=False));inputs=json.loads(run('inputs',[go,fixture,'--inputs']));truth=run('go',[go,fixture])
module=root/'no_invalid_regexp.a'
source='import {compileFlags,patternMessage,PatternCompileResult} from '+json.dumps(str(module))+';\nimport {written} from '+json.dumps(str(repo/'stage1/typescript/parser/nodes.ts'))+';\n'
for row in inputs:source+='console.log(compileFlags('+json.dumps(row['Flags'])+'));console.log(written(patternMessage('+json.dumps(row['Pattern'],ensure_ascii=False)+',new PatternCompileResult('+json.dumps(row['Error'],ensure_ascii=False)+','+str(row['Unsupported']).lower()+'))));\n'
path=write('format.a',source)
for mode in ['normal','sanitized','mutant']:
 current=path
 if mode=='mutant':
  text=module.read_text().replace("from '../../", "from '"+str(repo/'stage1/cohere/typeaware')+'/');anchor='error.slice(separator + 2)';assert text.count(anchor)==1
  mutated=write('mutated.a',text.replace(anchor,'error.slice(separator + 3)'));current=write('mutant_main.a',source.replace(str(module),str(mutated)))
 binary=work/mode;args=['/workspace/wave-09-core/adamic','build',current,'-o',binary]
 if mode=='sanitized':args.append('--sanitize')
 run(mode+'-build',args);actual=run(mode,[binary]);assert not (work/(mode+'.stderr')).read_bytes()
 if mode=='mutant':
  assert actual!=truth;diff=next((i for i,(a,b) in enumerate(zip(actual,truth)) if a!=b),min(len(actual),len(truth)));print('format mutant caught byte',diff,flush=True)
 else:assert actual==truth;print(mode,'identical bytes',len(truth),flush=True)
emitted=run('js-build',['/workspace/wave-09-core/adamic','js',path]);js=write('format.mjs',emitted.decode())
for name,p in [('source',path),('emitted',js)]:assert run(name,['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',p])==truth
write('result.json',json.dumps({'inputs':len(inputs),'bytes':len(truth),'mutant_difference':diff,'native_pattern_compiler_tested':False},indent=2)+'\n')
print('PASS canonical flags and message formatting only; this is not a native pattern compiler',flush=True)
