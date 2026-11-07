from pathlib import Path
import json,subprocess,shutil,os
s=Path('/workspace/scratch/arrayof');repo=Path('/workspace/adamic');own=repo/'stage3/adapt/99-correctness-core-arrayof';base=json.loads((s/'provenance.json').read_text())['base'];assert (s/'apply-before-proof.exit').read_text().strip()=='0';assert (s/'apply-after.exit').read_text().strip()=='0'
with (s/'frozen-adapter-check.log').open('w') as log:subprocess.run(['sha256sum','-c',str(s/'frozen-adapter-sha256.txt')],cwd=repo,stdout=log,stderr=subprocess.STDOUT,check=True)
key=(s/'after-product-key.txt').read_text().strip();shutil.copyfile(Path('/home/agent/.cache/adamic-stage3/products')/(key+'.manifest'),s/'after-product.manifest.json')
(own/'PROOF.md').write_text('''# Full-apply checkpoint

Base: origin/main f0271c6a7b2e92900eb7046d8246cfb89f52cfb6.
TypeScript pin: 050880ce59e30b356b686bd3144efe24f875ebc8.

| Check | Result |
| --- | --- |
| Whole-lane measurement | 398 units; 860 integer calls; 0 excluded-class calls; 0 counterexamples |
| Whole-lane control and reconciliation | Pass; exact test multiset and baseline diff |
| Baseline fresh full apply with product packaging | Exit 0; 81,504 inputs match built baseline |
| Main plus arrayof full apply with product packaging | Exit 0 |
| Owner and no-write contract | 11 checks pass |
| Before types | 79 roots; exit 0; 0 diagnostics |
| Before full-core census | Exit 0; F2 present at core.ts:1339:12 |
| Paired replay, after types, JavaScript, declarations, census and mutant | Pending |

The failed initial baseline product-packaging attempt and layout preflights are
retained separately. They are excluded from accepted measurement and apply
results. The frozen adapter and owner pins are unchanged by evidence packaging.
''')
with (s/'pack-checkpoint.log').open('w') as log:subprocess.run(['python3',str(s/'pack.py'),str(own/'checkpoint.tar.gz'),'apply-after.log','apply-after.exit','after-product-key.txt','after-product.manifest.json','frozen-adapter-sha256.txt','frozen-adapter-check.log','contract.json','contract.log','pack-measurement.log'],stdout=log,stderr=subprocess.STDOUT,check=True)
with (s/'push-checkpoint.log').open('w') as log:
 def git(*args,capture=False,input=None):
  r=subprocess.run(['git',*args],cwd=repo,text=True,input=input,stdout=subprocess.PIPE if capture else log,stderr=log,check=True);return r.stdout.strip() if capture else None
 git('add','--',str(own.relative_to(repo)),'stage3/upstream/LEDGER.md');paths=git('diff','--cached','--name-only',base,capture=True).splitlines();assert all(p.startswith(str(own.relative_to(repo))+'/') or p=='stage3/upstream/LEDGER.md' for p in paths),paths;tree=git('write-tree',capture=True);commit=git('commit-tree',tree,'-p',base,capture=True,input='checkpoint: arrayOf count and both full applies (b5w3ycg)\n');branch='typescript/core-arrayof-checkpoint-b5w3ycg';git('push','origin',commit+':refs/heads/'+branch);remote=git('ls-remote','--heads','origin',branch,capture=True);assert remote.split()[0]==commit
record=dict(commit=commit,branch=branch,base=base,files=len(paths),bothFullAppliesExit=0,remoteVerified=True);(s/'CHECKPOINT.json').write_text(json.dumps(record,indent=2)+'\n');shutil.copyfile(s/'CHECKPOINT.json',own/'CHECKPOINT.json');print(json.dumps(record),flush=True)
