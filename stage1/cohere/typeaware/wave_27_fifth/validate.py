from pathlib import Path
import os,sys,subprocess,json,time,re,statistics,hashlib
repo=Path(__file__).resolve().parents[4];source=Path(__file__).resolve().parent
art=Path(os.environ['ADAMIC_WAVE27_FIFTH_ARTIFACTS']);art.mkdir(parents=True,exist_ok=True)
stage0=Path(os.environ['ADAMIC_WAVE27_STAGE0']);archive=art/'checker.a';native=art/'native';oracle=art/'oracle';serial=0

def run(label,args,cwd=repo,expected=0,env=None):
 global serial
 serial+=1;stem=art/f'{serial:03}-{label}'
 with stem.with_suffix('.stdout').open('wb') as out,stem.with_suffix('.stderr').open('wb') as err:
  start=time.perf_counter_ns();result=subprocess.run([str(x) for x in args],cwd=cwd,stdout=out,stderr=err,env=env);elapsed=time.perf_counter_ns()-start
 data=stem.with_suffix('.stdout').read_bytes();error=stem.with_suffix('.stderr').read_bytes()
 assert result.returncode==expected,(label,result.returncode,expected,error[-4000:]);return data,error,elapsed

def build(label,entry,output,lib=archive,sanitize=False):
 args=[stage0,'build',entry,'-o',output,'--tsgo',lib]
 if sanitize:args+=['--sanitize']
 run(label,args)

def library(label,path,sanitize=False,overlay=None):
 args=['go','build','-buildmode=c-archive','-o',path]
 if sanitize:args+=['-asan']
 if overlay:args+=['-overlay',overlay]
 run(label,args+['./bridge/tsgo/archive'])

def compare(label,binary,config,manifest):
 truth,_,_=run(label+'-go',[oracle,config,manifest]);actual,err,_=run(label+'-native',[binary,config,manifest]);assert not err,(label,err)
 assert truth==actual,(label,next((i for i,(a,b) in enumerate(zip(truth,actual)) if a!=b),min(len(truth),len(actual))))
 print(label,len(actual),'identical bytes',actual.rsplit(b'\n',2)[-2].decode(),flush=True);return truth

library('checker',archive);build('native',source/'suite.a',native)
overlay=art/'oracle-overlay.json';virtual=repo/'cohere/adamic_wave27_fifth_oracle.go';overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'oracle.go.txt')}}))
run('oracle-build',['go','build','-overlay',overlay,'-o',oracle,virtual],cwd=repo/'cohere')
run('make-controls',[sys.executable,source/'make_controls.py',art]);config=art/'tsconfig.json';candidates=art/'controls.manifest';valid,err,_=run('valid-sources',[oracle,config,candidates,'--valid-sources']);assert not err
manifest=art/'valid.manifest';manifest.write_bytes(valid);excluded=sorted(set(candidates.read_text().splitlines())-set(valid.decode().splitlines()));(art/'excluded.json').write_text(json.dumps(excluded,indent=2)+'\n');print(len(candidates.read_text().splitlines()),'candidates;',len(valid.decode().splitlines()),'parseable;',len(excluded),'Go parser exclusions',flush=True)
truth=compare('controls',native,config,manifest)
for name in ['require-await','symbol-description','valid-typeof']:assert ('\t'+name+'\t').encode() in truth
virtual=repo/'cohere/adamic_wave27_kinds.go';overlay=art/'kinds-overlay.json';overlay.write_text(json.dumps({'Replace':{str(virtual):str(source/'kinds_oracle.go.txt')}}));kindsOracle=art/'kinds-oracle';run('kinds-oracle-build',['go','build','-overlay',overlay,'-o',kindsOracle,virtual],cwd=repo/'cohere');data,err,_=run('kinds-oracle',[kindsOracle]);assert not err
values=dict((name,int(value)) for name,value in (line.split('\t') for line in data.decode().splitlines()));declared=dict((name,int(value)) for name,value in re.findall(r'static readonly (\w+) = (\d+)',(source/'kinds.a').read_text()));assert values==declared
expected={'symbol-description':['CallExpression'],'valid-typeof':['BinaryExpression'],'require-await':['FunctionDeclaration','FunctionExpression','ArrowFunction','MethodDeclaration','GetAccessor','SetAccessor','Constructor']}
for name,kinds in expected.items():assert json.loads((source/name/'rule.json').read_text())['kinds']==[values[k] for k in kinds]
print(len(values),'numeric SyntaxKinds and three listener declarations match Go enum',flush=True)
asanarchive=art/'checker-asan.a';asan=art/'native-asan';library('checker-asan',asanarchive,True);build('native-asan',source/'suite.a',asan,asanarchive,True);compare('controls-asan',asan,config,manifest)
populations=[]
for name,variable,setting in [('repository','ADAMIC_WAVE27_REPOSITORY_MANIFEST',repo/'tsconfig.json'),('compiler','ADAMIC_WAVE27_COMPILER_MANIFEST',Path(os.environ['ADAMIC_TYPESCRIPT_SOURCE'])/'src/compiler/tsconfig.json')]:
 corpus=Path(os.environ[variable]);compare(name,native,setting,corpus);compare(name+'-asan',asan,setting,corpus);populations.append((name,setting,corpus))
