import concurrent.futures,json,re,subprocess,time
from pathlib import Path
root=Path.cwd();e=root/'review/compiler/loom-r-unions';known={r['package'] for r in json.loads((e/'consumers.json').read_text())};packages=[p for p in json.loads(Path('cloud/fast-gate/compiler-dependencies.json').read_text())['packages'] if p not in known]
selected={'internal/boundedrun':'TestCommandDefaultDeadlineIncludesStartup','cmd/adamic-admission-delta':'TestCompilerCacheIncludesSHA','internal/compilercensus':'TestDependenciesCoverTheDeclaredMap','cmd/adamic-gocacheprog':'TestTrustSplit','internal/buildcache/cmd/traced':'TestSettleParsesOnlyAskedCalls','internal/tracked':'TestChangedNamesEveryDifference','internal/toolchain':'TestAToolsEnvironmentIsWhatTheKeyReadsAndNothingElse','stage1/cohere/executable':'TestWriteLeavesNoWritableDescriptorForSiblingForksToExec','stage3/lane':'TestApplyPython'}
(e/'declared-readers.json').write_text(json.dumps(dict(packages=packages,selected=selected),indent=2)+'\n')
def run(p):
 test=selected.get(p);cmd=['go','test','./'+p,'-count=1','-timeout','60s','-json','-run','^'+test+'$' if test else '^$'];start=time.monotonic()
 with (e/(p.replace('/','-')+'-declared.log')).open('w') as log:
  try:code=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=180).returncode
  except subprocess.TimeoutExpired:code=124
 row=dict(package=p,test=test,command=cmd,seconds=round(time.monotonic()-start,3),exit=code);print(row,flush=True);return row
with concurrent.futures.ThreadPoolExecutor(max_workers=2) as pool:rows=list(pool.map(run,packages))
cmd=['go','vet',*['./'+p for p in packages]];start=time.monotonic()
with (e/'declared-readers-vet.log').open('w') as log:code=subprocess.run(cmd,stdout=log,stderr=subprocess.STDOUT,timeout=180).returncode
rows.append(dict(check='vet',command=cmd,seconds=round(time.monotonic()-start,3),exit=code));(e/'declared-reader-results.json').write_text(json.dumps(rows,indent=2)+'\n')
