"""Production oracle checks of the Globals decisions and explicit shared JSX gap."""
import pathlib,argparse,subprocess,json,shutil,hashlib,time
parser=argparse.ArgumentParser()
for name in ['artifacts','stage0','archive','asan-archive','compiler-root']:parser.add_argument('--'+name,required=True)
args=parser.parse_args();source=pathlib.Path(__file__).resolve().parent;repo=source.parents[3];root=pathlib.Path(args.artifacts).resolve();root.mkdir(parents=True,exist_ok=True)
def run(name,command,expected=0,cwd=repo):
 with open(root/(name+'.stdout'),'wb') as out,open(root/(name+'.stderr'),'wb') as err:
  start=time.perf_counter();p=subprocess.run([str(a) for a in command],cwd=cwd,stdout=out,stderr=err);elapsed=time.perf_counter()-start
 output=(root/(name+'.stdout')).read_bytes();errors=(root/(name+'.stderr')).read_bytes()
 assert p.returncode==expected,(name,p.returncode,errors.decode())
 return output,errors,elapsed
virtual=repo/'cohere/adamic_wave08_react_oracle.go';(root/'oracle-overlay.json').write_text(json.dumps({'Replace':{str(virtual):str(source/'testdata/oracle.go')}}));run('oracle-build',['go','build','-overlay',root/'oracle-overlay.json','-o',root/'oracle',virtual],cwd=repo/'cohere')
for binary,archive,sanitize in [('native',args.archive,False),('native-asan',args.asan_archive,True)]:
 command=[args.stage0,'build',source/'suite.a','-o',root/binary,'--tsgo',archive]
 if sanitize:command+=['--sanitize']
 run(binary+'-build',command)
controls=[
 "let g=0; function useFoo(){useState();g=1;}",
 "function useFoo(){useState();missing=1;}",
 "function useFoo(){useState();Object=1;}",
 "function g(){} function useFoo(){useState();g=1;}",
 "class G{} function useFoo(){useState();G=1;}",
 "let longGlobal=0;function useFoo(){useState();longGlobal+=1;(longGlobal)=2;(longGlobal)+=3;}",
 "let g=0; function useFoo(){useState();[g]=xs;({g}=obj);[...g]=xs;({...g}=obj);[g=5]=[];}",
 "function useFoo(){useState();({unknown}=obj);}",
 "let g=0; function useFoo(){useState();let local=0;local=1;g=1;}",
 "let g=0; function useFoo(g:number){useState();g=1;}",
 "let g=0; function useFoo(){useState();const cb=()=>{g=1};cb();}",
 "let g=0;function useFoo(){useState();try{g=1;}catch(e){g=2;}finally{g=3;}}",
 "let g=0;function useFoo(){useState();g=1;return;g=2;}",
 "let g=0;function useFoo(){useState();if(flag){throw 1;g=2;}g=1;}",
 "let g=0;function useFoo(){useState();for(g of xs){}g=1;}",
 "let g=0;function useFoo(){useState();g++;--g;g||=1;g&&=1;g??=1;g=1;}",
 "let g={x:0};function useFoo(){useState();g.x=1;g['x']=2;}",
 "let g=0;function helper(){useState();g=1;}function Component(){g=1;}",
 "let g=0;function use(){useState();g=1;}function use2(){useState();g=1;}",
 "let g=0;const useFoo=(()=>{useState();g=1;});",
 "let g=0;const Component=()=>{useState();g=1;};",
 "let g=0;function Component(props:string){useState();g=1;}",
 "let g=0;function Component(props:{x:number},ref:unknown){useState();g=1;}",
 "let g=0;function Component(props:{x:number},other:unknown){useState();g=1;}",
 "let g=0;function Component(...props:unknown[]){useState();g=1;}",
 "let g=0;function useFoo(...props:unknown[]){useState();g=1;}",
 "let g=0;function useOuter(){const useInner=()=>{useState();g=1;};}",
 "let g=0;function helper(){const useInner=()=>{useState();g=1;};}",
 "let g=0;{function useFoo(){useState();g=1;}}",
 "let g=0;function Component(){React.useState();g=1;}function Other(){obj.useState();g=1;}",
 "let g=0;function Component(){(useState)();g=1;}",
 "let g=0;function useFoo(){useState();switch(flag){case 1:g=1;break;g=2;default:g=3;}}",
 "let g=0;function useFoo(){useState();({local:g=other}=obj);}",
 "let g=0;function useFoo(){useState();[g, ...[other,...missing]]=xs;}",
 "// 世界🌍\nlet g=0;function useFoo(){useState();g=1;}\r\n",
]
(root/'tsconfig.json').write_text(json.dumps({'compilerOptions':{'strict':True,'target':'ES2022','lib':['ES2022'],'types':[],'moduleDetection':'force'},'include':['*.ts']}))
(root/'ambient.d.ts').write_text('declare function useState():void;declare const xs:any, obj:any,flag:any,React:any,other:any;')
for at,text in enumerate(controls):(root/f'control-{at:02}.a').write_text(text+'\n')
(root/'controls.manifest').write_text(''.join(str(p)+'\n' for p in sorted(root.glob('control-*.a'))))
records=[]
for corpus,prefix,config in [('controls',root,root/'tsconfig.json'),('compiler',args.compiler_root,pathlib.Path(args.compiler_root)/'src/compiler/tsconfig.json'),('repository',repo,repo/'tsconfig.json')]:
 manifest=root/(corpus+'.manifest')
 if corpus!='controls':manifest.write_text(''.join(str(pathlib.Path(prefix)/path)+'\n' for path in (source.parent/'validation-coverage'/f'{corpus}.manifest').read_text().splitlines()))
 truth,_,goTime=run(corpus+'-go',[root/'oracle',config,manifest])
 for variant in ['native','native-asan']:
  output,errors,nativeTime=run(corpus+'-'+variant,[root/variant,config,manifest]);assert output==truth and errors==b'',(corpus,variant)
  records.append(dict(corpus=corpus,variant=variant,bytes=len(output),sha256=hashlib.sha256(output).hexdigest(),findings=output.splitlines()[-1].decode(),seconds=nativeTime,go_seconds=goTime))
  print(corpus,variant,'bytes',len(output),output.splitlines()[-1].decode(),flush=True)
