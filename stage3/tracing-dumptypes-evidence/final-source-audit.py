from pathlib import Path
import json,hashlib,shutil
S=Path(__file__).resolve().parent;A=Path('/workspace/scratch/dp24r5m-after');B=Path('/workspace/scratch/dp24r5m-base');F=Path('/workspace/scratch/dp24r5m-fresh');D=Path('/workspace/adamic/stage3/adapt/99-correctness-tracing-dumptypes')
files=sorted(p.relative_to(A/'src') for p in (A/'src').rglob('*') if p.is_file());assert len(files)==747
assert sorted(p.relative_to(F/'src') for p in (F/'src').rglob('*') if p.is_file())==files
assert all((A/'src'/p).read_bytes()==(F/'src'/p).read_bytes() for p in files)
changed=[str(p) for p in files if (A/'src'/p).read_bytes()!=(B/'src'/p).read_bytes()];assert changed==['compiler/tracing.ts']
frozen=json.loads((S/'frozen-adapter.json').read_text());assert all(hashlib.sha256((D/n).read_bytes()).hexdigest()==h for n,h in frozen.items())
rows=[dict(path=str(p),sha256=hashlib.sha256((F/'src'/p).read_bytes()).hexdigest()) for p in files];(S/'fresh-source-hashes.json').write_text(json.dumps(rows,indent=2)+'\n')
r=dict(sourceFiles=747,changedFromBaseline=changed,normalBuiltCandidateMatchesFreshApply=True,adapterAndPinsFrozen=True);(S/'final-source-audit-results.json').write_text(json.dumps(r,indent=2)+'\n');shutil.copyfile(F/'patch-set.md',S/'fresh-patch-set.md');print(json.dumps(r))
