import hashlib,json,os,shutil,signal,subprocess,time,sys
from concurrent.futures import ThreadPoolExecutor,as_completed
from pathlib import Path
repo=Path('/workspace/adamic');root=Path('/workspace/scratch/takewhile');out=repo/'stage3/adapt/99-correctness-core-takewhile';results=root/'after-lane';results.mkdir()
plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());units=plan['units'];records=[];diff=b'';counts={'pass':0,'fail':0,'skip':0};pieces=[]
sys.path.insert(0,str(repo/'stage3/lane'))
from shard_plan import names_digest
manifest=lambda:{str(p.relative_to(root/'after')):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((root/'after'/base).rglob('*')) if p.is_file()}
frozen=manifest();(root/'raw/after-lane-source-manifest.json').write_text(json.dumps(frozen,indent=2)+'\n')
def run(unit):
    env=dict(os.environ,STAGE3_ADAPTED_TREE=str(root/'after'),STAGE3_SHARD_SECONDS='90',NODE_DISABLE_COMPILE_CACHE='1',NODE_OPTIONS='')
    env.pop('TW_CALL_DIR',None)
    with (results/(unit['name']+'.log')).open('w') as log:
        child=subprocess.Popen(['python3',str(repo/'stage3/lane/shard.py'),unit['name'],str(results)],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
        try: code=child.wait(timeout=100)
        finally:
            try: os.killpg(child.pid,signal.SIGKILL)
            except ProcessLookupError: pass
            child.wait()
    report=json.loads((results/unit['name']/'report.json').read_text());return unit,code,report
for start in range(0,len(units),20):
    begin=time.monotonic();subset=units[start:start+20]
    with ThreadPoolExecutor(max_workers=4) as pool: observations=[f.result() for f in as_completed([pool.submit(run,u) for u in subset])]
    for unit,code,report in observations:
        assert code==0 and report['status']=='pass',(unit['name'],report)
        observed=json.loads((results/unit['name']/'observations.json').read_text());assert names_digest(observed)==unit['names_digest'];records+=observed
        diff+=(results/unit['name']/'oracle/baseline.diff').read_bytes()
        for key,n in report['counts'].items():counts[key]+=n
        shutil.rmtree(results/unit['name']/'adapted-tree')
    assert frozen==manifest()
    piece={'piece':start//20+1,'units':[u['name'] for u in subset],'seconds':time.monotonic()-begin};assert piece['seconds']<1400;pieces.append(piece)
    (out/'AFTER-LANE.json').write_text(json.dumps({'complete':False,'completedUnits':sum(len(p['units']) for p in pieces),'pieces':pieces,'counts':counts},indent=2)+'\n');print(json.dumps(piece),flush=True)
counted=json.loads((out/'COUNT-VERIFIED.json').read_text());assert counts==counted['oracleCounts']==plan['counts'];assert names_digest(records)==counted['namesDigest'];assert hashlib.sha256(diff).hexdigest()==counted['baselineDiffSha256']
summary={'complete':True,'completedUnits':len(units),'pieces':pieces,'counts':counts,'namesDigest':names_digest(records),'identityMultisetMatchesCount':True,'baselineDiffMatchesCount':True,'sourceAndBuildUnchanged':True}
(out/'AFTER-LANE.json').write_text(json.dumps(summary,indent=2)+'\n');print('COMPLETE',json.dumps(summary),flush=True)
