import hashlib,json,tarfile
from pathlib import Path
out=Path(__file__).resolve().parent;scratch=Path('/workspace/scratch/after-wave')
status=json.loads((scratch/'entry-latent-status.json').read_text());assert status['exit_code']!=0
completed=[json.loads(l) for l in (scratch/'entry-unit-times.jsonl').read_text().splitlines()]
stop=dict(status=status,last_completed_timing=completed[-1] if completed else None,active_unit='not recorded; last completed unit is not the interrupted unit',raw_records=len((scratch/'artifacts/tsc/latent.jsonl').read_bytes().splitlines()))
(out/'ENTRY-STOP.json').write_text(json.dumps(stop,indent=2)+'\n')
archive=out/'entry-probe.tar.gz'
with tarfile.open(archive,'w:gz',compresslevel=9) as t:
 for name,path in [('entry-latent-status.json',scratch/'entry-latent-status.json'),('entry-census-status.json',scratch/'entry-census-status.json'),('latent.jsonl',scratch/'artifacts/tsc/latent.jsonl'),('census.jsonl',scratch/'artifacts/tsc/census.jsonl'),('unit-times.jsonl',scratch/'entry-unit-times.jsonl'),('entry-latent.log',scratch/'entry-latent.log')]:t.add(path,arcname=name,recursive=False)
(out/'entry-probe.tar.gz.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  entry-probe.tar.gz\n')
print(json.dumps(stop,indent=2))
