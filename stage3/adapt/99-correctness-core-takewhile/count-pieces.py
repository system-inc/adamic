import hashlib,json,os,shutil,signal,subprocess,tarfile,time
from collections import Counter
from concurrent.futures import ThreadPoolExecutor,as_completed
from pathlib import Path
repo=Path('/workspace/adamic');root=Path('/workspace/scratch/takewhile');out=repo/'stage3/adapt/99-correctness-core-takewhile';results=root/'count-results';results.mkdir(exist_ok=True)
plan=json.loads((repo/'stage3/lane/shard-plan.json').read_text());units=plan['units'];classes=['undefined','empty','all accepted','none accepted','middle rejection']
summary={'scope':'All unchanged lane manifest units, one canonical run per unit; seeds excluded','unitInventory':len(units),'completedUnits':[],'pieces':[],'calls':0,'classes':dict.fromkeys(classes,0),'acceptedRangeHoleCalls':0,'resultDifferences':0,'predicateSequenceDifferences':0,'unstableInputs':0,'oracleCounts':{'pass':0,'fail':0,'skip':0},'complete':False}
if (out/'COUNT.json').exists(): summary=json.loads((out/'COUNT.json').read_text())
frozen=json.loads((root/'raw/counted-manifest.json').read_text())
def snapshot():
    return {str(p.relative_to(root/'counted')):hashlib.sha256(p.read_bytes()).hexdigest() for base in ['src','built/local'] for p in sorted((root/'counted'/base).rglob('*')) if p.is_file()}
def run(unit,directory):
    name=unit['name'];env=dict(os.environ,STAGE3_ADAPTED_TREE=str(root/'counted'),STAGE3_SHARD_SECONDS='90',NODE_OPTIONS='--require='+str(out/'count-preload.cjs'),NODE_DISABLE_COMPILE_CACHE='1',TW_CALL_DIR=str(directory/'calls'),TW_SHARD=name)
    with (directory/(name+'.log')).open('w') as log:
        child=subprocess.Popen(['python3',str(repo/'stage3/lane/shard.py'),name,str(results)],cwd=repo,env=env,stdout=log,stderr=subprocess.STDOUT,start_new_session=True)
        try: code=child.wait(timeout=100)
        finally:
            try: os.killpg(child.pid,signal.SIGKILL)
            except ProcessLookupError: pass
            child.wait()
    report=json.loads((results/name/'report.json').read_text())
    return unit,code,report
for start in range(0,len(units),20):
    if start<len(summary['completedUnits']): continue
    begin=time.monotonic();piece=start//20+1;directory=root/f'count-{piece:02d}';directory.mkdir();(directory/'calls').mkdir();subset=units[start:start+20]
    with ThreadPoolExecutor(max_workers=4) as pool:
        observations=[f.result() for f in as_completed([pool.submit(run,u,directory) for u in subset])]
    failed=[]
    for unit,code,report in observations:
        if code!=0 or report['status']!='pass': failed.append({'unit':unit['name'],'exit':code,'report':report})
        shutil.copytree(results/unit['name'],directory/unit['name'],ignore=shutil.ignore_patterns('adapted-tree'))
    record={'piece':piece,'units':[u['name'] for u in subset],'calls':0,'classes':dict.fromkeys(classes,0),'acceptedRangeHoleCalls':0,'resultDifferences':0,'predicateSequenceDifferences':0,'unstableInputs':0,'failed':failed,'oracleCounts':dict.fromkeys(['pass','fail','skip'],0)}
    for file in sorted((directory/'calls').glob('*.jsonl')):
        for line in file.open():
            row=json.loads(line);record['calls']+=1
            category='undefined' if row['input'].get('undefined') else 'empty' if row['input']['length']==0 else 'all accepted' if all(row['answers']) else 'none accepted' if not row['answers'][0] else 'middle rejection'
            record['classes'][category]+=1
            record['acceptedRangeHoleCalls']+=row['acceptedRangeHasHole'];record['resultDifferences']+=not row['sameResult'];record['predicateSequenceDifferences']+=not row['sameSequence'];record['unstableInputs']+=not row['stableInput']
    for _,_,report in observations:
        for key,n in report['counts'].items(): record['oracleCounts'][key]+=n
    assert frozen==snapshot(),'counted source or build changed'
    record['seconds']=time.monotonic()-begin;assert record['seconds']<1400
    (directory/'piece.json').write_text(json.dumps(record,indent=2)+'\n')
    (directory/'SHA256SUMS').write_text(''.join(hashlib.sha256(p.read_bytes()).hexdigest()+'  '+str(p.relative_to(directory))+'\n' for p in sorted(directory.rglob('*')) if p.is_file()))
    archive=out/f'count-{piece:02d}.tar.gz'
    with tarfile.open(archive,'w:gz') as tar: tar.add(directory,arcname=directory.name)
    (out/f'count-{piece:02d}.sha256').write_text(hashlib.sha256(archive.read_bytes()).hexdigest()+'  '+archive.name+'\n')
    summary['pieces'].append(record)
    for key in ['calls','acceptedRangeHoleCalls','resultDifferences','predicateSequenceDifferences','unstableInputs']: summary[key]+=record[key]
    for key in classes: summary['classes'][key]+=record['classes'][key]
    if not failed:
        summary['completedUnits']+=record['units']
        for key,n in record['oracleCounts'].items(): summary['oracleCounts'][key]+=n
    summary['complete']=len(summary['completedUnits'])==len(units)
    (out/'COUNT.json').write_text(json.dumps(summary,indent=2)+'\n')
    print(json.dumps({key:record[key] for key in ['piece','calls','seconds','acceptedRangeHoleCalls','resultDifferences','predicateSequenceDifferences','unstableInputs']}),flush=True)
    for unit,_,_ in observations: shutil.rmtree(results/unit['name']/'adapted-tree')
    if failed or any(record[k] for k in ['acceptedRangeHoleCalls','resultDifferences','predicateSequenceDifferences','unstableInputs']): raise SystemExit('STOP: failed lane or measurement gate; raw piece retained')
assert summary['oracleCounts']==plan['counts']
print('COMPLETE',json.dumps({key:summary[key] for key in ['calls','classes','oracleCounts']}),flush=True)
