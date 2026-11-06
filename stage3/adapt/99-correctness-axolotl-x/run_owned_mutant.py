#!/usr/bin/env python3
"""Build the one-site restored ownership-revision fixture and require one detection."""
import difflib
import gzip
import json
import os
from pathlib import Path
import shutil
import subprocess
p=Path(__file__).resolve().parent;repo=p.parents[2];source=Path('/workspace/bind-owned-after');target=Path('/workspace/bind-owned-mutant');scratch=Path('/workspace/scratch');target.mkdir(exist_ok=False)
for item in source.iterdir():
    if item.name in ['built','.git']:continue
    if item.name in ['node_modules','tests','lib']:(target/item.name).symlink_to(item,target_is_directory=True)
    elif item.is_dir():shutil.copytree(item,target/item.name,symlinks=True)
    else:shutil.copyfile(item,target/item.name)
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
for label,command,cwd in [('restore',['node',str(p/'mutant.cjs'),str(target)],repo),('build',['npm','run','build:compiler'],target)]:
    with (scratch/('bind-owned-mutant-'+label+'.log')).open('w') as f:subprocess.run(command,cwd=cwd,env=env,stdout=f,stderr=subprocess.STDOUT,check=True)
with (scratch/'bind-owned-mutant-compare.log').open('w') as f:
    result=subprocess.run(['node',str(p/'compare.cjs'),'/workspace/bind-before',str(target),str(scratch/'bind-owned-mutant-compare')],cwd=repo,env=env,stdout=f,stderr=subprocess.STDOUT)
assert result.returncode==1
report=json.loads((scratch/'bind-owned-mutant-compare/comparison.json').read_text());assert not report['pass'] and len(report['unexpectedSites'])==1
assert report['unexpectedSites'][0]['owner']=='createBuilderProgram.emitBuildInfo'
for row in report['files']:
    if not row['byteIdentical']:assert row['boundSites']==27 and len(row['remainingOriginal'])==1
out=p/'owned';out.mkdir(exist_ok=True)
(out/'mutant-comparison.json').write_text(json.dumps(report,indent=2)+'\n')
for label in ['restore','build','compare']:
    with gzip.open(out/('mutant-'+label+'.log.gz'),'wb') as f:f.write((scratch/('bind-owned-mutant-'+label+'.log')).read_bytes())
a=(source/'src/compiler/builder.ts').read_text();b=(target/'src/compiler/builder.ts').read_text()
assert ''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='a/src/compiler/builder.ts',tofile='b/src/compiler/builder.ts'))==(p/'mutant.patch').read_text()
paths=[str(f.relative_to(repo)) for f in out.glob('mutant-*')]+[str((p/'run_owned_mutant.py').relative_to(repo)),str((p/'publish_paths.py').relative_to(repo))]
subprocess.run(['python3',str(p/'publish_paths.py'),'Pass the 28-site one-restored-site negative control',*paths],cwd=repo,check=True)
