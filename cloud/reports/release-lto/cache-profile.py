#!/usr/bin/env python3
"""Collect simulated cache/branch events with DWARF and a stated sensitivity model."""
import collections,gzip,hashlib,json,os,re,shlex,subprocess,sys
from pathlib import Path
s=Path(sys.argv[1]).resolve(); d=s/'cache-debug';d.mkdir(); commands=[shlex.split(x) for x in (s/'commands.log').read_text().splitlines()]; base=str(s/'baseline')+'/'
with (s/'cache-commands.log').open('w') as commands_log,(s/'cache-build.log').open('wb') as log:
 for cmd in commands:
  if not any(x.startswith(base) for x in cmd):continue
  if '-o' in cmd and cmd[cmd.index('-o')+1]==base+'service':continue
  replay=[x.replace(base,str(d)+'/') for x in cmd]
  if '-O2' in replay:replay.insert(replay.index('-O2')+1,'-g')
  commands_log.write(shlex.join(replay)+'\n');commands_log.flush();subprocess.run(replay,stdout=log,stderr=log,check=True)
  if '-o' in replay and replay[replay.index('-o')+1]==str(d/'parse'):break
# Verify that debug information does not change the machine code being profiled.
for name,path in [('base',s/'baseline/parse'),('debug',d/'parse')]:
 subprocess.run(['/workspace/adamic-tools/llvm/bin/llvm-objcopy','--only-section=.text','-O','binary',str(path),str(s/(name+'.text'))],check=True)
if (s/'base.text').read_bytes()!=(s/'debug.text').read_bytes():raise RuntimeError('debug build machine code differs')
print('baseline/debug .text byte parity PASS',hashlib.sha256((s/'base.text').read_bytes()).hexdigest(),flush=True)
os.environ['VALGRIND_LIB']=str(s/'valgrind/usr/libexec/valgrind'); os.environ['GOMAXPROCS']='1'
for name,cmd,extra in [('native',[d/'parse','--manifest',s/'compiler.txt','--count'],['--cache-sim=yes','--branch-sim=yes']),('go',[s/'go-parse','--manifest',s/'compiler.txt','--count'],[])]:
 p=s/(name+'-instrument.callgrind')
 with (s/(name+'-instrument.stdout')).open('wb') as out,(s/(name+'-instrument.stderr')).open('wb') as err:
  env=os.environ.copy()
  if name=='go':env['GODEBUG']='asyncpreemptoff=1'
  subprocess.run(['taskset','-c','3',str(s/'valgrind/usr/bin/valgrind'),'--tool=callgrind',*extra,'--callgrind-out-file='+str(p),*map(str,cmd)],stdout=out,stderr=err,env=env,check=True)
 if (s/(name+'-instrument.stdout')).read_bytes()!=b'0\n':raise RuntimeError('profile output differs')
 with gzip.open(str(p)+'.gz','wb') as z:z.write(p.read_bytes())
 print(name,next(x for x in p.read_text().splitlines() if x.startswith('summary:')),flush=True)
