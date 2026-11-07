"""Final scope, archive, source and evidence-integrity checks."""
from pathlib import Path
import json,hashlib,tarfile,subprocess,os
u=Path(__file__).resolve().parent;repo=u.parents[2];s=Path('/workspace/scratch/debug-messages');proof=json.loads((u/'PROOF.json').read_text());base=proof['base']
assert subprocess.check_output(['git','rev-parse','HEAD'],cwd=repo,text=True).strip()==base
assert all(proof['gates'].values());assert len([p for p in u.rglob('*') if p.is_file()])<120
archives=json.loads((u/'ARCHIVES.json').read_text())
for row in archives:
 p=u/row['file'];assert p.stat().st_size==row['bytes']<20_000_000;assert hashlib.sha256(p.read_bytes()).hexdigest()==row['sha256']
with tarfile.open(u/'raw-proof.tar.gz') as tf:
 manifest=json.loads(tf.extractfile('MANIFEST.json').read())
 for name,digest in manifest.items():assert hashlib.sha256(tf.extractfile(name).read()).hexdigest()==digest,name
 for name in ['raw/before-census.jsonl','raw/after-census.jsonl','raw/mutant-census.jsonl','raw/before-types.log','raw/after-types.log','raw/after-apply.json','raw/paired-replay-comparison.json','raw/final-witness.json']:assert name in manifest,name
for side in ['before','after']:
 with tarfile.open(u/(side+'-javascript.tar.gz')) as tf:
  manifest=json.loads(tf.extractfile('manifest.json').read())
  for name,digest in manifest.items():assert hashlib.sha256(tf.extractfile(name).read()).hexdigest()==digest
 with tarfile.open(u/(side+'-replay.tar.gz')) as tf:assert len(json.loads(tf.extractfile('records.json').read()))==96
before_ledger=subprocess.check_output(['git','show',base+':stage3/upstream/LEDGER.md'],cwd=repo);after_ledger=(repo/'stage3/upstream/LEDGER.md').read_bytes();assert after_ledger.startswith(before_ledger);assert len([l for l in after_ledger[len(before_ledger):].decode().splitlines() if l.strip()])==1
for p in u.iterdir():
 if p.suffix in ['.md','.py','.cjs','.json','.patch']:
  text=p.read_text();assert '\u2014' not in text and '\u2013' not in text,p.name
for side in ['before','after']:
 current={str(p.relative_to(s/side)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((s/side/'src').rglob('*')) if p.is_file()};assert current==json.loads((u/'build-comparison.json').read_text())['sourceManifests'][side]
changed=subprocess.check_output(['git','diff','HEAD','--name-only'],cwd=repo,text=True).splitlines();assert all(p.startswith('stage3/adapt/99-correctness-core-debug-messages/') or p=='stage3/upstream/LEDGER.md' for p in changed)
# Task descendants must be gone. Exclude this validator and its shell/tool ancestors.
ancestors={os.getpid()};parent=os.getppid()
while parent>1:
 ancestors.add(parent)
 try:parent=int(Path(f'/proc/{parent}/stat').read_text().split()[3])
 except (FileNotFoundError,ValueError):break
live=[]
for d in Path('/proc').iterdir():
 if not d.name.isdigit() or int(d.name) in ancestors:continue
 try:
  stat=(d/'stat').read_text().split();cmd=(d/'cmdline').read_bytes().replace(b'\0',b' ').decode(errors='replace');cwd=os.readlink(d/'cwd')
 except (FileNotFoundError,PermissionError,ProcessLookupError):continue
 if stat[2]=='Z':continue
 if '/workspace/scratch/debug-messages' in cmd or cwd.startswith('/workspace/scratch/debug-messages'):live.append({'pid':int(d.name),'command':cmd,'cwd':cwd})
assert not live,live
result={'allGatesPassed':True,'archivesVerified':len(archives),'largestArchiveBytes':max(r['bytes'] for r in archives),'rawArchiveManifestVerified':True,'allNormalJavaScriptBytesRetainedAndVerified':True,'sourceManifestsUnchangedSinceBuildComparison':True,'ledgerAppendsExactlyOneNonemptyLine':True,'scopeOnlyNewDirectoryAndLedger':True,'noLiveTaskProcesses':True,'noEmOrEnDashes':True}
(u/'validation.json').write_text(json.dumps(result,indent=2)+'\n')
(u/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+p.name+'\n' for p in sorted(u.iterdir()) if p.is_file() and p.name!='SHA256SUMS'))
print(json.dumps(result))
