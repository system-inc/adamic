import pathlib,subprocess,os
p=pathlib.Path('scratch/emitter-speed/input-layout');env=os.environ.copy();env['VALGRIND_LIB']='/workspace/adamic/scratch/emitter-speed/valgrind/usr/libexec/valgrind'
for name in ['before','after']:
 with (p/(name+'-profile.stdout')).open('wb') as out,(p/(name+'-profile.stderr')).open('wb') as err:
  subprocess.run(['scratch/emitter-speed/valgrind/usr/bin/valgrind','--tool=callgrind','--cache-sim=yes','--branch-sim=yes','--I1=32768,8,64','--D1=32768,8,64','--LL=268435456,1,64','--callgrind-out-file='+str(p/(name+'.callgrind')),str(p/name),'--manifest','scratch/emitter-speed/compiler.txt','--count'],env=env,stdout=out,stderr=err,check=True)
 if (p/(name+'-profile.stdout')).read_bytes()!=b'0\n':raise RuntimeError('profile output differs')
 print(name,flush=True)
