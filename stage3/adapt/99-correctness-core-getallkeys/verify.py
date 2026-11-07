"""Run the full checks only after both full-apply commands have exited zero."""
import difflib,json,os,pathlib,sys
from supervise import run,SCRATCH
repo=pathlib.Path('/workspace/adamic');owner=repo/'stage3/adapt/99-correctness-core-getallkeys'
for side in ['before','after']:
    status=json.loads((SCRATCH/('apply-'+side+'-status.json')).read_text())
    assert status['exit_code']==0 and status['stopping_reason'] is None
    status=run('checks-'+side,['python3',str(owner/'check-tree.py'),side],limit=1800)
    if status['exit_code']:raise SystemExit(1)
status=run('mutant',['python3',str(owner/'mutant.py')],limit=1800)
if status['exit_code']:raise SystemExit(1)
env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
for name,command in [
    ('site-compare',['node',str(owner/'site-compare.cjs'),'/workspace/getallkeys-before','/workspace/getallkeys-after',str(SCRATCH/'mutant'),str(SCRATCH),str(owner/'sites.json')]),
    ('compare',['node',str(owner/'compare.cjs'),'/workspace/getallkeys-before','/workspace/getallkeys-after',str(SCRATCH/'compare-source'),str(owner/'comparison.json')]),
]:
    status=run(name,command,env=env)
    if status['exit_code']:raise SystemExit(1)
before=pathlib.Path('/workspace/getallkeys-before');after=pathlib.Path('/workspace/getallkeys-after');hunks=[];records=[]
for a in sorted((before/'built/local').rglob('*')):
    if not a.is_file() or a.suffix!='.js':continue
    relative=a.relative_to(before/'built/local');b=after/'built/local'/relative
    diff=''.join(difflib.unified_diff(a.read_text().splitlines(True),b.read_text().splitlines(True),fromfile='before/'+str(relative),tofile='after/'+str(relative)))
    if diff:hunks.append(diff);records.append({'file':str(relative),'hunks':sum(l.startswith('@@') for l in diff.splitlines())})
(owner/'javascript-hunks.diff').write_text(''.join(hunks))
file='typescript.internal.d.ts';a=(before/'built/local'/file).read_text();b=(after/'built/local'/file).read_text()
(owner/'declaration-hunk.diff').write_text(''.join(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='before/'+file,tofile='after/'+file)))
source=[]
for file in ['src/compiler/core.ts','src/harness/fourslashImpl.ts']:
    a=(before/file).read_bytes().decode();b=(after/file).read_bytes().decode()
    source.extend(difflib.unified_diff(a.splitlines(True),b.splitlines(True),fromfile='before/'+file,tofile='after/'+file))
(owner/'source.diff').write_text(''.join(source))
(owner/'hunks.json').write_text(json.dumps(records,indent=2)+'\n');print(json.dumps({'javascriptHunks':records,'allChecksPassed':True}),flush=True)
