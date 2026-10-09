from pathlib import Path
import argparse
import json
import re
import subprocess
parser = argparse.ArgumentParser(description="Record independent Node and three-branch native observations.")
parser.add_argument('--main', type=Path, required=True)
parser.add_argument('--namespaces-tsc', type=Path, required=True)
parser.add_argument('--parameter-properties-namespaces', type=Path, required=True)
parser.add_argument('--logs', type=Path, required=True)
args = parser.parse_args()
root = args.main.resolve()
bucket = Path(__file__).resolve().parent
scratch = args.logs.resolve()
scratch.mkdir(parents=True, exist_ok=True)
manifest = []
for p in sorted(bucket.glob('[0-9][0-9]_*.a')):
    tsc = [line.split('TypeScript 6.0.3, ')[1] for line in p.read_text().splitlines() if line.startswith('// From TypeScript 6.0.3, ')]
    manifest.append({'file': p.name, 'tsc': tsc, 'reason': 'a namespace'})
results = {}
status = []
def run(cmd,cwd):
    p=subprocess.run(cmd,cwd=cwd,stdout=subprocess.PIPE,stderr=subprocess.PIPE)
    return {'stdout':p.stdout.decode(),'stderr':p.stderr.decode(),'exit':p.returncode}
def outcome(r):
    if r['exit']==0:return 'Compiles'
    if "stage 0 can't lower" in r['stderr'].lower():return 'NotYet'
    if 'Adamic 0.1 refuses' in r['stderr']:return 'Refused'
    if re.search(r'error TS[0-9]+:', r['stderr']):return 'Checker'
    raise RuntimeError(r)
for row in manifest:
    file=row['file']; p=bucket/file
    node=run(['node','--disable-warning=ExperimentalWarning',str(args.namespaces_tsc.resolve()/'oracle/node.mjs'),str(p)],root)
    (scratch/(file+'.node.log')).write_text(json.dumps(node,indent=2))
    if node['exit']!=0: raise RuntimeError(('NODE FAILED',file,node))
    results[file]={'node':node}
    for name,repo in [('main',root),('namespaces-tsc',args.namespaces_tsc.resolve()),('parameter-properties-namespaces',args.parameter_properties_namespaces.resolve())]:
        binary=scratch/(file+'.'+name)
        build=run(['go','run','./cmd/adamic','build',str(p),'-o',str(binary)],repo)
        (scratch/(file+'.'+name+'.build.log')).write_text(json.dumps(build,indent=2))
        kind=outcome(build); record={'outcome':kind,'what':build['stderr']+build['stdout']}
        if kind=='Compiles':
            record['run']=run([str(binary)],repo)
            record['matchesNode']=record['run']==node
            if not record['matchesNode']:print('SILENT MISCOMPILE',file,name,json.dumps(record),flush=True)
        results[file][name]=record
        print(file,name,kind,flush=True)
    status.append({k:row[k] for k in ['file','tsc','reason']}|{'node':node,'stage0':{k:results[file]['main'][k] for k in ['outcome','what']}})
(bucket/'status.json').write_text(json.dumps(status,indent=2)+'\n')
(bucket/'validation.json').write_text(json.dumps(results,indent=2)+'\n')
assert all(record.get('matchesNode', True) for row in results.values() for record in row.values()), 'native/Node mismatch'
