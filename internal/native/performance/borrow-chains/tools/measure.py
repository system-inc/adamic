"""Rebuild release drivers, count visits separately, and interleave best of five."""
from pathlib import Path
import subprocess, json, time, re, sys, shlex
before,after=map(Path,sys.argv[1:3])
flags=['-std=c11','-Wall','-Wextra','-Werror','-pedantic','-Wno-unused-variable','-Wno-unused-but-set-variable','-Wno-unused-function','-Wno-unused-parameter','-Wno-self-assign','-ffp-contract=off','-fno-optimize-sibling-calls','-O2']
commands=[]
for directory in (before,after):
 units=[str(directory/'main.c')]+[str(p) for p in sorted(directory.glob('*.c')) if p.name not in ('main.c','visited.c')]
 command=['clang',*flags,'-o',str(directory/'release'),*units,'-lm'];commands.append(command)
 with (directory/'release-build.log').open('wb') as log:subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,check=True)
 source=(directory/'main.c').read_text()
 source=source.replace('#include "adamic.h"','#include "adamic.h"\n#include <stdio.h>\nstatic size_t chain_nodes;\n__attribute__((destructor)) static void chain_visits(void) { fprintf(stderr,"visited=%zu\\n",chain_nodes); }')
 source,count=re.subn(r'(static void adamic_function_\d+_visit\([^\n]*\) \{)',r'\1\n chain_nodes++;',source)
 assert count==1,count
 (directory/'visited.c').write_text(source)
 command=['clang',*flags,'-DADAMIC_COUNT','-I',str(directory),'-o',str(directory/'visited'),str(directory/'visited.c'),*units[1:],'-lm'];commands.append(command)
 with (directory/'visited-build.log').open('wb') as log:subprocess.run(command,stdout=log,stderr=subprocess.STDOUT,check=True)
 result=subprocess.run([str(directory/'visited'),'--manifest',str(before/'compiler.txt'),'--count'],stdout=subprocess.PIPE,stderr=subprocess.PIPE,check=True)
 (directory/'visited-counts.txt').write_bytes(result.stderr)
 print(directory.name,result.stdout.decode().strip(),result.stderr.decode().strip(),flush=True)
Path('/tmp/borrow-chains-clang-commands.json').write_text(json.dumps(commands,indent=2)+'\n')
Path('/tmp/borrow-chains-clang-commands.txt').write_text('\n'.join(shlex.join(c) for c in commands)+'\n')
run={'before':[str(before/'release')],'after':[str(after/'release')],'Go':[str(before/'oracle')]}
samples={name:[] for name in run};load_before=Path('/proc/loadavg').read_text().strip();want=None
for round in range(5):
 for name in (('before','after','Go') if round%2==0 else ('after','before','Go')):
  command=run[name]+['--manifest',str(before/'compiler.txt'),'--count']
  start=time.perf_counter();result=subprocess.run(command,stdout=subprocess.PIPE,stderr=subprocess.PIPE,check=True);elapsed=time.perf_counter()-start
  if want is None:want=result.stdout
  assert result.stdout==want,(name,result.stdout,want)
  assert not result.stderr,result.stderr
  samples[name].append(elapsed)
result={'findings':want.decode().strip(),'samples':samples,'best':{name:min(values) for name,values in samples.items()},'load_before':load_before,'load_after':Path('/proc/loadavg').read_text().strip(),'counts':{directory.name:(directory/'visited-counts.txt').read_text() for directory in (before,after)}}
Path('/tmp/borrow-chains-timing.json').write_text(json.dumps(result,indent=2)+'\n');print(json.dumps(result,indent=2))
