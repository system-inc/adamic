from pathlib import Path
import hashlib,json,shutil
p=Path(__file__).parent;source=Path('/workspace/scratch/timer-ambient-delete-3-after');out=Path('/workspace/scratch/timer-ambient-delete-3-property-mutant');shutil.copytree(source/'src',out/'src')
f=out/'src/compiler/sys.ts';s=f.read_bytes();old=b'clearTimeout?(timeoutId: unknown): void;';new=b'clearTimeout?: (timeoutId: unknown) => void;';assert s.count(old)==1;assert s.count(b'=> { clearTimeout(timeoutId); }')==2;t=s.replace(old,new);f.write_bytes(t)
p.joinpath('property-slot-source.json').write_text(json.dumps({'only_mutation':'System clearTimeout method becomes an optional function property','before_sha256':hashlib.sha256(s).hexdigest(),'after_sha256':hashlib.sha256(t).hexdigest(),'block_forwarders_preserved':True},indent=2)+'\n')
print(out)
