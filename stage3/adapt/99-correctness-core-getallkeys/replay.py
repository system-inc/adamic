"""Retain paired whole-stack replay outcomes and prove source-coordinate changes."""
import difflib,hashlib,json,os,pathlib,re,shutil,sys
from supervise import run,SCRATCH
repo=pathlib.Path('/workspace/adamic'); main=pathlib.Path('/workspace/getallkeys-main'); output=SCRATCH/'replay';output.mkdir()
source={side:pathlib.Path('/workspace/getallkeys-'+side) for side in ['before','after']}
def snapshot(tree):return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for p in (tree/'src').rglob('*') if p.is_file()}
reports={}
for side,checkout in [('before',main),('after',repo)]:
    tree=output/side
    shutil.copytree(source[side],tree,symlinks=True,ignore=lambda d,n:[x for x in n if pathlib.Path(d)==source[side] and x in ['.git','tests','built','node_modules']])
    (tree/'node_modules').symlink_to(repo/'stage3/typescript/node_modules',target_is_directory=True)
    rows=[]
    for directory in sorted(p for p in (checkout/'stage3/adapt').iterdir() if p.is_dir()):
        name=directory.name;before=snapshot(tree);writes=output/(side+'-'+name+'-writes.jsonl')
        env=dict(os.environ,NODE_PATH=str(repo/'stage3/api/node_modules'),NODE_OPTIONS='--require='+str(repo/'stage3/adapt/95-checker-axolotl-h/replay-hook.cjs'),REPLAY_TREE=str(tree),REPLAY_LOG=str(writes))
        status=run('replay-'+side+'-'+name,['node',str(directory/'adapt.cjs'),str(tree)],cwd=str(checkout),env=env,limit=1800)
        after=snapshot(tree)
        text=(SCRATCH/('replay-'+side+'-'+name+'.log')).read_text().replace(str(tree),'TREE').replace(str(checkout),'REPO')
        text=text.replace(str(repo),'REPO').replace(str(main),'REPO')
        row={'adapter':name,'exit':status['exit_code'],'output':text,'changedFiles':sorted(k for k in before.keys()|after.keys() if before.get(k)!=after.get(k)),'writes':sorted(x for line in writes.read_text().splitlines() for x in json.loads(line)) if writes.exists() else []}
        rows.append(row);(output/(side+'.json')).write_text(json.dumps(rows,indent=2)+'\n')
        if status['stopping_reason']:raise SystemExit(status['stopping_reason'])
    reports[side]={r['adapter']:r for r in rows}
maps={};remaps=[]
def remap(file,line,column=None):
    if not file.startswith('src/'):file='src/compiler/'+file
    if file not in maps:
        a=(source['before']/file).read_text().splitlines();b=(source['after']/file).read_text().splitlines();mapping={}
        for tag,i,j,k,l in difflib.SequenceMatcher(None,a,b,autojunk=False).get_opcodes():
            if tag=='equal':mapping.update({v+1:u+1 for u,v in zip(range(i,j),range(k,l))})
        maps[file]=(a,b,mapping)
    a,b,mapping=maps[file];assert line in mapping,(file,line,'changed source line');old=mapping[line];assert a[old-1]==b[line-1]
    if old!=line:remaps.append({'file':file,'beforeLine':old,'afterLine':line,'column':column,'identicalCompleteLine':True})
    return old

def normalize(value,file=None):
    if isinstance(value,dict):
        file=value.get('file',file);result={k:normalize(v,file) for k,v in value.items()}
        if isinstance(value.get('line'),int):assert file;result['line']=remap(file,value['line'],value.get('column'))
        return result
    if isinstance(value,list):return [normalize(v,file) for v in value]
    if isinstance(value,str):return re.sub(r'(src/[^:\s]+\.ts):(\d+):(\d+)',lambda m:m[1]+':'+str(remap(m[1],int(m[2]),int(m[3])))+':'+m[3],value)
    return value
left,right=reports['before'],reports['after'];new='99-correctness-core-getallkeys';assert right.keys()-left.keys()=={new};differences=[]
for name,before in left.items():
    after=right[name];assert {k:v for k,v in before.items() if k!='output'}=={k:v for k,v in after.items() if k!='output'},name+' outcome changed'
    if before['output']==after['output']:continue
    try:expected=json.loads(before['output']);actual=json.loads(after['output'])
    except json.JSONDecodeError:raise AssertionError(name+' non-JSON replay difference')
    assert expected==normalize(actual),name+' unaccounted replay difference';differences.append(name)
assert right[new]['exit']==0 and not right[new]['writes'] and not right[new]['changedFiles']
summary={'passed':True,'baselineAdapters':len(left),'adaptedAdapters':len(right),'allOtherOutcomesIdentical':True,'coordinateOnlyDifferences':differences,'remaps':remaps,'newAdapterNoWrites':True}
(output/'summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps({k:v for k,v in summary.items() if k!='remaps'}),flush=True)
