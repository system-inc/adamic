"""Resume after the user's head replacement, retaining the in-flight A parent."""
import hashlib,json,os,subprocess,sys,time,shutil
from pathlib import Path
out=Path(__file__).resolve().parent;repo=out.parents[2];base='72f23939c773ba0045aeed5d6afb9b025b6473a0';head='5a86e3b3588914eb367f800eb7339db8d1bd70f9'
def ready(name):
 p=out/(name+'.status.json')
 return p.exists() and json.loads(p.read_text())['exit']==0
def hashes(root):return {str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for d in ['src/compiler','src/tsc'] for p in sorted((root/d).rglob('*')) if p.is_file() and p.suffix in ['.ts','.a']}
def run(name,cwd,cmd):
 rc=subprocess.run([sys.executable,str(out/'run.py'),name,str(cwd),*cmd]).returncode;assert rc==0,(name,rc)
while not ready('B-apply') or not ready('head-build'):time.sleep(2)
assert subprocess.check_output(['git','rev-parse','HEAD'],cwd='/tmp/fa-head',text=True).strip()==head
A=hashes(Path('/tmp/fa-adapted'));B=hashes(Path('/tmp/fa-B-adapted'));assert A==B
assert A==json.loads((out/'source-hashes.json').read_text())
(out/'B-source-hashes.json').write_text(json.dumps(B,indent=2)+'\n');(out/'B-patch-set.md').write_bytes(Path('/tmp/fa-B-adapted/patch-set.md').read_bytes())
for label,root in [('A',repo),('B',Path('/tmp/fa-B-report'))]:run(label+'-apply-key',root,['bash','stage3/apply.sh','--key'])
assert (out/'A-apply-key.log').read_text()==(out/'B-apply-key.log').read_text()
dirs=['stage3/adapt/99-correctness-core-or-predicate','stage3/adapt/99-correctness-core-multimap']
proof={'parent':base,'adapter':'eeb87728b743235cf1df68dfed17241f63416734','directories':{},'apply_keys_equal':True,'all_81_measured_source_hashes_equal':True,'reuse':'Both pairings have the same compiler pins and adapted bytes. B reuses A streams and focused observations; its pair tool runs separately over the verified B tree.'}
for d in dirs:
 a=subprocess.check_output(['git','rev-parse',base+':'+d],cwd=repo,text=True).strip();b=subprocess.check_output(['git','rev-parse','eeb87728:'+d],cwd=repo,text=True).strip();assert a==b
 proof['directories'][d]={'parent_tree':a,'adapter_tree':b}
(out/'AB-equivalence.json').write_text(json.dumps(proof,indent=2)+'\n')
prov=json.loads((out/'provenance.json').read_text());root=Path('/tmp/fa-head');scopes=['internal/lower','internal/load','stage3/census/latent','stage3/meter/entry.go.txt','stage3/meter/entry_overlay.py']
files=subprocess.check_output(['git','ls-tree','-r','--name-only',head,'--',*scopes],cwd=root,text=True).splitlines()
code={f:hashlib.sha256((root/f).read_bytes()).hexdigest() for f in files if f.endswith(('.go','.go.txt','.py','.cjs')) and '/data/' not in f and '/evidence/' not in f and '/testdata/' not in f}
prov['head']=dict(sha=head,source_hashes=code,binary_sha256=hashlib.sha256(Path('/tmp/fa-head-census').read_bytes()).hexdigest());(out/'provenance.json').write_text(json.dumps(prov,indent=2)+'\n')
meta=json.loads(Path('/tmp/fa-head-entry/overlay.json').read_text());dest=out/'overlays/head';shutil.rmtree(dest);dest.mkdir(parents=True);rows=[]
for i,(logical,actual) in enumerate(sorted(meta['Replace'].items())):
 name=f'{i:02d}-'+Path(actual).name
 if name.endswith('.go'):name+='.txt'
 data=Path(actual).read_bytes();(dest/name).write_bytes(data);rows.append(dict(logical=logical,evidence=str((dest/name).relative_to(out)),sha256=hashlib.sha256(data).hexdigest()))
(dest/'manifest.json').write_text(json.dumps(rows,indent=2)+'\n')
(out/'REPORT.md').write_text('# Find ancestor 5a86e3b3 census checkpoint\n\nThe parent run is retained for A. The replacement head is built from its own pin. B apply succeeded with both requested adaptation directories copied from eeb87728. A and B have identical apply keys and identical measured source hashes, so their census observations will be reused and paired separately. Findings are pending.\n')
subprocess.run(['git','add','-f','--',str(out)],cwd=repo,check=True)
tree=subprocess.check_output(['git','write-tree'],cwd=repo,text=True).strip();sha=subprocess.check_output(['git','commit-tree',tree,'-p',base],cwd=repo,input='census: replacement head and B apply checkpoint\n',text=True).strip();ref='refs/heads/typescript/find-ancestor-census-checkpoint-B-apply-5a86e3b3'
with (out/'checkpoint-B-apply.log').open('w') as log:rc=subprocess.run(['git','push','origin',sha+':'+ref],cwd=repo,stdout=log,stderr=subprocess.STDOUT).returncode
assert rc==0;(out/'checkpoint-B-apply.json').write_text(json.dumps(dict(sha=sha,ref=ref,exit=rc),indent=2)+'\n');print('B checkpoint',sha,'child_exit=0',flush=True)
shutil.rmtree('/tmp/fa-B-adapted/tests',ignore_errors=True)
run('head-run',repo,['env','-u','LATENT_FULL','-u','LATENT_SPECULATIVE','-u','LATENT_ONLY_FILE','LATENT_ASSERT_NO_OUTPUT=1','LATENT_TIMES='+str(out/'head-run-units.jsonl'),'/tmp/fa-head-census','/tmp/fa-adapted/src/tsc/tsc.ts',str(out/'head-run.jsonl')])
while not (out/'parent-run.status.json').exists():time.sleep(2)
assert ready('parent-run')
assert hashes(Path('/tmp/fa-adapted'))==A and hashes(Path('/tmp/fa-B-adapted'))==A
for label,adapted in [('A','/tmp/fa-adapted'),('B','/tmp/fa-B-adapted')]:
 run('pair' if label=='A' else 'B-pair',repo,['python3',str(repo/'stage3/census/pair/pair.py'),str(out/'parent-run.jsonl'),str(out/'head-run.jsonl'),adapted,'--typescript',str(repo/'stage3/api/node_modules/typescript'),'--json',str(out/('paired.json' if label=='A' else 'B-paired.json')),'--report',str(out/('paired.md' if label=='A' else 'B-paired.md'))])
run('focus-all',repo,['python3',str(out/'focus.py')])
print('both pairings complete child_exit=0',flush=True)
