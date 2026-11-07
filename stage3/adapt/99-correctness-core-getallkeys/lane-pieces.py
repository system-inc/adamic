"""Run the whole unchanged lane inventory in bounded pieces and retain every unit."""
import collections,concurrent.futures,hashlib,json,os,pathlib,shutil,sys,time
from supervise import run,SCRATCH
repo=pathlib.Path('/workspace/adamic');tree=pathlib.Path('/workspace/getallkeys-after');output=SCRATCH/'lane';output.mkdir(exist_ok=True)
units=json.loads((repo/'stage3/lane/shards.json').read_text())['units']
# Shard views supply the locked packages themselves. The completed build used this link.
modules=tree/'node_modules'
assert modules.is_symlink(), 'expected the build dependency link'
modules.unlink()
def manifest():return {str(p.relative_to(tree)):hashlib.sha256(p.read_bytes()).hexdigest() for folder in ['src','built','tests/cases','tests/baselines/reference'] for p in sorted((tree/folder).rglob('*')) if p.is_file()}
frozen=manifest();(output/'input-manifest.json').write_text(json.dumps(frozen,indent=2)+'\n')
env=dict(os.environ,STAGE3_ADAPTED_TREE=str(tree),STAGE3_RESULTS=str(output/'results'),STAGE3_SHARD_SECONDS='90',NODE_DISABLE_COMPILE_CACHE='1')
rows=[]
def unit_run(unit):
    name=unit['name'];status=run('lane-'+name,['python3',str(repo/'stage3/lane/shard.py'),name,str(output/'results')],env=env,limit=120)
    report=output/'results'/name/'report.json';report=json.loads(report.read_text()) if report.exists() else {}
    view=output/'results'/name/'adapted-tree'
    if view.exists():shutil.rmtree(view)
    row={'unit':name,'exit_code':status['exit_code'],'stopping_reason':status['stopping_reason'],'report':report}
    (output/(name+'.json')).write_text(json.dumps(row,indent=2)+'\n')
    return row
for start in range(0,len(units),40):
    began=time.monotonic()
    with concurrent.futures.ThreadPoolExecutor(max_workers=4) as pool:piece=list(pool.map(unit_run,units[start:start+40]))
    rows+=piece
    result={'start':start,'end':start+len(piece),'seconds':time.monotonic()-began,'units':piece}
    (output/('piece-'+str(start)+'.json')).write_text(json.dumps(result,indent=2)+'\n')
    failed=[r['unit'] for r in piece if r['exit_code'] or r['stopping_reason'] or r['report'].get('status')!='pass']
    print(json.dumps({'piece':start,'completeUnits':len(rows),'failed':failed}),flush=True)
    if failed:raise SystemExit('lane piece failed: '+str(failed))
assert manifest()==frozen,'source, build, or oracle inputs changed'
assert [r['unit'] for r in rows]==[u['name'] for u in units]
counts=collections.Counter()
for row,unit in zip(rows,units):
    assert row['report']['counts']==unit['expect'],unit['name']
    counts.update(row['report']['counts'])
assert dict(counts)=={'pass':106366,'fail':1,'skip':0},dict(counts)
summary={'passed':True,'units':len(rows),'pieces':(len(rows)+39)//40,'counts':dict(counts),'allUnitIdentityAndBaselineGuardsPassed':True,'inputsUnchanged':True}
(output/'summary.json').write_text(json.dumps(summary,indent=2)+'\n');print(json.dumps(summary),flush=True)
