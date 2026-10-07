#!/usr/bin/env python3
"""Compare stage1 files added after the first complete corpus manifest froze."""
import subprocess, sys, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[5]
owned=Path(__file__).resolve().parent
artifacts=Path(sys.argv[1])
scratch=Path(tempfile.mkdtemp(prefix='wave14-delta-'))
log=(owned/'evidence/delta.log').open('w',buffering=1)
old={row.split('\t')[0] for row in (artifacts/'corpus-prefer-as-const.manifest').read_text().splitlines()}
files=sorted(p for p in (root/'stage1').rglob('*') if p.suffix in ('.ts','.a'))
delta=[p for p in files if str(p) not in old]
print(f'original compiler=77 stage1={len(old)-77}; added stage1={len(delta)}; final stage1={len(files)}',file=log)
def observe(args,label):
 output=scratch/(label+'.out');errors=scratch/(label+'.err')
 with output.open('wb') as out,errors.open('wb') as err:
  result=subprocess.run([str(a) for a in args],cwd=root,stdout=out,stderr=err)
 assert result.returncode==0 and errors.stat().st_size==0,(args,result.returncode,errors.read_text())
 return output.read_bytes()
node=['node','--disable-warning=ExperimentalWarning',root/'oracle/node.mjs']
for name in ['no-unnecessary-type-constraint','prefer-as-const','prefer-enum-initializers']:
 manifest=scratch/(name+'.manifest');manifest.write_text(''.join(f'{p}\t@typescript-eslint/{name}\t{p}\n' for p in delta))
 want=observe([artifacts/'oracle',manifest],name+'-Go')
 for side,args in [('Node',node+[owned/'complete_runner.a',manifest]),('native',[artifacts/'native',manifest]),('emitted JavaScript',node+[artifacts/'emitted.mjs',manifest])]:
  assert observe(args,name+'-'+side)==want
 count=observe([artifacts/'oracle',manifest,'--count'],name+'-count').decode().strip()
 print(f'{name}: all four identical {len(want)} bytes; findings={count}; files={len(delta)}',file=log,flush=True)
print('PASS complete source coverage after the claimed backlog ports were added',file=log)
log.close()
