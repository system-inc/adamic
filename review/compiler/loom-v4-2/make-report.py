import ast,csv,json,re,subprocess
from pathlib import Path
e=Path('review/compiler/loom-v4-2')
def jsonl(p):return [json.loads(l) for l in p.read_text().splitlines()]
focused=[r for r in jsonl(e/'focused-results.jsonl') if 'test' in r];builds=[r for r in jsonl(e/'consumer-results.jsonl') if r.get('build')];consumer=[]
for l in (e/'consumer-tests.log').read_text().splitlines():
 try:r=ast.literal_eval(l)
 except (ValueError,SyntaxError):continue
 if isinstance(r,dict) and 'test' in r:consumer.append(r)
clean=json.loads((e/'clean-consumer-results.json').read_text());mutants=jsonl(e/'source-mutants.jsonl');checks=json.loads((e/'final-checks.json').read_text())
a=subprocess.check_output(['git','show','HEAD:internal/oracle/counts.md'],text=True);b=Path('internal/oracle/counts.md').read_text()
def rows(s):return {l.split('|')[1].strip():l for l in s.splitlines() if re.match(r'^\| (internal/|stage3/)',l)}
x,y=rows(a),rows(b);moves=[dict(fixture=k,before=x.get(k),after=y.get(k)) for k in sorted(x.keys()|y.keys()) if x.get(k)!=y.get(k)];assert len(moves)==30
(e/'counts-diff.json').write_text(json.dumps(moves,indent=2)+'\n')
records=[]
for category,rs in [('focused',focused),('consumer',consumer),('clean-consumer',clean),('mutant',mutants),('consumer-build',builds)]:
 for r in rs:records.append(dict(category=category,package=r.get('package','internal/oracle' if r.get('test')!='TestPortElementMetadataEmission' else 'stage1/cohere/json'),test=r.get('test',r.get('mutant','build')),seconds=r['seconds'],measurement='process wall',status='caught' if r.get('killed') else ('skip' if r.get('skip') else ('pass' if r['exit']==0 else 'fail')),grain=r.get('grain',r['seconds']<60),mutant=r.get('mutant','')))
for file,category in [('counts-check.jsonl','counts'),('restored-controls.log','restored-control')]:
 for line in (e/file).read_text().splitlines():
  try:r=json.loads(line)
  except json.JSONDecodeError:continue
  if r.get('Action') in ['pass','fail','skip'] and 'Test' in r:records.append(dict(category=category,package=r['Package'],test=r['Test'],seconds=r.get('Elapsed',0),measurement='Go elapsed',status=r['Action'],grain=r.get('Elapsed',0)<60,mutant=''))
(e/'all-results.json').write_text(json.dumps(dict(records=records,focused=focused,consumer=consumer,clean_consumer=clean,consumer_builds=builds,mutants=mutants,final_checks=checks),indent=2)+'\n')
with (e/'test-seconds.csv').open('w') as f:
 writer=csv.DictWriter(f,fieldnames=['category','package','test','seconds','measurement','status','grain','mutant']);writer.writeheader();writer.writerows(records)
