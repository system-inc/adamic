import pathlib,json,subprocess,shlex
root=pathlib.Path('/workspace/adamic');p=root/'review/test-defend/stage1-cohere-yaml-gaps'
rows={};subcases={};failures=[];commands=[]
for f in sorted(p.glob('D02-*.log')):
 if f.name=='D02-driver.log':continue
 for line in f.read_text().splitlines():
  try:x=json.loads(line)
  except:continue
  if x.get('Test') and x['Action'] in ['pass','fail','skip']:
   name=x['Test'];subcases.setdefault(name,[]).append(x['Action']);top=name.split('/')[0]
   if '/' not in name:rows.setdefault(top,[]).append(x['Action'])
  if x.get('Test','').startswith('TestSharedSliceAppendMatchesNode') and ('scalar_runtime_gap_test.go:' in x.get('Output','') or 'ERROR: AddressSanitizer' in x.get('Output','')):failures.append({'log':f.name,'test':x['Test'],'line':x['Output'].strip()})
(p/'D02-matrix.json').write_text(json.dumps({'top_level':rows,'subcases':subcases},indent=2))
expected={x for x in (p/'list.log').read_text().splitlines() if x.startswith('Test')};missing=sorted(expected-rows.keys());failed=sorted(k for k,v in rows.items() if 'fail' in v);passed=sorted(k for k,v in rows.items() if v and set(v)=={'pass'});skipped=sorted(k for k,v in rows.items() if 'skip' in v)
assert not missing,(missing,'missing matrix rows');assert failed==['TestSharedSliceAppendMatchesNode'],failed;assert not skipped,skipped
(p/'D02-passed-rows.json').write_text(json.dumps(passed,indent=2));(p/'D02-failing-lines.json').write_text(json.dumps(failures,indent=2))
core=next(x for x in json.loads((p/'D02-runs.json').read_text()) if x['group']=='core')
command='ADAMIC_YAML_LIBRARY=/tmp/u152/library ADAMIC_BUILD_CACHE_DIR=/tmp/defend-yaml/cache/D02 '+shlex.join(core['command'])+' > D02-core.log 2>&1'
fail=next(x['line'] for x in failures if 'scalar_runtime_gap_test.go:' in x['line'])
item=dict(test='TestSharedSliceAppendMatchesNode',package='stage1/cohere/yaml',prior_verdict='subsumed',subsumed_by=['TestPropsMatchGo'],defense='defended',unique_mutant='D02 internal/native/runtime/string_share.c:59',attempts=[dict(mutant='D01',file_line='internal/native/runtime/string_share.c:59',change='Give every shared slice capacity size + 1.',rows_failed=['TestSharedSliceAppendMatchesNode','TestFormatterMatchesGo']),dict(mutant='D02',file_line='internal/native/runtime/string_share.c:59',change='Give a shared slice capacity size + 1 only when it ends at its ultimate owner byte boundary.',rows_failed=failed)],evidence=command+'; '+fail+'; AddressSanitizer: heap-buffer-overflow')
(p/'rows.json').write_text(json.dumps([item],indent=2)+'\n')
for ident in ['D01','D02']:
 r=subprocess.run(['git','apply','--check',str(p/(ident+'.diff'))],cwd=root,capture_output=True,text=True);(p/(ident+'-apply-check.log')).write_text(r.stdout+r.stderr);assert r.returncode==0
