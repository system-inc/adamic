#!/usr/bin/env python3
"""Bounded witnesses for split dispatch, compiler budgets and archive scheduling."""
import json, os, subprocess, tempfile
from pathlib import Path
root=Path(__file__).resolve().parents[3]
mutants=[
 ('no-instrumented-default','native.go','return options.Sanitize || options.ThreadSanitize || options.Count','return false','TestSplitInstrumentedDefaults'),
 ('ignore-single-override','native.go','case "0":\n\t\treturn false','case "0":\n\t\treturn true','TestSplitOverrideSingle'),
 ('ignore-feature-preamble','native.go','allowed = allowed || line == "#define "+feature+" 1"','allowed = allowed || line == "#define "+feature+" 0"','TestSplitSourcePolicy'),
 ('split-handwritten-harness','native.go','if !found || !allowed {','if !found {','TestSplitSourcePolicy'),
 ('go-tests-use-all-cpus','units.go','if test {','if false && test {','TestSplitJobsPolicy'),
 ('no-default-dispatch','native.go','if useSplitSource(source, options) {','if false && useSplitSource(source, options) {','TestSplitFrameTypedefPreservesState'),
 ('no-shared-pool','jobserver.go','release := acquireClang()\n\tdefer release()','// no shared pool','TestClangPoolBoundsCompiles'),
 ('failed-compile-keeps-slot','jobserver.go','defer release()','defer func(){if false {release()}}()','TestClangFailureReturnsBudget'),
 ('ignore-runtime-workers','library.go','compileRuntime(files, flags, compiler, temporary, jobs)','compileRuntime(files, flags, compiler, temporary, 1)','TestParallelRuntimeKeepsArchiveOrderAndState'),
 ('reverse-archive-order','library.go','compiles.Wait()','compiles.Wait(); for left,right:=0,len(objects)-1;left<right;left,right=left+1,right-1 {objects[left],objects[right]=objects[right],objects[left]}','TestParallelRuntimeKeepsArchiveOrderAndState'),
 ('function-pointer-is-prototype','units.go','tokens[i+1].text != "*"','true','TestExternalPrototypeStructure'),
]
selected=os.environ.get('ADAMIC_POLICY_MUTANTS','').split(',')
with tempfile.TemporaryDirectory(prefix='split-policy-mutants-') as d:
 for name,file,before,after,test in mutants:
  if selected!=[''] and name not in selected:continue
  path=root/'internal/native'/file;source=path.read_text()
  if before not in source:raise RuntimeError('anchor '+name)
  changed=source.replace(before,after,1)
  if name=='split-handwritten-harness':changed=changed.replace('if !found {','_ = allowed\n if !found {',1)
  f=Path(d)/(name+'.go.txt');f.write_text(changed);overlay=Path(d)/'overlay.json';overlay.write_text(json.dumps({'Replace':{str(path):str(f)}}))
  log=Path('/tmp/policy-mutant-'+name+'.log')
  with log.open('w') as out:r=subprocess.run(['go','test','-overlay='+str(overlay),'./internal/native','-run','^'+test+'$','-count=1','-timeout=55s','-v'],cwd=root,stdout=out,stderr=subprocess.STDOUT,timeout=58)
  text=log.read_text()
  if r.returncode==0 or '--- FAIL: '+test not in text or '[build failed]' in text:raise RuntimeError(name+' not caught: '+text)
  print(name+' caught by '+test,flush=True)
