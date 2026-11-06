#!/usr/bin/env python3
"""Record every baseline adapter outcome on both trees, then compare them all."""
from concurrent.futures import ThreadPoolExecutor
import base64
import difflib
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys

repo,baseline,accepted,output=map(lambda s:Path(s).resolve(),sys.argv[1:5])
output.mkdir(parents=True,exist_ok=False)
base_env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
hook=Path(__file__).resolve().with_name('replay_io.cjs')
def manifest(tree):
    return {str(p.relative_to(tree)):dict(sha256=hashlib.sha256(p.read_bytes()).hexdigest(),mtime_ns=p.stat().st_mtime_ns)
            for p in tree.rglob('*') if p.is_file() and '.git' not in p.relative_to(tree).parts}
initial={}
for label,tree in [('before',baseline),('after',accepted)]:
    initial[label]=manifest(tree)
    (output/(label+'-initial-manifest.json')).write_text(json.dumps(initial[label],sort_keys=True)+'\n')
def run(label,tree,directory):
    logs=output/label;logs.mkdir(exist_ok=True)
    journal=logs/(directory.name+'.writes.jsonl');journal.write_text('')
    env=dict(base_env,BIND_REPLAY_ROOT=str(tree),BIND_REPLAY_JOURNAL=str(journal))
    env['NODE_OPTIONS']=env.get('NODE_OPTIONS','')+' --require='+str(hook)
    with (logs/(directory.name+'.log')).open('w') as stream:
        r=subprocess.run(['node',str(directory/'adapt.cjs'),str(tree)],cwd=repo,env=env,stdout=stream,stderr=subprocess.STDOUT)
    events=[json.loads(s) for s in journal.read_text().splitlines()]
    files=sorted({e['file'] for e in events});edits=[]
    for file in files:
        rows=[e for e in events if e['file']==file]
        if rows[0]['before']==rows[-1]['after']:continue
        changed=[e for e in rows if e['before']!=e['after']];assert changed
        old=base64.b64decode(changed[0]['beforeData']) if changed[0]['beforeData'] is not None else b''
        new=base64.b64decode(changed[-1]['afterData']) if changed[-1]['afterData'] is not None else b''
        x=old.decode(errors='backslashreplace').splitlines(True);y=new.decode(errors='backslashreplace').splitlines(True)
        diff=list(difflib.unified_diff(x,y,fromfile='before/'+file,tofile='after/'+file))
        changes=[dict(before=''.join(x[i:j]),after=''.join(y[k:l])) for op,i,j,k,l in difflib.SequenceMatcher(None,x,y).get_opcodes() if op!='equal']
        edits.append(dict(file=file,changes=changes,added=sum(s.startswith('+') and not s.startswith('+++') for s in diff),removed=sum(s.startswith('-') and not s.startswith('---') for s in diff)))
        with (logs/(directory.name+'.patch')).open('a') as p:p.write(''.join(diff))
    text=(logs/(directory.name+'.log')).read_text()
    return dict(adapter=directory.name,exit=r.returncode,files_written=files,edits=edits,refusal=text.strip().replace(str(tree),'<tree>').replace(str(repo),'<repo>').replace('/workspace/32-main','<repo>') if r.returncode else None)
records={'before':[],'after':[]};differences=[]
directories=sorted(p for p in (repo/'stage3/adapt').iterdir() if p.is_dir())
baseline_repo=Path('/workspace/32-main')
for directory in directories:
    with ThreadPoolExecutor(max_workers=2) as pool:
        futures=[pool.submit(run,label,tree,(baseline_repo/'stage3/adapt'/directory.name) if label=='before' else directory) for label,tree in [('before',baseline),('after',accepted)]]
        b,a=[f.result() for f in futures]
    if b!=a:differences.append(dict(adapter=directory.name,before=b,after=a))
    records['before'].append(b);records['after'].append(a)
    for label,rows in records.items():(output/(label+'.json')).write_text(json.dumps(rows,indent=2)+'\n')
    rendered={label:json.dumps(rows,indent=2).splitlines(True) for label,rows in records.items()}
    (output/'records.diff').write_text(''.join(difflib.unified_diff(rendered['before'],rendered['after'],fromfile='before.json',tofile='after.json')))
    (output/'status.json').write_text(json.dumps(dict(complete=False,adapters_recorded=len(records['after']),differences=differences),indent=2)+'\n')
    print(json.dumps(dict(adapter=directory.name,before_exit=b.get('exit'),after_exit=a['exit'],differences=len(differences))),flush=True)
coverage={}
for label,tree in [('before',baseline),('after',accepted)]:
    final=manifest(tree);(output/(label+'-final-manifest.json')).write_text(json.dumps(final,sort_keys=True)+'\n')
    changed={f for f in initial[label].keys()|final.keys() if initial[label].get(f)!=final.get(f)}
    observed={f for r in records[label] for f in r.get('files_written',[])}
    untracked=sorted(changed-observed);assert not untracked,untracked
    coverage[label]=dict(final_changes=sorted(changed),journaled_files=sorted(observed),untracked=[])
(output/'coverage.json').write_text(json.dumps(coverage,indent=2)+'\n')
(output/'status.json').write_text(json.dumps(dict(complete=True,adapters_recorded=len(records['after']),baseline_adapters_compared=len(directories),differences=differences),indent=2)+'\n')
raise SystemExit(1 if any(r['adapter']!='32-indexed-reads-program' for r in differences) else 0)
