import sys
sys.dont_write_bytecode=True
import hashlib,json,tarfile
from pathlib import Path
here=Path(__file__).resolve().parent;scratch=Path('/workspace/scratch/clear-generic')
# Keep transformed sources and generated proof streams, never node_modules,
# Git object stores or compiler/native binaries. Large trees stay outside Git.
selected=[]
for p in scratch.rglob('*'):
 if not p.is_file() or p.is_symlink():continue
 rel=p.relative_to(scratch)
 if rel.parts[0] in ['base','before','baseline','mutant']:continue
 if rel.parts[0]=='replay' and len(rel.parts)>1 and rel.parts[1].startswith('replay-') and p.suffix not in ['.log']:continue
 if p.suffix in ['.log','.json','.jsonl','.txt','.py','.cjs','.ts','.mjs'] or p.name.endswith('.go.txt'):selected.append(p)
manifest={str(p.relative_to(scratch)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(selected)}
(here/'clear-evidence-manifest.json').write_text(json.dumps(manifest,indent=2)+'\n')
archive=here/'clear-evidence.tar.gz'
with tarfile.open(archive,'w:gz') as tar:
 for p in sorted(selected):tar.add(p,arcname=str(p.relative_to(scratch)),recursive=False)
(here/'clear-evidence.tar.gz.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
print(json.dumps({'packedFiles':len(selected),'archiveBytes':archive.stat().st_size,'archiveSha256':hashlib.sha256(archive.read_bytes()).hexdigest()},indent=2))
