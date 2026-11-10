import concurrent.futures,gzip,hashlib,json,os,pathlib,re,subprocess,sys,threading
root=pathlib.Path(sys.argv[1]).resolve();unit=pathlib.Path(__file__).resolve().parent
work=pathlib.Path('/workspace/scratch/debug-display-2-witness');work.mkdir(exist_ok=True)
prior=json.loads(gzip.decompress((unit/'debug-display-witness.json.gz').read_bytes()))
cases=[]
for p in sorted((root/'tests/cases/compiler').glob('*.ts')):
    if re.search(r'@(?:filename|link|symlink|libFiles|includeBuiltFile)\s*:',p.read_text(errors='replace'),re.I):continue
    cases.append(p)
    if len(cases)==200:break
assert len(cases)==200
assert [hashlib.sha256(p.read_bytes()).hexdigest() for p in cases]==[r['case_sha256'] for r in prior['rows']], 'Change 1 case inventory differs'
base=dict(os.environ)
for key in ['NODE_OPTIONS','VSCODE_INSPECTOR_OPTIONS','NODE_INSPECTOR_IPC']:base.pop(key,None)
stop=threading.Event()
def run(item):
    if stop.is_set():return None
    index,p=item;command=['node',str(root/'built/local/tsc.js'),'--ignoreConfig','--pretty','false','--noEmit','--skipLibCheck',str(p)]
    pair=[];raw=[]
    for debug in [False,True]:
        env=dict(base)
        if debug:env['VSCODE_INSPECTOR_OPTIONS']='{}'
        child=subprocess.run(command,cwd=root,env=env,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
        row={'child_exit':child.returncode}
        for stream in ['stdout','stderr']:
            data=getattr(child,stream);(work/f'{index:03d}-{int(debug)}.{stream}').write_bytes(data)
            row[stream+'_sha256']=hashlib.sha256(data).hexdigest();row[stream+'_bytes']=len(data)
        pair.append(row);raw.append((child.returncode,child.stdout,child.stderr))
    identical=raw[0]==raw[1]
    result={'index':index,'case':str(p.relative_to(root)),'case_sha256':hashlib.sha256(p.read_bytes()).hexdigest(),'plain':pair[0],'debug':pair[1],'identical':identical}
    (work/f'{index:03d}.json').write_text(json.dumps(result,indent=2)+'\n')
    print(json.dumps({'index':index,'plain_child_exit':pair[0]['child_exit'],'debug_child_exit':pair[1]['child_exit'],'identical':identical}),flush=True)
    if not identical:stop.set();raise RuntimeError('witness differs at '+str(index))
    return result
with concurrent.futures.ThreadPoolExecutor(max_workers=3) as pool:rows=list(pool.map(run,enumerate(cases)))
assert len(rows)==200 and all(r and r['identical'] for r in rows)
report={'status':'pass','cases':200,'all_byte_identical':True,'selection':'Same first 200 sorted single-file compiler inputs and fixed CLI flags as Change 1.','rows':rows}
(unit/'debug-display-2-witness.json').write_text(json.dumps(report,indent=2)+'\n')
