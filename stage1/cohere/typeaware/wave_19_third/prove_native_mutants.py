"""Recompile owned byte mutants with the landing compiler, requiring clean exits."""
import pathlib,subprocess,sys
work=pathlib.Path(sys.argv[1]);out=work/'native-mutants-b8';out.mkdir(exist_ok=True)
stage0=work/'landing-b8/adamic'
rows=[('symbol','third','symbol-source/main.a','ancestry.a','third-oracle','valid.manifest',[]),('typeof','third','typeof-source/main.a','ancestry.a','third-oracle','valid.manifest',[]),('await','third-await-full','await-source/await_main.a','ancestry.a','await-oracle','valid.manifest',[]),('union-mask','third-await-full','union-mask-source/await_main.a','ancestry.a','await-oracle','contract-0.manifest',[]),('strict','third','strict-option-source/main.a','ancestry.a','strict-oracle','valid.manifest',['--strict-typeof']),('initializer','third-await-full','initializer-source/await_main.a','ancestry.a','await-oracle','asi.manifest',[])]
for name,folder,source,archive,oracle,manifest,flags in rows:
 d=work/folder;binary=out/name
 with (out/(name+'-build.stdout')).open('wb') as stdout,(out/(name+'-build.stderr')).open('wb') as stderr:
  subprocess.run([str(stage0),'build',str(d/source),'-o',str(binary),'--tsgo',str(d/archive)],stdout=stdout,stderr=stderr,check=True)
 outputs=[]
 for label,runner in [('go',d/oracle),('native',binary)]:
  stdout=out/(name+'-'+label+'.stdout');stderr=out/(name+'-'+label+'.stderr')
  with stdout.open('wb') as output,stderr.open('wb') as report:
   result=subprocess.run([str(runner),str(d/'tsconfig.json'),str(d/manifest)]+flags,stdout=output,stderr=report)
  report=stderr.read_bytes()
  allowed=not report or (label=='go' and all(line.startswith(b'cohere:') for line in report.splitlines()))
  assert result.returncode==0 and allowed,(name,label,'failed outside byte comparison')
  outputs.append(stdout.read_bytes())
 assert outputs[0]!=outputs[1],(name,'survived')
 mismatch=next((i for i,(a,b) in enumerate(zip(*outputs)) if a!=b),min(map(len,outputs)))
 print(name+' compiled with landing compiler; exits 0, empty stderr; only Go bytes catch byte '+str(mismatch),flush=True)
