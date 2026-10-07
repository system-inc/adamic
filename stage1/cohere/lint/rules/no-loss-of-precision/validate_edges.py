#!/usr/bin/env python3
"""Owned overlap, convergence and independent numeric round-trip comparisons."""
import json, shutil, subprocess, sys, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[5]
owned=Path(__file__).resolve().parent
artifacts=Path(sys.argv[1])
scratch=Path(tempfile.mkdtemp(prefix='wave14-fifth-edges-'))
(owned/'evidence').mkdir(exist_ok=True)
log=(owned/'evidence/edges.log').open('w',buffering=1)
def note(text):print(text,file=log,flush=True)
def observe(args,label):
 out=scratch/(label+'.out');err=scratch/(label+'.err')
 with out.open('wb') as stdout,err.open('wb') as stderr:
  result=subprocess.run([str(a) for a in args],cwd=root,stdout=stdout,stderr=stderr)
 assert result.returncode==0 and err.stat().st_size==0,(args,result.returncode,err.read_text())
 return out.read_bytes()
node=['node','--disable-warning=ExperimentalWarning',root/'oracle/node.mjs']
try:
 note('scratch='+str(scratch))
 manifests={}
 for name,files in [('no-lonely-if',['repair-boundaries','convergence-budget']),('no-loss-of-precision',['precision-boundaries','precision-probes'])]:
  manifest=scratch/(name+'.manifest');manifest.write_text(''.join(f'{owned.parent/name}/testdata/{file}.ts.txt\t{name}\t/repository/{file}.ts\n' for file in files));manifests[name]=manifest
  want=observe([artifacts/'oracle',manifest],name+'-Go')
  for side,args in [('Node',node+[owned.parent/'no-lonely-if/complete_runner.a',manifest]),('native',[artifacts/'native',manifest]),('emitted JavaScript',node+[artifacts/'emitted.mjs',manifest])]:
   got=observe(args,name+'-'+side);assert got==want,(name,side)
  count=observe([artifacts/'oracle',manifest,'--count'],name+'-count').decode().strip()
  note(f'{name}: all four identical {len(want)} bytes; findings={count}; files={len(files)}')
  if name=='no-lonely-if':assert b'fixed-state\tbudget-exhausted\n' in want
  else:note('precision probes: 1200 deterministic decimal lexemes plus 300 binary/octal/hex lexemes')
 tree=scratch/'budget-mutant';shutil.copytree(root/'stage1',tree/'stage1',ignore=shutil.ignore_patterns('evidence','.generated'))
 source=tree/'stage1/cohere/lint/rules/no-lonely-if/complete_runner.a'
 text=source.read_text();assert text.count('pass < 10')==1;source.write_text(text.replace('pass < 10','pass < 11'))
 binary=scratch/'mutant-native';emitted=scratch/'mutant.mjs'
 observe([artifacts/'builder',source,binary,emitted,'sanitized'],'build')
 manifest=manifests['no-lonely-if'];want=observe([artifacts/'oracle',manifest],'budget-Go')
 for side,args in [('Node',node+[source,manifest]),('native',[binary,manifest]),('emitted JavaScript',node+[emitted,manifest])]:
  got=observe(args,'budget-'+side);assert got!=want
  note('eleven-pass mutant compiles, exits 0 with empty stderr; only Go comparison catches it on '+side)
 note('PASS owned edge comparisons and compiling convergence-budget mutant')
except Exception as error:note('FAIL '+repr(error));raise
finally:log.close()
