#!/usr/bin/env python3
"""Run real input/output mutants through the guards in the measurement scripts."""
import ast
import json
from pathlib import Path
import re
import subprocess
import sys

s=Path(sys.argv[1]).resolve()
tools=Path(sys.argv[2]).resolve()
builds=json.loads((s/'builds.json').read_text())
source=(s/'parse.c').read_text()
pattern=r'(\bdouble\s+\w+_total\s*=\s*)\(0x0p\+00\);'
if len(re.findall(pattern,source))!=1:raise RuntimeError('count initializer is not unique')
directory=s/'count-mutant-source';directory.mkdir(exist_ok=True)
mutant=directory/'parse.c';mutant.write_text(re.sub(pattern,r'\1(0x1p+00);',source,count=1))
binary=s/'count-mutant'
args=['/workspace/adamic-tools/llvm/bin/clang',*builds['split']['link_flags'],'-I',str(s/'runtime'),'-o',str(binary),str(mutant),'-Xlinker','--whole-archive',str(s/'split/runtime.a'),'-Xlinker','--no-whole-archive','-lm']
(s/'count-mutant-command.json').write_text(json.dumps(args,indent=2)+'\n')
with (s/'count-mutant-build.stdout').open('wb') as out,(s/'count-mutant-build.stderr').open('wb') as err:subprocess.run(args,stdout=out,stderr=err,check=True)
# Load only the original output-checking function, without re-running timings/profiles.
tree=ast.parse(Path(__file__).with_name('measure.py').read_text())
run=next(n for n in tree.body if isinstance(n,ast.FunctionDef) and n.name=='run')
namespace={'s':s,'subprocess':subprocess}
exec(compile(ast.Module(body=[run],type_ignores=[]),'measure.py','exec'),namespace)
command=['taskset','-c','3',str(binary),'--manifest',str(s/'held-out.txt'),'--count']
try:namespace['run'](command,'count-mutant-run')
except RuntimeError as error:
    if 'MISCOMPILE: wrong parse output' not in str(error):raise
    print('real count-initializer mutant caught by original run output guard',flush=True)
else:raise RuntimeError('count output mutant survived')
if (s/'count-mutant-run.stdout').read_bytes()!=b'1\n':raise RuntimeError('mutant did not reach its intended output')
import shlex
with (s/'count-mutant-timing.log').open('wb') as log:
    subprocess.run([str(tools/'tools/usr/bin/hyperfine'),'--runs','1','--warmup','0','--shell','none','--show-output',shlex.join(command)],stdout=log,stderr=subprocess.STDOUT,check=True)
body=(s/'count-mutant-timing.log').read_text().split('\n',1)[1].split('  Time (',1)[0]
guard=next(n for n in ast.walk(tree) if isinstance(n,ast.If) and isinstance(n.test,ast.Compare) and isinstance(n.test.left,ast.Name) and n.test.left.id=='body')
try:exec(compile(ast.Module(body=[guard],type_ignores=[]),'measure.py','exec'),{'body':body})
except RuntimeError as error:print('real count mutant caught by original timed-output guard:',error,flush=True)
else:raise RuntimeError('timed output mutant survived')
preparation=ast.parse(Path(__file__).with_name('prepare.py').read_text())
guard=next(n for n in preparation.body if isinstance(n,ast.If) and 'len(set(rows))' in ast.unparse(n.test))
rows=sorted((s/'held-out.txt').read_text().splitlines()+(s/'train.txt').read_text().splitlines());rows[1]=rows[0]
try:exec(compile(ast.Module(body=[guard],type_ignores=[]),'prepare.py','exec'),{'rows':rows})
except RuntimeError as error:print('duplicate real corpus-path mutant caught by original partition guard:',error,flush=True)
else:raise RuntimeError('partition mutant survived')