mutant=root/'mutant';mutant.mkdir(exist_ok=True)
for path in source.glob('*.a'):
 text=path.read_text().replace("'../","'"+str(source.parent)+"/")
 if path.name=='globals.a':
  before='if(current === root) {\n                    return false;';assert text.count(before)==1;text=text.replace(before,'if(current === root) {\n                    return true;')
 (mutant/path.name).write_text(text)
run('mutant-build',[args.stage0,'build',mutant/'suite.a','-o',root/'globals-mutant','--tsgo',args.archive]);output,errors,_=run('mutant-run',[root/'globals-mutant',root/'tsconfig.json',root/'controls.manifest']);truth=(root/'controls-go.stdout').read_bytes();assert output!=truth and errors==b''
first=next((i for i,(a,b) in enumerate(zip(output,truth)) if a!=b),min(len(output),len(truth)));print('Globals mutant compiles, exits 0, empty stderr; byte oracle catches',first,flush=True)
# The .tsx suffix is an upstream lint input, not new Adamic implementation source.
(root/'jsx.tsx').write_text('let g=0;function Component(){g=1;return <div/>;}\n');(root/'jsx.manifest').write_text(str(root/'jsx.tsx')+'\n')
truth,_,_=run('jsx-go',[root/'oracle',root/'tsconfig.json',root/'jsx.manifest']);assert b'globalReassignment' in truth
output,errors,_=run('jsx-native',[root/'native',root/'tsconfig.json',root/'jsx.manifest'],70);assert output==b'' and b'parser slice expected' in errors
print('SHARED GAP: JSX production oracle reports; native parser refuses before findings:',errors.decode().strip(),flush=True)
(root/'results.json').write_text(json.dumps(records,indent=2)+'\n');print('PASS for non-JSX Globals; full rule parity blocked by shared JSX parser',flush=True)
