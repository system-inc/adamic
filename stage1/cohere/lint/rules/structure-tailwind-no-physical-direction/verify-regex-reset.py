#!/usr/bin/env python3
"""Prove the published global regexp's fresh-search reset is load-bearing."""
import argparse,json,shutil,subprocess
from pathlib import Path
owned=Path(__file__).resolve().parent
repo=owned.parents[4]
p=argparse.ArgumentParser();p.add_argument('--scratch',type=Path,required=True);args=p.parse_args()
scratch=args.scratch.resolve()
def run(cmd,label):
 with (scratch/(label+'.log')).open('wb') as out:
  result=subprocess.run(list(map(str,cmd)),cwd=repo,stdout=out,stderr=subprocess.PIPE)
 (scratch/(label+'.stderr')).write_bytes(result.stderr)
 assert result.returncode==0 and not result.stderr,(label,result.returncode,result.stderr[:800])
 return (scratch/(label+'.log')).read_bytes()
tree=scratch/'reset-tree'
for source in (repo/'stage1').rglob('*'):
 if source.suffix in ['.a','.ts']:
  target=tree/source.relative_to(repo);target.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(source,target)
rule=tree/'stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/rule.a'
s=rule.read_text();assert s.count('pattern.lastIndex = 0;')==1;rule.write_text(s.replace('pattern.lastIndex = 0;','pattern.lastIndex = 1;'))
entry=tree/(owned/'standalone.a').relative_to(repo);binary=scratch/'regex-reset-mutant'
run(['go','run','-overlay='+str(scratch/'builder-overlay.json'),repo/'wave15_owned_build.go',entry,binary],'regex-reset-mutant-build')
manifest=scratch/'regex-reset.manifest'
rows=(scratch/'fixtures.manifest').read_text().splitlines();manifest.write_text('\n'.join(r for r in rows if r.split('\t')[1]=='structure/tailwind-no-physical-direction')+'\n')
expected=run([scratch/'oracle',manifest],'regex-reset-Go')
for side,cmd in [('Node',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',entry,manifest]),('JavaScript',['node','--disable-warning=ExperimentalWarning',repo/'oracle/node.mjs',str(binary)+'.mjs',manifest]),('native',[binary,manifest])]:
 assert run(cmd,'regex-reset-'+side)!=expected,'reset mutant survived on '+side
 print('lastIndex 0 -> 1 mutant caught on '+side+' only by output comparison; successful compile and exit',flush=True)
