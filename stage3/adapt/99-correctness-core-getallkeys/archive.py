"""Pack bulky raw evidence with internal and external SHA-256 manifests."""
import hashlib,io,json,pathlib,shutil,tarfile
from supervise import SCRATCH
owner=pathlib.Path(__file__).resolve().parent
hashfile=lambda p:hashlib.sha256(p.read_bytes()).hexdigest()
for side in ['before','after']:
    tree=pathlib.Path('/workspace/getallkeys-'+side)
    manifest={str(p.relative_to(tree)):hashfile(p) for p in sorted((tree/'src').rglob('*')) if p.is_file()}
    (SCRATCH/('source-'+side+'-manifest.json')).write_text(json.dumps(manifest,indent=2)+'\n')
    shutil.copyfile(tree/'patch-set.md',SCRATCH/('patch-set-'+side+'.md'))
    for file in ['src/compiler/core.ts','src/harness/fourslashImpl.ts']:
        dest=SCRATCH/'owner-snapshots'/(side+'-'+pathlib.Path(file).name+'.txt');dest.parent.mkdir(exist_ok=True);shutil.copyfile(tree/file,dest)
for source,destination in [(SCRATCH/'replay/summary.json',owner/'replay-summary.json'),(SCRATCH/'lane/summary.json',owner/'lane-summary.json')]:
    if source.exists():shutil.copyfile(source,destination)
def include(p):
    relative=p.relative_to(SCRATCH)
    if relative.parts[0] in ['mutant','compare-source']:return False
    if relative.parts[:2] in [('replay','before'),('replay','after')]:return False
    if p.name in ['adamic','latent'] or '__pycache__' in relative.parts:return False
    if 'adapted-tree' in relative.parts or '.git' in relative.parts or 'node_modules' in relative.parts:return False
    return p.is_file() and not p.is_symlink()
files=sorted(p for p in SCRATCH.rglob('*') if include(p));manifest=''.join(hashfile(p)+'  '+str(p.relative_to(SCRATCH))+'\n' for p in files)
archive=owner/'proof-evidence.tar.gz'
with tarfile.open(archive,'w:gz') as tar:
    for p in files:tar.add(p,arcname=str(p.relative_to(SCRATCH)),recursive=False)
    data=manifest.encode();member=tarfile.TarInfo('SHA256SUMS');member.size=len(data);tar.addfile(member,io.BytesIO(data))
with tarfile.open(archive) as tar:
    expected=dict(line.split('  ',1)[::-1] for line in tar.extractfile('SHA256SUMS').read().decode().splitlines())
    actual={m.name:hashlib.sha256(tar.extractfile(m).read()).hexdigest() for m in tar.getmembers() if m.isfile() and m.name!='SHA256SUMS'}
    assert actual==expected,'archive digest verification failed'
    assert all(not n.endswith('.go') for n in actual),'a physical Go file lacks .txt'
sha=hashfile(archive);(owner/'proof-evidence.sha256').write_text(sha+'  proof-evidence.tar.gz\n')
index={'archive':archive.name,'sha256':sha,'bytes':archive.stat().st_size,'evidenceFiles':len(files),'internalManifest':'SHA256SUMS','internalManifestVerified':True,'excludedSourceCopiesAndBinaries':True,'ordinaryCompilerAndCensusBinaryHashes':'provenance.json','excludedAttempts':'excluded/ retains packaging overlap and proof-comparison corrections; no failed result is counted as accepted evidence'}
(owner/'EVIDENCE.json').write_text(json.dumps(index,indent=2)+'\n');print(json.dumps(index),flush=True)
