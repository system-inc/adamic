#!/usr/bin/env python3
"""Archive owned command outputs and source pins without executable binaries."""
import gzip
import hashlib
import json
import pathlib
import subprocess
UNIT = pathlib.Path(__file__).resolve().parent
ROOT = UNIT.parents[3]
OUT = pathlib.Path('/workspace/wave-07-jsx')
EVIDENCE = UNIT/'evidence'
EVIDENCE.mkdir(exist_ok=True)
metadata = dict(source_commit=subprocess.check_output(['git','rev-parse','HEAD'],cwd=ROOT,text=True).strip(), main=subprocess.check_output(['git','rev-parse','origin/main'],cwd=ROOT,text=True).strip(), files={}, owned_sources={})
for path in UNIT.rglob('*'):
    if path.is_file() and 'evidence' not in path.parts:
        metadata['owned_sources'][str(path.relative_to(UNIT))] = hashlib.sha256(path.read_bytes()).hexdigest()
for label, directory in [('jsx', OUT), ('landing', pathlib.Path('/workspace/wave-07-latest-landing'))]:
    for path in directory.rglob('*'):
        if not path.is_file() or (label=='jsx' and 'native' in path.relative_to(directory).parts): continue
        if label=='landing' and path.parent!=directory: continue
        if path.suffix not in ['.stdout','.stderr','.json','.log','.manifest','.tsx','.cjs','.a','.ts']: continue
        data=path.read_bytes()
        target=EVIDENCE/label/path.relative_to(directory)
        target=target.with_name(target.name+'.gz')
        target.parent.mkdir(parents=True,exist_ok=True)
        target.write_bytes(gzip.compress(data,mtime=0))
        metadata['files'][str(target.relative_to(EVIDENCE))]=dict(bytes=len(data),sha256=hashlib.sha256(data).hexdigest())
for name in ['jsx-setup.log','jsx-selection.json','jsx-rebase.log','jsx-validation-rebased.log','jsx-landing-gates.log','jsx-benchmark.log','jsx-source-lint-final.log','jsx-library-proof.log','jsx-main-build.log','jsx-gofmt.log','jsx-python-check.log','named-listener-gate.log','named-source-lint.log','named-landing-gates.log','next-setup.log','next-nproc.log','final-trio-selection.json','named-landing-rebase.log','rebase-b8.log','named-listener-b8-gate.log','named-library-b8-gate.log','named-landing-resumed.log','named-node-resumed.log','named-b8-benchmark.log']:
    path=pathlib.Path('/workspace/wave-07-artifacts')/name
    if path.exists():
        data=path.read_bytes();target=EVIDENCE/(name+'.gz');target.write_bytes(gzip.compress(data,mtime=0))
        metadata['files'][target.name]=dict(bytes=len(data),sha256=hashlib.sha256(data).hexdigest())
(EVIDENCE/'metadata.json').write_text(json.dumps(metadata,indent=2)+'\n')
print('Archived',len(metadata['files']),'streams; source',metadata['source_commit'],'main',metadata['main'])
