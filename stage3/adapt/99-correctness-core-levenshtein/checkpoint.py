import hashlib,json,tarfile,io,subprocess
from pathlib import Path
s=Path('/workspace/scratch/levenshtein');repo=Path('/workspace/adamic');out=repo/'stage3/adapt/99-correctness-core-levenshtein'
after=json.loads((s/'after-apply.json').read_text());assert after['exit']==0
before=json.loads((s/'prepare.json').read_text())['apply-before'];assert before['exit']==0
pin=json.loads((out/'owners.json').read_text())[0];a=(s/'before/src/compiler/core.ts').read_bytes();b=(s/'after/src/compiler/core.ts').read_bytes();newline='\r\n' if b'\r\n' in a else '\n';assert a.replace(pin['before'].replace('\n',newline).encode(),pin['after'].replace('\n',newline).encode())==b
names=sorted(p.relative_to(s/'before/src') for p in (s/'before/src').rglob('*') if p.is_file());changed=[str(n) for n in names if (s/'before/src'/n).read_bytes()!=(s/'after/src'/n).read_bytes()];assert changed==['compiler/core.ts']
(out/'source-scope.json').write_text(json.dumps({'sourceFiles':len(names),'changed':changed,'exactPinnedOwnerOnly':True},indent=2)+'\n')
(out/'apply-results.json').write_text(json.dumps({'before':before,'after':after},indent=2)+'\n')
(out/'PROOF.md').write_text('''# Post-apply checkpoint

| Check | Result |
| --- | --- |
| Real-call count | 798965 calls, 0 differences, all 398 accepted shards |
| Unchanged full lane control | Pass; identities, counts and raw baseline diff match |
| Owner contract | All 29 checks pass |
| Full main apply | Exit 0 |
| Full main plus this adapter apply | Exit 0 |
| Source scope | All 746 sources compared; exact pinned owner is the only change |
| 79-root types, full paired replay, JavaScript/declarations, census and mutant | Pending final proof run |

Both complete apply logs, results and frozen adapter hashes are packed in
apply-evidence.tar.gz. This checkpoint preserves the completed count and
implementation before the remaining proofs. The initial report remains
historical in initial-report.tar.gz. No completed proof is claimed for the
pending items. No earlier adapter or compiler source changes.
''')
files=[s/n for n in ['apply-before.log','apply-after.log','after-apply.json','prepare.json','adapter-frozen.json','reconciled.json','contract-before-apply-checkpoint.log']]
archive=out/'apply-evidence.tar.gz'
with tarfile.open(archive,'w:gz') as tar:
 for p in files:tar.add(p,arcname=p.name)
 data=''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in files).encode();info=tarfile.TarInfo('SHA256SUMS');info.size=len(data);tar.addfile(info,io.BytesIO(data))
(out/'apply-evidence.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
print('post-apply checkpoint ready',len(list(out.iterdir())),flush=True)
