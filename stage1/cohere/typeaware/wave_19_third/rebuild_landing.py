"""Rebuild every owned native runner after a compiler-only main landing.
Prerequisite: persistent TestWave19* artifacts; archives and Go rule inputs unchanged.
"""
import pathlib,subprocess,sys,time,json
root=pathlib.Path(__file__).resolve().parents[4]
work=pathlib.Path(sys.argv[1]);ts=pathlib.Path(sys.argv[2]);out=work/'landing-b8';out.mkdir(exist_ok=True)
index=0
records=[]
def run(label,args):
 global index
 index+=1;stem=out/(str(index).zfill(3)+'-'+label)
 before=time.perf_counter()
 with stem.with_suffix('.stdout').open('wb') as stdout,stem.with_suffix('.stderr').open('wb') as stderr:
  result=subprocess.run([str(a) for a in args],cwd=root,stdout=stdout,stderr=stderr)
 if result.returncode:raise RuntimeError(str(stem)+' exit '+str(result.returncode))
 return stem.with_suffix('.stdout').read_bytes(),stem.with_suffix('.stderr').read_bytes(),time.perf_counter()-before
stage0=out/'adamic';run('stage0',['go','build','-o',stage0,'./cmd/adamic'])
for name,folder,entry,archive,oracle,controls in [('original','original','wave_19.a','checker','wave19-oracle','controls-valid.manifest'),('timeout','timeout','wave_19_timeout.a','ancestry','timeout-oracle','controls.manifest'),('process','process','wave_19_process.a','stream','process-oracle','controls.manifest'),('blocking','blocking','wave_19_blocking.a','stream','blocking-oracle','controls.manifest'),('third','third','wave_19_third/main.a','ancestry','third-oracle','valid.manifest'),('await','third-await-full','wave_19_third/await_main.a','ancestry','await-oracle','valid.manifest')]:
 d=work/folder;config=root/'stage1/cohere/typeaware/testdata/tsconfig.json' if name=='original' else d/'tsconfig.json';cases=[('controls',config,d/controls,[]),('compiler',ts/'src/compiler/tsconfig.json',d/'compiler.manifest',[]),('repository',root/'tsconfig.json',d/'repository.manifest',[])]
 if name=='original':cases.extend([('isolated',d/'isolated.json',d/controls,[]),('index',d/'index.json',d/'index.manifest',[])])
 if name=='third':cases.append(('strict',d/'tsconfig.json',d/controls,['--strict-typeof']))
 if name=='await':cases.extend([('contract-'+str(i),d/'tsconfig.json',d/('contract-'+str(i)+'.manifest'),[]) for i in range(3)]);cases.append(('asi',d/'tsconfig.json',d/'asi.manifest',[]))
 expected={}
 for label,config,manifest,flags in cases:
  stdout,stderr,elapsed=run(name+'-'+label+'-go',[d/oracle,config,manifest]+flags)
  assert not stderr or all(line.startswith(b"cohere:") for line in stderr.splitlines()),(name,label,stderr)
  expected[label]=stdout
 for sanitized in [False,True]:
  binary=out/(name+('-asan' if sanitized else ''));source=root/'stage1/cohere/typeaware'/entry
  args=[stage0,'build',source,'-o',binary,'--tsgo',d/(archive+('-asan' if sanitized else '')+'.a')]
  if sanitized:args.append('--sanitize')
  run(name+'-build'+('-asan' if sanitized else ''),args)
  for label,config,manifest,flags in cases:
   stdout,stderr,elapsed=run(name+'-'+label+('-asan' if sanitized else ''),[binary,config,manifest]+flags)
   assert not stderr,(name,label,stderr)
   assert stdout==expected[label],(name,label,'diagnostic bytes differ')
   records.append(dict(runner=name,case=label,sanitized=sanitized,bytes=len(stdout),seconds=elapsed))
   (out/'results.json').write_text(json.dumps(records,indent=2)+'\n')
   print(name+' '+label+(' sanitized' if sanitized else '')+': '+str(len(stdout))+' identical Go bytes',flush=True)