assert len(focused)==287 and all(r['exit']==0 and r['grain'] for r in focused);assert len(builds)==45 and all(r['exit']==0 for r in builds);assert len(mutants)==17 and all(r['killed'] for r in mutants)
base=subprocess.check_output(['git','rev-parse','origin/main'],text=True).strip();initial=subprocess.check_output(['git','rev-parse','HEAD'],text=True).strip();runtime=json.loads((e/'runtime-delta-verified.json').read_text())['files']
lines=[
'Built: parked V4 checked callables and carried V3 arrays re-cut as compiler/loom-v4-2.',
'Commits: source 0998767348c4c20292cf989363dfc0822b0eb8f3; initial main '+initial+'; final main base '+base+'.',
'Commands and outputs: 287 focused tests/products green with seven carried skips; 45 consumer packages compiled; counts, reader checks and vet passed.',
'Mutants: all 17 independent source/witness mutants caught and restored; exact catchers and seconds below.',
'Not covered: full gate or exhaustive consumer families; seven consumer checks exceed 60 seconds and one was interrupted at the driver limit. First-green push within 30 minutes was missed.',
'',
'The branch tip is the single final commit above the reported base. It is rebased once after main advanced only in loom/rules.go, loom/rules.json and loom/rules_test.go. No compiler, runtime or counts changes landed during this validation. No PR or Loom submission is made.',
'',
'The squash of origin/compiler/loom-v4 onto current main conflicted only in these two files, each resolved once:',
'',
'- internal/flow/build.go: main c6e58480 (Three lowering gaps, on main) recognizes throwing object spreads through program.ObjectSpreadMayThrow. V4 recognizes ir.ArrayHoles as throwing. Keep both independent struct checks so getter exceptions and sparse-array RangeError propagate.',
'- internal/lower/exceptions.go: the same main change adds ObjectLiteral -> ObjectSpreadMayThrow to throwsOut. Keep that switch arm and V4\'s ir.Throw/ir.ArrayHoles arm plus checkArrayHoles, preserving both exception paths.',
'',
'Main\'s C portability change 5f9f4b45 and getter-copy cleanup auto-merged with V4. ADAMIC_STRING_BYTES keeps main\'s unsigned-byte cast; object copying keeps main\'s adamic_thrown cleanup. No additional runtime code was introduced. Normalized runtime additions/deletions exactly equal the parked V4 commit\'s delta, with hunk context excluded. Every runtime file\'s bytes also match the pre-mutant snapshot after restoration. Evidence: runtime-delta-verified.json, runtime-hashes.json and runtime-delta.txt.',
'',
'Runtime files changed against main:',
'']+['- '+p for p in runtime]
lines += ['',
'Counts: the auto-merged table passed the generated TestCountsAreRecorded shards without changes. No regeneration was necessary or performed in this recut. Command: timeout 210 go test ./internal/oracle -run ^TestCountsAreRecorded -count=1 -timeout 180s -parallel 4 -json; output counts-check.jsonl. Package elapsed 142.158 s, command wall 152.211 s; 1,078 pass records including shard parents, maximum leaf 8.670 s. The old exact anchor no longer selects main\'s generated top-level shards; use the prefix conservatively. Against main the 30 carried V4 rows are unchanged from parked V4: 26 V3 additions, three V4 compile-time escape refusals removed, and namespace_method_receiver retains/releases each decrease by two because declaration-only this is excluded from runtime argument slots. That attribution is a code reading. Every before/after row is in counts-diff.json.',
'',
'Focused validation: run-focused.py compiled selected test binaries and executed each exact top-level name with -test.timeout=90s, -test.parallel=1 and an outer 95-second limit. Native products are built before consumer tests; products themselves have measured setup time. All 287 selected tests/products completed below 60 seconds, maximum 51.484 s. The selection includes V4 positives/refusals, array tests, embedded compiler/private-runtime mutants, RuntimeSetsAreEachUsed, product key checks, TestCallTargetReaders, new main getter-spread and accessor tests and hash-owned native Node fixture shards. Native and backend fixtures retain their source Node, ASan/UBSan and leak checks. Deliberate checked-view refusals retain their own documented diagnostic/exit assertions. Seven existing pending tests are explicitly skipped, not coverage claimed.',
'',
'Rules 11 and 12: no new tests or product recipes were added by this recut, no changed compiler consumer package needs a declaration. V4 test splitting and placeless product declarations from the parked commit are retained. gofmt reported no changed Go file needing formatting. No other package\'s tests were weakened. No cohere implementation was copied; clean validation references the existing pinned submodule through a symlink.',
'',
'Consumer validation: go list -json ./... identified 45 packages directly importing internal/flow or internal/lower in production/test imports. All 45 compiled within the 180-second build bound, and vet covered those 45 plus internal/javascript and internal/load. Two packages have no tests. The 43 test packages have very large inventories, so run-consumers.py runs up to three product preparations and three representative exact top-level tests per package; selected-consumer-tests.json names the actual selection. This is bounded representative coverage, not an exhaustive package test run. Consumer driver timeout was 840 seconds; each test was bounded by 90 seconds internally and 95 seconds externally. There are 168 completed consumer records. Parent partial skips are preserved in logs; the CSV skip field can indicate a skipped subtest rather than the entire parent. Product/test checks over 60 seconds are red under the grain law even where their assertions passed, and were never rerun.',
'',
'Consumer limit observations:',
'',
'| Package | Test | Seconds | Observation |','|---|---|---:|---|']
reasons={'bridge/tsgo':'80-second product deadline killed bridge-test-tsgo-asan.a; cgroup oom_kill=0, not a demonstrated runtime defect','stage1/cohere/lint/helpers':'Go 90-second test timeout','stage1/cohere/lint':'assertions passed, exceeds 60-second law','stage1/cohere/typeaware':'Go 90-second test timeout','stage1/cohere/static_single_assignment':'assertions passed, exceeds 60-second law','stage1/cohere/suppression':'assertions passed, exceeds 60-second law','stage1/cohere/values':'assertions passed with partial skips, exceeds 60-second law'}
for r in consumer:
 if not r['grain']:lines.append(f"| {r['package']} | {r['test']} | {r['seconds']:.3f} | {reasons[r['package']]} |")
