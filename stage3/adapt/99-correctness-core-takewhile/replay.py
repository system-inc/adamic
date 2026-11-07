import hashlib,json,os,shutil,signal,subprocess,time
from concurrent.futures import ThreadPoolExecutor
from pathlib import Path
repo=Path('/workspace/adamic');root=Path('/workspace/scratch/takewhile');raw=root/'raw';own='99-correctness-core-takewhile';env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'))
for side in ['before','after']:
    tree=root/('replay-'+side)
    shutil.copytree(root/side,tree,symlinks=True,ignore=shutil.ignore_patterns('node_modules','built','.git'))
    (tree/'node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
rows=[]
for directory in sorted(p for p in (repo/'stage3/adapt').iterdir() if p.is_dir()):
    def run(side):
        tree=root/('replay-'+side);begin=time.monotonic()
        with (raw/('replay-'+side+'.log')).open('a') as log:
            log.write(directory.name+'\n');log.flush()
            child=subprocess.Popen(['node',str(directory/'adapt.cjs'),str(tree)],env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True,start_new_session=True)
            try: stdout,stderr=child.communicate(timeout=900)
            finally:
                try: os.killpg(child.pid,signal.SIGKILL)
                except ProcessLookupError: pass
                child.wait()
            log.write(stdout+stderr)
        return side,{'exit':child.returncode,'stdout':stdout.replace(str(tree),'TREE'),'stderr':stderr.replace(str(tree),'TREE')}
    with ThreadPoolExecutor(max_workers=2) as pool: row={'adapter':directory.name,**dict(pool.map(run,['before','after']))}
    rows.append(row);(raw/'replay.json').write_text(json.dumps(rows,indent=2)+'\n')
    print(directory.name,row['before']['exit'],row['after']['exit'],flush=True)
    if directory.name!=own: assert row['before']==row['after'],row
    else:
        assert row['before']['exit']==row['after']['exit']==0
        assert json.loads(row['before']['stdout'])=={'files':1,'edits':1,'javascriptChanging':True}
        assert json.loads(row['after']['stdout'])=={'files':0,'edits':0,'javascriptChanging':True}
manifest=lambda tree:{str(p.relative_to(tree/'src')):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((tree/'src').rglob('*')) if p.is_file()}
assert manifest(root/'replay-before')==manifest(root/'replay-after')
summary={'adapters':len(rows),'allOtherOutcomesIdentical':True,'onlyOwnApplyVsNoop':True,'finalReplayedSourcesIdentical':True,'nonzeroPairedAdapters':[row['adapter'] for row in rows if row['before']['exit']!=0]}
(raw/'replay-summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary),flush=True)