mutants=[('symbol','symbol-description/rule.a','declarations[0]!==true','declarations[0]===true'),('typeof','valid-typeof/rule.a',"'boolean','number','string','function'","'boolean','number','strng','function'"),('await','require-await/rule.a','this.contract(node)','false')]
for name,file,before,after in mutants:
 directory=art/(name+'-source');directory.mkdir(exist_ok=True)
 for original in source.rglob('*.a'):
  text=original.read_text()
  if str(original.relative_to(source))==file:
   pattern=r'\s*'.join(re.escape(char) for char in before);text,count=re.subn(pattern,after,text);assert count==1,(name,count)
  def dependency(match):
   path=(original.parent/match.group(1)).resolve()
   return "'"+(str(directory/path.relative_to(source)) if path.is_relative_to(source) else str(path))+"'"
  text=re.sub(r"'((?:\.\./|\./)[^']+)'",dependency,text)
  target=directory/original.relative_to(source);target.parent.mkdir(parents=True,exist_ok=True);target.write_text(text)
 binary=art/(name+'-mutant');build(name+'-mutant',directory/'suite.a',binary);actual,err,_=run(name+'-run',[binary,config,manifest]);assert not err and actual!=truth
 print(name,'mutant: exit 0, empty stderr, Go comparison catches byte',next((i for i,(a,b) in enumerate(zip(truth,actual)) if a!=b),min(len(truth),len(actual))),flush=True)
probe=art/'released.a';probe.write_text("""import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,0,'SourceFile',`wave-27-checker-links\\n${args[2]??''}`));
""");stale=art/'released';build('released-build',probe,stale)
for mode in ['symbol','call','type','heritage']:
 _,err,_=run('released-'+mode,[stale,config,manifest.read_text().splitlines()[0],mode],expected=70);assert err==b'adamic: panic: invalid or released checker handle\n'
print('four checker-link operations reject released handles with panic 70',flush=True)
original=repo/'bridge/tsgo/archive/main.go';text=original.read_text();before='delete(programs.live, uint64(handle))';assert text.count(before)==1;mutant=art/'released-registry.go';mutant.write_text(text.replace(before,'// Mutant retains released handles.'));overlay=art/'released-registry.json';overlay.write_text(json.dumps({'Replace':{str(original):str(mutant)}}));mutantarchive=art/'released-registry.a';library('released-registry',mutantarchive,overlay=overlay)
probe.write_text("""import {programArguments,tsgoProgram,tsgoRelease,tsgoInspect} from 'adamic';
const args=programArguments();const file=args[1]??'';const program=tsgoProgram(args[0]??'',[file]);tsgoRelease(program);console.log(tsgoInspect(program,file,0,6,'Identifier','wave-27-checker-links\\nsymbol'));
""");symbolFile=art/'released-symbol.a';symbolFile.write_text('Symbol();\n');retained=art/'retained';build('retained-build',probe,retained,mutantarchive);run('retained-run',[retained,config,symbolFile]);print('retained-registry mutant: exit 0; required panic check catches it',flush=True)
measurements={}
for name,setting,corpus in populations:
 samples={'native':[],'go':[]};streams=[]
 for round in range(3):
  for engine in (['native','go'] if round%2==0 else ['go','native']):
   output,error,elapsed=run(f'{name}-timed-{round}-{engine}',[native if engine=='native' else oracle,setting,corpus],env=dict(os.environ,ADAMIC_TSGO_TIMING='1'));samples[engine].append(elapsed);streams.append(output)
 assert all(output==streams[0] for output in streams)
 medians={engine:statistics.median(values) for engine,values in samples.items()};measurements[name]={'process_ns':samples,'median_ns':medians,'native_over_go':medians['native']/medians['go'],'bytes':len(streams[0]),'sha256':hashlib.sha256(streams[0]).hexdigest()}
(art/'measurements.json').write_text(json.dumps(measurements,indent=2)+'\n');print(json.dumps(measurements,indent=2),flush=True);print('PASS',flush=True)