lines += ['',
'For product failures/overruns, remaining selected checks in that package were blocked. stage3/fixtures TestFixtureOracleProductIgnoresGoCacheLocation started but has no completed result before the 840-second driver deadline; it is not counted as passing or rerun. Its three preceding selected tests/products passed. All other selected completed failures were setup issues corrected as follows.',
'',
'MarkdownInline_0000 and _0001 initially failed in 0.032/0.033 s because staged incoming Markdown differs from main HEAD. A clean detached commit-tree snapshot of the staged code/docs (not a branch commit) fixed the validation context. They then passed in 26.275/22.347 s. Scanner shards 000/001/002 initially failed in 10.120/6.933/11.030 s because their in-test clone could not connect to GitHub. An explicit bounded git clone of v6.0.3 succeeded; ADAMIC_TYPESCRIPT_SOURCE names that verified pinned checkout for validation only. The clean snapshot probes then passed in 6.648/2.424/9.515 s. No machine path was added to a product/test source. Snapshot metadata and exact rerun commands are in clean-snapshot.json and run-clean-consumers.py. The detached snapshot includes all incoming V4 Markdown, excluding this new evidence directory.',
'',
'Final restored checks: '+str(checks[0]['seconds'])+' s for the exact source-mutant catcher union plus TestCallTargetReaders/TestProduct_CallTargetExports, all pass. Vet: '+str(checks[1]['seconds'])+' s, exit 0, no diagnostics across 47 packages. final-checks.json records the full argument vectors; restored-controls.log contains every Go leaf\'s elapsed seconds. test-seconds.csv and all-results.json contain every completed focused, consumer, mutant and count record, with build and process wall times labeled separately.',
'',
'Independent mutants, all exit 1 due to test assertions (no build failures):',
'',
'| Mutant | Catcher | Seconds | What failed |','|---|---|---:|---|']
catchers={'surface-borrow':'ASan adapter root use-after-free','bind-name-borrow':'ASan name string use-after-free','identity-root':'Node stdout comparison','metadata-inline':'2,885 inline stores violate emission bound','metadata-omitted':'zero runtime certificates instead of 2,885','native-order':'argument output missing after premature callee check','javascript-order':'argument output missing after premature callee check','unsupported-domain':'required recursive-domain refusal missing','unsupported-producer':'required producer refusal missing','construct-witness':'required construct-signature refusal missing','boxed-witness':'required unproven-read refusal missing','factory-witness':'required wider-write path/fix refusal missing'}
for r in mutants:lines.append(f"| {r['mutant']} | {r['test']} | {r['seconds']:.3f} | {catchers.get(r['mutant'],'required escaping-read path/fix refusal missing')} |")
lines += ['',
'All source bytes restored after each mutant, patches stored as .patch. The driver initially used the wrong witness directory after the 14 compiler/runtime mutations; it stopped with FileNotFoundError. Only the three not-yet-run witness mutations were run after correcting the evidence script. No successful mutant was rerun.',
'',
'Setup: GOPROXY=https://proxy.golang.org|direct; timeout 600 bash cloud/setup.sh; source /workspace/adamic-tools/env.sh. Completed 51.744 s. nproc=5; cpu.max=400000 100000 means four CPU shares. Setup timing lines follow.',
'', '```text']
lines += [l for l in (e/'setup.log').read_text().splitlines() if l.startswith('setup:')]
lines += ['```','']
(e/'REPORT.md').write_text('\n'.join(lines)+'\n')
# Delete only this driver\'s explicitly named owned binary artifacts.
for package in json.loads((e/'consumers.json').read_text()):
 path=e/(package['package'].replace('/','-')+'-consumer.test')
 if path.exists():path.unlink()
print('focused',len(focused),'consumer records',len(consumer),'compiled',len(builds),'mutants',len(mutants),'counts delta',len(moves))
