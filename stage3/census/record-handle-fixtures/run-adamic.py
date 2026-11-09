#!/usr/bin/env python3
"""Compare the exact receiver fixtures on two prebuilt, pinned Adamic compilers."""
from pathlib import Path
from concurrent.futures import ThreadPoolExecutor
import hashlib, json, os, resource, signal, subprocess

folder=Path(__file__).resolve().parent
root=folder.parents[2]
logs=folder/'adamic-logs'; logs.mkdir(exist_ok=True)
node=os.environ.get('FIXTURE_NODE','node')
tsc=os.environ.get('FIXTURE_TSC','/workspace/adamic/stage3/api/node_modules/typescript/bin/tsc')
revisions=[
 {'label':'main','sha':'86fc5a546bc85325756b472a4ba3d24306ec9a5c','checkout':'/workspace/ny0c3gb-main','binary':os.environ.get('ADAMIC_MAIN','/tmp/ny0c3gb-adamic-main')},
 {'label':'lower','sha':'554335247893a0488673a1b5076bb3edc18b2ca7','checkout':'/workspace/ny0c3gb-lower','binary':os.environ.get('ADAMIC_LOWER','/tmp/ny0c3gb-adamic-lower')},
]
# A crashing native test must not leave a core dump. Direct children retain their own exit codes.
resource.setrlimit(resource.RLIMIT_CORE,(0,0))
def run(args,tag,timeout=180,js=False):
 stem=logs/tag; stem.parent.mkdir(parents=True,exist_ok=True)
 stdout=stem.with_name(stem.name+('.mjs' if js else '.stdout.log'))
 stderr=stem.with_name(stem.name+'.stderr.log')
 with stdout.open('w') as out, stderr.open('w') as err:
  proc=subprocess.Popen(args,cwd=root,stdout=out,stderr=err,start_new_session=True)
  timed_out=False
  try: code=proc.wait(timeout=timeout)
  except subprocess.TimeoutExpired:
   timed_out=True
   os.killpg(proc.pid,signal.SIGKILL)
   code=proc.wait()
 result=dict(command=args,cwd=str(root),exit=code,timed_out=timed_out,
             stdout=stdout.read_text(errors='replace'),stderr=stderr.read_text(errors='replace'),
             stdout_file=str(stdout.relative_to(folder)),stderr_file=str(stderr.relative_to(folder)))
 print(json.dumps(dict(command=args,exit=code,timed_out=timed_out)),flush=True)
 return result

def outcome(result):
 if result['exit']==0: return 'compiled'
 text=result['stderr']
 if 'Adamic 0.1 refuses' in text: return 'Refused'
 if "stage 0 can't lower" in text: return 'NotYet'
 if 'compiler bug:' in text or result['exit']==4: return 'Compiler panic'
 if result['timed_out']: return 'Timed out'
 if ': error TS' in text or ': TS' in text or 'error TS' in text: return 'Checker rejection'
 return 'Other rejection'

versions={}
versions['node']=run([node,'--version'],Path('versions/node'))
versions['tsc']=run([node,tsc,'--version'],Path('versions/tsc'))
assert versions['node']['stdout'].startswith('v24.')
assert versions['tsc']['stdout'].strip()=='Version 6.0.3'
for revision in revisions:
 binary=Path(revision['binary'])
 revision['binary_sha256']=hashlib.sha256(binary.read_bytes()).hexdigest()
 versions[revision['label']]=run([str(binary),'--version'],Path('versions')/revision['label'])
 versions[revision['label']+'-help']=run([str(binary),'--help'],Path('versions')/(revision['label']+'-help'))

receivers=folder/'receivers'
files=sorted(receivers.glob('*--assign-function.ts'))
kinds=sorted({f.name.split('--')[0] for f in files})
assert len(files)==30 and len(kinds)==10
for kind in kinds:
 files.extend(receivers/(kind+'--path--'+road+'.ts') for road in ['direct','predicate'])
files.extend(sorted((folder/'class-copies').glob('*.ts')))
files=sorted(files)
assert len(files)==52
if os.environ.get("ADAMIC_LOGGING_ONLY")=="1":
 files=sorted((folder/"adamic-logging").glob("*.ts"))
 assert len(files)==31

def baseline(file):
 label=('logging-' if file.parent.name=='adamic-logging' else '')+file.stem
 result=dict(fixture=str(file.relative_to(folder)),sha256=hashlib.sha256(file.read_bytes()).hexdigest())
 if file.parent.name in ('class-copies','adamic-logging'):
  args=[node,tsc,'--ignoreConfig','--strict','--noUncheckedIndexedAccess','--exactOptionalPropertyTypes','--noEmit','--target','es2024','--lib','es2020,dom','--typeRoots',str(folder/'empty-types'),str(file)]
  result['tsc']=run(args,Path('tsc')/label)
 else:
  prior=json.loads((folder/'receiver-results.json').read_text())
  row=next(r for r in prior['rows'] if r['fixture']==file.name)
  assert result['sha256']==row['sha256']
  result['prior_tsc']=dict(exit=row['exit'],diagnostic=row['diagnostic'],command=row['command'])
 result['node']=run([node,'--experimental-strip-types',str(file)],Path('node')/label,timeout=15)
 return result
with ThreadPoolExecutor(max_workers=4) as pool:
 baselines=list(pool.map(baseline,files))

def check(pair):
 revision,file=pair
 label=Path(revision['label'])/(('logging-' if file.parent.name=='adamic-logging' else '')+file.stem)
 binary=Path('/tmp')/('ny0c3gb-native-'+revision['label']+'-'+file.stem)
 native_compile=run([revision['binary'],'build',str(file),'-o',str(binary)],label.with_name(label.name+'.build'))
 native_run=run([str(binary)],label.with_name(label.name+'.native'),timeout=15) if native_compile['exit']==0 else None
 js_compile=run([revision['binary'],'js',str(file)],label.with_name(label.name+'.javascript'),js=True)
 js_run=run([node,'--disable-warning=ExperimentalWarning',str(Path(revision['checkout'])/'oracle/node.mjs'),str(folder/js_compile['stdout_file'])],label.with_name(label.name+'.js-run'),timeout=15) if js_compile['exit']==0 else None
 return dict(revision=revision['label'],sha=revision['sha'],fixture=str(file.relative_to(folder)),
             native_outcome=outcome(native_compile),native_compile=native_compile,native_run=native_run,
             javascript_outcome=outcome(js_compile),javascript_compile=js_compile,javascript_run=js_run)
with ThreadPoolExecutor(max_workers=4) as pool:
 rows=list(pool.map(check,[(revision,file) for revision in revisions for file in files]))
Path(os.environ.get('ADAMIC_RESULTS',str(folder/'adamic-results.json'))).write_text(json.dumps(dict(revisions=revisions,versions=versions,baselines=baselines,rows=rows),indent=2)+'\n')
print('Completed '+str(len(files))+' fixtures at two compiler revisions.',flush=True)
