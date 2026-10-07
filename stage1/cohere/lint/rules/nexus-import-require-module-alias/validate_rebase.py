#!/usr/bin/env python3
"""Refresh fixtures and mutants when a rebase preserves all compiled inputs."""
import os, subprocess, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[5]
owned=Path(__file__).resolve().parent
scratch=Path(tempfile.mkdtemp(prefix='wave14-identity-recheck-'))
log=(owned/'evidence/parking-latest-nine-rules.log').open('w',buffering=1)
def note(text):print(text,file=log,flush=True)
def observe(args):
 out=scratch/'out';err=scratch/'err'
 with out.open('wb') as stdout,err.open('wb') as stderr:
  p=subprocess.run(list(map(str,args)),cwd=root,stdout=stdout,stderr=stderr)
 assert p.returncode==0 and err.stat().st_size==0,(args,p.returncode,err.read_text())
 return out.read_bytes()
changed=subprocess.check_output(['git','diff','--name-only','b8fb957aa','origin/main','--','cmd','internal','go.mod','go.sum','cohere','oracle','stage1'],cwd=root,text=True).splitlines()
assert changed==[],changed
note('Main changes only uncompiled stage3/record paths; compiler, dependencies, runners, tools and stage1 corpus retain exact input identity.')
node=['node','--disable-warning=ExperimentalWarning',root/'oracle/node.mjs']
groups=[
 ('/tmp/adamic-gate/wave14-complete-17q4r2di','typescript-eslint-prefer-as-const/complete_runner.a',['no-unnecessary-type-constraint','prefer-as-const','prefer-enum-initializers']),
 ('/tmp/adamic-gate/wave14-backlog-x_cf6lnp','typescript-eslint-prefer-as-const/backlog_runner.a',['no-extra-non-null-assertion','no-confusing-non-null-assertion','no-unnecessary-parameter-property-assignment']),
 ('/tmp/adamic-gate/wave14-fifth-gp7bfc3k','no-lonely-if/complete_runner.a',['no-lone-blocks','no-lonely-if','no-loss-of-precision'])]
for directory,relative,names in groups:
 artifacts=Path(directory)
 # These source trees are immutable copies of the same .a/.ts inputs verified above.
 snapshot=Path('/tmp/adamic-gate/wave14-parking-99sn1qgz')
 runner=snapshot/'stage1/cohere/lint/rules'/relative
 for name in names:
  manifest=artifacts/('fixtures-'+name+'.manifest')
  want=observe([artifacts/'oracle',manifest])
  for side,args in [('Node',node+[runner,manifest]),('native',[artifacts/'native',manifest]),('emitted JavaScript',node+[artifacts/'emitted.mjs',manifest])]:
   assert observe(args)==want,(name,side)
  note(name+': fresh four-way fixture bytes PASS '+str(len(want)))
  manifest=artifacts/('mutant-'+name+'.manifest')
  want=observe([artifacts/'oracle',manifest])
  source=artifacts/('mutant-'+name)/'stage1/cohere/lint/rules'/relative
  for side,args in [('Node',node+[source,manifest]),('native',[artifacts/('mutant-'+name+'-native'),manifest]),('emitted JavaScript',node+[artifacts/('mutant-'+name+'.mjs'),manifest])]:
   assert observe(args)!=want,(name,side)
   note(name+': clean compiling mutant caught by Go comparison on '+side)
note('PASS current-main identity recheck, fresh fixtures and all nine mutants. Full corpus evidence reused only because compiler, runners, dependencies, tools and corpus are unchanged.')
log.close()
