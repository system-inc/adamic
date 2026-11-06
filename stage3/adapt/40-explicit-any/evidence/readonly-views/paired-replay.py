"""Replay every adapter on disposable equal-dependency tree copies, continuing after refusals."""
import difflib,gzip,hashlib,json,os,shutil,subprocess,sys
from pathlib import Path
repo=Path(__file__).resolve().parents[5]
before,after,work=map(lambda x:Path(x).resolve(),sys.argv[1:])
work.mkdir(parents=True,exist_ok=True)
trace=Path(__file__).with_name('replay-trace.cjs')
adapters=sorted((repo/'stage3/adapt').glob('*/adapt.cjs'))
records={}
for label,source in [('before',before),('after',after)]:
    tree=work/('tree-'+label)
    def ignore(directory,names):
        return [n for n in ('node_modules','built') if Path(directory)==source and n in names]
    shutil.copytree(source,tree,ignore=ignore)
    (tree/'node_modules').symlink_to(before/'node_modules',target_is_directory=True)
    logdir=work/label;logdir.mkdir()
    rows=[]
    def normalize(s):return s.replace(str(tree),'<tree>').replace(str(source),'<tree>')
    for adapter in adapters:
        writes=logdir/(adapter.parent.name+'.writes.jsonl')
        writes.write_text('')
        env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),NODE_OPTIONS='--require='+str(trace),RO_REPLAY_TRACE=str(writes))
        p=subprocess.run(['node',str(adapter),str(tree)],env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE,text=True)
        (logdir/(adapter.parent.name+'.stdout.log.gz')).write_bytes(gzip.compress(p.stdout.encode(),mtime=0))
        (logdir/(adapter.parent.name+'.stderr.log.gz')).write_bytes(gzip.compress(p.stderr.encode(),mtime=0))
        operations=[json.loads(line) for line in writes.read_text().splitlines()]
        files=sorted({normalize(row['file']) for row in operations})
        row={'adaptation':adapter.parent.name,'exit':p.returncode,'files_written':files,'write_calls':len(operations),'stdout':normalize(p.stdout),'stderr':normalize(p.stderr)}
        rows.append(row)
        (logdir/'outcomes.json').write_text(json.dumps(rows,indent=2)+'\n')
        print(json.dumps({'tree':label,'adaptation':adapter.parent.name,'exit':p.returncode,'files_written':files}),flush=True)
    records[label]=rows
left=json.dumps(records['before'],indent=2).splitlines(keepends=True)
right=json.dumps(records['after'],indent=2).splitlines(keepends=True)
(work/'outcomes.diff').write_text(''.join(difflib.unified_diff(left,right,fromfile='before',tofile='after')))
differences=[{'adaptation':a['adaptation'],'before':a,'after':b} for a,b in zip(records['before'],records['after']) if a!=b]
summary={'adapters_each':len(adapters),'differences':differences,'unexpected_differences':[d['adaptation'] for d in differences if d['adaptation']!='40-explicit-any']}
(work/'comparison.json').write_text(json.dumps(summary,indent=2)+'\n')
print(json.dumps({'adapters_each':len(adapters),'differences':[d['adaptation'] for d in differences],'unexpected_differences':summary['unexpected_differences']}),flush=True)
sys.exit(bool(summary['unexpected_differences']))