library=pathlib.Path('/tmp/u152/library');(p/'library-package.json').write_bytes((library/'package.json').read_bytes());(p/'library-package-lock.json').write_bytes((library/'package-lock.json').read_bytes())
base=(p/'starting-commit.txt').read_text().strip()
report=f'''TestSharedSliceAppendMatchesNode is defended by D02, a production C runtime mutant.
All 72 current top-level tests were replayed in bounded groups; only the target failed and 71 others passed.
Starting main: {base}; standalone diffs, clang checks, coverage profiles, matrices and logs are preserved here.

CODE UNDER TEST: Adamic lowering/native emission and the C runtime's adamic_string_share, adamic_string_append, and string slice operations. The mutated implementation is string_share.c; no test input, oracle, harness, Go cohere implementation, or port source was changed. ORACLE: live Node stdout pinned to a\\nx\\n for the target, plus successful native execution under ASan/UBSan/LSan. TestPropsMatchGo compares the port with live Go cohere, Node, emitted JavaScript and yaml@2.9.0. Source Node stays unchanged.

The target and subsumer test files were read whole after reading the prior audit REPORT.md, report.json and M3.diff. Discovery verifies both names exist and records 72 names, unchanged from the audit's complete inventory. The audit itself ran only its assigned slice; this defense replays every discovered current name. Families do not change D02 uniqueness because exactly one top-level function fails.

Coverage commands: ADAMIC_YAML_LIBRARY=/tmp/u152/library timeout 120 go test -count=1 -timeout 90s ./stage1/cohere/yaml/ -run '^NAME$' -coverpkg=./internal/lower,./internal/native -coverprofile=review/test-defend/stage1-cohere-yaml-gaps/NAME.cover > NAME.coverage.log 2>&1, for the target and TestPropsMatchGo. Both pass. There are two target-only covered blocks: emit_strings.go:63 (repeat emission) and native.go:106 (unsanitized O2 flag). All other covered compiler/native blocks are shared. Go profiles cannot instrument the embedded C runtime, so no exclusive C-line coverage claim is made.

Semantic lead: sharedSliceAppend.ts builds a 128-byte heap owner, shares an 80-byte view, and appends one byte. At offset 0 it checks an observable owner byte remains 'a'. At offset 48 the view ends exactly at the owner's byte allocation boundary. A slice header's capacity must be zero even when its own reference count is one. The property resolver builds different string histories and does not exercise this owner-end append boundary. The production mutation outcomes below establish that difference empirically; shared compiler lines alone do not prove it.

D01 replaces zero shared capacity with size + 1. Target offset 0 prints x\\nx\\n instead of Node's a\\nx\\n; offset 48 triggers ASan. TestPropsMatchGo passes. TestFormatterMatchesGo also fails, so the broad mutant is not unique. Its remaining replay was stopped after that decisive counterexample; D01 has unknown results outside its completed groups. D01-stop.json records task-owned process cleanup and source restoration. D01's partial matrix never supports uniqueness.

D02 makes that capacity mistake only for a slice whose end pointer equals its ultimate owner's end pointer. This is a change to the capacity option, aimed at the semantic owner-end boundary, without matching filenames, test names or literal input bytes. Offset 0 passes; offset 48 reaches the illegal append and ASan reports heap-buffer-overflow in adamic_string_append. Only TestSharedSliceAppendMatchesNode fails. TestPropsMatchGo, TestFormatterMatchesGo, all scalar/file execution shards and every other current row pass. D02-passed-rows.json gives the complete list; D02-matrix.json records top-level and subcase outcomes; D02-runs.json gives every exact command and duration.

Each mutant uses ADAMIC_BUILD_CACHE_DIR=/tmp/defend-yaml/cache/ID. Runtime library caching additionally hashes runtime contents, flags and compiler identity, so each changed C implementation rebuilds. D01.clang.log and D02.clang.log preserve successful compilation of each modified translation unit using native.Flags' ordinary C11 warning/error, floating-point and optimization flags. The matrix also builds and executes native products with ordinary and sanitizer options. After restoring production, both standalone diffs pass git apply --check against the starting main worktree. No production changes remain.

Brief costs and ambiguities:
- The mandatory whole-package baseline exhausted the 90-second binary budget without an assertion failure. Grouping all rows together is unsuitable for this package. Additional aggregate groups also timed out, so the final matrix selects rows or complete subcase sets separately. groups.json lists all selections; all 72 names have observed final results. The successful uniqueness claim is package-wide over those observations, not a slice extrapolation.
- The audit's warm library contained only yaml@2.9.0 and prettier@3.9.6. Enabling newer oracles exposed missing yaml-unist-parser, with ERR_MODULE_NOT_FOUND. No mutation was planted while this baseline was red. Installing yaml-unist-parser@3.2.1 resolved it; the repaired unist control and every mutant-witness subcase pass cleanly. Exact dependency manifests are saved.
- TestMain preflight and build-only product tests complicate the outer timeout. Listing suppresses preflight. File-driver setup, shard execution, formatter witness products and scalar shard products are all included rather than assuming a union test executes every shard.
- Go -coverprofile with -coverpkg measures compiler and emitter blocks, not C runtime execution. The defense therefore uses an explicit input-history/boundary argument and actual native mutation evidence. LLVM C coverage tools were unavailable; no C exclusive-line claim was fabricated.
- The broad D01 replay was stopped as soon as a second catcher settled that attempt. This saves expensive native rebuilds but means its failed-row list is observed, not exhaustive. The successful D02 replay is complete.
- An attempted read of the clang executable produced an exec-server output-recovery error. No source or evidence was changed by that read; normal clang compilation subsequently succeeded.
- This unit required substantially more baseline work than the audit's original assigned slice. Native C changes invalidate cached integration products, and witness/shard families need separate 90-second runs. Commands and wall durations are saved rather than conflating setup, build and test time.

Owner finding: the target's name is supported. It compares native and Node output for both offsets and also requires sanitized native execution to succeed. D02 is detected by the memory-safety leg rather than a changed unsanitized stdout; that is the demonstrated limit and strength of this defense. The test is not being defended merely by an admission check or a snapshot.

Toolchain setup skipped, env.sh works, nproc=5. API npm ci ran before baseline. No skipped final matrix rows, no test/oracle/harness changes, no other packages tested, no main push and no pull request. A unique defense was found on the second attempt, so a third mutant was unnecessary.

```json
{json.dumps([item],indent=2)}
```
'''
(p/'REPORT.md').write_text(report);print(json.dumps([item],indent=2));print('passed',len(passed))
