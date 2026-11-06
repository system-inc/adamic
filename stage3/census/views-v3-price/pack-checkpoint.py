#!/usr/bin/env python3
"""Pack the incomplete source investigation without implying census completion."""
import hashlib, io, json, pathlib, tarfile
root=pathlib.Path(__file__).resolve().parent
scratch=pathlib.Path('/workspace/scratch')
items={}
for name in ['v3-setup.log','v3-npm.log','v3-fetch.log','v3-worktree.log','v3-ts-clone.log','v3-ts-shallow.log','v3-fixture-proof.log','v3-main-apply.log','v3-source.log','v3-source.json','v3-source.jsonl.gz']:
    p=scratch/name
    if p.exists(): items[name]=p.read_bytes()
items['provenance.json']=json.dumps(dict(box='4b7d656a5150',head='9626d0113f08269e519ba216ff7f95ff84af16db',main='2c3f322975b2dee68889e82eab0bbca324d417a9',report_base='12fec995dcadc7f1ae94e4ba63835884451378ae',typescript='050880ce59e30b356b686bd3144efe24f875ebc8',status='incomplete; no certified ceiling or lowered totals'),indent=2).encode()+b'\n'
source=pathlib.Path('/workspace/v3-main-adapted/src')
if source.exists():
    items['adapted-src.sha256']=''.join(f'{hashlib.sha256(p.read_bytes()).hexdigest()}  {p.relative_to(source.parent)}\n' for p in sorted(source.rglob('*')) if p.is_file()).encode()
items['SHA256SUMS']=''.join(f'{hashlib.sha256(data).hexdigest()}  {name}\n' for name,data in sorted(items.items())).encode()
archive=root/'raw-checkpoint.tar.gz'
with tarfile.open(archive,'w:gz') as tar:
    for name,data in sorted(items.items()):
        info=tarfile.TarInfo(name);info.size=len(data);info.mtime=0
        tar.addfile(info,io.BytesIO(data))
(root/'SHA256SUMS').write_text(f'{hashlib.sha256(archive.read_bytes()).hexdigest()}  {archive.name}\n')
print(f'Packed {len(items)} files')
