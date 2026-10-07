#!/bin/bash
set -eu
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
python3 - <<'PY'
from pathlib import Path
import subprocess,json
p=Path('internal/native/runtime/heap.c');original=p.read_text();results=[]
try:
 for name,old,new in [('split-skip-last','release_last(value);','(void)value;'),('split-double-decrement','--heap->references != 0','(heap->references -= 2) != 0')]:
  assert original.count(old)==1,(name,original.count(old))
  p.write_text(original.replace(old,new,1))
  path='/tmp/lint-'+name+'-mutant.log'
  with open(path,'wb') as log:r=subprocess.run(['go','test','./internal/native','-run','^TestRuntimeReleasePaths$','-count=1'],stdout=log,stderr=log)
  results.append({'mutant':name,'exit':r.returncode,'log':path});assert r.returncode!=0
finally:p.write_text(original)
Path('/tmp/lint-split-mutants.json').write_text(json.dumps(results,indent=2)+'\n')
PY
go test ./internal/native -count=1 -timeout 30m > /tmp/lint-native-package.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestRuntimeLastIndexOfMatchesNode|TestWeakReadsUndefinedOnceFreed|TestNativeAgreesWithNode/internal/oracle/testdata/(runtime_last_index_of|search_halves|search_from_sweep|shared_slices|shared_slice_append|strings|lone_surrogates|long_chain|weak_parent|doubly_linked|exceptions|map_iteration|sort_releases|class_oct6_release)\.a$' -count=1 -timeout 30m > /tmp/lint-filtered-oracle.log 2>&1
go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 30m -args -update-counts > /tmp/lint-counts-update.log 2>&1
go vet ./internal/native ./internal/oracle > /tmp/lint-vet.log 2>&1
gofmt -l internal/native/runtime_profile_test.go internal/oracle/runtime_last_index_test.go internal/oracle/oracle_test.go > /tmp/lint-gofmt.log
git diff --check > /tmp/lint-diff-check.log
