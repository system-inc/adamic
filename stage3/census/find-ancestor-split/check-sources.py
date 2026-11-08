import hashlib,json,sys
from pathlib import Path
out=Path(__file__).resolve().parent;root=Path('/workspace/scratch/fa-adapted')
actual={str(p.relative_to(root)):hashlib.sha256(p.read_bytes()).hexdigest() for d in ['src/compiler','src/tsc'] for p in sorted((root/d).rglob('*')) if p.is_file() and p.suffix in ['.ts','.a']}
expected=json.loads((out/'baseline-source-hashes.json').read_text())
(out/'source-hashes.json').write_text(json.dumps(actual,indent=2)+'\n')
diff={p:dict(expected=expected.get(p),actual=actual.get(p)) for p in sorted(expected.keys()|actual.keys()) if expected.get(p)!=actual.get(p)}
(out/'source-hash-check.json').write_text(json.dumps(dict(match=not diff,expected_files=len(expected),actual_files=len(actual),differences=diff),indent=2)+'\n')
print('source hash check child exit: '+str(bool(diff)*1),flush=True)
sys.exit(bool(diff))
