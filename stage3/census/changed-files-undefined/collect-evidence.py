import gzip,hashlib,json,shutil
from pathlib import Path
out=Path(__file__).resolve().parent; dest=out/'evidence';dest.mkdir(exist_ok=True)
plain=['inputs.json','provenance.json','source-manifests.json','census-delta.json','reader-inventory.json','javascript.json','main-witness.json','after-witness.json','amended-owner-replay.json','contract.log','latent-audit.log','lane-inputs.log','main-receipt.json','after-receipt.json','restoring-mutant-receipt.json','main-types-receipt.json','after-types-receipt.json','env.sh','prepare-baseline.sh','prepare-pair.py','amend-owner.py','build-tools.sh','run-probe.py','run-types.py','restore-mutant.py','compare-census.py','collect-evidence.py','verification.json','fresh-apply.json','implementation-delta.json','finish-verification.py','write-report.py','REPORT.md','COMMANDS.md']
compressed=['main.jsonl','after.jsonl','restoring-mutant.jsonl','main-times.jsonl','after-times.jsonl','restoring-mutant-times.jsonl','main.log','after.log','restoring-mutant.log','main-types.log','after-types.log','main-sys.js','after-sys.js','sys-restoring-mutant.ts.txt','main-apply.log','amended-apply.log','lane-product.jsonl','lane-whole.jsonl','lane-npm-ci.log','overlay.log','latent-build.log','adamic-build.log']
for n in plain:
 p=out/n
 if p.exists():shutil.copy2(p,dest/n)
for n in compressed:
 p=out/n
 if p.exists():(dest/(n+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
overlay=dest/'overlay';overlay.mkdir(exist_ok=True)
for p in (out/'overlay').iterdir():
 if p.is_file():shutil.copy2(p,overlay/p.name)
for n in ['main','after']:
 (dest/(n+'-sys.ts.txt.gz')).write_bytes(gzip.compress((out/(n+'-tree/src/compiler/sys.ts')).read_bytes(),mtime=0))
# The invalid preview is retained separately and explicitly excluded from the authoritative pair.
preview=dest/'excluded-preview';preview.mkdir(exist_ok=True)
for p in out.glob('invalid-lf-preview*'):
 if p.is_file():(preview/(p.name+'.gz')).write_bytes(gzip.compress(p.read_bytes(),mtime=0))
files={str(p.relative_to(dest)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted(dest.rglob('*')) if p.is_file() and p.name!='SHA256.json'}
(dest/'SHA256.json').write_text(json.dumps(files,indent=2)+'\n')
print(len(files),'evidence files')
